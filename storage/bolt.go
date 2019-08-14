package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
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
	bucketName := s.queueBucket(name)
	log.Printf("Creating queue %s (bucket %s)", name, bucketName)
	return s.db.Update(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		// Upsert a bucket for the queue
		_, err := sqs.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		_, err = sqs.CreateBucketIfNotExists(s.queueUUIDIndexBucket(name))
		if err != nil {
			return err
		}
		// Update the list of queues
		rawList := sqs.Get(queueList)
		list := QueueList{Queues: map[string]Queue{}}
		if rawList != nil {
			err = json.NewDecoder(bytes.NewReader(rawList)).Decode(&list)
			if err != nil {
				return err
			}
		}
		list.Queues[name] = Queue{name}
		var newRawList bytes.Buffer
		err = json.NewEncoder(&newRawList).Encode(list)
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
			err := json.NewDecoder(bytes.NewReader(rawList)).Decode(&list)
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
	var sequence uint64
	bucketName := s.queueBucket(queueName)
	uuidIndexbucketName := s.queueUUIDIndexBucket(queueName)
	err = s.db.Update(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		bucket := sqs.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", bucketName)
		}
		sequence, _ = bucket.NextSequence()
		return nil
	})
	if err != nil {
		return "", err
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	payloadKey := &PayloadKey{
		Sequence:     sequence,
		VisibleAfter: now, // TODO: Add delay if queue has one setup
	}
	payload := &MessagePayload{
		UUID:      id.String(),
		CreatedAt: now,
		Payload:   body,
		Key:       payloadKey.String(),
	}
	var payloadJSON bytes.Buffer
	err = json.NewEncoder(&payloadJSON).Encode(payload)
	if err != nil {
		return "", err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		bucket := sqs.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", bucketName)
		}
		err = bucket.Put([]byte(payloadKey.String()), payloadJSON.Bytes())
		if err != nil {
			return err
		}
		uuidIndexBucket := sqs.Bucket(uuidIndexbucketName)
		if uuidIndexBucket == nil {
			return fmt.Errorf("Bucket not found: %s", uuidIndexbucketName)
		}
		return uuidIndexBucket.Put([]byte(id.String()), []byte(payloadKey.String()))
	})
	if err != nil {
		return "", err
	}
	return id.String(), nil
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

func (s *BoltStorage) queueBucket(queueName string) []byte {
	return []byte(fmt.Sprintf("_queue:%s", queueName))
}

// UUID to payloadKey "index"
func (s *BoltStorage) queueUUIDIndexBucket(queueName string) []byte {
	return []byte(fmt.Sprintf("_queue_uuid_idx:%s", queueName))
}

func (s *BoltStorage) debug(msg string, args ...interface{}) {
	if s.config.Debug {
		log.Printf("DEBUG: %s", fmt.Sprintf(msg, args...))
	}
}
