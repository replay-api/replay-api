package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// RateLimitTier defines different rate limiting tiers
type RateLimitTier string

const (
	TierAnonymous   RateLimitTier = "anonymous"
	TierFree        RateLimitTier = "free"
	TierPro         RateLimitTier = "pro"
	TierEnterprise  RateLimitTier = "enterprise"
	TierInternal    RateLimitTier = "internal"
	TierWhitelisted RateLimitTier = "whitelisted"
)

// TierConfig defines rate limits for each tier
type TierConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	CooldownPeriod    time.Duration `json:"cooldown_period"`
	MaxConcurrent     int           `json:"max_concurrent"`
}

// DefaultTierConfigs provides sensible defaults for each tier
var DefaultTierConfigs = map[RateLimitTier]TierConfig{
	TierAnonymous:   {RequestsPerMinute: 30, BurstSize: 5, CooldownPeriod: 5 * time.Minute, MaxConcurrent: 3},
	TierFree:        {RequestsPerMinute: 60, BurstSize: 10, CooldownPeriod: 2 * time.Minute, MaxConcurrent: 5},
	TierPro:         {RequestsPerMinute: 300, BurstSize: 50, CooldownPeriod: 30 * time.Second, MaxConcurrent: 20},
	TierEnterprise:  {RequestsPerMinute: 1000, BurstSize: 200, CooldownPeriod: 10 * time.Second, MaxConcurrent: 100},
	TierInternal:    {RequestsPerMinute: 10000, BurstSize: 1000, CooldownPeriod: 0, MaxConcurrent: 500},
	TierWhitelisted: {RequestsPerMinute: 100000, BurstSize: 10000, CooldownPeriod: 0, MaxConcurrent: 1000},
}

// EndpointSensitivity defines how sensitive an endpoint is
type EndpointSensitivity int

const (
	SensitivityLow      EndpointSensitivity = 1
	SensitivityMedium   EndpointSensitivity = 2
	SensitivityHigh     EndpointSensitivity = 3
	SensitivityCritical EndpointSensitivity = 4
)

// ThreatLevel indicates the current threat assessment
type ThreatLevel int

const (
	ThreatNone     ThreatLevel = 0
	ThreatLow      ThreatLevel = 1
	ThreatMedium   ThreatLevel = 2
	ThreatHigh     ThreatLevel = 3
	ThreatCritical ThreatLevel = 4
)

// ClientState tracks the state of a client for adaptive limiting
type ClientState struct {
	mu               sync.RWMutex
	ClientID         string
	UserID           *uuid.UUID
	Tier             RateLimitTier
	
	// Token bucket
	Tokens           float64
	LastRefill       time.Time
	
	// Request tracking
	RequestCount     int64
	SuccessCount     int64
	FailureCount     int64
	Last4xxCount     int64
	Last5xxCount     int64
	
	// Timing
	FirstRequestAt   time.Time
	LastRequestAt    time.Time
	AverageLatency   time.Duration
	
	// Threat assessment
	ThreatScore      float64
	ThreatLevel      ThreatLevel
	ConsecutiveErrors int
	
	// Blocking
	BlockedUntil     *time.Time
	BlockCount       int
	
	// Concurrent requests
	ActiveRequests   int32
	
	// Fingerprinting
	UserAgents       map[string]int
	Endpoints        map[string]int
	Methods          map[string]int
}

// AdaptiveRateLimiter implements intelligent rate limiting with DDoS protection
type AdaptiveRateLimiter struct {
	mu             sync.RWMutex
	clients        map[string]*ClientState
	tierConfigs    map[RateLimitTier]TierConfig
	
	// Global metrics
	globalRequests int64
	globalBlocked  int64
	globalErrors   int64
	startTime      time.Time
	
	// Adaptive thresholds
	systemLoad     float64
	threatLevel    ThreatLevel
	
	// Configuration
	cleanupInterval     time.Duration
	blockEscalation     []time.Duration
	sensitiveEndpoints  map[string]EndpointSensitivity
	
	// Circuit breaker
	circuitOpen     bool
	circuitOpenedAt *time.Time
	circuitTimeout  time.Duration
	
	// Callbacks
	onBlock         func(clientID string, reason string)
	onThreatDetect  func(clientID string, level ThreatLevel, indicators []string)
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(tierConfigs map[RateLimitTier]TierConfig) *AdaptiveRateLimiter {
	if tierConfigs == nil {
		tierConfigs = DefaultTierConfigs
	}

	arl := &AdaptiveRateLimiter{
		clients:         make(map[string]*ClientState),
		tierConfigs:     tierConfigs,
		startTime:       time.Now(),
		cleanupInterval: 5 * time.Minute,
		blockEscalation: []time.Duration{
			1 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
			1 * time.Hour,
			24 * time.Hour,
		},
		sensitiveEndpoints: map[string]EndpointSensitivity{
			"/auth/signin":           SensitivityCritical,
			"/auth/signup":           SensitivityHigh,
			"/auth/password-reset":   SensitivityCritical,
			"/auth/mfa":              SensitivityCritical,
			"/wallet/withdraw":       SensitivityCritical,
			"/payments":              SensitivityHigh,
			"/admin":                 SensitivityCritical,
		},
		circuitTimeout: 30 * time.Second,
	}

	// Start background tasks
	go arl.runCleanup()
	go arl.runThreatAssessment()

	return arl
}

// Allow checks if a request should be allowed
func (arl *AdaptiveRateLimiter) Allow(ctx context.Context, req *RateLimitRequest) *RateLimitResult {
	clientState := arl.getOrCreateClient(req.ClientID, req.Tier)

	// Check circuit breaker
	if arl.isCircuitOpen() {
		return &RateLimitResult{
			Allowed:      false,
			Reason:       "Service temporarily unavailable due to high load",
			RetryAfter:   arl.circuitTimeout,
			ThreatLevel:  arl.threatLevel,
		}
	}

	// Check if client is blocked
	if blocked, remaining := arl.isBlocked(clientState); blocked {
		return &RateLimitResult{
			Allowed:      false,
			Reason:       "Rate limit exceeded. Client temporarily blocked.",
			RetryAfter:   remaining,
			BlockedUntil: clientState.BlockedUntil,
			ThreatLevel:  clientState.ThreatLevel,
		}
	}

	// Check concurrent request limit
	config := arl.tierConfigs[req.Tier]
	if int(atomic.LoadInt32(&clientState.ActiveRequests)) >= config.MaxConcurrent {
		return &RateLimitResult{
			Allowed:     false,
			Reason:      "Too many concurrent requests",
			RetryAfter:  100 * time.Millisecond,
			ThreatLevel: clientState.ThreatLevel,
		}
	}

	// Token bucket check with adaptive rate
	allowed := arl.consumeToken(clientState, req)
	if !allowed {
		arl.handleRateLimitExceeded(clientState, req)
		return &RateLimitResult{
			Allowed:     false,
			Reason:      "Rate limit exceeded",
			RetryAfter:  arl.calculateRetryAfter(clientState, config),
			Remaining:   0,
			Limit:       config.RequestsPerMinute,
			ThreatLevel: clientState.ThreatLevel,
		}
	}

	// Update metrics
	arl.recordRequest(clientState, req)
	atomic.AddInt32(&clientState.ActiveRequests, 1)

	return &RateLimitResult{
		Allowed:     true,
		Remaining:   int(clientState.Tokens),
		Limit:       config.RequestsPerMinute,
		ThreatLevel: clientState.ThreatLevel,
		RequestID:   uuid.New().String(),
	}
}

// Complete marks a request as complete
func (arl *AdaptiveRateLimiter) Complete(clientID string, statusCode int, latency time.Duration) {
	arl.mu.RLock()
	state, exists := arl.clients[clientID]
	arl.mu.RUnlock()

	if !exists {
		return
	}

	atomic.AddInt32(&state.ActiveRequests, -1)

	state.mu.Lock()
	defer state.mu.Unlock()

	// Update latency average
	alpha := 0.1 // Exponential moving average factor
	state.AverageLatency = time.Duration(float64(state.AverageLatency)*(1-alpha) + float64(latency)*alpha)

	// Track response codes
	if statusCode >= 200 && statusCode < 400 {
		state.SuccessCount++
		state.ConsecutiveErrors = 0
	} else if statusCode >= 400 && statusCode < 500 {
		state.Last4xxCount++
		state.FailureCount++
		state.ConsecutiveErrors++
		arl.assessThreat(state, statusCode)
	} else if statusCode >= 500 {
		state.Last5xxCount++
		state.FailureCount++
		atomic.AddInt64(&arl.globalErrors, 1)
	}
}

// consumeToken attempts to consume a token from the bucket
func (arl *AdaptiveRateLimiter) consumeToken(state *ClientState, req *RateLimitRequest) bool {
	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	config := arl.tierConfigs[state.Tier]

	// Calculate adaptive rate based on threat level and system load
	adaptiveRate := float64(config.RequestsPerMinute)
	adaptiveRate *= arl.getThreatMultiplier(state.ThreatLevel)
	adaptiveRate *= arl.getLoadMultiplier()
	adaptiveRate *= arl.getEndpointMultiplier(req.Endpoint)

	// Refill tokens
	elapsed := now.Sub(state.LastRefill)
	tokensToAdd := elapsed.Seconds() * (adaptiveRate / 60.0)
	state.Tokens = math.Min(float64(config.BurstSize), state.Tokens+tokensToAdd)
	state.LastRefill = now

	// Consume token
	if state.Tokens >= 1 {
		state.Tokens--
		return true
	}

	return false
}

// getThreatMultiplier reduces rate limit based on threat level
func (arl *AdaptiveRateLimiter) getThreatMultiplier(level ThreatLevel) float64 {
	switch level {
	case ThreatLow:
		return 0.8
	case ThreatMedium:
		return 0.5
	case ThreatHigh:
		return 0.2
	case ThreatCritical:
		return 0.05
	default:
		return 1.0
	}
}

// getLoadMultiplier adjusts rate based on system load
func (arl *AdaptiveRateLimiter) getLoadMultiplier() float64 {
	if arl.systemLoad > 0.9 {
		return 0.3
	} else if arl.systemLoad > 0.7 {
		return 0.6
	} else if arl.systemLoad > 0.5 {
		return 0.8
	}
	return 1.0
}

// getEndpointMultiplier adjusts rate based on endpoint sensitivity
func (arl *AdaptiveRateLimiter) getEndpointMultiplier(endpoint string) float64 {
	sensitivity, exists := arl.sensitiveEndpoints[endpoint]
	if !exists {
		return 1.0
	}

	switch sensitivity {
	case SensitivityCritical:
		return 0.1 // Only 10% of normal rate for critical endpoints
	case SensitivityHigh:
		return 0.3
	case SensitivityMedium:
		return 0.6
	default:
		return 1.0
	}
}

// assessThreat updates threat assessment for a client
func (arl *AdaptiveRateLimiter) assessThreat(state *ClientState, statusCode int) {
	// Increase threat score based on behavior
	if statusCode == 401 || statusCode == 403 {
		state.ThreatScore += 10 // Auth failures are suspicious
	} else if statusCode == 400 {
		state.ThreatScore += 2 // Bad requests
	} else if statusCode == 404 {
		state.ThreatScore += 1 // Probing
	}

	// Detect patterns
	indicators := []string{}

	// High error rate
	if state.RequestCount > 10 {
		errorRate := float64(state.FailureCount) / float64(state.RequestCount)
		if errorRate > 0.5 {
			state.ThreatScore += 20
			indicators = append(indicators, "high_error_rate")
		}
	}

	// Many consecutive errors (brute force pattern)
	if state.ConsecutiveErrors > 5 {
		state.ThreatScore += float64(state.ConsecutiveErrors) * 5
		indicators = append(indicators, "consecutive_errors")
	}

	// Request velocity (too fast)
	if state.RequestCount > 0 && state.FirstRequestAt != state.LastRequestAt {
		requestsPerSecond := float64(state.RequestCount) / state.LastRequestAt.Sub(state.FirstRequestAt).Seconds()
		if requestsPerSecond > 10 {
			state.ThreatScore += requestsPerSecond
			indicators = append(indicators, "high_velocity")
		}
	}

	// Multiple user agents (bot rotation)
	if len(state.UserAgents) > 5 {
		state.ThreatScore += float64(len(state.UserAgents)) * 3
		indicators = append(indicators, "ua_rotation")
	}

	// Update threat level
	previousLevel := state.ThreatLevel
	switch {
	case state.ThreatScore >= 100:
		state.ThreatLevel = ThreatCritical
	case state.ThreatScore >= 50:
		state.ThreatLevel = ThreatHigh
	case state.ThreatScore >= 25:
		state.ThreatLevel = ThreatMedium
	case state.ThreatScore >= 10:
		state.ThreatLevel = ThreatLow
	default:
		state.ThreatLevel = ThreatNone
	}

	// Callback on threat level increase
	if state.ThreatLevel > previousLevel && arl.onThreatDetect != nil {
		go arl.onThreatDetect(state.ClientID, state.ThreatLevel, indicators)
	}
}

// handleRateLimitExceeded handles when rate limit is exceeded
func (arl *AdaptiveRateLimiter) handleRateLimitExceeded(state *ClientState, req *RateLimitRequest) {
	state.mu.Lock()
	defer state.mu.Unlock()

	atomic.AddInt64(&arl.globalBlocked, 1)

	// Increase threat score
	state.ThreatScore += 5

	// Determine block duration based on escalation
	blockIndex := state.BlockCount
	if blockIndex >= len(arl.blockEscalation) {
		blockIndex = len(arl.blockEscalation) - 1
	}

	blockDuration := arl.blockEscalation[blockIndex]

	// Apply threat level multiplier to block duration
	blockDuration = time.Duration(float64(blockDuration) * (1 + float64(state.ThreatLevel)*0.5))

	blockedUntil := time.Now().Add(blockDuration)
	state.BlockedUntil = &blockedUntil
	state.BlockCount++

	slog.Warn("Client rate limited",
		"client_id", state.ClientID,
		"threat_level", state.ThreatLevel,
		"block_duration", blockDuration,
		"block_count", state.BlockCount,
	)

	if arl.onBlock != nil {
		go arl.onBlock(state.ClientID, fmt.Sprintf("Rate limit exceeded, blocked for %s", blockDuration))
	}
}

// isBlocked checks if a client is currently blocked
func (arl *AdaptiveRateLimiter) isBlocked(state *ClientState) (bool, time.Duration) {
	state.mu.RLock()
	defer state.mu.RUnlock()

	if state.BlockedUntil == nil {
		return false, 0
	}

	remaining := time.Until(*state.BlockedUntil)
	if remaining <= 0 {
		return false, 0
	}

	return true, remaining
}

// isCircuitOpen checks if the circuit breaker is open
func (arl *AdaptiveRateLimiter) isCircuitOpen() bool {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	if !arl.circuitOpen {
		return false
	}

	if arl.circuitOpenedAt != nil && time.Since(*arl.circuitOpenedAt) > arl.circuitTimeout {
		return false
	}

	return true
}

// OpenCircuit opens the circuit breaker
func (arl *AdaptiveRateLimiter) OpenCircuit(reason string) {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	now := time.Now()
	arl.circuitOpen = true
	arl.circuitOpenedAt = &now
	arl.threatLevel = ThreatCritical

	slog.Error("Circuit breaker opened",
		"reason", reason,
		"timeout", arl.circuitTimeout,
	)
}

// CloseCircuit closes the circuit breaker
func (arl *AdaptiveRateLimiter) CloseCircuit() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.circuitOpen = false
	arl.circuitOpenedAt = nil
	arl.threatLevel = ThreatNone

	slog.Info("Circuit breaker closed")
}

// getOrCreateClient gets or creates a client state
func (arl *AdaptiveRateLimiter) getOrCreateClient(clientID string, tier RateLimitTier) *ClientState {
	arl.mu.RLock()
	state, exists := arl.clients[clientID]
	arl.mu.RUnlock()

	if exists {
		return state
	}

	arl.mu.Lock()
	defer arl.mu.Unlock()

	// Double-check after acquiring write lock
	if state, exists = arl.clients[clientID]; exists {
		return state
	}

	config := arl.tierConfigs[tier]
	now := time.Now()

	state = &ClientState{
		ClientID:       clientID,
		Tier:           tier,
		Tokens:         float64(config.BurstSize),
		LastRefill:     now,
		FirstRequestAt: now,
		UserAgents:     make(map[string]int),
		Endpoints:      make(map[string]int),
		Methods:        make(map[string]int),
	}

	arl.clients[clientID] = state
	return state
}

// recordRequest records request metadata
func (arl *AdaptiveRateLimiter) recordRequest(state *ClientState, req *RateLimitRequest) {
	state.mu.Lock()
	defer state.mu.Unlock()

	atomic.AddInt64(&state.RequestCount, 1)
	atomic.AddInt64(&arl.globalRequests, 1)
	state.LastRequestAt = time.Now()

	if req.UserAgent != "" {
		state.UserAgents[req.UserAgent]++
	}
	if req.Endpoint != "" {
		state.Endpoints[req.Endpoint]++
	}
	if req.Method != "" {
		state.Methods[req.Method]++
	}
}

// calculateRetryAfter calculates the retry-after duration
func (arl *AdaptiveRateLimiter) calculateRetryAfter(state *ClientState, config TierConfig) time.Duration {
	// Base retry is cooldown period
	retry := config.CooldownPeriod

	// Add time based on threat level
	retry += time.Duration(state.ThreatLevel) * 30 * time.Second

	// If blocked, return remaining block time
	if state.BlockedUntil != nil {
		remaining := time.Until(*state.BlockedUntil)
		if remaining > retry {
			return remaining
		}
	}

	return retry
}

// runCleanup periodically cleans up stale entries
func (arl *AdaptiveRateLimiter) runCleanup() {
	ticker := time.NewTicker(arl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		arl.mu.Lock()
		threshold := time.Now().Add(-30 * time.Minute)

		cleaned := 0
		for id, state := range arl.clients {
			state.mu.RLock()
			if state.LastRequestAt.Before(threshold) && state.BlockedUntil == nil {
				delete(arl.clients, id)
				cleaned++
			}
			state.mu.RUnlock()
		}
		arl.mu.Unlock()

		if cleaned > 0 {
			slog.Debug("Rate limiter cleanup", "cleaned_clients", cleaned)
		}
	}
}

// runThreatAssessment runs periodic global threat assessment
func (arl *AdaptiveRateLimiter) runThreatAssessment() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		arl.assessGlobalThreat()
	}
}

// assessGlobalThreat assesses overall system threat level
func (arl *AdaptiveRateLimiter) assessGlobalThreat() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	if arl.globalRequests == 0 {
		return
	}

	// Check error rate
	errorRate := float64(arl.globalErrors) / float64(arl.globalRequests)
	blockRate := float64(arl.globalBlocked) / float64(arl.globalRequests)

	previousLevel := arl.threatLevel

	if blockRate > 0.3 || errorRate > 0.5 {
		arl.threatLevel = ThreatCritical
	} else if blockRate > 0.2 || errorRate > 0.3 {
		arl.threatLevel = ThreatHigh
	} else if blockRate > 0.1 || errorRate > 0.1 {
		arl.threatLevel = ThreatMedium
	} else if blockRate > 0.05 {
		arl.threatLevel = ThreatLow
	} else {
		arl.threatLevel = ThreatNone
	}

	if arl.threatLevel != previousLevel {
		slog.Info("Global threat level changed",
			"previous", previousLevel,
			"current", arl.threatLevel,
			"error_rate", errorRate,
			"block_rate", blockRate,
		)
	}

	// Auto-circuit breaker
	if arl.threatLevel == ThreatCritical && !arl.circuitOpen {
		now := time.Now()
		arl.circuitOpen = true
		arl.circuitOpenedAt = &now
		slog.Warn("Auto-opening circuit breaker due to critical threat level")
	}
}

// SetSystemLoad updates the current system load (0.0 - 1.0)
func (arl *AdaptiveRateLimiter) SetSystemLoad(load float64) {
	arl.mu.Lock()
	arl.systemLoad = load
	arl.mu.Unlock()
}

// GetStats returns current rate limiter statistics
func (arl *AdaptiveRateLimiter) GetStats() *RateLimiterStats {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	return &RateLimiterStats{
		TotalRequests:   atomic.LoadInt64(&arl.globalRequests),
		TotalBlocked:    atomic.LoadInt64(&arl.globalBlocked),
		TotalErrors:     atomic.LoadInt64(&arl.globalErrors),
		ActiveClients:   len(arl.clients),
		ThreatLevel:     arl.threatLevel,
		CircuitOpen:     arl.circuitOpen,
		SystemLoad:      arl.systemLoad,
		Uptime:          time.Since(arl.startTime),
	}
}

// Request/Response types

// RateLimitRequest contains request information for rate limiting
type RateLimitRequest struct {
	ClientID  string
	UserID    *uuid.UUID
	Tier      RateLimitTier
	Endpoint  string
	Method    string
	UserAgent string
	IP        string
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed      bool          `json:"allowed"`
	Reason       string        `json:"reason,omitempty"`
	Remaining    int           `json:"remaining"`
	Limit        int           `json:"limit"`
	RetryAfter   time.Duration `json:"retry_after,omitempty"`
	BlockedUntil *time.Time    `json:"blocked_until,omitempty"`
	ThreatLevel  ThreatLevel   `json:"threat_level"`
	RequestID    string        `json:"request_id,omitempty"`
}

// RateLimiterStats contains rate limiter statistics
type RateLimiterStats struct {
	TotalRequests int64         `json:"total_requests"`
	TotalBlocked  int64         `json:"total_blocked"`
	TotalErrors   int64         `json:"total_errors"`
	ActiveClients int           `json:"active_clients"`
	ThreatLevel   ThreatLevel   `json:"threat_level"`
	CircuitOpen   bool          `json:"circuit_open"`
	SystemLoad    float64       `json:"system_load"`
	Uptime        time.Duration `json:"uptime"`
}

// Middleware creates an HTTP middleware for the adaptive rate limiter
func (arl *AdaptiveRateLimiter) Middleware(getTier func(r *http.Request) RateLimitTier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract client ID
			clientID := getClientIdentifier(r)
			tier := getTier(r)

			req := &RateLimitRequest{
				ClientID:  clientID,
				Tier:      tier,
				Endpoint:  r.URL.Path,
				Method:    r.Method,
				UserAgent: r.UserAgent(),
				IP:        getClientIP(r),
			}

			result := arl.Allow(ctx, req)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-Threat-Level", fmt.Sprintf("%d", result.ThreatLevel))

			if result.RequestID != "" {
				w.Header().Set("X-Request-ID", result.RequestID)
			}

			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       result.Reason,
					"code":        "RATE_LIMIT_EXCEEDED",
					"retry_after": int(result.RetryAfter.Seconds()),
				})
				return
			}

			// Track request completion
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			arl.Complete(clientID, rw.statusCode, time.Since(start))
		})
	}
}

// Helper to get client identifier
func getClientIdentifier(r *http.Request) string {
	// Try to get user ID from context
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok && userID != uuid.Nil {
		return "user:" + userID.String()
	}

	// Fall back to IP + User-Agent hash
	ip := getClientIP(r)
	return "ip:" + ip
}

// Helper to get client IP
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

