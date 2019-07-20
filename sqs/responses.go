package sqs

import "fmt"

// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/sqs-api.pdf

const (
	// DummyRequestID to be used in aws like responses
	DummyRequestID = "00000000-0000-0000-0000-000000000000"
)

func NewErrorResponse(errType string, errCode string) *ErrorResponse {
	return &ErrorResponse{
		Error: Error{
			Type:    errType,
			Code:    errCode,
			Message: fmt.Sprintf("%s; ; see the SQS docs.", errCode),
		},
		RequestId: DummyRequestID,
	}
}

type ErrorResponse struct {
	Error     Error
	RequestId string
}

type Error struct {
	Type    string
	Code    string
	Message string
	Detail  string
}

type CreateQueueResponse struct {
	CreateQueueResult CreateQueueResult
	ResponseMetadata  ResponseMetadata
}

type CreateQueueResult struct {
	QueueUrl string
}

type ResponseMetadata struct {
	RequestId string
}

func NewListQueueResponse(config *Config, queues []string) *ListQueuesResponse {
	queueUrls := []string{}
	for _, queue := range queues {
		queueUrls = append(queueUrls, fmt.Sprintf("http://%s:%d/queue/%s", config.Host, config.Port, queue))
	}
	return &ListQueuesResponse{
		ListQueuesResult: ListQueuesResult{
			QueueUrl: queueUrls,
		},
		ResponseMetadata: ResponseMetadata{
			RequestId: DummyRequestID,
		},
	}
}

type ListQueuesResponse struct {
	ListQueuesResult ListQueuesResult
	ResponseMetadata ResponseMetadata
}

type ListQueuesResult struct {
	QueueUrl []string
}
