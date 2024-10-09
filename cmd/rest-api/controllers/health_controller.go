package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golobby/container/v3"
)

type HealthController struct {
	Container container.Container
}

func NewHealthController(container container.Container) *HealthController {
	return &HealthController{Container: container}
}

func (hc *HealthController) HealthCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(`{ status: "ok" }`)
	}
}
