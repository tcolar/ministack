package main

import (
	"os"
	"sync"

	"github.com/tcolar/ministack/sns"
	"github.com/tcolar/ministack/sqs"
	"github.com/tcolar/ministack/storage"
)

var config = Config{
	Sns: sns.SnsConfig{
		Enabled: false,
		Port:    3000,
		Host:    "localhost",
	},
	Sqs: sqs.SqsConfig{
		Enabled: true,
		Port:    3001,
		Host:    "localhost",
	},
}

func main() {
	host := os.Getenv("HOSTNAME_EXTERNAL")
	if len(host) > 0 {
		config.Sqs.Host = host
		config.Sns.Host = host
	}
	var wg sync.WaitGroup
	store := storage.NewBoltStorage()
	defer store.Close()
	if config.Sns.Enabled {
		//wg.Add(1)
		//go sns.NewServer(&config.sns).Start()
	}
	if config.Sqs.Enabled {
		wg.Add(1)
		go sqs.NewServer(&config.Sqs, store).Start()
	}
	wg.Wait()
}
