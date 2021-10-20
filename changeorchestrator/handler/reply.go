package handler

import (
	"encoding/json"
	"log"

	polhandler "github.com/tgmendes/saga-go/policy/handler"
	phandler "github.com/tgmendes/saga-go/profile/handler"
	vhandler "github.com/tgmendes/saga-go/vehicle/handler"
	vphandler "github.com/tgmendes/saga-go/vehicleprofile/handler"
)

func (c ChangeOrchestrator) HandleReply(cmd []byte, reqData CreateRequest) {
	switch string(cmd) {
	case string(phandler.UpdatedCmd):
		c.UpdateVehicle(reqData.VehicleChange)
	case string(vhandler.UpdatedCmd):
		c.UpdateVehicleProfile(reqData.VehicleProfileChange)
	case string(vphandler.UpdatedCmd):
		data := polhandler.PolicyData{
			UserID:         reqData.UserID,
			Profile:        reqData.ProfileChange,
			Vehicle:        reqData.VehicleChange,
			VehicleProfile: reqData.VehicleProfileChange,
		}
		c.UpdatePolicy(data)
	case string(polhandler.UpdatedCmd):
		log.Println("all changes successful")
	case string(polhandler.UpdateFailedCmd):
		log.Println("rolling back vehicle profile...")
		fallthrough
	case string(vphandler.UpdateFailedCmd):
		log.Println("rolling back vehicle...")
		fallthrough
	case string(vhandler.UpdateFailedCmd):
		log.Println("rolling back profile")
		prollback := phandler.Message{
			Command: phandler.RollbackCmd,
			Payload: phandler.ProfileData{UserID: reqData.UserID},
		}

		b, _ := json.Marshal(prollback)
		_ = c.mqClient.Publish("profile", b)
	case string(phandler.UpdateFailedCmd):
		log.Println("failed updating profile")
	default:
		log.Printf("unrecognized command: %s\n", cmd)
	}
}

func (c ChangeOrchestrator) UpdateProfile(data phandler.ProfileData) {
	pupdate := phandler.Message{
		Command: phandler.UpdateCmd,
		Payload: data,
	}

	// ignoring error for simplicity
	pUpdateReq, _ := json.Marshal(pupdate)

	err := c.mqClient.Publish("profile", pUpdateReq)
	if err != nil {
		log.Printf("error publishing change: %v", err)
	}
}

func (c ChangeOrchestrator) UpdateVehicle(data vhandler.VehicleData) {
	vupdate := vhandler.Message{
		Command: vhandler.UpdateCmd,
		Payload: data,
	}

	// ignoring error for simplicity
	vUpdateReq, _ := json.Marshal(vupdate)

	err := c.mqClient.Publish("vehicle", vUpdateReq)
	if err != nil {
		log.Printf("error publishing change: %v", err)
	}
}

func (c ChangeOrchestrator) UpdateVehicleProfile(data vphandler.VehicleProfileData) {
	pupdate := vphandler.Message{
		Command: vphandler.UpdateCmd,
		Payload: data,
	}

	// ignoring error for simplicity
	pUpdateReq, _ := json.Marshal(pupdate)

	err := c.mqClient.Publish("vehicle_profile", pUpdateReq)
	if err != nil {
		log.Printf("error publishing change: %v", err)
	}
}

func (c ChangeOrchestrator) UpdatePolicy(data polhandler.PolicyData) {
	pupdate := polhandler.Message{
		Command: polhandler.UpdateCmd,
		Payload: data,
	}

	// ignoring error for simplicity
	pupdateReq, _ := json.Marshal(pupdate)

	err := c.mqClient.Publish("policy", pupdateReq)
	if err != nil {
		log.Printf("error publishing change: %v", err)
	}
}
