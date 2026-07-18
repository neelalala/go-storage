package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type MetadataService struct {
	bucketRepo domain.BucketRepository
	uploadRepo domain.UploadRepository
	objRepo    domain.ObjectRepository
	storage    domain.Storage
	hasher     domain.Hasher

	log *slog.Logger
}

func NewMetadataService(
	bucketRepo domain.BucketRepository,
	uploadRepo domain.UploadRepository,
	objRepo domain.ObjectRepository,
	storage domain.Storage,
	hasher domain.Hasher,
	log *slog.Logger,
) *MetadataService {
	return &MetadataService{
		bucketRepo: bucketRepo,
		uploadRepo: uploadRepo,
		objRepo:    objRepo,
		storage:    storage,
		hasher:     hasher,
		log:        log,
	}
}

func (s *MetadataService) ListBuckets(ctx context.Context, limit, offset int) ([]domain.Bucket, error) {
	s.log.Debug("metadata service",
		"method", "list buckets",
		"limit", limit,
		"offset", offset,
	)

	return s.bucketRepo.GetBuckets(ctx, limit, offset)
}

func (s *MetadataService) CreateBucket(ctx context.Context, name string) (domain.Bucket, error) {
	s.log.Debug("metadata service",
		"method", "create bucket",
		"name", name,
	)

	return s.bucketRepo.CreateBucket(ctx, name)
}

func (s *MetadataService) GetBucket(ctx context.Context, name string) (domain.Bucket, error) {
	s.log.Debug("metadata service",
		"method", "get bucket",
		"name", name,
	)

	return s.bucketRepo.GetBucket(ctx, name)
}

func (s *MetadataService) DeleteBucket(ctx context.Context, name string) error {
	s.log.Debug("metadata service",
		"method", "delete bucket",
		"name", name,
	)

	return s.bucketRepo.DeleteBucket(ctx, name)
}

func (s *MetadataService) InitUpload(ctx context.Context, bucket, key string, size uint64) (domain.Upload, domain.Storage, error) {
	s.log.Debug("metadata service",
		"method", "init upload",
		"bucket", bucket,
		"key", key,
		"size", size,
	)

	// TODO: what if bucket = "bucket/"? it has to be the same as "bucket"
	objPath := fmt.Sprintf("%X", s.hasher.Hash([]byte(bucket+key)))

	upload := domain.Upload{
		Bucket:        bucket,
		Key:           key,
		ObjectPath:    objPath,
		Size:          size,
		StorageNodeID: s.storage.ID,
	}

	saved, err := s.uploadRepo.CreateUpload(ctx, upload)
	if err != nil {
		s.log.Error("metadata service",
			"method", "init upload",
			"context", "UploadRepository.CreateUpload",
			"upload", fmt.Sprintf("%+v", upload),
			"error", err,
		)

		return domain.Upload{}, domain.Storage{}, fmt.Errorf("error starting upload transaction")
	}

	s.log.Debug("metadata service",
		"method", "init upload",
		"bucket", bucket,
		"key", key,
		"object_path", saved.ObjectPath,
		"message", "successful",
	)

	return saved, s.storage, nil
}

func (s *MetadataService) CommitUpload(ctx context.Context, uploadID uuid.UUID, checksum uint32) error {
	s.log.Debug("metadata service",
		"method", "commit upload",
		"upload_id", uploadID,
		"checksum", checksum,
	)

	err := s.uploadRepo.CommitUpload(ctx, uploadID, checksum)
	if err != nil {
		if !errors.Is(err, domain.ErrUploadNotFound) {
			s.log.Error("metadata service",
				"method", "commit upload",
				"context", "UploadRepository.CommitUpload",
				"error", err,
			)
		}

		return err
	}

	s.log.Debug("metadata service",
		"method", "commit upload",
		"upload_id", uploadID,
		"message", "successful",
	)

	return nil
}

func (s *MetadataService) AbortUpload(ctx context.Context, uploadID uuid.UUID) error {
	s.log.Debug("metedata service",
		"method", "abort upload",
		"upload_id", uploadID,
	)

	err := s.uploadRepo.DeleteUpload(ctx, uploadID)
	if err != nil {
		if !errors.Is(err, domain.ErrUploadNotFound) {
			s.log.Error("metadata service",
				"method", "abort upload",
				"context", "UploadRepository.DeleteUpload",
				"error", err,
			)
		}

		return err
	}

	s.log.Debug("metadata service",
		"method", "abort upload",
		"upload_id", uploadID,
		"message", "successful",
	)

	return nil
}

func (s *MetadataService) GetObject(ctx context.Context, bucket, key string) (domain.Object, domain.Storage, error) {
	s.log.Debug("metadata service",
		"method", "get object",
		"bucket", bucket,
		"key", key,
	)

	obj, err := s.objRepo.GetObject(ctx, bucket, key)
	if err != nil {
		if !errors.Is(err, domain.ErrObjectNotFound) {
			s.log.Error("metadata service",
				"method", "get object",
				"context", "ObjectRepository.GetObject",
				"error", err,
			)
		}

		return domain.Object{}, domain.Storage{}, err
	}

	s.log.Debug("metadata service",
		"method", "get object",
		"bucket", bucket,
		"key", key,
		"message", "successful",
	)

	return obj, s.storage, nil
}

func (s *MetadataService) GetObjects(ctx context.Context, bucket, prefix, delimiter string, limit, offset int) ([]domain.Object, error) {
	s.log.Debug("metadata service",
		"method", "get objects",
		"bucket", bucket,
		"prefix", prefix,
		"delimiter", delimiter,
		"limit", limit,
		"offset", offset,
	)

	return s.objRepo.GetObjects(ctx, bucket, prefix, delimiter, limit, offset)
}

func (s *MetadataService) DeleteObject(ctx context.Context, bucket, key string) error {
	s.log.Debug("metadata service",
		"method", "delete object",
		"bucket", bucket,
		"key", key,
	)

	err := s.objRepo.SoftDeleteObject(ctx, bucket, key)
	if err != nil {
		if !errors.Is(err, domain.ErrObjectNotFound) {
			s.log.Error("metadata service",
				"method", "delete object",
				"context", "ObjectRepository.DeleteObject",
				"error", err,
			)
		}

		return err
	}

	s.log.Debug("metadata service",
		"method", "delete object",
		"bucket", bucket,
		"key", key,
		"message", "successful",
	)

	return nil
}

func (s *MetadataService) HeadObject(ctx context.Context, bucket, key string) (domain.Object, error) {
	s.log.Debug("metadata service",
		"method", "head object",
		"bucket", bucket,
		"key", key,
	)

	return s.objRepo.GetObject(ctx, bucket, key)
}
