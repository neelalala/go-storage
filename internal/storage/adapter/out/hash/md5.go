package hash

import (
	"crypto/md5"
	"fmt"

	"github.com/neelalala/go-storage/internal/storage/domain"
)

var _ domain.Hasher = MD5{}

type MD5 struct {
}

func NewMD5() MD5 {
	return MD5{}
}

func (_ MD5) Hash(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}
