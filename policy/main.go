package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tgmendes/saga-go/rabbitmq"

	"github.com/tgmendes/saga-go/policy/handler"
)

func main() {
	if err := run(); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run() error {
	declaredQueues := []string{"policy", "reply"}
	mqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		declaredQueues,
	)
	if err != nil {
		return fmt.Errorf("failed to start rabbitmq: %w", err)
	}

	p := handler.NewPolicy(mqClient)
	err = mqClient.Consume("policy", p.Update)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	return nil
}
