package application

import (
	"context"

	"github.com/neelalala/go-storage/internal/users/domain"
)

type UsersService struct {
	userRepo domain.UserRepository
}

func NewUsersService(userRepo domain.UserRepository) *UsersService {
	return &UsersService{
		userRepo: userRepo,
	}
}

func (s *UsersService) CreateUser(ctx context.Context, name string) (domain.User, error) {
	return s.userRepo.CreateUser(ctx, name)
}

func (s *UsersService) GetUserByName(ctx context.Context, name string) (domain.User, error) {
	return s.userRepo.GetUserByName(ctx, name)
}
