package main

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

const (
	projectID = "umroo-app-414608"
	topicName = "devices"
	subName   = "device-id"
)

func PublishToPubSub(data []byte) {

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	topic := client.Topic(topicName)

	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Printf("topic %v does not exist - creating it", topicName)
		_, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Publish a text message on the topic.
	_, err = topic.Publish(ctx, &pubsub.Message{
		Data: data,
	}).Get(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Published a message to the topic %v", topicName)

}
