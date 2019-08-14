package storage

import "fmt"

// MessagePayload represents how a message is stored in the DB
type MessagePayload struct {
	UUID      string
	CreatedAt int64
	Payload   string
	Key       string
}

// PayloadKey helps creating sortable keys for the payload
type PayloadKey struct {
	VisibleAfter int64
	Sequence     uint64 // in case of ties on timestamp
}

// Byte ordered key (sortable)
func (p *PayloadKey) String() string {
	return fmt.Sprintf("%010d_%019d", p.VisibleAfter, p.Sequence)
}
