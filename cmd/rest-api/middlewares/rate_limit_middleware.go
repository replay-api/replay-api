package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter using token bucket algorithm
type RateLimiter struct {
	mu          sync.Mutex
	clients     map[string]*clientBucket
	rate        int           // requests per interval
	interval    time.Duration // time interval for rate limit
	cleanupTick time.Duration // cleanup interval for expired entries
}

type clientBucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: maximum requests per interval
// interval: time window for rate limiting
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string]*clientBucket),
		rate:        rate,
		interval:    interval,
		cleanupTick: time.Minute * 5,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes stale client entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		threshold := time.Now().Add(-rl.interval * 2)
		for ip, bucket := range rl.clients {
			if bucket.lastRefill.Before(threshold) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxy/load balancer setups)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Allow checks if the client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[clientIP]

	if !exists {
		rl.clients[clientIP] = &clientBucket{
			tokens:     rl.rate - 1, // Use one token for this request
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed / rl.interval) * rl.rate

	if tokensToAdd > 0 {
		bucket.tokens = min(rl.rate, bucket.tokens+tokensToAdd)
		bucket.lastRefill = now
	}

	// Check if we have tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates an HTTP middleware for rate limiting
type RateLimitMiddleware struct {
	limiter *RateLimiter
}

// NewRateLimitMiddleware creates a new rate limit middleware
// Default: 100 requests per minute
func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: NewRateLimiter(100, time.Minute),
	}
}

// NewRateLimitMiddlewareWithConfig creates a rate limit middleware with custom config
func NewRateLimitMiddlewareWithConfig(rate int, interval time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: NewRateLimiter(rate, interval),
	}
}

// Handler is the HTTP middleware handler
func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rlm.limiter.Allow(clientIP) {
			slog.Warn("rate limit exceeded",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"method", r.Method,
			)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Rate limit exceeded. Please wait before making more requests.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
