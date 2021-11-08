package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tgmendes/saga-go/sagaorchestrator/repo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tgmendes/saga-go/rabbitmq"

	"github.com/tgmendes/saga-go/sagaorchestrator/handler"
)

func main() {
	logger := logrus.New()

	declaredQueues := []string{"saga_reply", "profile", "vehicle", "policy"}
	mqURL := os.Getenv("MQ_URL")
	if mqURL == "" {
		mqURL = "amqp://guest:guest@localhost:5672/"
	}

	mqClient, err := rabbitmq.NewClient(mqURL, declaredQueues)
	if err != nil {
		logger.Fatalf("rabbit: %s", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "mongodb://root:example@localhost:27017"
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(dbURL))
	if err != nil {
		log.Fatalf("mongo: %s", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("mongo connect: %s", err)
	}
	defer client.Disconnect(ctx)

	db := repo.NewMongo(client)

	coord := handler.NewOrchestrator(mqClient, db, logger)

	err = coord.ConsumeReply()
	if err != nil {
		log.Fatalf("consumer: %s", err)
	}
}
