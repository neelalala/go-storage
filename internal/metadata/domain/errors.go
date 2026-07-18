package domain

import "errors"

var (
	ErrBucketExists    = errors.New("error: bucket already exists")
	ErrBucketNotExists = errors.New("error: bucket not exists")
	ErrBucketNotEmpty  = errors.New("error: bucket bot empty")
	ErrObjectNotFound  = errors.New("error: object not found")
	ErrUploadNotFound  = errors.New("error: upload not found")
)
