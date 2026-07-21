package domain

import "errors"

var (
	ErrAccessDenied        = errors.New("access denied")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrBucketNotExists     = errors.New("bucket not exists")
	ErrBucketNotEmpty      = errors.New("bucket not empty")
	ErrKeyNotExists        = errors.New("key not exists in this bucket")
	ErrUploadNotExists     = errors.New("upload not exists")
	ErrInvalidRequest      = errors.New("invalid request")
)
