package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

type SimpleVerifier struct {
	userService domain.UserService
}

func NewSimpleVerifier(userService domain.UserService) *SimpleVerifier {
	return &SimpleVerifier{
		userService: userService,
	}
}

func (v *SimpleVerifier) Verify(r *http.Request) (domain.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return domain.User{}, errors.New("missing authorization header")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "Username" {
		return domain.User{}, errors.New("invalid authorization header format")
	}

	username := parts[1]
	user, err := v.userService.GetUserByName(r.Context(), username)
	if err != nil {
		return domain.User{}, errors.New("user not found")
	}

	return user, nil
}
