package storage

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type BoltStorage struct {
	db *bolt.DB
}

func NewBoltStorage() Store {
	db, err := bolt.Open("ministack.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &BoltStorage{
		db: db,
	}
}
func (s *BoltStorage) Close() {
	defer s.db.Close()
}

func (s *BoltStorage) CreateQueue(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		log.Printf("Creating queue %s\n", name)
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("Error creating queue %s: %s", name, err)
		}
		return nil
	})
}
