package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/adapter/in/http/middleware"
	"github.com/neelalala/go-storage/internal/gateway/application"
	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const (
	DefaultBucketsLimit  = 100
	DefaultBucketsOffset = 0
	DefaultObjectsLimit  = 100
	DefaultObjectsOffset = 0
)

type Handler struct {
	gateway    *application.Gateway
	marshaller Marshaller

	log *slog.Logger
}

func NewHandler(gateway *application.Gateway, marshaller Marshaller, log *slog.Logger) *Handler {
	return &Handler{
		gateway:    gateway,
		marshaller: marshaller,
		log:        log,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, req *http.Request) {
	username, ok := h.extractPathValue(w, req, "username")
	if !ok {
		return
	}

	_, err := h.gateway.CreateUser(req.Context(), username)
	if err != nil {
		h.handleError(w, req, err, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListBuckets(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	limit := parseQueryInt(req, "limit", DefaultBucketsLimit)
	offset := parseQueryInt(req, "offset", DefaultBucketsOffset)

	buckets, err := h.gateway.ListBuckets(req.Context(), user.ID, limit, offset)
	if err != nil {
		h.handleError(w, req, err, "/")
		return
	}

	resp, err := h.marshaller.ListBuckets(user, buckets)
	if err != nil {
		h.handleError(w, req, err, "/")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) CreateBucket(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	if _, err := h.gateway.CreateBucket(req.Context(), user.ID, bucket); err != nil {
		h.handleError(w, req, err, "/"+bucket)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HeadBucket(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	meta, err := h.gateway.HeadBucket(req.Context(), user.ID, bucket)
	if err != nil {
		h.handleError(w, req, err, "/"+bucket)
		return
	}

	w.Header().Set("X-Go-Bucket-Name", meta.Name)
	w.Header().Set("X-Go-Owner-Id", meta.OwnerID.String())
	w.Header().Set("X-Go-Creation-Date", meta.CreatedAt.Format(time.RFC3339))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	if err := h.gateway.DeleteBucket(req.Context(), user.ID, bucket); err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s", bucket))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListObjects(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	prefix := req.URL.Query().Get("prefix")

	delimiter := req.URL.Query().Get("delimiter")
	if delimiter == "" {
		delimiter = "/"
	}

	limit := parseQueryInt(req, "limit", DefaultObjectsLimit)
	offset := parseQueryInt(req, "offset", DefaultObjectsOffset)

	objs, prefixes, err := h.gateway.ListObjects(req.Context(), user.ID, bucket, prefix, delimiter, limit+1, offset)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s", bucket))
		return
	}

	var isTruncated bool
	if len(objs)+len(prefixes) > limit {
		isTruncated = true

		if len(objs) == 0 {
			prefixes = prefixes[:len(prefixes)-1]
		} else if len(prefixes) == 0 {
			objs = objs[:len(objs)-1]
		} else {
			if objs[len(objs)-1].Key < prefixes[len(prefixes)-1] {
				prefixes = prefixes[:len(prefixes)-1]
			} else {
				objs = objs[:len(objs)-1]
			}
		}
	}

	resp, err := h.marshaller.ListObjectsV2(bucket, prefix, delimiter, limit, objs, prefixes, isTruncated)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s", bucket))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) PutObject(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	key, ok := h.extractPathValue(w, req, "key")
	if !ok {
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s/%s", bucket, key))
		return
	}

	contentType := req.Header.Get("Content-Type")
	systemMetadata, userMetadata := ExtractMetadata(req)

	err = h.gateway.PutObject(req.Context(), user.ID, bucket, key, data, contentType, systemMetadata, userMetadata)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s/%s", bucket, key))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HeadObject(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	key, ok := h.extractPathValue(w, req, "key")
	if !ok {
		return
	}

	meta, err := h.gateway.HeadObject(req.Context(), user.ID, bucket, key)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s/%s", bucket, key))
		return
	}

	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(meta.Size), 10))
	w.Header().Set("ETag", fmt.Sprintf("\"%s\"", meta.Hash))
	w.Header().Set("X-Go-Owner-Id", meta.OwnerID.String())
	w.Header().Set("X-Go-Storage-Node-Id", meta.StorageNodeID.String())
	w.Header().Set("X-Go-Created-At", meta.CreatedAt.Format(time.RFC3339))
	w.Header().Set("X-Go-Updated-At", meta.UpdatedAt.Format(time.RFC3339))

	for header, value := range meta.SystemMetadata {
		w.Header().Set(header, value)
	}

	for header, value := range meta.UserMetadata {
		w.Header().Set(header, value)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	key, ok := h.extractPathValue(w, req, "key")
	if !ok {
		return
	}

	meta, data, err := h.gateway.GetObject(req.Context(), user.ID, bucket, key)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s/%s", bucket, key))
		return
	}

	w.Header().Set("Content-Length", strconv.FormatInt(int64(meta.Size), 10))

	if meta.ContentType != "" {
		w.Header().Set("Content-Type", meta.ContentType)
	}

	if meta.Hash != "" {
		w.Header().Set("ETag", fmt.Sprintf("\"%s\"", meta.Hash))
	}

	for header, value := range meta.SystemMetadata {
		w.Header().Set(header, value)
	}

	for header, value := range meta.UserMetadata {
		w.Header().Set(header, value)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) DeleteObject(w http.ResponseWriter, req *http.Request) {
	user, ok := h.getUserFromContext(w, req)
	if !ok {
		return
	}

	bucket, ok := h.extractPathValue(w, req, "bucket")
	if !ok {
		return
	}

	key, ok := h.extractPathValue(w, req, "key")
	if !ok {
		return
	}

	err := h.gateway.DeleteObject(req.Context(), user.ID, bucket, key)
	if err != nil {
		h.handleError(w, req, err, fmt.Sprintf("/%s/%s", bucket, key))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getUserFromContext(w http.ResponseWriter, req *http.Request) (domain.User, bool) {
	reqID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("missing request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("internal server error"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return domain.User{}, false
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("missing user in context", "error", err, "request_id", reqID)
		resp, status := h.marshaller.Error(domain.ErrAccessDenied, "", reqID)
		w.WriteHeader(status)
		w.Write(resp)
		return domain.User{}, false
	}

	return user, true
}

func (h *Handler) extractPathValue(w http.ResponseWriter, req *http.Request, name string) (string, bool) {
	value := req.PathValue(name)

	if value == "" {
		h.handleError(w, req, fmt.Errorf("%w: %s is missing", domain.ErrInvalidRequest, name), "/")
		return "", false
	}
	return value, true
}

func (h *Handler) handleError(w http.ResponseWriter, req *http.Request, err error, resourcePath string) {
	reqID, _ := middleware.GetRequestID(req.Context())

	if isBusinessError(err) {
		h.log.Debug("client error", "method", req.Method, "path", req.URL.Path, "error", err, "request_id", reqID)
	} else {
		h.log.Error("internal error", "method", req.Method, "path", req.URL.Path, "error", err, "request_id", reqID)
	}

	resp, status := h.marshaller.Error(err, resourcePath, reqID)
	w.WriteHeader(status)
	w.Write(resp)
}

func isBusinessError(err error) bool {
	return errors.Is(err, domain.ErrAccessDenied) ||
		errors.Is(err, domain.ErrInvalidRequest) ||
		errors.Is(err, domain.ErrBucketNotExists) ||
		errors.Is(err, domain.ErrKeyNotExists) ||
		errors.Is(err, domain.ErrBucketAlreadyExists) ||
		errors.Is(err, domain.ErrBucketNotEmpty) ||
		errors.Is(err, domain.ErrUserAlreadyExists)
}

func parseQueryInt(req *http.Request, key string, defaultVal int) int {
	valStr := req.URL.Query().Get(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
