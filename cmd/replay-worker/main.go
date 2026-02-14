package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	shared "github.com/resource-ownership/go-common/pkg/common"

	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	"github.com/replay-api/replay-api/pkg/infra/ioc"
	"github.com/replay-api/replay-api/pkg/infra/kafka"
	segkafka "github.com/segmentio/kafka-go"
)

// ReplayWorkerConfig holds configuration for the replay worker.
type ReplayWorkerConfig struct {
	KafkaBrokers   string
	GroupID        string
	ProcessorCount int
	RetryAttempts  int
	RetryBackoff   time.Duration
	HealthPort     int
	MetricsPort    int
}

// NewReplayWorkerConfigFromEnv reads worker configuration from environment variables with sensible defaults.
func NewReplayWorkerConfigFromEnv() *ReplayWorkerConfig {
	cfg := &ReplayWorkerConfig{
		KafkaBrokers:   envOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		GroupID:        envOrDefault("KAFKA_CONSUMER_GROUP", "replay-worker-group"),
		ProcessorCount: envIntOrDefault("WORKER_PROCESSOR_COUNT", 4),
		RetryAttempts:  envIntOrDefault("WORKER_RETRY_ATTEMPTS", 3),
		RetryBackoff:   envDurationOrDefault("WORKER_RETRY_BACKOFF", 5*time.Second),
		HealthPort:     envIntOrDefault("HEALTH_PORT", 8081),
		MetricsPort:    envIntOrDefault("METRICS_PORT", 9090),
	}

	// Fallback for KAFKA_BOOTSTRAP_SERVERS
	if cfg.KafkaBrokers == "localhost:9092" {
		if v := os.Getenv("KAFKA_BROKERS"); v != "" {
			cfg.KafkaBrokers = v
		}
	}

	return cfg
}

// --- Prometheus Metrics ---

var (
	replaysProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "replay_worker_replays_processed_total",
			Help: "Total number of replays processed",
		},
		[]string{"status", "game_id"},
	)
	replayProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "replay_worker_processing_duration_seconds",
			Help:    "Duration of replay processing in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 12), // 0.1s to ~204s
		},
		[]string{"game_id"},
	)
	replaysInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "replay_worker_replays_in_flight",
			Help: "Number of replays currently being processed",
		},
	)
	kafkaPublishErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "replay_worker_kafka_publish_errors_total",
			Help: "Total number of Kafka publish errors",
		},
		[]string{"event_type"},
	)
	retryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "replay_worker_retry_attempts_total",
			Help: "Total number of retry attempts",
		},
		[]string{"result"},
	)
)

func init() {
	prometheus.MustRegister(
		replaysProcessedTotal,
		replayProcessingDuration,
		replaysInFlight,
		kafkaPublishErrors,
		retryAttempts,
	)
}

// --- Health Check Server ---

type healthServer struct {
	ready   atomic.Bool
	alive   atomic.Bool
	mu      sync.RWMutex
	details map[string]string
}

func newHealthServer() *healthServer {
	hs := &healthServer{
		details: make(map[string]string),
	}
	hs.alive.Store(true)
	return hs
}

func (h *healthServer) setReady(ready bool)       { h.ready.Store(ready) }
func (h *healthServer) setAlive(alive bool)        { h.alive.Store(alive) }

func (h *healthServer) setDetail(key, value string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.details[key] = value
}

func (h *healthServer) serve(port int) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if h.alive.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"alive"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"shutting_down"}`))
		}
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if h.ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ready"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready"}`))
		}
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("Health/metrics server started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Health server error", "error", err)
		}
	}()

	return srv
}

// --- Main ---

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseSlogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(logger)

	cfg := NewReplayWorkerConfigFromEnv()

	// Start health/metrics server
	health := newHealthServer()
	health.setDetail("version", "1.0.0")
	healthSrv := health.serve(cfg.HealthPort)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = healthSrv.Shutdown(shutdownCtx)
	}()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal, draining...", "signal", sig)
		health.setAlive(false)
		health.setReady(false)
		// Give K8s time to stop sending traffic / drain in-flight
		time.Sleep(5 * time.Second)
		cancel()
	}()

	// Initialize IoC container
	builder := ioc.NewContainerBuilder()
	c := builder.WithEnvFile().WithEventPublisher().Build()

	if err := ioc.InjectMongoDB(c); err != nil {
		slog.Error("Failed to inject MongoDB services", "error", err)
		os.Exit(1)
	}

	builder.WithInboundPorts()
	defer builder.Close(c)

	slog.Info("Starting replay worker",
		"brokers", cfg.KafkaBrokers,
		"topic", kafka.TopicReplaysUploaded,
		"group_id", cfg.GroupID,
		"processor_count", cfg.ProcessorCount,
		"retry_attempts", cfg.RetryAttempts,
		"retry_backoff", cfg.RetryBackoff.String(),
	)

	// Create Kafka client
	kafkaConfig := kafka.NewConfigFromEnv()
	kafkaConfig.BootstrapServers = cfg.KafkaBrokers

	kafkaClient, err := kafka.NewClient(kafkaConfig)
	if err != nil {
		slog.Error("Failed to create Kafka client", "error", err)
		os.Exit(1)
	}
	defer kafkaClient.Close()

	eventPublisher := kafka.NewEventPublisher(kafkaClient)

	// Resolve process command
	var processCommand replay_in.ProcessReplayFileCommand
	if err := c.Resolve(&processCommand); err != nil {
		slog.Error("Failed to resolve ProcessReplayFileCommand", "error", err)
		os.Exit(1)
	}

	// Create consumer
	consumerConfig := kafka.DefaultConsumerConfig(cfg.GroupID, []string{kafka.TopicReplaysUploaded})
	consumer := kafka.NewConsumer(kafkaClient, consumerConfig)

	// Build handler with retry logic
	handler := &replayHandler{
		processCommand: processCommand,
		eventPublisher: eventPublisher,
		retryAttempts:  cfg.RetryAttempts,
		retryBackoff:   cfg.RetryBackoff,
	}

	consumer.RegisterHandler(kafka.TopicReplaysUploaded, handler.handle)

	// Mark as ready
	health.setReady(true)
	health.setDetail("status", "consuming")
	slog.Info("Replay worker ready, listening for uploaded replays...")

	// Start consuming (blocks until ctx cancelled)
	if err := consumer.Start(ctx); err != nil && ctx.Err() == nil {
		slog.Error("Consumer error", "error", err)
		os.Exit(1)
	}

	slog.Info("Replay worker shutdown complete")
}

// --- Handler ---

type replayHandler struct {
	processCommand replay_in.ProcessReplayFileCommand
	eventPublisher *kafka.EventPublisher
	retryAttempts  int
	retryBackoff   time.Duration
}

func (h *replayHandler) handle(ctx context.Context, msg *segkafka.Message) error {
	var event kafka.ReplayUploadedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("Failed to unmarshal replay uploaded event",
			"error", err,
			"offset", msg.Offset,
			"partition", msg.Partition,
		)
		// Permanent failure — don't retry bad payload
		replaysProcessedTotal.WithLabelValues("unmarshal_error", "unknown").Inc()
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	slog.Info("Processing replay",
		"replay_file_id", event.ReplayFileID,
		"game_id", event.GameID,
		"file_size", event.FileSize,
	)

	replaysInFlight.Inc()
	defer replaysInFlight.Dec()

	startTime := time.Now()
	gameID := event.GameID

	// Set up context with resource owner
	processCtx := context.WithValue(ctx, shared.TenantIDKey, event.TenantID)
	processCtx = context.WithValue(processCtx, shared.UserIDKey, event.UserID)
	processCtx = context.WithValue(processCtx, shared.AuthenticatedKey, true)

	// Publish processing started event
	if err := h.eventPublisher.PublishReplayProcessing(ctx, &kafka.ReplayProcessingEvent{
		ReplayFileID: event.ReplayFileID,
		EventType:    kafka.EventTypeReplayProcessing,
		Progress:     0,
		Stage:        "starting",
	}); err != nil {
		slog.Warn("Failed to publish processing event", "error", err, "replay_file_id", event.ReplayFileID)
		kafkaPublishErrors.WithLabelValues("processing").Inc()
	}

	// Process with retry
	var lastErr error
	for attempt := 0; attempt <= h.retryAttempts; attempt++ {
		if attempt > 0 {
			backoff := h.retryBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
			slog.Info("Retrying replay processing",
				"replay_file_id", event.ReplayFileID,
				"attempt", attempt,
				"backoff", backoff.String(),
			)
			retryAttempts.WithLabelValues("retry").Inc()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := h.processCommand.Exec(processCtx, event.ReplayFileID)
		if err == nil {
			// Success
			processingDuration := time.Since(startTime)
			replayProcessingDuration.WithLabelValues(gameID).Observe(processingDuration.Seconds())
			replaysProcessedTotal.WithLabelValues("success", gameID).Inc()

			playerCount := 0
			if result.Scoreboard.TeamScoreboards != nil {
				for _, ts := range result.Scoreboard.TeamScoreboards {
					playerCount += len(ts.Players)
				}
			}

			slog.Info("Replay processed successfully",
				"replay_file_id", event.ReplayFileID,
				"match_id", result.ID,
				"duration_ms", processingDuration.Milliseconds(),
				"event_count", result.EventCount,
				"player_count", playerCount,
				"match_duration_s", result.Duration,
			)

			// Publish completed event
			if err := h.eventPublisher.PublishReplayCompleted(ctx, &kafka.ReplayCompletedEvent{
				ReplayFileID:  event.ReplayFileID,
				MatchID:       result.ID,
				GameID:        event.GameID,
				EventCount:    result.EventCount,
				PlayerCount:   playerCount,
				Duration:      processingDuration.Milliseconds(),
				MatchDuration: int64(result.Duration),
			}); err != nil {
				slog.Error("Failed to publish completed event", "error", err, "replay_file_id", event.ReplayFileID)
				kafkaPublishErrors.WithLabelValues("completed").Inc()
			}

			return nil
		}

		lastErr = err
		slog.Warn("Replay processing attempt failed",
			"replay_file_id", event.ReplayFileID,
			"attempt", attempt+1,
			"max_attempts", h.retryAttempts+1,
			"error", err,
		)
	}

	// All retries exhausted
	processingDuration := time.Since(startTime)
	replayProcessingDuration.WithLabelValues(gameID).Observe(processingDuration.Seconds())
	replaysProcessedTotal.WithLabelValues("failed", gameID).Inc()
	retryAttempts.WithLabelValues("exhausted").Inc()

	slog.Error("Failed to process replay after all retries",
		"replay_file_id", event.ReplayFileID,
		"error", lastErr,
		"attempts", h.retryAttempts+1,
		"duration_ms", processingDuration.Milliseconds(),
	)

	// Publish failed event
	if err := h.eventPublisher.PublishReplayFailed(ctx, &kafka.ReplayFailedEvent{
		ReplayFileID: event.ReplayFileID,
		GameID:       event.GameID,
		Stage:        "processing",
		ErrorType:    "processing_error",
		ErrorMessage: lastErr.Error(),
		Retryable:    false,
		RetryCount:   h.retryAttempts,
	}); err != nil {
		slog.Error("Failed to publish failed event", "error", err, "replay_file_id", event.ReplayFileID)
		kafkaPublishErrors.WithLabelValues("failed").Inc()
	}

	return fmt.Errorf("failed to process replay after %d attempts: %w", h.retryAttempts+1, lastErr)
}

// --- Helpers ---

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func parseSlogLevel(level string) slog.Level {
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
