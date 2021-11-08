package main

import (
	"fmt"
	"log"
	"os"

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
	declaredQueues := []string{"profile", "saga_reply"}
	mqURL := os.Getenv("MQ_URL")
	if mqURL == "" {
		mqURL = "amqp://guest:guest@localhost:5672/"
	}
	mqClient, err := rabbitmq.NewClient(mqURL, declaredQueues)
	if err != nil {
		return fmt.Errorf("failed to start rabbitmq: %w", err)
	}

	v := handler.NewProfile(mqClient)
	err = mqClient.Consume("profile", v.Update)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	return nil
}
