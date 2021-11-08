package handler

import (
	"encoding/json"
	"log"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Transition string

const (
	Next       Transition = "next"
	Compensate Transition = "compensate"
)

type VehicleData struct {
	VehicleID string `json:"vehicle_id"`
	Colour    string `json:"colour"`
	VRM       string `json:"vrm"`
}

type SagaMsg struct {
	TransactionID string     `json:"transaction_id"`
	Command       Transition `json:"command"`
	Payload       []byte     `json:"payload"`
}

type Vehicle struct {
	mqClient *rabbitmq.Client
}

func NewVehicle(mqClient *rabbitmq.Client) Vehicle {
	v := Vehicle{
		mqClient: mqClient,
	}

	return v
}

func (v Vehicle) Update(req []byte) {
	var msg SagaMsg
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	var payload VehicleData
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		v.handleError(msg.TransactionID, err)
	}

	if msg.Command == Next {
		log.Printf("[TX_%s] updating vehicle with data: %+v\n", msg.TransactionID, payload)

		replyMsg := SagaMsg{
			TransactionID: msg.TransactionID,
			Command:       Next,
		}
		replyMsgB, err := json.Marshal(replyMsg)
		if err != nil {
			log.Printf("error %s", err)
		}
		_ = v.mqClient.Publish("saga_reply", replyMsgB)
		return
	}

	if msg.Command == Compensate {
		log.Printf("[TX_%s] compensating vehicle\n", msg.TransactionID)
		replyMsg := SagaMsg{
			TransactionID: msg.TransactionID,
			Command:       Compensate,
		}
		replyMsgB, err := json.Marshal(replyMsg)
		if err != nil {
			log.Printf("error %s", err)
		}
		_ = v.mqClient.Publish("saga_reply", replyMsgB)
	}
}

func (v Vehicle) handleError(txID string, err error) {
	log.Printf("[TX_%s] something happened: %s", txID, err)
	msg := SagaMsg{TransactionID: txID, Command: Compensate}
	msgB, _ := json.Marshal(msg)
	_ = v.mqClient.Publish("saga_reply", msgB)

}
