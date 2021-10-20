package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tgmendes/saga-go/profile/repo"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tgmendes/saga-go/profile/handler"

	"github.com/tgmendes/saga-go/rabbitmq"
)

func main() {
	if err := run(); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run() error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	mongoRepo := repo.NewMongo(client)
	declaredQueues := []string{"profile", "reply"}
	mqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		declaredQueues,
	)
	if err != nil {
		return fmt.Errorf("failed to start rabbitmq: %w", err)
	}

	v := handler.NewProfile(mqClient, mongoRepo)
	err = mqClient.Consume("profile", v.Update)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	return nil
}
