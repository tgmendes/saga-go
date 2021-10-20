package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Client struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewClient(url string, queues []string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	for _, q := range queues {
		_, err = ch.QueueDeclare(
			q,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare queue: %w", err)
		}
	}

	return &Client{
		connection: conn,
		channel:    ch,
	}, nil
}

func (r *Client) Publish(queueName string, body []byte) error {
	err := r.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}
	return nil
}

func (r *Client) Consume(queueName string, consumeFn func(body []byte)) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			consumeFn(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}
