package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/replay-api/replay-api/pkg/infra/observability"
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
	Container     container.Container
	healthService *observability.HealthService
	appMetrics    *observability.ApplicationMetrics
}

func NewHealthController(c container.Container) *HealthController {
	// Initialize observability services
	healthService := observability.NewHealthService("1.0.0")
	appMetrics := observability.NewApplicationMetrics()

	// Register MongoDB health checker
	var mongoClient *mongo.Client
	if err := c.Resolve(&mongoClient); err == nil && mongoClient != nil {
		healthService.RegisterMongoDBChecker(func(ctx context.Context) error {
			return mongoClient.Ping(ctx, nil)
		})
	}

	return &HealthController{
		Container:     c,
		healthService: healthService,
		appMetrics:    appMetrics,
	}
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
		result := hc.healthService.Check(r.Context())

		// Map to simpler response format for backward compatibility
		checks := make(map[string]string)
		for name, component := range result.Components {
			checks[name] = string(component.Status)
			if component.Message != "" {
				checks[name] += ": " + component.Message
			}
		}

		statusCode := http.StatusOK
		if result.Status == observability.HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		response := HealthResponse{
			Status:    string(result.Status),
			Timestamp: result.Timestamp.Format(time.RFC3339),
			Uptime:    result.Uptime.Round(time.Second).String(),
			Version:   result.Version,
			Checks:    checks,
		}

		json.NewEncoder(w).Encode(response)
	}
}

// DetailedHealthCheck returns full health check result with all metrics
func (hc *HealthController) DetailedHealthCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := hc.healthService.Check(r.Context())

		statusCode := http.StatusOK
		if result.Status == observability.HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if result.Status == observability.HealthStatusDegraded {
			statusCode = http.StatusOK // Still 200 but status shows degraded
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(result)
	}
}

// LivenessCheck returns simple liveness status for Kubernetes
func (hc *HealthController) LivenessCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hc.healthService.Liveness(r.Context()) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT OK"))
		}
	}
}

// MetricsHandler returns the Prometheus metrics endpoint handler
func (hc *HealthController) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// GetAppMetrics returns the application metrics instance for middleware use
func (hc *HealthController) GetAppMetrics() *observability.ApplicationMetrics {
	return hc.appMetrics
}

// GetHealthService returns the health service for registering additional checkers
func (hc *HealthController) GetHealthService() *observability.HealthService {
	return hc.healthService
}

// ComponentHealth returns health for a specific component
func (hc *HealthController) ComponentHealth(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get component name from URL path or query param
		componentName := r.URL.Query().Get("component")
		if componentName == "" {
			http.Error(w, "component parameter required", http.StatusBadRequest)
			return
		}

		result := hc.healthService.Check(r.Context())
		component, exists := result.Components[componentName]
		if !exists {
			http.Error(w, "component not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(component)
	}
}
