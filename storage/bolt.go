package storage

import (
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

	// TODO: job to delete messages after a certain time
	// TODO: job to flush invisible messages

	return store, err
}

// Close should be called upon program termination
func (s *BoltStorage) Close() {
	defer s.db.Close()
}

// CreateQueue creates a new SQS queue
func (s *BoltStorage) CreateQueue(name string) error {
	bucketName := s.queueBucket(name)
	s.debug("Creating queue %s (bucket %s)", name, bucketName)
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
			err = s.decode(rawList, &list)
			if err != nil {
				return err
			}
		}
		list.Queues[name] = Queue{name}
		newRawList, err := s.encode(list)
		if err != nil {
			return err
		}
		return sqs.Put(queueList, newRawList)
	})
}

// ListQueues lists the SQS queues
func (s *BoltStorage) ListQueues() (QueueList, error) {
	s.debug("Listing queues")
	list := QueueList{}
	err := s.db.View(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		rawList := sqs.Get(queueList)
		if rawList != nil {
			err := s.decode(rawList, &list)
			if err != nil {
				return err
			}
		}
		return nil
	})
	s.debug("Queue List: %v", list.Keys())
	return list, err
}

// ReceiveMessage receives messages from the queue
func (s *BoltStorage) ReceiveMessage(queueName string, maxMessages int, visibilityTimeoutSeconds int) ([]string, error) {
	s.debug("Receive messages from %s", queueName)
	messages := []string{}
	bucketName := s.queueBucket(queueName)
	uuidIndexbucketName := s.queueUUIDIndexBucket(queueName)
	err := s.db.Update(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		queueBucket := sqs.Bucket(bucketName)
		if queueBucket == nil {
			return fmt.Errorf("Bucket not found for %s", queueName)
		}
		uuidIndexBucket := sqs.Bucket(uuidIndexbucketName)
		if uuidIndexBucket == nil {
			return fmt.Errorf("Bucket not found: %s", uuidIndexbucketName)
		}
		now := time.Now().Unix()
		cursor := queueBucket.Cursor()
		for key, payload := cursor.First(); key != nil && len(messages) < maxMessages; key, payload = cursor.Next() {
			if ParsePayloadKey(key).VisibleAfter > now {
				break // We reached a message that is not visible yet, since it's sorted, we are done.
			}
			var message MessagePayload
			s.decode(payload, &message)
			messages = append(messages, string(payload))
			if visibilityTimeoutSeconds > 0 {
				// Make the message invisible for a while by updating it's timestamp based key
				newKey := ParsePayloadKey(key).ExtendInvisibility(visibilityTimeoutSeconds)
				queueBucket.Put(newKey.Bytes(), payload)
				queueBucket.Delete(key)
				uuidIndexBucket.Put([]byte(message.UUID), newKey.Bytes())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.debug("Received %d messages from %s", len(messages), queueName)
	return messages, nil
}

// SendMessage sends a SQS message to the queue
func (s *BoltStorage) SendMessage(queueName, body string) (messageID string, err error) {
	s.debug("Send message to %s with body size: %d", queueName, len(body))
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
	payloadKey := PayloadKey{
		Sequence:     sequence,
		VisibleAfter: now, // TODO: Add delay if queue has one setup
	}.Bytes()
	payload := &MessagePayload{
		UUID:      id.String(),
		CreatedAt: now,
		Payload:   body,
	}
	payloadBytes, err := s.encode(payload)
	if err != nil {
		return "", err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		sqs := tx.Bucket(bucketSqs)
		bucket := sqs.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", bucketName)
		}
		err = bucket.Put([]byte(payloadKey), payloadBytes)
		if err != nil {
			return err
		}
		uuidIndexBucket := sqs.Bucket(uuidIndexbucketName)
		if uuidIndexBucket == nil {
			return fmt.Errorf("Bucket not found: %s", uuidIndexbucketName)
		}
		return uuidIndexBucket.Put([]byte(id.String()), []byte(payloadKey))
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

// Use JSON for redability, but could swicth to something faster such as gob
func (s *BoltStorage) encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *BoltStorage) decode(payload []byte, v interface{}) error {
	return json.Unmarshal(payload, v)
}
