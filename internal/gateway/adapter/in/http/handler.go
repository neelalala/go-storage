package http

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const (
	DefaultBucketsLimit  = 100
	DefaultBucketsOffset = 0
	DefaultObjectsLimit  = 100
	DefaultObjectsOffset = 0
)

type Handler struct {
	metadata domain.MetadataService
	gateway  Gateway

	marshaller Marshaller

	log *slog.Logger
}

func NewHandler(metadata domain.MetadataService, gateway Gateway, marshaller Marshaller, log *slog.Logger) *Handler {
	return &Handler{
		metadata:   metadata,
		gateway:    gateway,
		marshaller: marshaller,
		log:        log,
	}
}

func (h *Handler) ListBuckets(w http.ResponseWriter, req *http.Request) {
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
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

	buckets, err := h.metadata.ListBuckets(req.Context(), limit, offset)
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	resp, err := h.marshaller.ListBuckets(buckets)
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) CreateBucket(w http.ResponseWriter, req *http.Request) {
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	name := req.PathValue("bucket")
	if name == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	if _, err := h.metadata.CreateBucket(req.Context(), name); err != nil {
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s", name), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListObjects(w http.ResponseWriter, req *http.Request) {
	// TODO: should list all objects when prefix=""?
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
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

	objs, err := h.metadata.ListObjects(req.Context(), bucket, prefix, delimiter, limit, offset)
	if err != nil {
		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, prefix), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	resp, err := h.marshaller.ListObjectsV2(bucket, prefix, delimiter, objs)
	if err != nil {
		resp, status := h.marshaller.Error(err, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, req *http.Request) {
	// TODO: not idempotent; returns 500 if bucket did not exists and if bucket wasnt empty
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	if err := h.metadata.DeleteBucket(req.Context(), bucket); err != nil {
		_, status := h.marshaller.Error(err, fmt.Sprintf("/%s", bucket), requestID)
		w.WriteHeader(status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PutObject(w http.ResponseWriter, req *http.Request) {
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
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
		)

		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	err = h.gateway.PutObject(req.Context(), bucket, key, data)
	if err != nil {
		h.log.Error(
			"put object",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, req *http.Request) {
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	object, err := h.gateway.GetObject(req.Context(), bucket, key)
	if err != nil {
		h.log.Error(
			"get object",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	filename := path.Base(key)

	w.Header().Set("Content-Type", "image/jpeg") // TODO: save mime type in object metadata
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(object)), 10))

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	w.WriteHeader(http.StatusOK)
	w.Write(object)
}

func (h *Handler) DeleteObject(w http.ResponseWriter, req *http.Request) {
	requestID, err := uuid.NewV7()
	if err != nil {
		resp, status := h.marshaller.Error(err, "/", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	bucket := req.PathValue("bucket")
	if bucket == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	key := req.PathValue("key")
	if key == "" {
		resp, status := h.marshaller.Error(domain.ErrInvalidRequest, "", requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	err = h.gateway.DeleteObject(req.Context(), bucket, key)
	if err != nil {
		h.log.Error(
			"delete objec",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		resp, status := h.marshaller.Error(err, fmt.Sprintf("/%s/%s", bucket, key), requestID)
		w.WriteHeader(status)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
