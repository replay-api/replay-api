package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	"github.com/replay-api/replay-api/pkg/infra/ioc"
	"github.com/replay-api/replay-api/pkg/infra/kafka"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	segkafka "github.com/segmentio/kafka-go"
)

// systemContext creates a context with system-level RLS credentials.
// Workers need TenantID, ClientID, and UserID set for RLS-filtered DB queries.
func systemContext(parent context.Context) context.Context {
	ctx := context.WithValue(parent, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, replay_common.TeamPROAppClientID) // system user
	return ctx
}

// OracleWorkerConfig holds configuration for the oracle worker.
type OracleWorkerConfig struct {
	KafkaBrokers       string
	GroupID            string
	HealthPort         int
	MetricsPort        int
	IngestionInterval  time.Duration
	PublicationInterval time.Duration
	FinalizationDelay  time.Duration
}

// NewOracleWorkerConfigFromEnv reads worker configuration from environment.
func NewOracleWorkerConfigFromEnv() *OracleWorkerConfig {
	cfg := &OracleWorkerConfig{
		KafkaBrokers:       envOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		GroupID:            envOrDefault("KAFKA_CONSUMER_GROUP", "oracle-worker-group"),
		HealthPort:         envIntOrDefault("HEALTH_PORT", 8081),
		MetricsPort:        envIntOrDefault("METRICS_PORT", 9090),
		IngestionInterval:  envDurationOrDefault("ORACLE_INGESTION_INTERVAL", 30*time.Second),
		PublicationInterval: envDurationOrDefault("ORACLE_PUBLICATION_INTERVAL", 60*time.Second),
		FinalizationDelay:  envDurationOrDefault("ORACLE_FINALIZATION_DELAY", 72*time.Hour),
	}

	if cfg.KafkaBrokers == "localhost:9092" {
		if v := os.Getenv("KAFKA_BROKERS"); v != "" {
			cfg.KafkaBrokers = v
		}
	}

	return cfg
}

// --- Prometheus Metrics ---

var (
	oracleScoresIngested = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oracle_worker_scores_ingested_total",
			Help: "Total number of scores ingested from external providers",
		},
		[]string{"source_type", "game_id", "status"},
	)
	oracleConsensusComputed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oracle_worker_consensus_computed_total",
			Help: "Total number of consensus computations",
		},
		[]string{"result", "game_id"},
	)
	oraclePublicationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oracle_worker_publications_total",
			Help: "Total number of chain publications",
		},
		[]string{"chain_id", "status"},
	)
	oracleFinalizationsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "oracle_worker_finalizations_total",
			Help: "Total number of score finalizations",
		},
	)
	oracleProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "oracle_worker_processing_duration_seconds",
			Help:    "Duration of oracle processing operations",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 14),
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(
		oracleScoresIngested,
		oracleConsensusComputed,
		oraclePublicationsTotal,
		oracleFinalizationsTotal,
		oracleProcessingDuration,
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

func (h *healthServer) setReady(ready bool) { h.ready.Store(ready) }
func (h *healthServer) setAlive(alive bool)  { h.alive.Store(alive) }

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

// --- Oracle Event Handlers ---

type oracleEventHandler struct {
	commandHandler oracle_in.OracleCommandHandler
	queryHandler   oracle_in.OracleQueryHandler
	repository     oracle_out.OracleResultRepository
	chainGateway   oracle_out.ChainScoreGateway
}

// handleConsensusReached processes oracle.consensus.reached events
// triggers chain publication for results reaching consensus
func (h *oracleEventHandler) handleConsensusReached(ctx context.Context, msg *segkafka.Message) error {
	start := time.Now()
	defer func() {
		oracleProcessingDuration.WithLabelValues("publish_after_consensus").Observe(time.Since(start).Seconds())
	}()

	var event kafka.OracleEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal consensus reached event", "error", err)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	slog.InfoContext(ctx, "processing consensus reached event",
		slog.String("oracle_result_id", event.OracleResultID.String()),
		slog.String("game_id", event.GameID),
		slog.Int("confidence_level", event.ConfidenceLevel),
	)

	// Auto-publish to chain if confidence is high enough
	if event.ConfidenceLevel >= 2 { // confidence ≥ 0.80
		cmd := oracle_in.PublishToChainCommand{
			OracleResultID: event.OracleResultID,
		}

		if err := h.commandHandler.PublishToChain(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to auto-publish to chain", "error", err)
			oraclePublicationsTotal.WithLabelValues("auto", "failed").Inc()
			return fmt.Errorf("auto-publish failed: %w", err)
		}

		oraclePublicationsTotal.WithLabelValues("auto", "success").Inc()
		slog.InfoContext(ctx, "auto-published to chain",
			slog.String("oracle_result_id", event.OracleResultID.String()),
		)
	}

	return nil
}

// handleExternalIngested processes oracle.external.ingested events
// can trigger re-evaluation or cross-validation workflows
func (h *oracleEventHandler) handleExternalIngested(ctx context.Context, msg *segkafka.Message) error {
	start := time.Now()
	defer func() {
		oracleProcessingDuration.WithLabelValues("process_ingestion").Observe(time.Since(start).Seconds())
	}()

	var event kafka.OracleEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal external ingested event", "error", err)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	oracleScoresIngested.WithLabelValues(event.SourceType, event.GameID, "received").Inc()

	slog.InfoContext(ctx, "external score ingested",
		slog.String("oracle_result_id", event.OracleResultID.String()),
		slog.String("source_type", event.SourceType),
		slog.String("game_id", event.GameID),
	)

	return nil
}

// handleScorePublished processes oracle.published events
// starts the finalization countdown
func (h *oracleEventHandler) handleScorePublished(ctx context.Context, msg *segkafka.Message) error {
	var event kafka.OracleEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	slog.InfoContext(ctx, "score published on-chain",
		slog.String("oracle_result_id", event.OracleResultID.String()),
		slog.Int("chain_publications", len(event.ChainPublications)),
	)

	oraclePublicationsTotal.WithLabelValues("event", "published").Inc()
	return nil
}

// --- Background Jobs ---

// publishPendingResults polls for results that reached consensus but aren't published yet
func publishPendingResults(ctx context.Context, commandHandler oracle_in.OracleCommandHandler, repo oracle_out.OracleResultRepository) {
	results, err := repo.FindPendingPublication(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pending publications", "error", err)
		return
	}

	for _, result := range results {
		cmd := oracle_in.PublishToChainCommand{
			OracleResultID: result.ID,
		}

		if err := commandHandler.PublishToChain(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "failed to publish pending result",
				slog.String("oracle_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		oraclePublicationsTotal.WithLabelValues("scheduled", "success").Inc()
	}
}

// finalizeExpiredResults checks for published results past the dispute window and finalizes them
func finalizeExpiredResults(ctx context.Context, repo oracle_out.OracleResultRepository, finalizationDelay time.Duration) {
	before := time.Now().Add(-finalizationDelay)
	results, err := repo.FindPublishedBefore(ctx, before)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find results for finalization", "error", err)
		return
	}

	for _, result := range results {
		if err := result.Finalize(); err != nil {
			slog.WarnContext(ctx, "cannot finalize result",
				slog.String("oracle_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := repo.Update(ctx, result); err != nil {
			slog.ErrorContext(ctx, "failed to update finalized result",
				slog.String("oracle_result_id", result.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		oracleFinalizationsTotal.Inc()
		slog.InfoContext(ctx, "oracle result finalized",
			slog.String("oracle_result_id", result.ID.String()),
		)
	}
}

// --- Main ---

func main() {
	ctx, cancel := context.WithCancel(systemContext(context.Background()))
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseSlogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(logger)

	cfg := NewOracleWorkerConfigFromEnv()

	// Start health/metrics server
	health := newHealthServer()
	health.setDetail("version", "1.0.0")
	health.setDetail("worker", "oracle-worker")
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

	slog.Info("Starting oracle worker",
		"brokers", cfg.KafkaBrokers,
		"group_id", cfg.GroupID,
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

	// Resolve oracle dependencies
	var commandHandler oracle_in.OracleCommandHandler
	if err := c.Resolve(&commandHandler); err != nil {
		slog.Error("Failed to resolve OracleCommandHandler", "error", err)
		os.Exit(1)
	}

	var queryHandler oracle_in.OracleQueryHandler
	if err := c.Resolve(&queryHandler); err != nil {
		slog.Error("Failed to resolve OracleQueryHandler", "error", err)
		os.Exit(1)
	}

	var repository oracle_out.OracleResultRepository
	if err := c.Resolve(&repository); err != nil {
		slog.Error("Failed to resolve OracleResultRepository", "error", err)
		os.Exit(1)
	}

	var chainGateway oracle_out.ChainScoreGateway
	if err := c.Resolve(&chainGateway); err != nil {
		slog.Error("Failed to resolve ChainScoreGateway", "error", err)
		os.Exit(1)
	}

	handler := &oracleEventHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
		repository:     repository,
		chainGateway:   chainGateway,
	}

	// Create Kafka consumers for oracle topics
	consensusConsumerConfig := kafka.DefaultConsumerConfig(
		cfg.GroupID+"-consensus",
		[]string{kafka.TopicOracleConsensusReached},
	)
	consensusConsumer := kafka.NewConsumer(kafkaClient, consensusConsumerConfig)
	consensusConsumer.RegisterHandler(kafka.TopicOracleConsensusReached, handler.handleConsensusReached)

	ingestedConsumerConfig := kafka.DefaultConsumerConfig(
		cfg.GroupID+"-ingested",
		[]string{kafka.TopicOracleExternalIngested},
	)
	ingestedConsumer := kafka.NewConsumer(kafkaClient, ingestedConsumerConfig)
	ingestedConsumer.RegisterHandler(kafka.TopicOracleExternalIngested, handler.handleExternalIngested)

	publishedConsumerConfig := kafka.DefaultConsumerConfig(
		cfg.GroupID+"-published",
		[]string{kafka.TopicOraclePublished},
	)
	publishedConsumer := kafka.NewConsumer(kafkaClient, publishedConsumerConfig)
	publishedConsumer.RegisterHandler(kafka.TopicOraclePublished, handler.handleScorePublished)

	// Mark as ready
	health.setReady(true)
	health.setDetail("status", "consuming")
	slog.Info("Oracle worker ready, listening for oracle events...")

	// Start Kafka consumers in goroutines
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting consensus consumer...")
		if err := consensusConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			slog.Error("Consensus consumer error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting ingested consumer...")
		if err := ingestedConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			slog.Error("Ingested consumer error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting published consumer...")
		if err := publishedConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			slog.Error("Published consumer error", "error", err)
		}
	}()

	// Start background schedulers
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.PublicationInterval)
		defer ticker.Stop()

		slog.Info("Starting publication scheduler", "interval", cfg.PublicationInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				publishPendingResults(ctx, commandHandler, repository)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		slog.Info("Starting finalization scheduler", "delay", cfg.FinalizationDelay)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				finalizeExpiredResults(ctx, repository, cfg.FinalizationDelay)
			}
		}
	}()

	// Wait for all goroutines
	wg.Wait()

	slog.Info("Oracle worker shut down gracefully")
}

// --- Utility helpers ---

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envIntOrDefault(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func envDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func parseSlogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Suppress unused import warnings
var (
	_ = uuid.Nil
	_ = oracle_vo.OracleStatusPending
)
