package marshal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const errorMarshallingErrorResponse = `{"code": "InternalError", "message": "error marshalling error", "resource": "%s", "requestID": "%s"}`

type JSONMarshaller struct{}

func (_ JSONMarshaller) ListBuckets(owner domain.User, buckets []domain.BucketMetadata) ([]byte, error) {
	type Bucket struct {
		Name         string `json:"name"`
		CreationDate string `json:"creationDate"`
	}

	type Response struct {
		ListAllMyBucketsResult struct {
			Owner struct {
				ID          string `json:"id"`
				DisplayName string `json:"displayName"`
			} `json:"owner"`
			Buckets struct {
				Bucket []Bucket `json:"bucket"`
			} `json:"buckets"`
		} `json:"listAllMyBucketsResult"`
	}

	var resp Response

	resp.ListAllMyBucketsResult.Owner.ID = owner.ID.String()
	resp.ListAllMyBucketsResult.Owner.DisplayName = owner.DisplayName

	resp.ListAllMyBucketsResult.Buckets.Bucket = make([]Bucket, 0, len(buckets))

	for _, bucket := range buckets {
		resp.ListAllMyBucketsResult.Buckets.Bucket = append(resp.ListAllMyBucketsResult.Buckets.Bucket, Bucket{
			Name:         bucket.Name,
			CreationDate: bucket.CreatedAt.Format(time.RFC3339),
		})
	}

	return json.MarshalIndent(resp, "", " ")
}

func (_ JSONMarshaller) ListObjectsV2(
	name, prefix, delimiter string,
	limit int,
	objects []domain.ObjectMetadata,
	prefixes []string,
	isTruncated bool,
) ([]byte, error) {
	type Content struct {
		Key          string `json:"key"`
		LastModified string `json:"lastModified"`
		ETag         string `json:"ETag"`
		Size         int64  `json:"size"`
		StorageClass string `json:"storageClass"` // TODO:
	}

	type CommonPrefix struct {
		Prefix string `json:"prefix"`
	}

	type Response struct {
		ListBucketResult struct {
			Name        string `json:"name"`
			Prefix      string `json:"prefix"`
			KeyCount    int    `json:"keyCount"`
			MaxKeys     int    `json:"maxKeys"`
			Delimiter   string `json:"delimiter"`
			IsTruncated bool   `json:"isTruncated"`

			Contents []Content `json:"contents"`

			CommonPrefixes []CommonPrefix `json:"commonPrefixes"`
		} `json:"listBucketResult"`
	}

	var resp Response

	resp.ListBucketResult.Name = name
	resp.ListBucketResult.Prefix = prefix
	resp.ListBucketResult.KeyCount = len(objects)
	resp.ListBucketResult.MaxKeys = limit
	resp.ListBucketResult.Delimiter = delimiter
	resp.ListBucketResult.IsTruncated = isTruncated

	resp.ListBucketResult.Contents = make([]Content, 0, len(objects))

	for _, object := range objects {
		resp.ListBucketResult.Contents = append(resp.ListBucketResult.Contents, Content{
			Key:          object.Key,
			LastModified: object.UpdatedAt.Format(time.RFC3339),
			ETag:         fmt.Sprintf("\"%s\"", object.Hash),
			Size:         int64(object.Size),
			StorageClass: "STANDARD",
		})
	}

	resp.ListBucketResult.CommonPrefixes = make([]CommonPrefix, 0)

	for _, pref := range prefixes {
		resp.ListBucketResult.CommonPrefixes = append(resp.ListBucketResult.CommonPrefixes, CommonPrefix{
			Prefix: pref,
		})
	}

	return json.MarshalIndent(resp, "", " ")
}

func (_ JSONMarshaller) Error(err error, resource string, requestID uuid.UUID) ([]byte, int) {
	type Response struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			Resource  string `json:"resource"`
			RequestID string `json:"requestId"`
		} `json:"error"`
	}

	var status int
	var resp Response
	resp.Error.Code, status = ErrorToCode(err)
	resp.Error.Message = err.Error()
	resp.Error.Resource = resource
	resp.Error.RequestID = requestID.String()

	if bytes, err := json.MarshalIndent(resp, "", " "); err == nil {
		return bytes, status
	}
	return []byte(fmt.Sprintf(errorMarshallingErrorResponse, resource, requestID.String())), http.StatusInternalServerError
}
