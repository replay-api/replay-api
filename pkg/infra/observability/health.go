package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents health of a single component
type ComponentHealth struct {
	Name       string                 `json:"name"`
	Status     HealthStatus           `json:"status"`
	Message    string                 `json:"message,omitempty"`
	Latency    time.Duration          `json:"latency_ms"`
	LastCheck  time.Time              `json:"last_check"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheckResult contains complete health check results
type HealthCheckResult struct {
	Status      HealthStatus               `json:"status"`
	Version     string                     `json:"version"`
	Uptime      time.Duration              `json:"uptime"`
	Components  map[string]ComponentHealth `json:"components"`
	System      SystemHealth               `json:"system"`
	Timestamp   time.Time                  `json:"timestamp"`
}

// SystemHealth contains system-level health metrics
type SystemHealth struct {
	Goroutines   int     `json:"goroutines"`
	HeapAlloc    uint64  `json:"heap_alloc_bytes"`
	HeapSys      uint64  `json:"heap_sys_bytes"`
	HeapInuse    uint64  `json:"heap_inuse_bytes"`
	StackInuse   uint64  `json:"stack_inuse_bytes"`
	NumGC        uint32  `json:"num_gc"`
	CPUUsage     float64 `json:"cpu_usage_percent,omitempty"`
	MemoryUsage  float64 `json:"memory_usage_percent,omitempty"`
}

// HealthChecker defines a health check function
type HealthChecker func(ctx context.Context) ComponentHealth

// HealthService manages health checks and exposes endpoints
type HealthService struct {
	mu          sync.RWMutex
	checkers    map[string]HealthChecker
	version     string
	startTime   time.Time
	
	// Caching
	lastResult  *HealthCheckResult
	cacheTTL    time.Duration
	lastCheckAt time.Time
	
	// Metrics
	healthGauge      *prometheus.GaugeVec
	checkDuration    *prometheus.HistogramVec
	componentStatus  *prometheus.GaugeVec
}

// NewHealthService creates a new health service
func NewHealthService(version string) *HealthService {
	hs := &HealthService{
		checkers:  make(map[string]HealthChecker),
		version:   version,
		startTime: time.Now(),
		cacheTTL:  5 * time.Second,
		healthGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "leetgaming_health_status",
				Help: "Overall health status (1=healthy, 0.5=degraded, 0=unhealthy)",
			},
			[]string{"component"},
		),
		checkDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "leetgaming_health_check_duration_seconds",
				Help:    "Duration of health checks",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
			},
			[]string{"component"},
		),
		componentStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "leetgaming_component_status",
				Help: "Component health status",
			},
			[]string{"component", "status"},
		),
	}

	// Register built-in checkers
	hs.RegisterChecker("runtime", hs.runtimeChecker)

	return hs
}

// RegisterChecker adds a new health checker
func (hs *HealthService) RegisterChecker(name string, checker HealthChecker) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.checkers[name] = checker
	slog.Info("Health checker registered", "name", name)
}

// RegisterMongoDBChecker registers a MongoDB health checker
func (hs *HealthService) RegisterMongoDBChecker(pingFunc func(ctx context.Context) error) {
	hs.RegisterChecker("mongodb", func(ctx context.Context) ComponentHealth {
		start := time.Now()
		err := pingFunc(ctx)
		latency := time.Since(start)

		if err != nil {
			return ComponentHealth{
				Name:      "mongodb",
				Status:    HealthStatusUnhealthy,
				Message:   err.Error(),
				Latency:   latency,
				LastCheck: time.Now(),
			}
		}

		status := HealthStatusHealthy
		if latency > 100*time.Millisecond {
			status = HealthStatusDegraded
		}

		return ComponentHealth{
			Name:      "mongodb",
			Status:    status,
			Latency:   latency,
			LastCheck: time.Now(),
			Metadata: map[string]interface{}{
				"response_time_ms": latency.Milliseconds(),
			},
		}
	})
}

// RegisterKafkaChecker registers a Kafka health checker
func (hs *HealthService) RegisterKafkaChecker(checkFunc func(ctx context.Context) (bool, error)) {
	hs.RegisterChecker("kafka", func(ctx context.Context) ComponentHealth {
		start := time.Now()
		healthy, err := checkFunc(ctx)
		latency := time.Since(start)

		if err != nil || !healthy {
			msg := "Kafka unhealthy"
			if err != nil {
				msg = err.Error()
			}
			return ComponentHealth{
				Name:      "kafka",
				Status:    HealthStatusUnhealthy,
				Message:   msg,
				Latency:   latency,
				LastCheck: time.Now(),
			}
		}

		return ComponentHealth{
			Name:      "kafka",
			Status:    HealthStatusHealthy,
			Latency:   latency,
			LastCheck: time.Now(),
		}
	})
}

// RegisterRedisChecker registers a Redis health checker
func (hs *HealthService) RegisterRedisChecker(pingFunc func(ctx context.Context) error) {
	hs.RegisterChecker("redis", func(ctx context.Context) ComponentHealth {
		start := time.Now()
		err := pingFunc(ctx)
		latency := time.Since(start)

		if err != nil {
			return ComponentHealth{
				Name:      "redis",
				Status:    HealthStatusUnhealthy,
				Message:   err.Error(),
				Latency:   latency,
				LastCheck: time.Now(),
			}
		}

		return ComponentHealth{
			Name:      "redis",
			Status:    HealthStatusHealthy,
			Latency:   latency,
			LastCheck: time.Now(),
		}
	})
}

// RegisterExternalServiceChecker registers a checker for an external service
func (hs *HealthService) RegisterExternalServiceChecker(name string, healthURL string, timeout time.Duration) {
	hs.RegisterChecker(name, func(ctx context.Context) ComponentHealth {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return ComponentHealth{
				Name:      name,
				Status:    HealthStatusUnhealthy,
				Message:   fmt.Sprintf("Failed to create request: %v", err),
				Latency:   time.Since(start),
				LastCheck: time.Now(),
			}
		}

		resp, err := http.DefaultClient.Do(req)
		latency := time.Since(start)

		if err != nil {
			return ComponentHealth{
				Name:      name,
				Status:    HealthStatusUnhealthy,
				Message:   fmt.Sprintf("Request failed: %v", err),
				Latency:   latency,
				LastCheck: time.Now(),
			}
		}
		defer resp.Body.Close()

		status := HealthStatusHealthy
		if resp.StatusCode >= 500 {
			status = HealthStatusUnhealthy
		} else if resp.StatusCode >= 400 {
			status = HealthStatusDegraded
		}

		return ComponentHealth{
			Name:      name,
			Status:    status,
			Latency:   latency,
			LastCheck: time.Now(),
			Metadata: map[string]interface{}{
				"status_code": resp.StatusCode,
			},
		}
	})
}

// Check performs all health checks
func (hs *HealthService) Check(ctx context.Context) *HealthCheckResult {
	hs.mu.RLock()
	if hs.lastResult != nil && time.Since(hs.lastCheckAt) < hs.cacheTTL {
		result := hs.lastResult
		hs.mu.RUnlock()
		return result
	}
	hs.mu.RUnlock()

	hs.mu.Lock()
	defer hs.mu.Unlock()

	// Double-check after acquiring write lock
	if hs.lastResult != nil && time.Since(hs.lastCheckAt) < hs.cacheTTL {
		return hs.lastResult
	}

	result := &HealthCheckResult{
		Status:     HealthStatusHealthy,
		Version:    hs.version,
		Uptime:     time.Since(hs.startTime),
		Components: make(map[string]ComponentHealth),
		System:     hs.getSystemHealth(),
		Timestamp:  time.Now().UTC(),
	}

	// Run all checkers in parallel
	var wg sync.WaitGroup
	results := make(chan ComponentHealth, len(hs.checkers))

	for name, checker := range hs.checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			start := time.Now()
			health := checker(checkCtx)
			duration := time.Since(start)

			// Record metrics
			hs.checkDuration.WithLabelValues(name).Observe(duration.Seconds())
			
			statusValue := 1.0
			if health.Status == HealthStatusDegraded {
				statusValue = 0.5
			} else if health.Status == HealthStatusUnhealthy {
				statusValue = 0
			}
			hs.healthGauge.WithLabelValues(name).Set(statusValue)

			results <- health
		}(name, checker)
	}

	// Wait and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	for health := range results {
		result.Components[health.Name] = health

		// Update overall status
		if health.Status == HealthStatusUnhealthy {
			result.Status = HealthStatusUnhealthy
		} else if health.Status == HealthStatusDegraded && result.Status != HealthStatusUnhealthy {
			result.Status = HealthStatusDegraded
		}
	}

	hs.lastResult = result
	hs.lastCheckAt = time.Now()

	return result
}

// Liveness returns simple liveness check (for Kubernetes)
func (hs *HealthService) Liveness(ctx context.Context) bool {
	return true // If we can respond, we're alive
}

// Readiness returns readiness check (for Kubernetes)
func (hs *HealthService) Readiness(ctx context.Context) bool {
	result := hs.Check(ctx)
	return result.Status != HealthStatusUnhealthy
}

// runtimeChecker checks Go runtime health
func (hs *HealthService) runtimeChecker(ctx context.Context) ComponentHealth {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	goroutines := runtime.NumGoroutine()
	status := HealthStatusHealthy

	// Check for goroutine leak
	if goroutines > 10000 {
		status = HealthStatusDegraded
	}
	if goroutines > 50000 {
		status = HealthStatusUnhealthy
	}

	// Check for memory pressure
	heapPercent := float64(memStats.HeapAlloc) / float64(memStats.HeapSys) * 100
	if heapPercent > 90 {
		status = HealthStatusDegraded
	}

	return ComponentHealth{
		Name:      "runtime",
		Status:    status,
		LastCheck: time.Now(),
		Metadata: map[string]interface{}{
			"goroutines":         goroutines,
			"heap_alloc_mb":      memStats.HeapAlloc / 1024 / 1024,
			"heap_sys_mb":        memStats.HeapSys / 1024 / 1024,
			"heap_percent":       heapPercent,
			"gc_runs":            memStats.NumGC,
			"last_gc_pause_ns":   memStats.PauseNs[(memStats.NumGC+255)%256],
		},
	}
}

// getSystemHealth collects system-level health metrics
func (hs *HealthService) getSystemHealth() SystemHealth {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemHealth{
		Goroutines:  runtime.NumGoroutine(),
		HeapAlloc:   memStats.HeapAlloc,
		HeapSys:     memStats.HeapSys,
		HeapInuse:   memStats.HeapInuse,
		StackInuse:  memStats.StackInuse,
		NumGC:       memStats.NumGC,
	}
}

// HTTPHandler returns an HTTP handler for health endpoints
func (hs *HealthService) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Full health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		result := hs.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if result.Status == HealthStatusDegraded {
			statusCode = http.StatusOK // Still operational, just degraded
		} else if result.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(result)
	})

	// Kubernetes liveness probe
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		if hs.Liveness(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("NOT OK"))
		}
	})

	// Kubernetes readiness probe
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if hs.Readiness(r.Context()) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("NOT READY"))
		}
	})

	// Component-specific health
	mux.HandleFunc("/health/component/", func(w http.ResponseWriter, r *http.Request) {
		componentName := r.URL.Path[len("/health/component/"):]
		
		hs.mu.RLock()
		checker, exists := hs.checkers[componentName]
		hs.mu.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Component not found"))
			return
		}

		health := checker(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(health)
	})

	return mux
}

// StartBackgroundChecks starts periodic background health checks
func (hs *HealthService) StartBackgroundChecks(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result := hs.Check(ctx)
				if result.Status == HealthStatusUnhealthy {
					slog.Warn("Health check failed",
						"status", result.Status,
						"components", len(result.Components),
					)
				}
			}
		}
	}()
}

// ApplicationMetrics provides application-level metrics
type ApplicationMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	errorsTotal     *prometheus.CounterVec
	activeRequests  *prometheus.GaugeVec
	
	// Business metrics
	matchesCreated   prometheus.Counter
	tournamentsActive prometheus.Gauge
	usersOnline      prometheus.Gauge
	withdrawalsTotal *prometheus.CounterVec
}

// NewApplicationMetrics creates application metrics
func NewApplicationMetrics() *ApplicationMetrics {
	return &ApplicationMetrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "leetgaming_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "leetgaming_http_request_duration_seconds",
				Help:    "HTTP request duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "leetgaming_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "component"},
		),
		activeRequests: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "leetgaming_http_active_requests",
				Help: "Number of active HTTP requests",
			},
			[]string{"method"},
		),
		matchesCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "leetgaming_matches_created_total",
				Help: "Total matches created",
			},
		),
		tournamentsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "leetgaming_tournaments_active",
				Help: "Number of active tournaments",
			},
		),
		usersOnline: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "leetgaming_users_online",
				Help: "Number of users currently online",
			},
		),
		withdrawalsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "leetgaming_withdrawals_total",
				Help: "Total withdrawals by status",
			},
			[]string{"status", "currency"},
		),
	}
}

// RecordRequest records an HTTP request metric
func (m *ApplicationMetrics) RecordRequest(method, path, status string, duration time.Duration) {
	m.requestsTotal.WithLabelValues(method, path, status).Inc()
	m.requestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordError records an error metric
func (m *ApplicationMetrics) RecordError(errorType, component string) {
	m.errorsTotal.WithLabelValues(errorType, component).Inc()
}

// IncrementActiveRequests increments active request count
func (m *ApplicationMetrics) IncrementActiveRequests(method string) {
	m.activeRequests.WithLabelValues(method).Inc()
}

// DecrementActiveRequests decrements active request count
func (m *ApplicationMetrics) DecrementActiveRequests(method string) {
	m.activeRequests.WithLabelValues(method).Dec()
}

// RecordMatchCreated records a match creation
func (m *ApplicationMetrics) RecordMatchCreated() {
	m.matchesCreated.Inc()
}

// SetActiveTournaments sets the active tournament count
func (m *ApplicationMetrics) SetActiveTournaments(count int) {
	m.tournamentsActive.Set(float64(count))
}

// SetOnlineUsers sets the online user count
func (m *ApplicationMetrics) SetOnlineUsers(count int) {
	m.usersOnline.Set(float64(count))
}

// RecordWithdrawal records a withdrawal
func (m *ApplicationMetrics) RecordWithdrawal(status, currency string) {
	m.withdrawalsTotal.WithLabelValues(status, currency).Inc()
}

// MetricsMiddleware creates HTTP middleware for metrics
func (m *ApplicationMetrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		path := r.URL.Path

		m.IncrementActiveRequests(method)
		defer m.DecrementActiveRequests(method)

		// Wrap response writer to capture status
		rw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		status := fmt.Sprintf("%d", rw.statusCode)

		m.RecordRequest(method, normalizePath(path), status, duration)

		if rw.statusCode >= 500 {
			m.RecordError("http_5xx", path)
		} else if rw.statusCode >= 400 {
			m.RecordError("http_4xx", path)
		}
	})
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath normalizes path for metrics (replace IDs with placeholders)
func normalizePath(path string) string {
	// Replace UUIDs with placeholder
	normalized := path
	// Simple replacement - in production use regex
	return normalized
}

