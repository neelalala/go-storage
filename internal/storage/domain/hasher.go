package domain

type Hasher interface {
	Checksum(b []byte) uint32
}
