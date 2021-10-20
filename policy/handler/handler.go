package handler

import (
	"encoding/json"
	"log"

	vhandler "github.com/tgmendes/saga-go/vehicle/handler"
	vphandler "github.com/tgmendes/saga-go/vehicleprofile/handler"

	phandler "github.com/tgmendes/saga-go/profile/handler"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Command string

const (
	UpdateCmd       Command = "update_policy"
	UpdatedCmd      Command = "policy_updated"
	UpdateFailedCmd Command = "policy_update_failed"
	RollbackCmd     Command = "rollback_policy"
)

type PolicyData struct {
	UserID         string                       `json:"user_id"`
	Profile        phandler.ProfileData         `json:"profile"`
	Vehicle        vhandler.VehicleData         `json:"vehicle"`
	VehicleProfile vphandler.VehicleProfileData `json:"vehicle_profile"`
}

type Message struct {
	Command Command    `json:"command"`
	Payload PolicyData `json:"payload,omitempty"`
}

type Policy struct {
	mqClient *rabbitmq.Client
}

func NewPolicy(mqClient *rabbitmq.Client) Policy {
	vp := Policy{
		mqClient: mqClient,
	}

	return vp
}

func (p Policy) Update(req []byte) {
	var msg Message
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	if msg.Command == UpdateCmd {
		log.Printf("updating policy with data: %+v\n", msg.Payload)
		_ = p.mqClient.Publish("reply", []byte(UpdatedCmd))
		return
	}

	if msg.Command == RollbackCmd {
		log.Printf("rolling back vehicle profile changes to previous state\n")
		return
	}
}
