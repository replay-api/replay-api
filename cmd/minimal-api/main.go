package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/replay-api/replay-api/pkg/infra/metrics"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "healthy",
		Service: "replay-api",
		Version: "1.0.0-minimal",
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "LeetGaming Replay API - Minimal Mode",
			"status":  "running",
		})
	})

	mux.Handle("/metrics", metrics.Handler())

	handler := metrics.Middleware(mux)

	log.Printf("Starting minimal API server on port %s", port)
	log.Printf("Prometheus metrics available at /metrics")
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
