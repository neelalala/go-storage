package hasher

import (
	"crypto/sha256"
)

type SHA256 struct {
}

func NewSHA256() SHA256 {
	return SHA256{}
}

func (_ SHA256) Hash(b []byte) []byte {
	h := sha256.New()

	h.Write(b)
	return h.Sum(nil)
}
