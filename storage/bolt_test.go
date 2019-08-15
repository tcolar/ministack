package storage

import (
	"fmt"
	"log"
	"os"
	"testing"
)

var store Store

var storageConfig = Config{
	Debug:  true,
	DbName: "ministack_test",
}

func TestMain(m *testing.M) {
	var err error
	file := fmt.Sprintf("%s.db", storageConfig.DbName)
	os.Remove(file)
	store, err = NewBoltStorage(&storageConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	result := m.Run()
	// os.Remove(file)
	os.Exit(result)
}

func TestQueues(t *testing.T) {
	queues, err := store.SqsListQueues()
	if err != nil {
		t.Error(err)
	}
	if len(queues.Queues) != 0 {
		t.Errorf("Expected 0 queues")
	}
	err = store.SqsCreateQueue("queue1")
	if err != nil {
		t.Error(err)
	}
	queues, err = store.SqsListQueues()
	fmt.Println(queues)
	if err != nil {
		t.Error(err)
	}
	if len(queues.Queues) != 1 {
		t.Errorf("Expected 1 queues")
	}
	err = store.SqsCreateQueue("queue2")
	if err != nil {
		t.Error(err)
	}
	queues, err = store.SqsListQueues()
	fmt.Println(queues.Keys())
	if err != nil {
		t.Error(err)
	}
	queue1, found := queues.Queues["queue1"]
	if !found {
		t.Errorf("Queue 1 not found")
	}
	if queue1.Name != "queue1" {
		t.Errorf("Queue 1 should be named queue1")
	}
	queue2, found := queues.Queues["queue2"]
	if !found {
		t.Errorf("Queue 1 not found")
	}
	if queue2.Name != "queue2" {
		t.Errorf("Queue 2 should be named queue2")
	}
	store.SqsSendMessage("queue2", "This is a test")
	store.SqsSendMessage("queue2", "This is another test")

	queues, err = store.SqsListQueues()
	fmt.Println(queues.Keys())
	if err != nil {
		t.Error(err)
	}

	msg, err := store.SqsReceiveMessage("queue2", 1, 30)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(msg)
	msg, err = store.SqsReceiveMessage("queue2", 1, 30)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(msg)
	msg, err = store.SqsReceiveMessage("queue2", 1, 30)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(msg)
}
