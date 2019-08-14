package storage

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/google/uuid"
)

func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	encoded := base64.StdEncoding.EncodeToString(b)
	return []byte(encoded)
}

func uint64ToUUID(v uint64) uuid.UUID {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[8:], v)
	u, _ := uuid.FromBytes(b)
	return u
}
