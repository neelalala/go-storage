package domain

import "github.com/google/uuid"

type Storage struct {
	ID      uuid.UUID
	Address string
}
