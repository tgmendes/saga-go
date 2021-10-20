package repo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/options"

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
		collection: client.Database("profile").Collection("profile"),
	}
}

type Profile struct {
	UserID             string `bson:"user_id"`
	Name               string `bson:"name"`
	Surname            string `bson:"surname"`
	ResidentialAddress string `bson:"residential_address"`
	Occupation         string `bson:"occupation"`
	Revision           int    `bson:"revision"`
}

type ProfileUpdate struct {
	Surname            string `bson:"surname"`
	ResidentialAddress string `bson:"residential_address"`
	Occupation         string `bson:"occupation"`
}

func (m *Mongo) Update(ctx context.Context, userID string, profUpdate ProfileUpdate) error {
	findOptions := options.FindOneOptions{}
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"revision", -1}})

	var prevProfile Profile
	if err := m.collection.FindOne(ctx, bson.M{"user_id": userID}, &findOptions).Decode(&prevProfile); err != nil {
		return fmt.Errorf("could not find document: %w", err)
	}

	newProfile := prevProfile
	if profUpdate.Surname != "" {
		newProfile.Surname = profUpdate.Surname
	}

	if profUpdate.ResidentialAddress != "" {
		newProfile.ResidentialAddress = profUpdate.ResidentialAddress
	}

	if profUpdate.Occupation != "" {
		newProfile.Occupation = profUpdate.Occupation
	}

	newProfile.Revision = prevProfile.Revision + 1
	_, err := m.collection.InsertOne(ctx, newProfile)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) Rollback(ctx context.Context, userID string) error {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"revision", -1}})
	findOptions.SetLimit(2)

	cur, err := m.collection.Find(ctx, bson.M{"user_id": userID}, findOptions)
	if err != nil {
		return fmt.Errorf("could not find document: %w", err)
	}

	var profiles []Profile
	if err = cur.All(ctx, &profiles); err != nil {
		return fmt.Errorf("could not retrieve document")
	}

	// set the new profile to the previous revision
	newProfile := profiles[1]
	newProfile.Revision = profiles[0].Revision + 1
	_, err = m.collection.InsertOne(ctx, newProfile)
	if err != nil {
		return err
	}
	return nil
}
