package domain

type Hasher interface {
	Hash(b []byte) string
}
