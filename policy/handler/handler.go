package handler

import (
	"encoding/json"
	"log"

	vhandler "github.com/tgmendes/saga-go/vehicle/handler"

	profhandler "github.com/tgmendes/saga-go/profile/handler"

	"github.com/tgmendes/saga-go/rabbitmq"
)

type Transition string

const (
	Next       Transition = "next"
	Compensate Transition = "compensate"
)

type PolicyData struct {
	UserID      string                  `json:"user_id"`
	ProfileData profhandler.ProfileData `json:"profile_data"`
	VehicleData vhandler.VehicleData    `json:"vehicle_data"`
}

type SagaMsg struct {
	TransactionID string     `json:"transaction_id"`
	Command       Transition `json:"command"`
	Payload       []byte     `json:"payload"`
}

type Policy struct {
	mqClient *rabbitmq.Client
}

func NewPolicy(mqClient *rabbitmq.Client) Policy {
	v := Policy{
		mqClient: mqClient,
	}

	return v
}

func (p Policy) Update(req []byte) {
	var msg SagaMsg
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	var payload PolicyData
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		p.handleError(msg.TransactionID, err)
	}

	if msg.Command == Next {
		log.Printf("[TX_%s] updating policy with data: %+v\n", msg.TransactionID, payload)

		replyMsg := SagaMsg{
			TransactionID: msg.TransactionID,
			Command:       Next,
		}
		if payload.UserID == "userABC_3" {
			replyMsg.Command = Compensate
		}
		replyMsgB, err := json.Marshal(replyMsg)
		if err != nil {
			log.Printf("error %s", err)
		}
		_ = p.mqClient.Publish("saga_reply", replyMsgB)
		return
	}

	if msg.Command == Compensate {
		log.Printf("[TX_%s] compensating policy\n", msg.TransactionID)

		replyMsg := SagaMsg{
			TransactionID: msg.TransactionID,
			Command:       Compensate,
		}
		replyMsgB, err := json.Marshal(replyMsg)
		if err != nil {
			log.Printf("error %s", err)
		}
		_ = p.mqClient.Publish("saga_reply", replyMsgB)
	}
}

func (p Policy) handleError(txID string, err error) {
	log.Printf("[TX_%s] something happened: %s", txID, err)
	msg := SagaMsg{TransactionID: txID, Command: Compensate}
	msgB, _ := json.Marshal(msg)
	_ = p.mqClient.Publish("saga_reply", msgB)

}
