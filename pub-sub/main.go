package main

import (
	"context"
	"fmt"
	"pubsub/broker"
	"pubsub/client"
	"sync"
	"time"
)

func main() {
	// 1. Initialize the Broker
	b := broker.GetBroker()

	// 2. Create a Topic with 3 partitions
	topicName := "user_events"
	err := b.CreateTopic(topicName, 3)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Topic '%s' created successfully.\n", topicName)

	// 3. Initialize Producer
	producer := client.NewProducer(b)

	// 4. Initialize Consumer Group
	cg := client.NewConsumerGroup("group-1", b)

	// Context to handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup to wait for simulated work to finish
	var wg sync.WaitGroup

	// 5. Start Consumer Group
	fmt.Println("Starting Consumer Group...")
	err = cg.Subscribe(ctx, topicName, func(partition int, key []byte, value []byte) {
		fmt.Printf("[Consumer Group] Processed message from partition %d | Key: %s | Value: %s\n", partition, string(key), string(value))
	})
	if err != nil {
		panic(err)
	}

	// Give consumers a moment to start up
	time.Sleep(100 * time.Millisecond)

	// 6. Produce some messages
	fmt.Println("Producing messages...")
	messages := []struct {
		Key   string
		Value string
	}{
		{"user1", "login"},
		{"user2", "purchase"},
		{"user3", "logout"},
		{"user1", "view_item"}, // Same key 'user1', should go to the same partition as previous 'user1'
	}

	for _, m := range messages {
		wg.Add(1)
		go func(msg struct{ Key, Value string }) {
			defer wg.Done()
			part, offset, err := producer.Publish(topicName, []byte(msg.Key), []byte(msg.Value))
			if err != nil {
				fmt.Printf("Error publishing: %v\n", err)
				return
			}
			fmt.Printf("[Producer] Sent message to partition %d at offset %d | Key: %s\n", part, offset, msg.Key)
		}(m)
	}

	// Wait for all messages to be produced
	wg.Wait()

	// Wait a moment for consumers to process the messages
	time.Sleep(1 * time.Second)
	fmt.Println("Shutting down...")
}
