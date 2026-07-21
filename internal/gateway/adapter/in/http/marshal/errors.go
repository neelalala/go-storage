package marshal

import (
	"errors"
	"net/http"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const (
	AccessDenied        = "AccessDenied"
	BucketAlreadyExists = "BucketAlreadyExists"
	BucketNotEmpty      = "BucketNotEmpty"
	EndpointNotFound    = "EndpointNotFound"
	InternalError       = "InternalError"
	InvalidPrefix       = "InvalidPrefix"
	InvalidRequest      = "InvalidRequest"
	NoSuchBucket        = "NoSuchBucket"
	NoSuchKey           = "NoSuchKey"
	NoSuchUpload        = "NoSuchUpload"
)

func ErrorToCode(err error) (string, int) {
	switch {
	default:
		return InternalError, http.StatusInternalServerError
	case errors.Is(err, domain.ErrAccessDenied):
		return AccessDenied, http.StatusForbidden
	case errors.Is(err, domain.ErrBucketAlreadyExists):
		return BucketAlreadyExists, http.StatusConflict
	case errors.Is(err, domain.ErrBucketNotEmpty):
		return BucketNotEmpty, http.StatusConflict
	case errors.Is(err, domain.ErrBucketNotExists):
		return NoSuchBucket, http.StatusNotFound
	case errors.Is(err, domain.ErrKeyNotExists):
		return NoSuchKey, http.StatusNotFound
	case errors.Is(err, domain.ErrUploadNotExists):
		return NoSuchUpload, http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidRequest):
		return InvalidRequest, http.StatusBadRequest
	}
}
