package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserByName(ctx context.Context, name string) (User, error)
}

type User struct {
	ID          uuid.UUID
	DisplayName string
}
