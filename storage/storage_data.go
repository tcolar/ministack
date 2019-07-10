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
	Name string
}
