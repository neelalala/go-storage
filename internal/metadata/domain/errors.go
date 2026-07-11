package domain

import "errors"

var (
	ErrObjectNotFound = errors.New("error: object not found")
	ErrUploadNotFound = errors.New("error: upload not found")
)
