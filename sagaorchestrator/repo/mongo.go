package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongo(client *mongo.Client) *Mongo {
	return &Mongo{
		client:     client,
		collection: client.Database("saga").Collection("saga_log"),
	}
}

type SagaLog struct {
	TransactionID string `bson:"transaction_id"`
	CurrentState  string `bson:"current_state"`
	Payload       []byte `bson:"payload"`
}

func (m Mongo) GetSagaLog(ctx context.Context, txID string) (SagaLog, error) {
	var log SagaLog
	if err := m.collection.FindOne(ctx, bson.M{"transaction_id": txID}).Decode(&log); err != nil {
		return SagaLog{}, err
	}

	return log, nil
}

func (m *Mongo) CreateSagaLog(ctx context.Context, log SagaLog) error {
	_, err := m.collection.InsertOne(ctx, log)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) UpdateSagaState(ctx context.Context, txID string, newState string) error {
	_, err := m.collection.UpdateOne(ctx, bson.M{"transaction_id": txID}, bson.D{
		{
			"$set", bson.D{{"current_state", newState}},
		},
	})

	if err != nil {
		return err
	}

	return nil
}
