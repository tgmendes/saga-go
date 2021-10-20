package handler

import (
	"encoding/json"
	"log"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Command string

const (
	UpdateCmd       Command = "update_vehicle"
	UpdatedCmd      Command = "vehicle_updated"
	UpdateFailedCmd Command = "vehicle_update_failed"
	RollbackCmd     Command = "rollback_vehicle"
)

type VehicleData struct {
	VehicleID string `json:"vehicle_id"`
	Colour    string `json:"colour"`
	VRM       string `json:"vrm"`
}

type Message struct {
	Command Command     `json:"command"`
	Payload VehicleData `json:"payload,omitempty"`
}

type Vehicle struct {
	mqClient *rabbitmq.Client
}

func NewVehicle(mqClient *rabbitmq.Client) Vehicle {
	vp := Vehicle{
		mqClient: mqClient,
	}

	return vp
}

func (vp Vehicle) Update(req []byte) {
	var msg Message
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	if msg.Command == UpdateCmd {
		log.Printf("updating vehicle with data: %+v\n", msg.Payload)
		log.Printf("error: not implemented")
		_ = vp.mqClient.Publish("reply", []byte(UpdateFailedCmd))
		return
	}

	if msg.Command == RollbackCmd {
		log.Printf("rolling back vehicle changes to previous state")
		return
	}
}
