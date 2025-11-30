package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Business metrics
	PaymentAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_attempts_total",
			Help: "Total number of payment attempts",
		},
		[]string{"provider", "type"},
	)

	PaymentFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_failures_total",
			Help: "Total number of payment failures",
		},
		[]string{"provider", "type", "reason"},
	)

	PaymentProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_processing_duration_seconds",
			Help:    "Payment processing duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "type"},
	)

	StripeWebhookFailuresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "stripe_webhook_failures_total",
			Help: "Total number of Stripe webhook processing failures",
		},
	)

	TournamentParticipants = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tournament_participants",
			Help: "Current number of tournament participants",
		},
		[]string{"tournament_id", "status"},
	)

	LobbyActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "lobby_active_connections",
			Help: "Number of active lobby WebSocket connections",
		},
	)

	MatchmakingQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "matchmaking_queue_size",
			Help: "Current matchmaking queue size",
		},
		[]string{"game", "mode", "region"},
	)

	WalletBalanceTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wallet_balance_total",
			Help: "Total wallet balance across all users (in cents)",
		},
		[]string{"currency"},
	)

	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "collection"},
	)

	CacheHitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hit_total",
			Help: "Total cache hits",
		},
		[]string{"cache"},
	)

	CacheMissTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_miss_total",
			Help: "Total cache misses",
		},
		[]string{"cache"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		path := normalizePath(r.URL.Path)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func normalizePath(path string) string {
	if len(path) > 50 {
		return path[:50]
	}
	return path
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func RecordDBOperation(operation, collection string, duration time.Duration) {
	DatabaseOperationDuration.WithLabelValues(operation, collection).Observe(duration.Seconds())
}

func RecordCacheHit(cache string) {
	CacheHitTotal.WithLabelValues(cache).Inc()
}

func RecordCacheMiss(cache string) {
	CacheMissTotal.WithLabelValues(cache).Inc()
}

func RecordPaymentAttempt(provider, paymentType string) {
	PaymentAttemptsTotal.WithLabelValues(provider, paymentType).Inc()
}

func RecordPaymentFailure(provider, paymentType, reason string) {
	PaymentFailuresTotal.WithLabelValues(provider, paymentType, reason).Inc()
}

func RecordPaymentDuration(provider, paymentType string, duration time.Duration) {
	PaymentProcessingDuration.WithLabelValues(provider, paymentType).Observe(duration.Seconds())
}
