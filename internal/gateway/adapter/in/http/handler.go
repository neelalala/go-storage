package http

import (
	"io"
	"log/slog"
	"net/http"
)

type Handler struct {
	gateway Gateway

	log *slog.Logger
}

func NewHandler(gateway Gateway, log *slog.Logger) *Handler {
	return &Handler{
		gateway: gateway,
		log:     log,
	}
}

func (h *Handler) PutObject(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	if bucket == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing bucket"))
		return
	}

	key := req.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing key"))
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		h.log.Error("read request body",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't read request body"))
		return
	}

	err = h.gateway.PutObject(bucket, key, data)
	if err != nil {
		h.log.Error("put object",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't save object"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	if bucket == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing bucket"))
		return
	}

	key := req.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing key"))
		return
	}

	object, err := h.gateway.GetObject(bucket, key)
	if err != nil {
		h.log.Error("get object",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't get object"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(object)
}

func (h *Handler) DeleteObject(w http.ResponseWriter, req *http.Request) {
	bucket := req.PathValue("bucket")
	if bucket == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing bucket"))
		return
	}

	key := req.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing key"))
		return
	}

	err := h.gateway.DeleteObject(bucket, key)
	if err != nil {
		h.log.Error("delete objec",
			"bucket", bucket,
			"key", key,
			"error", err,
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't delete object"))
		return
	}

	w.WriteHeader(http.StatusOK)
}
