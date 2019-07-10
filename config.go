package main

import (
	"github.com/tcolar/ministack/sns"
	"github.com/tcolar/ministack/sqs"
)

// Config for ministack
type Config struct {
	Sns sns.SnsConfig
	Sqs sqs.SqsConfig
}
