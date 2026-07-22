package hash

import (
	"hash/crc32"

	"github.com/neelalala/go-storage/internal/storage/domain"
)

var _ domain.Hasher = MD5{}

type MD5 struct {
}

func NewMD5() MD5 {
	return MD5{}
}

func (_ MD5) Hash(b []byte) string {
	return crc32.ChecksumIEEE(b)
}
