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
	root string
}

func New(root string) FileStore {
	root += func() string {
		if strings.HasSuffix(root, "/") {
			return ""
		}
		return root + "/"
	}()

	return FileStore{
		root: root,
	}
}

func (s FileStore) Save(ctx context.Context, obj domain.Object) error {
	// TODO: if root dir not exists always an error
	if err := os.WriteFile(s.root+obj.Name, obj.Data, 0644); err != nil {
		return fmt.Errorf("error saving object: %w", err)
	}

	return nil
}

func (s FileStore) Get(ctx context.Context, name string) (domain.Object, error) {
	data, err := os.ReadFile(s.root + name)
	if errors.Is(err, os.ErrNotExist) {
		return domain.Object{}, fmt.Errorf("%w: %s", domain.ErrFileNotFound, name)
	}

	return domain.Object{
		Name: name,
		Data: data,
	}, nil
}

func (s FileStore) Delete(ctx context.Context, name string) error {
	if err := os.Remove(s.root + name); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", domain.ErrFileNotFound, name)
		}

		return fmt.Errorf("unexpected error: %v", err)
	}

	return nil
}
