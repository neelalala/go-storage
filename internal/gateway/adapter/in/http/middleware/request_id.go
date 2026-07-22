package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, err := uuid.NewV7()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("x-amz-request-id", requestID.String())

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetRequestID(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(requestIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
