package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tgmendes/saga-go/profile/repo"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Command string

const (
	UpdateCmd       Command = "update_profile"
	UpdatedCmd      Command = "profile_updated"
	UpdateFailedCmd Command = "profile_update_failed"
	RollbackCmd     Command = "rollback_profile"
)

type ProfileData struct {
	UserID             string `json:"user_id"`
	Surname            string `json:"surname"`
	ResidentialAddress string `json:"residential_address"`
	Occupation         string `json:"occupation"`
}

type Message struct {
	Command Command     `json:"command"`
	Payload ProfileData `json:"payload,omitempty"`
}

type Storer interface {
	Update(ctx context.Context, userID string, profUpdate repo.ProfileUpdate) error
	Rollback(ctx context.Context, userID string) error
}
type Profile struct {
	mqClient *rabbitmq.Client
	repo     Storer
}

func NewProfile(mqClient *rabbitmq.Client, repo Storer) Profile {
	p := Profile{
		mqClient: mqClient,
		repo:     repo,
	}

	return p
}

func (p Profile) Update(req []byte) {
	var msg Message
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	if msg.Command == UpdateCmd {
		log.Printf("updating profile with data: %+v\n", msg.Payload)
		profUpdate := repo.ProfileUpdate{
			Surname:            msg.Payload.Surname,
			ResidentialAddress: msg.Payload.ResidentialAddress,
			Occupation:         msg.Payload.Occupation,
		}
		err := p.repo.Update(context.Background(), msg.Payload.UserID, profUpdate)
		if err != nil {
			log.Printf("error updating profile: %v", err)
			_ = p.mqClient.Publish("reply", []byte(UpdateFailedCmd))
			return
		}
		_ = p.mqClient.Publish("reply", []byte(UpdatedCmd))
		return
	}

	if msg.Command == RollbackCmd {
		log.Printf("rolling back profile changes to previous state\n")
		_ = p.repo.Rollback(context.Background(), msg.Payload.UserID)
	}
}
