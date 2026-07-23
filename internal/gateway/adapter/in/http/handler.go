package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/adapter/in/http/middleware"
	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const (
	DefaultBucketsLimit  = 100
	DefaultBucketsOffset = 0
	DefaultObjectsLimit  = 100
	DefaultObjectsOffset = 0
)

type Handler struct {
	gateway    Gateway
	marshaller Marshaller

	log *slog.Logger
}

func NewHandler(gateway Gateway, marshaller Marshaller, log *slog.Logger) *Handler {
	return &Handler{
		gateway:    gateway,
		marshaller: marshaller,
		log:        log,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	name := req.URL.Query().Get("name")
	if name == "" {
		h.log.Debug("name is required", "name", name, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("name is required"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	_, err = h.gateway.CreateUser(req.Context(), name)
	if err != nil {
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			h.log.Error("error creating user", "error", err, "request_id", requestID)
		}
		resp, status := h.marshaller.Error(err, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListBuckets(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	limit := DefaultBucketsLimit
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 0); err == nil {
			limit = int(l)
		}
	}

	offset := DefaultBucketsOffset
	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 0); err == nil {
			offset = int(o)
		}
	}

	buckets, err := h.gateway.ListBuckets(req.Context(), user.ID, limit, offset)
	if err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) {
			h.log.Error("error listing buckets", "request_id", requestID, "error", err)
		} else {
			h.log.Debug("error listing buckets", "request_id", requestID, "error", err)
		}
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	resp, err := h.marshaller.ListBuckets(user, buckets)
	if err != nil {
		h.log.Error("error marshalling response",
			"method", "list buckets",
			"request_id", requestID,
			"error", err,
		)
		resp, status := h.marshaller.Error(errors.New("error marshalling response"), "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) CreateBucket(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	name := req.PathValue("bucket")
	if name == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no name specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	if _, err := h.gateway.CreateBucket(req.Context(), user.ID, name); err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) && !errors.Is(err, domain.ErrBucketAlreadyExists) {
			h.log.Error("error creating bucket", "error", err, "request_id", requestID)
		} else {
			h.log.Debug("error creating bucket", "error", err, "request_id", requestID)
		}
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s", name), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListObjects(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no bucket specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	prefix := req.URL.Query().Get("prefix")

	delimiter := req.URL.Query().Get("delimiter")
	if delimiter == "" {
		delimiter = "/"
	}

	limit := DefaultObjectsLimit
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 0); err == nil {
			limit = int(l)
		}
	}

	offset := DefaultObjectsOffset
	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 0); err == nil {
			offset = int(o)
		}
	}

	objs, prefixes, err := h.gateway.ListObjects(req.Context(), user.ID, bucket, prefix, delimiter, limit+1, offset)
	if err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) && !errors.Is(err, domain.ErrBucketNotExists) {
			h.log.Error("error listing objects", "request_id", requestID, "error", err)
		} else {
			h.log.Debug("error listing objects", "request_id", requestID, "error", err)
		}
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, prefix), requestID)
		w.WriteHeader(status)
		w.Write(resp)
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
		h.log.Error("error marshalling response",
			"method", "list objects v2",
			"request_id", requestID,
			"error", err,
		)
		resp, status := h.marshaller.Error(errors.New("error marshalling response"), "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no bucket specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	if err := h.gateway.DeleteBucket(req.Context(), user.ID, bucket); err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) &&
			!errors.Is(err, domain.ErrBucketNotExists) &&
			!errors.Is(err, domain.ErrBucketNotEmpty) {
			h.log.Error("error deleting bucket", "request_id", requestID, "error", err)
		} else {
			h.log.Debug("error deleting bucket", "request_id", requestID, "error", err)
		}
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s", bucket), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PutObject(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no bucket specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no key specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		h.log.Error(
			"read request body",
			"bucket", bucket,
			"key", key,
			"error", err,
			"request_id", requestID,
		)
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	contentType := req.Header.Get("Content-Type")
	systemMetadata, userMetadata := ExtractMetadata(req)

	err = h.gateway.PutObject(req.Context(), user.ID, bucket, key, data, contentType, systemMetadata, userMetadata)
	if err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) &&
			!errors.Is(err, domain.ErrBucketNotExists) {
			h.log.Error(
				"put object",
				"bucket", bucket,
				"key", key,
				"error", err,
			)
		} else {
			h.log.Debug(
				"put object",
				"bucket", bucket,
				"key", key,
				"error", err,
			)
		}
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, req *http.Request) {
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no bucket specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no key specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	meta, data, err := h.gateway.GetObject(req.Context(), user.ID, bucket, key)
	if err != nil {
		if !errors.Is(err, domain.ErrAccessDenied) &&
			!errors.Is(err, domain.ErrBucketNotExists) &&
			!errors.Is(err, domain.ErrKeyNotExists) {
			h.log.Error("error getting object", "error", err, "request_id", requestID)
		} else {
			h.log.Debug("error getting object", "error", err, "request_id", requestID)
		}
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
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
	requestID, err := middleware.GetRequestID(req.Context())
	if err != nil {
		h.log.Error("error getting request id", "error", err)
		resp, status := h.marshaller.Error(errors.New("error getting request id"), "", uuid.Nil)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	user, err := middleware.GetUser(req.Context())
	if err != nil {
		h.log.Error("error getting user", "error", err, "request_id", requestID)
		resp, status := h.marshaller.Error(errors.New("error getting user id"), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no bucket specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(fmt.Errorf("%w: no key specified in path", domain.ErrInvalidRequest), "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	err = h.gateway.DeleteObject(req.Context(), user.ID, bucket, key)
	if err != nil && !errors.Is(err, domain.ErrKeyNotExists) {
		if !errors.Is(err, domain.ErrAccessDenied) &&
			!errors.Is(err, domain.ErrBucketNotExists) {
			h.log.Error("error deleting object", "error", err, "request_id", requestID)
		} else {
			h.log.Debug("error deleting object", "error", err, "request_id", requestID)
		}

		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
