package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Header names for request signing
	SignatureHeader   = "X-Signature"
	TimestampHeader   = "X-Timestamp"
	NonceHeader       = "X-Nonce"

	// Maximum age for requests (5 minutes)
	MaxRequestAge = 5 * time.Minute
)

// RequestSigningMiddleware validates HMAC signatures for sensitive operations
type RequestSigningMiddleware struct {
	signingKey     []byte
	enabledPaths   map[string]bool
	nonceCache     map[string]time.Time
	nonceCacheTTL  time.Duration
}

// NewRequestSigningMiddleware creates a new request signing middleware
func NewRequestSigningMiddleware() *RequestSigningMiddleware {
	signingKey := os.Getenv("REQUEST_SIGNING_KEY")
	if signingKey == "" {
		slog.Warn("REQUEST_SIGNING_KEY not set, request signing disabled")
	}

	rsm := &RequestSigningMiddleware{
		signingKey:    []byte(signingKey),
		nonceCacheTTL: MaxRequestAge * 2,
		nonceCache:    make(map[string]time.Time),
		enabledPaths: map[string]bool{
			"/payments":         true,
			"/wallet/withdraw":  true,
			"/wallet/transfer":  true,
		},
	}

	// Start nonce cleanup goroutine
	go rsm.cleanupNonces()

	return rsm
}

// cleanupNonces removes expired nonces from cache
func (rsm *RequestSigningMiddleware) cleanupNonces() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for nonce, expiry := range rsm.nonceCache {
			if now.After(expiry) {
				delete(rsm.nonceCache, nonce)
			}
		}
	}
}

// isSignatureRequired checks if the request path requires signature validation
func (rsm *RequestSigningMiddleware) isSignatureRequired(path string, method string) bool {
	// Only POST/PUT/DELETE on sensitive endpoints require signing
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
		return false
	}

	// Check exact match
	if rsm.enabledPaths[path] {
		return true
	}

	// Check prefix match
	for enabledPath := range rsm.enabledPaths {
		if strings.HasPrefix(path, enabledPath) {
			return true
		}
	}

	return false
}

// verifySignature validates the HMAC signature of a request
func (rsm *RequestSigningMiddleware) verifySignature(body []byte, timestamp, nonce, providedSignature string) bool {
	if len(rsm.signingKey) == 0 {
		// Signing key not configured, skip validation
		return true
	}

	// Create the message to sign: timestamp + nonce + body
	message := timestamp + "." + nonce + "." + string(body)

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, rsm.signingKey)
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(expectedSignature), []byte(providedSignature))
}

// Handler is the HTTP middleware handler
func (rsm *RequestSigningMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Skip if signature not required for this endpoint
		if !rsm.isSignatureRequired(path, method) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip if signing key not configured (dev mode)
		if len(rsm.signingKey) == 0 {
			slog.WarnContext(r.Context(), "request signing disabled, allowing unsigned request",
				"path", path,
				"method", method,
			)
			next.ServeHTTP(w, r)
			return
		}

		// Get signature headers
		signature := r.Header.Get(SignatureHeader)
		timestamp := r.Header.Get(TimestampHeader)
		nonce := r.Header.Get(NonceHeader)

		// Validate required headers
		if signature == "" || timestamp == "" || nonce == "" {
			slog.WarnContext(r.Context(), "missing signature headers",
				"path", path,
				"has_signature", signature != "",
				"has_timestamp", timestamp != "",
				"has_nonce", nonce != "",
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Request signing required for this endpoint",
				"code":    "SIGNATURE_REQUIRED",
			})
			return
		}

		// Validate timestamp (prevent replay attacks)
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid timestamp format",
				"code":    "INVALID_TIMESTAMP",
			})
			return
		}

		requestTime := time.Unix(ts, 0)
		age := time.Since(requestTime)
		if age > MaxRequestAge || age < -MaxRequestAge {
			slog.WarnContext(r.Context(), "request timestamp out of range",
				"path", path,
				"request_time", requestTime,
				"age", age,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Request timestamp expired or invalid",
				"code":    "TIMESTAMP_EXPIRED",
			})
			return
		}

		// Check nonce hasn't been used (prevent replay attacks)
		if _, exists := rsm.nonceCache[nonce]; exists {
			slog.WarnContext(r.Context(), "duplicate nonce detected",
				"path", path,
				"nonce", nonce,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Request nonce already used",
				"code":    "DUPLICATE_NONCE",
			})
			return
		}

		// Read and restore body for signature verification
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to read request body",
				"code":    "BODY_READ_ERROR",
			})
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Verify signature
		if !rsm.verifySignature(body, timestamp, nonce, signature) {
			slog.WarnContext(r.Context(), "invalid request signature",
				"path", path,
				"method", method,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request signature",
				"code":    "INVALID_SIGNATURE",
			})
			return
		}

		// Store nonce to prevent replay
		rsm.nonceCache[nonce] = time.Now().Add(rsm.nonceCacheTTL)

		slog.InfoContext(r.Context(), "request signature verified",
			"path", path,
			"method", method,
		)

		next.ServeHTTP(w, r)
	})
}
