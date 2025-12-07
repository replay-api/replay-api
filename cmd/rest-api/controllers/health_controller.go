package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

var startTime = time.Now()

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

type HealthController struct {
	Container container.Container
}

func NewHealthController(container container.Container) *HealthController {
	return &HealthController{Container: container}
}

// HealthCheck returns a simple health check for liveness probes
func (hc *HealthController) HealthCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Uptime:    time.Since(startTime).Round(time.Second).String(),
		}

		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessCheck returns a comprehensive health check including dependencies
func (hc *HealthController) ReadinessCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)
		overallStatus := "ok"

		// Check MongoDB connection
		var mongoClient *mongo.Client
		if err := hc.Container.Resolve(&mongoClient); err == nil && mongoClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := mongoClient.Ping(ctx, nil); err != nil {
				checks["mongodb"] = "error: " + err.Error()
				overallStatus = "degraded"
			} else {
				checks["mongodb"] = "ok"
			}
		} else {
			checks["mongodb"] = "not configured"
		}

		// Set status code based on health
		statusCode := http.StatusOK
		if overallStatus != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		response := HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Uptime:    time.Since(startTime).Round(time.Second).String(),
			Checks:    checks,
		}

		json.NewEncoder(w).Encode(response)
	}
}
