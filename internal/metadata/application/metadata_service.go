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
	uploadRepo domain.UploadRepository
	objRepo    domain.ObjectRepository
	storage    domain.Storage

	log *slog.Logger
}

func NewMetadataService(uploadRepo domain.UploadRepository, objRepo domain.ObjectRepository, storage domain.Storage, log *slog.Logger) *MetadataService {
	return &MetadataService{
		uploadRepo: uploadRepo,
		objRepo:    objRepo,
		storage:    storage,
		log:        log,
	}
}

func (s *MetadataService) InitUpload(ctx context.Context, bucket, key string, size uint64) (uuid.UUID, domain.Storage, error) {
	s.log.Debug("metadata service",
		"method", "init upload",
		"bucket", bucket,
		"key", key,
		"size", size,
	)

	upload := domain.Upload{Bucket: bucket, Key: key, Size: size, StorageNodeID: s.storage.ID}
	saved, err := s.uploadRepo.CreateUpload(ctx, upload)
	if err != nil {
		s.log.Error("metadata service",
			"method", "init upload",
			"context", "UploadRepository.CreateUpload",
			"upload", fmt.Sprintf("%+v", upload),
			"error", err,
		)

		return uuid.UUID{}, domain.Storage{}, fmt.Errorf("error starting upload transaction")
	}

	s.log.Debug("metadata service",
		"method", "init upload",
		"bucket", bucket,
		"key", key,
		"message", "successful",
	)

	return saved.UploadID, s.storage, nil
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

	return *obj, s.storage, nil
}

func (s *MetadataService) DeleteObject(ctx context.Context, bucket, key string) (domain.Object, domain.Storage, error) {
	s.log.Debug("metadata service",
		"method", "delete object",
		"bucket", bucket,
		"key", key,
	)

	obj, err := s.objRepo.SoftDeleteObject(ctx, bucket, key)
	if err != nil {
		if !errors.Is(err, domain.ErrObjectNotFound) {
			s.log.Error("metadata service",
				"method", "delete object",
				"context", "ObjectRepository.DeleteObject",
				"error", err,
			)
		}

		return domain.Object{}, domain.Storage{}, err
	}

	s.log.Debug("metadata service",
		"method", "delete object",
		"bucket", bucket,
		"key", key,
		"message", "successful",
	)

	return *obj, s.storage, nil
}
