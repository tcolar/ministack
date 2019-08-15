package storage

// Store is the generic interface to the backend storage used to implement functionality
type Store interface {
	// Close should be called upon program termination
	Close()
	// SqsCreateQueue creates a new SQS queue
	SqsCreateQueue(name string) error
	// SqsListQueues lists the SQS queues
	SqsListQueues() (QueueList, error)
	// SqsReceiveMessage receives messages from the queue
	SqsReceiveMessage(queueName string, maxMessages int, visibilityTimeout int) ([]string, error)
	// SqsSendMessage SQS Message
	SqsSendMessage(queueName, body string) (messageID string, err error)
}
