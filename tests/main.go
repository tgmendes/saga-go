package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	profhandler "github.com/tgmendes/saga-go/profile/handler"
	vehhandler "github.com/tgmendes/saga-go/vehicle/handler"

	"github.com/google/uuid"
	"github.com/tgmendes/saga-go/rabbitmq"
	"github.com/tgmendes/saga-go/sagaorchestrator/handler"
	"github.com/tgmendes/saga-go/sagaorchestrator/repo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx := context.Background()
	declaredQueues := []string{"saga_reply"}
	mqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		declaredQueues,
	)
	if err != nil {
		log.Fatalf("rabbit: %s", err)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
	if err != nil {
		log.Fatalf("mongo: %s", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("mongo connect: %s", err)
	}
	defer client.Disconnect(ctx)

	db := repo.NewMongo(client)

	for i := 0; i < 5; i++ {
		payload := handler.SagaPayload{
			UserID: fmt.Sprintf("userABC_%d", i),
			ProfileData: profhandler.ProfileData{
				UserID:             fmt.Sprintf("userABC_%d", i),
				Surname:            "Bond",
				ResidentialAddress: "Bond street",
				Occupation:         "Spy",
			},
			VehicleData: vehhandler.VehicleData{
				VehicleID: fmt.Sprintf("veh_%d", i),
				Colour:    "red",
				VRM:       "vrm123",
			},
		}
		pb, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("payload: %s", err)
		}

		reqID := uuid.NewString()
		slog := repo.SagaLog{
			TransactionID: reqID,
			CurrentState:  string(handler.Start),
			Payload:       pb,
		}
		err = db.CreateSagaLog(ctx, slog)
		if err != nil {
			log.Fatalf("create log: %s", err)
		}

		publishMsg := handler.SagaMsg{
			TransactionID: reqID,
			Command:       handler.Next,
		}
		pmsgB, err := json.Marshal(publishMsg)
		if err != nil {
			log.Fatalf("reply msg: %s", err)
		}
		_ = mqClient.Publish("saga_reply", pmsgB)
	}
}
