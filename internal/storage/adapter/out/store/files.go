package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/neelalala/go-storage/internal/storage/domain"
)

type FileStore struct {
	hasher domain.Hasher
	root   string
}

func New(root string, hasher domain.Hasher) (FileStore, error) {
	root += func() string {
		if strings.HasSuffix(root, "/") {
			return ""
		}
		return root + "/"
	}()
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return FileStore{}, err
	}

	return FileStore{
		hasher: hasher,
		root:   root,
	}, nil
}

func (s FileStore) Save(_ context.Context, obj domain.Object) (string, error) {
	if err := os.WriteFile(s.root+obj.Name, obj.Data, 0644); err != nil {
		return "", fmt.Errorf("error saving object: %w", err)
	}

	etag := s.hasher.Hash(obj.Data)

	return etag, nil
}

func (s FileStore) Get(_ context.Context, name string) ([]byte, error) {
	data, err := os.ReadFile(s.root + name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", domain.ErrFileNotFound, name)
	}

	return data, nil
}

func (s FileStore) Delete(_ context.Context, name string) error {
	if err := os.Remove(s.root + name); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", domain.ErrFileNotFound, name)
		}

		return fmt.Errorf("unexpected error: %v", err)
	}

	return nil
}
