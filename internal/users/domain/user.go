package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	DisplayName string
	CreatedAt   time.Time
}

type UserRepository interface {
	CreateUser(ctx context.Context, name string) (User, error)
	GetUserByName(ctx context.Context, name string) (User, error)
}
