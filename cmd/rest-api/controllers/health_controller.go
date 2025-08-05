package controllers

import (
	"context"
	"net/http"

	"github.com/golobby/container/v3"
)

type HealthController struct {
	Container container.Container
	helper    *ControllerHelper
}

func NewHealthController(container container.Container) *HealthController {
	return &HealthController{
		Container: container,
		helper:    NewControllerHelper(),
	}
}

func (hc *HealthController) HealthCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthStatus := map[string]string{"status": "ok"}
		hc.helper.WriteOK(w, r, healthStatus)
	}
}
