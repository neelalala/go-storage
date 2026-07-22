package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

const userKey contextKey = "user"

type Verifier interface {
	Verify(r *http.Request) (domain.User, error)
}

func Auth(next http.HandlerFunc, verifier Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := verifier.Verify(r)
		if err != nil {
			http.Error(w, "couldn't verify", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUser(ctx context.Context) (domain.User, error) {
	user, ok := ctx.Value(userKey).(domain.User)
	if !ok {
		return domain.User{}, errors.New("couldn't get user")
	}
	return user, nil
}
