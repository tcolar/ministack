package storage

import (
	"fmt"
	"strconv"
	"strings"
)

// MessagePayload represents how a message is stored in the DB
type MessagePayload struct {
	UUID      string
	CreatedAt int64
	Payload   string
}

// PayloadKey helps creating sortable keys for the payload
type PayloadKey struct {
	VisibleAfter int64
	Sequence     uint64 // for sorting and uniqueness in case of ties on VisibleAfter
}

// ParsePayloadKey parses a payloadKey from it's []byte representation
func ParsePayloadKey(key []byte) PayloadKey {
	parts := strings.Split(string(key), "_")
	v, _ := strconv.ParseInt(parts[0], 10, 64)
	s, _ := strconv.ParseUint(parts[1], 10, 64)
	return PayloadKey{
		VisibleAfter: v,
		Sequence:     s,
	}
}

// ExtendInvisibility returns a copy of the key with the Visibility extended by x seconds
func (p PayloadKey) ExtendInvisibility(seconds int) PayloadKey {
	return PayloadKey{
		VisibleAfter: p.VisibleAfter + int64(seconds*1000),
		Sequence:     p.Sequence,
	}
}

// Bytes return a  ordered key (sortable)
func (p PayloadKey) Bytes() []byte {
	return []byte(fmt.Sprintf("%010d_%019d", p.VisibleAfter, p.Sequence))
}
