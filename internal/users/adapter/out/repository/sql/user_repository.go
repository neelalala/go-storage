package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neelalala/go-storage/internal/users/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, name string) (domain.User, error) {
	query := `
		INSERT INTO users (id, display_name)
		VALUES ($1, $2)
		RETURNING id, display_name, created_at
	`

	userID, err := uuid.NewRandom()
	if err != nil {
		return domain.User{}, fmt.Errorf("error: create new user_id: %v", err)
	}

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, userID, name).Scan(
		&user.ID,
		&user.DisplayName,
		&user.CreatedAt,
	); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return domain.User{}, fmt.Errorf("%w: %s", domain.ErrUserAlreadyExists, name)
			}
		}
		return domain.User{}, fmt.Errorf("error: create new user: %v", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByName(ctx context.Context, name string) (domain.User, error) {
	query := `
		SELECT id, display_name, created_at
		FROM users
		WHERE display_name = $1
	`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, name).Scan(
		&user.ID,
		&user.DisplayName,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("%w: %s", domain.ErrUserNotFound, name)
		}
		return domain.User{}, fmt.Errorf("error: get user by name: %v", err)
	}

	return user, nil
}
