package storage

// Store is the generic interface to the backend storage used to implement functionality
type Store interface {
	// Close should be called upon program termination
	Close()
	// CreateQueue creates a new SQS queue
	CreateQueue(name string) error
	// ListQueues lists the SQS queues
	ListQueues() (QueueList, error)
	// Send SQS Message
	SendMessage(url, body string) (messageID string, err error)
}
