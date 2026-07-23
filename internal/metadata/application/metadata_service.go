package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type MetadataService struct {
	transactor domain.Transactor
	bucketRepo domain.BucketRepository
	uploadRepo domain.UploadRepository
	objRepo    domain.ObjectRepository
	storage    domain.Storage
	hasher     domain.Hasher

	log *slog.Logger
}

func NewMetadataService(
	transactor domain.Transactor,
	bucketRepo domain.BucketRepository,
	uploadRepo domain.UploadRepository,
	objRepo domain.ObjectRepository,
	storage domain.Storage,
	hasher domain.Hasher,
	log *slog.Logger,
) *MetadataService {
	return &MetadataService{
		transactor: transactor,
		bucketRepo: bucketRepo,
		uploadRepo: uploadRepo,
		objRepo:    objRepo,
		storage:    storage,
		hasher:     hasher,
		log:        log,
	}
}

func (s *MetadataService) ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Bucket, error) {
	return s.bucketRepo.GetBuckets(ctx, userID, limit, offset)
}

func (s *MetadataService) CreateBucket(ctx context.Context, userID uuid.UUID, name string) (domain.Bucket, error) {
	return s.bucketRepo.CreateBucket(ctx, userID, name)
}

func (s *MetadataService) DeleteBucket(ctx context.Context, userID uuid.UUID, name string) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		bucket, err := s.bucketRepo.GetBucket(ctx, name)
		if err != nil {
			return err
		}
		if bucket.OwnerID != userID {
			return domain.ErrAccessDenied
		}
		return s.bucketRepo.DeleteBucket(ctx, name)
	})
}

func (s *MetadataService) InitUpload(
	ctx context.Context,
	userID uuid.UUID,
	bucket, key string,
	size uint64,
	contentType string,
	systemMetadata map[string]string,
	userMetadata map[string]string,
) (domain.Upload, domain.Storage, error) {
	var saved domain.Upload
	err := s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		bucketMeta, err := s.bucketRepo.GetBucket(ctx, bucket)
		if err != nil {
			return err
		}
		if bucketMeta.OwnerID != userID {
			return domain.ErrAccessDenied
		}

		objPath := fmt.Sprintf("%X", s.hasher.Hash([]byte(bucket+key)))

		upload := domain.Upload{
			Bucket:         bucket,
			Key:            key,
			ObjectPath:     objPath,
			Size:           size,
			StorageNodeID:  s.storage.ID,
			ContentType:    contentType,
			SystemMetadata: systemMetadata,
			UserMetadata:   userMetadata,
			OwnerID:        userID,
		}

		saved, err = s.uploadRepo.CreateUpload(ctx, upload)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return domain.Upload{}, domain.Storage{}, err
	}

	return saved, s.storage, nil
}

func (s *MetadataService) CommitUpload(ctx context.Context, userID, uploadID uuid.UUID, hash string) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		upload, err := s.uploadRepo.GetUpload(ctx, uploadID)
		if err != nil {
			return err
		}
		if upload.OwnerID != userID {
			return domain.ErrAccessDenied
		}
		err = s.uploadRepo.CommitUpload(ctx, uploadID, hash)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *MetadataService) AbortUpload(ctx context.Context, userID, uploadID uuid.UUID) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		upload, err := s.uploadRepo.GetUpload(ctx, uploadID)
		if err != nil {
			return err
		}
		if upload.OwnerID != userID {
			return domain.ErrAccessDenied
		}
		err = s.uploadRepo.DeleteUpload(ctx, uploadID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *MetadataService) GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.Object, domain.Storage, error) {
	obj, err := s.objRepo.GetObject(ctx, userID, bucket, key)
	if err != nil {
		return domain.Object{}, domain.Storage{}, err
	}
	return obj, s.storage, nil
}

func (s *MetadataService) GetObjects(ctx context.Context, userID uuid.UUID, bucket, prefix, delimiter string, limit, offset int) ([]domain.Object, []string, error) {
	bucketMeta, err := s.bucketRepo.GetBucket(ctx, bucket)
	if err != nil {
		return nil, nil, err
	}
	if bucketMeta.OwnerID != userID {
		return nil, nil, domain.ErrAccessDenied
	}
	objs, commonPrefixes, err := s.objRepo.GetObjects(ctx, bucket, prefix, delimiter, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	return objs, commonPrefixes, nil
}

func (s *MetadataService) DeleteObject(ctx context.Context, userID uuid.UUID, bucket, key string) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		bucketMeta, err := s.bucketRepo.GetBucket(ctx, bucket)
		if err != nil {
			return err
		}
		if bucketMeta.OwnerID != userID {
			return domain.ErrAccessDenied
		}
		return s.objRepo.SoftDeleteObject(ctx, bucket, key)
	})
}

func (s *MetadataService) HeadBucket(ctx context.Context, userID uuid.UUID, bucket string) (domain.Bucket, error) {
	meta, err := s.bucketRepo.GetBucket(ctx, bucket)
	if err != nil {
		return domain.Bucket{}, err
	}
	if meta.OwnerID != userID {
		return domain.Bucket{}, domain.ErrAccessDenied
	}
	return meta, nil
}

func (s *MetadataService) HeadObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.Object, error) {
	return s.objRepo.GetObject(ctx, userID, bucket, key)
}
