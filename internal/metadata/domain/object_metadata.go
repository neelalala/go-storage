package domain

import "time"

type ObjectMetadata struct {
	Bucket        string
	Key           string
	Size          uint64
	Checksum      uint32
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StorageNodeID string
}
