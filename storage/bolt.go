package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

var bucketSqs = []byte("sqs")
var queueList = []byte("queue_list")

// BoltStorage is a Storage impl backed by Bolt DB
type BoltStorage struct {
	db     *bolt.DB
	config *Config
}

// NewBoltStorage creates the BoltDB storage instance
func NewBoltStorage(config *Config) (Store, error) {
	file := fmt.Sprintf("%s.db", config.DbName)
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &BoltStorage{
		db:     db,
		config: config,
	}
	err = store.initBuckets()
	return store, err
}

// Close should be called upon program termination
func (s *BoltStorage) Close() {
	defer s.db.Close()
}

// CreateQueue creates a new SQS queue
func (s *BoltStorage) CreateQueue(name string) error {
	bucket := s.toBucketName(name)
	log.Printf("Creating queue %s (bucket %s)", name, bucket)
	return s.db.Update(func(tx *bolt.Tx) error {
		// Upsert a bucket for the queue
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		// Update the list of queues
		sqs := tx.Bucket(bucketSqs)
		rawList := sqs.Get(queueList)
		list := QueueList{Queues: map[string]Queue{}}
		if rawList != nil {
			err = gob.NewDecoder(bytes.NewReader(rawList)).Decode(&list)
			if err != nil {
				return err
			}
		}
		list.Queues[name] = Queue{name}
		var newRawList bytes.Buffer
		err = gob.NewEncoder(&newRawList).Encode(list)
		if err != nil {
			return err
		}
		return sqs.Put(queueList, newRawList.Bytes())
	})
}

// ListQueues lists the SQS queues
func (s *BoltStorage) ListQueues() (QueueList, error) {
	log.Println("Listing queues")
	list := QueueList{}
	err := s.db.View(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		rawList := sqs.Get(queueList)
		if rawList != nil {
			err := gob.NewDecoder(bytes.NewReader(rawList)).Decode(&list)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if s.config.Debug {
		s.debug("Queue List: %v", list.Keys())
	}
	return list, err
}

// SendMessage sends a SQS message to the queue
func (s *BoltStorage) SendMessage(queueName, body string) (messageID string, err error) {
	var id uint64
	bucketName := s.toBucketName(queueName)
	s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", bucketName)
		}
		id, _ = bucket.NextSequence()
		return bucket.Put(uint64ToBytes(id), []byte(body))
	})
	messageID = uint64ToUUID(id).String()
	return messageID, nil
}

func (s *BoltStorage) initBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketSqs)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *BoltStorage) toBucketName(queueName string) []byte {
	return []byte(fmt.Sprintf("_queue_%s", queueName))
}

func (s *BoltStorage) debug(msg string, args ...interface{}) {
	if s.config.Debug {
		log.Printf("DEBUG: %s", fmt.Sprintf(msg, args...))
	}
}
