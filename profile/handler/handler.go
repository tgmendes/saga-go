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

type ProfileData struct {
	UserID             string `json:"user_id"`
	Surname            string `json:"surname"`
	ResidentialAddress string `json:"residential_address"`
	Occupation         string `json:"occupation"`
}

type SagaMsg struct {
	TransactionID string     `json:"transaction_id"`
	Command       Transition `json:"command"`
	Payload       []byte     `json:"payload"`
}

type Profile struct {
	mqClient *rabbitmq.Client
}

func NewProfile(mqClient *rabbitmq.Client) Profile {
	v := Profile{
		mqClient: mqClient,
	}

	return v
}

func (p Profile) Update(req []byte) {
	var msg SagaMsg
	if err := json.Unmarshal(req, &msg); err != nil {
		log.Printf("some error")
	}

	var payload ProfileData
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		p.handleError(msg.TransactionID, err)
	}

	if msg.Command == Next {
		log.Printf("[TX_%s] updating profile with data: %+v\n", msg.TransactionID, payload)

		replyMsg := SagaMsg{
			TransactionID: msg.TransactionID,
			Command:       Next,
		}
		replyMsgB, err := json.Marshal(replyMsg)
		if err != nil {
			log.Printf("error %s", err)
		}
		_ = p.mqClient.Publish("saga_reply", replyMsgB)
		return
	}

	if msg.Command == Compensate {
		log.Printf("[TX_%s] compensating profile\n", payload)

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

func (p Profile) handleError(txID string, err error) {
	log.Printf("[TX_%s] something happened: %s", txID, err)
	msg := SagaMsg{TransactionID: txID, Command: Compensate}
	msgB, _ := json.Marshal(msg)
	_ = p.mqClient.Publish("saga_reply", msgB)

}
