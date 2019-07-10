package main

import (
	"github.com/tcolar/ministack/sns"
	"github.com/tcolar/ministack/sqs"
	"github.com/tcolar/ministack/storage"
)

// Config for ministack
type Config struct {
	Sns     sns.Config
	Sqs     sqs.Config
	Storage storage.Config
}

var defaultConfig = Config{
	Sns: sns.Config{
		Enabled: false,
		Port:    3000,
		Host:    "localhost",
		Debug:   false,
	},
	Sqs: sqs.Config{
		Enabled: true,
		Port:    3001,
		Host:    "localhost",
		Debug:   false,
	},
	Storage: storage.Config{
		Debug: false,
	},
}
