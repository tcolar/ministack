package storage

// Store is the generic interface to the backend storage used to implement functionality
type Store interface {
	// Close should be called upon program termination
	Close()
	// CreateQueue creates a new SQS queue
	CreateQueue(name string) error
	// ListQueues lists the SQS queues
	ListQueues() (QueueList, error)
	// ReceiveMessage receives messages from the queue
	ReceiveMessage(queueName string, maxMessages int, visibilityTimeout int) ([]string, error)
	// Send SQS Message
	SendMessage(queueName, body string) (messageID string, err error)
}
