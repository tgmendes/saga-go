package handler

import (
	"net/http"

	"github.com/tgmendes/saga-go/rabbitmq"

	"github.com/julienschmidt/httprouter"
)

type ChangeOrchestrator struct {
	mqClient *rabbitmq.Client
}

func NewChangeOrchestrator(mqClient *rabbitmq.Client) http.Handler {
	chg := ChangeOrchestrator{
		mqClient: mqClient,
	}

	router := httprouter.New()
	router.POST("/create", chg.Create)

	return router
}
