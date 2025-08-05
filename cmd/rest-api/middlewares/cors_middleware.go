package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type CORSMiddleware struct {
	config common.CORSConfig
}

func NewCORSMiddleware(config common.CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

func (cm *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if cm.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if cm.allowsAllOrigins() {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set allowed methods
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.config.AllowedMethods, ", "))

		// Set allowed headers
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.config.AllowedHeaders, ", "))

		// Set credentials if allowed
		if cm.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set max age for preflight requests
		if cm.config.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cm.config.MaxAge))
		}

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains like *.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

func (cm *CORSMiddleware) allowsAllOrigins() bool {
	for _, origin := range cm.config.AllowedOrigins {
		if origin == "*" {
			return true
		}
	}
	return false
}
