package handler

import (
	"encoding/json"
	"log"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Command string

const (
	UpdateCmd       Command = "update_vehicle_profile"
	UpdatedCmd      Command = "vehicle_profile_updated"
	UpdateFailedCmd Command = "vehicle_profile_update_failed"
	RollbackCmd     Command = "rollback_vehicle_profile"
)

type VehicleProfileData struct {
	UserID         string `json:"user_id"`
	VehicleID      string `json:"vehicle_id"`
	AnnualMileage  int    `json:"annual_mileage"`
	EstimatedValue int    `json:"estimated_value"`
}

type Message struct {
	Command Command            `json:"command"`
	Payload VehicleProfileData `json:"payload,omitempty"`
}

type VehicleProfile struct {
	mqClient *rabbitmq.Client
}

func NewVehicleProfile(mqClient *rabbitmq.Client) VehicleProfile {
	vp := VehicleProfile{
		mqClient: mqClient,
	}

	return vp
}

func (vp VehicleProfile) Update(req []byte) {
	var msg Message
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	if msg.Command == UpdateCmd {
		log.Printf("updating vehicle profile with data: %+v\n", msg.Payload)
		_ = vp.mqClient.Publish("reply", []byte(UpdatedCmd))
		return
	}

	if msg.Command == RollbackCmd {
		log.Printf("rolling back vehicle profile changes to previous state")
		return
	}
}
