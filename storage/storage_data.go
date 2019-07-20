package storage

type QueueList struct {
	Queues map[string]Queue
}

func (q *QueueList) Keys() []string {
	keys := []string{}
	for k := range q.Queues {
		keys = append(keys, k)
	}
	return keys
}

type Queue struct { // TBD
	Name string // TODO: validate max length 80, alphanum _ and -
	// TODO ---------
	//DelaySeconds                  int
	//MaximumMessageSize            int
	//MessageRetentionPeriod        int //seconds
	//ReceiveMessageWaitTimeSeconds int
	//RedrivePolicy                 string
	//VisibilityTimeout             int
	//FifoQueue                     bool // name must end with .fifo
	//ContentBasedDeduplication     bool // MessageDeduplicationId
}
