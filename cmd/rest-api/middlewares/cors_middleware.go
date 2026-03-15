package middlewares

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware handles CORS for all requests
// Supports multiple origins via CORS_ALLOWED_ORIGINS env var (comma-separated)
// Falls back to CORS_ALLOWED_ORIGIN for single origin compatibility
type CORSMiddleware struct {
	allowedOrigins map[string]bool
	defaultOrigin  string
}

// NewCORSMiddleware creates a new CORS middleware
// Environment variables:
// - CORS_ALLOWED_ORIGINS: Comma-separated list of allowed origins (preferred)
// - CORS_ALLOWED_ORIGIN: Single origin (fallback for compatibility)
// - Default: http://localhost:3030
func NewCORSMiddleware() *CORSMiddleware {
	m := &CORSMiddleware{
		allowedOrigins: make(map[string]bool),
		defaultOrigin:  "http://localhost:3030",
	}

	// Check for multiple origins (preferred)
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		for _, origin := range strings.Split(origins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				m.allowedOrigins[origin] = true
			}
		}
	}

	// Fallback to single origin
	if singleOrigin := os.Getenv("CORS_ALLOWED_ORIGIN"); singleOrigin != "" {
		m.allowedOrigins[singleOrigin] = true
		m.defaultOrigin = singleOrigin
	}

	// SECURITY: Only add localhost origins in non-production environments
	if env := os.Getenv("APP_ENV"); env == "" || env == "development" || env == "local" {
		m.allowedOrigins["http://localhost:3030"] = true
		m.allowedOrigins["http://localhost:3000"] = true
	}

	return m
}

// isOriginAllowed checks if the request origin is in the allowed list
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	return m.allowedOrigins[origin]
}

// Handler is the middleware function that adds CORS headers to all responses
func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Determine which origin to allow
		allowedOrigin := m.defaultOrigin
		if origin != "" && m.isOriginAllowed(origin) {
			allowedOrigin = origin
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		// IMPORTANT: Only X-Resource-Owner-ID (RID token) is required for authentication.
		// Identity fields (user_id, tenant_id, etc.) are resolved server-side from the RID token.
		// DO NOT add X-User-ID, X-Tenant-ID, X-Client-ID, X-Group-ID — they are not sent by the frontend.
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Resource-Owner-ID, X-Intended-Audience, X-Request-ID, X-Search, x-search, X-API-Key")
		w.Header().Set("Access-Control-Expose-Headers", "X-Resource-Owner-ID, X-Intended-Audience")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
