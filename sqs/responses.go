package sqs

import "fmt"

// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/sqs-api.pdf

const (
	// DummyRequestID to be used in aws like responses
	DummyRequestID = "00000000-0000-0000-0000-000000000000"
)

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

type AddPermissionResponse struct {
	ResponseMetadata ResponseMetadata
}

type CreateQueueResponse struct {
	CreateQueueResult CreateQueueResult
	ResponseMetadata  ResponseMetadata
}

type CreateQueueResult struct {
	QueueUrl string
}

type RemovePermissionResponse struct {
	ResponseMetadata ResponseMetadata
}

type ResponseMetadata struct {
	RequestId string
}

func NewAddPermissionResponse() *AddPermissionResponse {
	return &AddPermissionResponse{
		ResponseMetadata: ResponseMetadata{
			RequestId: DummyRequestID,
		},
	}
}

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

func NewRemovePermissionResponse() *RemovePermissionResponse {
	return &RemovePermissionResponse{
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
