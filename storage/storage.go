package storage

type Store interface {
	Close()
	CreateQueue(name string) error
}
