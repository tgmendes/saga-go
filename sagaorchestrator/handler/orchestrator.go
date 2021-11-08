package handler

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	profhandler "github.com/tgmendes/saga-go/profile/handler"
	"github.com/tgmendes/saga-go/rabbitmq"
	"github.com/tgmendes/saga-go/sagaorchestrator/repo"
	vehhandler "github.com/tgmendes/saga-go/vehicle/handler"
)

type StateType string

const (
	Start         StateType = ""
	ProfileUpdate StateType = "profile_update"
	VehicleUpdate StateType = "vehicle_update"
	PolicyUpdate  StateType = "policy_update"
	Success       StateType = "success"
	Failure       StateType = "failure"
)

type SagaStorer interface {
	GetSagaLog(ctx context.Context, txID string) (repo.SagaLog, error)
	CreateSagaLog(ctx context.Context, log repo.SagaLog) error
	UpdateSagaState(ctx context.Context, txID string, newState string) error
}

type SagaPayload struct {
	UserID      string                  `json:"user_id"`
	ProfileData profhandler.ProfileData `json:"profile_data"`
	VehicleData vehhandler.VehicleData  `json:"vehicle_data"`
}

type SagaMsg struct {
	TransactionID string     `json:"transaction_id"`
	Command       Transition `json:"command"`
	Payload       []byte     `json:"payload"`
}

type Orchestrator struct {
	mqClient *rabbitmq.Client
	dbClient SagaStorer
	logger   *log.Logger

	stateTable StateTable
}

func NewOrchestrator(mqClient *rabbitmq.Client, dbClient SagaStorer, logger *log.Logger) *Orchestrator {
	o := Orchestrator{
		mqClient: mqClient,
		dbClient: dbClient,
		logger:   logger,
	}

	tt := StateTable{
		string(Start): {
			Next: {
				State:          string(ProfileUpdate),
				TransitionFunc: o.UpdateProfile,
			},
		},
		string(ProfileUpdate): {
			Next: {
				State:          string(VehicleUpdate),
				TransitionFunc: o.UpdateVehicle,
			},
			Compensate: {
				State: string(Failure),
			},
		},
		string(VehicleUpdate): {
			Next: {
				State:          string(PolicyUpdate),
				TransitionFunc: o.UpdatePolicy,
			},
			Compensate: {
				State:          string(ProfileUpdate),
				TransitionFunc: o.RollbackProfile,
			},
		},
		string(PolicyUpdate): {
			Next: {
				State: string(Success),
			},
			Compensate: {
				State:          string(VehicleUpdate),
				TransitionFunc: o.RollbackVehicle,
			},
		},
	}

	o.stateTable = tt
	return &o
}

func (o *Orchestrator) ConsumeReply() error {
	err := o.mqClient.Consume("saga_reply", func(b []byte) {
		go o.processReply(b)
	})

	return err
}

func (o *Orchestrator) processReply(msgB []byte) {
	ctx := context.Background()
	msg := SagaMsg{}
	if err := json.Unmarshal(msgB, &msg); err != nil {
		o.logger.WithError(err).Error("unable to unmarshal reply")
		return
	}

	slog, err := o.dbClient.GetSagaLog(ctx, msg.TransactionID)
	if err != nil {
		o.logger.WithField(
			"transaction_id", msg.TransactionID,
		).WithError(err).Error("unable to get saga log")
		return
	}

	nextState, err := o.stateTable.NextState(slog.CurrentState, msg.Command)
	if err != nil {
		o.logger.WithError(err).Error("couldn't get next state")
		return
	}

	err = o.dbClient.UpdateSagaState(ctx, msg.TransactionID, nextState)
	if err != nil {
		o.logger.WithError(err).Error("couldn't update saga")
		return
	}

	if nextState == string(Success) {
		o.logger.WithField("transaction_id", msg.TransactionID).Info("saga successful")
		return
	}

	if nextState == string(Failure) {
		o.logger.WithField("transaction_id", msg.TransactionID).Error("saga failed")
		return
	}

	updateMsg := SagaMsg{
		TransactionID: msg.TransactionID,
		Payload:       slog.Payload,
	}

	uMsgB, err := json.Marshal(updateMsg)
	if err != nil {
		o.logger.WithError(err).Error("couldn't marshal update message")
		return
	}

	err = o.stateTable.ExecuteNext(slog.CurrentState, msg.Command, uMsgB)
	if err != nil {
		o.logger.WithField("current_state", slog.CurrentState).WithError(err).Error("couldn't proceed to next step")
		return
	}
	o.logger.WithFields(
		log.Fields{
			"transaction_id": msg.TransactionID,
			"next_state":     nextState,
		},
	).Info("published message")
}

func (o *Orchestrator) UpdateProfile(b []byte) {
	txID, fullPload, err := unmarshalMessage(b)
	if err != nil {
		o.logger.WithError(err).Error("couldn't unmarshal payload")
		return
	}

	profPayload, err := json.Marshal(fullPload.ProfileData)
	if err != nil {
		o.logger.WithError(err).Error("couldn't marshal profile payload")
		return
	}

	msg := SagaMsg{TransactionID: txID, Command: Next, Payload: profPayload}
	err = o.publishMessage("profile", msg)
	if err != nil {
		o.logger.WithError(err).Error("couldn't publish profile update message")
	}
}

func (o *Orchestrator) UpdateVehicle(b []byte) {
	txID, fullPload, err := unmarshalMessage(b)
	if err != nil {
		o.logger.WithError(err).Error("couldn't unmarshal payload")
		return
	}

	vehPayload, err := json.Marshal(fullPload.VehicleData)
	if err != nil {
		o.logger.WithError(err).Error("couldn't marshal vehicle payload")
		return
	}

	msg := SagaMsg{TransactionID: txID, Command: Next, Payload: vehPayload}
	err = o.publishMessage("vehicle", msg)
	if err != nil {
		o.logger.WithError(err).Error("couldn't publish vehicle update message")
	}
}

func (o *Orchestrator) UpdatePolicy(b []byte) {
	txID, fullPload, err := unmarshalMessage(b)
	if err != nil {
		o.logger.WithError(err).Error("couldn't unmarshal payload")
		return
	}

	polPayload, err := json.Marshal(fullPload)
	if err != nil {
		o.logger.WithError(err).Error("couldn't marshal policy payload")
		return
	}

	msg := SagaMsg{TransactionID: txID, Command: Next, Payload: polPayload}
	err = o.publishMessage("policy", msg)
	if err != nil {
		o.logger.WithError(err).Error("couldn't publish policy update message")
	}
}

func (o *Orchestrator) RollbackVehicle(b []byte) {
	if err := o.rollback(b, "vehicle"); err != nil {
		o.logger.WithError(err).Error("couldn't publish vehicle rollback message")
	}
}

func (o *Orchestrator) RollbackProfile(b []byte) {
	if err := o.rollback(b, "profile"); err != nil {
		o.logger.WithError(err).Error("couldn't publish profile rollback message")
	}
}

func (o Orchestrator) rollback(b []byte, queue string) error {
	txID, _, err := unmarshalMessage(b)
	if err != nil {
		return err
	}

	msg := SagaMsg{TransactionID: txID, Command: Compensate}

	err = o.publishMessage(queue, msg)
	if err != nil {
		o.logger.WithError(err).Error("couldn't publish profile update message")
	}
	return nil
}

func (o Orchestrator) publishMessage(queue string, msg SagaMsg) error {
	msgB, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = o.mqClient.Publish(queue, msgB)
	if err != nil {
		return err
	}
	return nil
}

func unmarshalMessage(b []byte) (string, SagaPayload, error) {
	var msg SagaMsg
	if err := json.Unmarshal(b, &msg); err != nil {
		log.Printf("couldn't unmarshal message: %s", err)
		return "", SagaPayload{}, nil
	}

	var pload SagaPayload
	if err := json.Unmarshal(msg.Payload, &pload); err != nil {
		log.Printf("couldn't message unmarshal payload: %s", err)
		return "", SagaPayload{}, nil
	}

	return msg.TransactionID, pload, nil
}
