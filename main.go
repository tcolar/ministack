package main

import (
	"flag"
	"log"
	"sync"

	"github.com/tcolar/ministack/sqs"
	"github.com/tcolar/ministack/storage"
)

func main() {
	config := parseConfig()

	var wg sync.WaitGroup
	store, err := storage.NewBoltStorage(&config.Storage)
	if err != nil {
		log.Fatal(err)
	}
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

func parseConfig() *Config {
	config := &defaultConfig
	var externalName = flag.String("e", "localhost", "External server name")
	var debug = flag.Bool("debug", false, "Print debug statements")
	var snsPort = flag.Int("snsPort", 4575, "SNS Port")
	var sqsPort = flag.Int("sqsPort", 4576, "SQS Port")

	flag.Parse()

	config.Sqs.Host = *externalName
	config.Sns.Host = *externalName
	config.Sns.Port = *snsPort
	config.Sqs.Port = *sqsPort
	config.Sqs.Debug = *debug
	config.Sns.Debug = *debug
	config.Storage.Debug = *debug

	return config
}
