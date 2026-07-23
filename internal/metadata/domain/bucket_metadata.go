package domain

import (
	"time"

	"github.com/google/uuid"
)

type Bucket struct {
	Name      string
	CreatedAt time.Time
	OwnerID   uuid.UUID
}
