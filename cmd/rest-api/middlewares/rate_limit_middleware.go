package middlewares

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool
	Remaining     int
	Limit         int
	ResetAfter    time.Duration
	RetryAfterSec int
}

// Allow checks if the client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string) bool {
	result := rl.Check(clientIP)
	return result.Allowed
}

// Check performs a rate limit check and returns detailed information
func (rl *RateLimiter) Check(clientIP string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[clientIP]

	if !exists {
		rl.clients[clientIP] = &clientBucket{
			tokens:     rl.rate - 1, // Use one token for this request
			lastRefill: now,
		}
		return RateLimitResult{
			Allowed:    true,
			Remaining:  rl.rate - 1,
			Limit:      rl.rate,
			ResetAfter: rl.interval,
		}
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
		return RateLimitResult{
			Allowed:    true,
			Remaining:  bucket.tokens,
			Limit:      rl.rate,
			ResetAfter: rl.interval - elapsed,
		}
	}

	// Calculate time until next token refill
	timeUntilRefill := rl.interval - elapsed
	if timeUntilRefill < 0 {
		timeUntilRefill = rl.interval
	}
	retryAfterSec := int(timeUntilRefill.Seconds())
	if retryAfterSec < 1 {
		retryAfterSec = 1
	}

	return RateLimitResult{
		Allowed:       false,
		Remaining:     0,
		Limit:         rl.rate,
		ResetAfter:    timeUntilRefill,
		RetryAfterSec: retryAfterSec,
	}
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

// RateLimitErrorResponse represents the error response for rate limiting
type RateLimitErrorResponse struct {
	Success bool              `json:"success"`
	Error   *RateLimitError   `json:"error"`
}

// RateLimitError contains detailed rate limit error information
type RateLimitError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	RetryAfter int    `json:"retry_after_seconds"`
	Limit      int    `json:"limit_per_minute"`
	Remaining  int    `json:"remaining"`
}

// Handler is the HTTP middleware handler
func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for CORS preflight requests
		// OPTIONS requests are lightweight and should not consume rate limit tokens
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)
		result := rlm.limiter.Check(clientIP)

		// Always set rate limit headers for visibility
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(result.ResetAfter).Unix(), 10))

		if !result.Allowed {
			slog.Warn("rate limit exceeded",
				"client_ip", clientIP,
				"path", r.URL.Path,
				"method", r.Method,
				"limit", result.Limit,
				"retry_after_sec", result.RetryAfterSec,
			)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfterSec))
			w.WriteHeader(http.StatusTooManyRequests)

			errorResponse := RateLimitErrorResponse{
				Success: false,
				Error: &RateLimitError{
					Code:       "RATE_LIMIT_EXCEEDED",
					Message:    "Too many requests. Please slow down and try again.",
					Details:    fmt.Sprintf("You have exceeded the rate limit of %d requests per minute. Please wait %d seconds before retrying.", result.Limit, result.RetryAfterSec),
					RetryAfter: result.RetryAfterSec,
					Limit:      result.Limit,
					Remaining:  0,
				},
			}
			_ = json.NewEncoder(w).Encode(errorResponse)
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
