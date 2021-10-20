package handler

import (
	"encoding/json"
	"log"
	"net/http"

	phandler "github.com/tgmendes/saga-go/profile/handler"
	vhandler "github.com/tgmendes/saga-go/vehicle/handler"
	vphandler "github.com/tgmendes/saga-go/vehicleprofile/handler"

	"github.com/julienschmidt/httprouter"
)

type CreateRequest struct {
	UserID               string                       `json:"user_id"`
	ProfileChange        phandler.ProfileData         `json:"profile_change"`
	VehicleChange        vhandler.VehicleData         `json:"vehicle_change"`
	VehicleProfileChange vphandler.VehicleProfileData `json:"vehicle_profile_change"`
}

func (c ChangeOrchestrator) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req CreateRequest
	decode := json.NewDecoder(r.Body)
	if err := decode.Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.UpdateProfile(req.ProfileChange)

	go func() {
		err := c.mqClient.Consume("reply", func(b []byte) {
			c.HandleReply(b, req)
		})
		log.Fatalf("unable to consume: %v\n", err)
	}()

	w.WriteHeader(http.StatusNoContent)
}
