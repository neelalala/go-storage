package hash

import "hash/crc32"

type CRC32 struct {
}

func NewCRC32() CRC32 {
	return CRC32{}
}

func (_ CRC32) Checksum(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}
