package domain

import "errors"

var (
	ErrAccessDenied        = errors.New("access denied")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrBucketNotExists     = errors.New("bucket not exists")
	ErrBucketNotEmpty      = errors.New("bucket bot empty")
	ErrObjectNotFound      = errors.New("object not found")
	ErrUploadNotExists     = errors.New("upload not exists")
)
