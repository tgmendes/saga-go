package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tgmendes/saga-go/rabbitmq"

	"github.com/tgmendes/saga-go/changeorchestrator/handler"
)

func main() {
	if err := run(); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run() error {
	declaredQueues := []string{"profile", "vehicle_profile", "vehicle", "reply"}
	mqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		declaredQueues,
	)
	if err != nil {
		log.Fatalf("error %v", err)
	}

	// On Interrupt/SIGTERM tell net/http server to shutdown gracefully.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	server := http.Server{
		Addr:         ":5000",
		Handler:      handler.NewChangeOrchestrator(mqClient),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main: API listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("%s: %v", "server error", err)

	case <-shutdown:
		log.Println("Starting API Shutdown")

		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			log.Printf("Graceful shutdown did not complete in %v: %v", timeout, err)
			err = server.Close()
		}

		if err != nil {
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
	}

	return nil
}
