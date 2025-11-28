package routing

import (
	"net/http"
	"os"
)

func EnableCors(w *http.ResponseWriter) {
	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3030"
	}
	(*w).Header().Set("Access-Control-Allow-Origin", corsOrigin)
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Resource-Owner-ID, X-Intended-Audience")
	(*w).Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")
}

func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	w.WriteHeader(http.StatusOK)
}
