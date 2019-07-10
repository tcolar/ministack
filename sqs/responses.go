package sqs

import "fmt"

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
