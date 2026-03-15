package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/replay-api/replay-api/pkg/infra/ioc"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// systemContext creates a context with system-level RLS credentials.
func systemContext(parent context.Context) context.Context {
	ctx := context.WithValue(parent, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, replay_common.TeamPROAppClientID)
	return ctx
}

// --- Config ---

type DiscoveryWorkerConfig struct {
	HealthPort        int
	MetricsPort       int
	PollingInterval   time.Duration
	LookbackWindow    time.Duration
	MaxMatchesPerPoll int
	SupportedGames    []string
	TriggerOCR        bool
	TriggerAPIIngest  bool
}

func NewDiscoveryWorkerConfigFromEnv() *DiscoveryWorkerConfig {
	games := envListOrDefault("SUPPORTED_GAMES", []string{"cs2"})

	return &DiscoveryWorkerConfig{
		HealthPort:        envIntOrDefault("HEALTH_PORT", 8083),
		MetricsPort:       envIntOrDefault("METRICS_PORT", 9092),
		PollingInterval:   envDurationOrDefault("DISCOVERY_POLLING_INTERVAL", 5*time.Minute),
		LookbackWindow:    envDurationOrDefault("DISCOVERY_LOOKBACK_WINDOW", 24*time.Hour),
		MaxMatchesPerPoll: envIntOrDefault("DISCOVERY_MAX_PER_POLL", 50),
		SupportedGames:    games,
		TriggerOCR:        os.Getenv("DISCOVERY_TRIGGER_OCR") != "false",
		TriggerAPIIngest:  os.Getenv("DISCOVERY_TRIGGER_API_INGEST") != "false",
	}
}

// --- Prometheus Metrics ---

var (
	discoveryMatchesFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discovery_worker_matches_found_total",
			Help: "Total matches discovered from external providers",
		},
		[]string{"provider", "game_id", "is_new"},
	)
	discoveryImportsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discovery_worker_imports_total",
			Help: "Total matches successfully imported",
		},
		[]string{"game_id", "status"},
	)
	discoveryPollDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "discovery_worker_poll_duration_seconds",
			Help:    "Duration of discovery polling operations",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
		},
		[]string{"game_id"},
	)
	discoveryLastPollTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "discovery_worker_last_poll_timestamp",
			Help: "Timestamp of last discovery poll",
		},
	)
)

func init() {
	prometheus.MustRegister(
		discoveryMatchesFound,
		discoveryImportsTotal,
		discoveryPollDuration,
		discoveryLastPollTime,
	)
}

// --- Health Check ---

type healthServer struct {
	ready   atomic.Bool
	alive   atomic.Bool
	mu      sync.RWMutex
	details map[string]string
}

func newHealthServer() *healthServer {
	h := &healthServer{details: make(map[string]string)}
	h.alive.Store(true)
	return h
}

func (h *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/healthz":
		if h.alive.Load() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "alive")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "not alive")
		}
	case "/readyz":
		if h.ready.Load() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ready")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "not ready")
		}
	case "/status":
		h.mu.RLock()
		defer h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.details)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// --- Main ---

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("game discovery worker starting...")

	cfg := NewDiscoveryWorkerConfigFromEnv()

	// IoC container
	builder := ioc.NewContainerBuilder()
	c := builder.WithEnvFile().WithEventPublisher().Build()

	if err := ioc.InjectMongoDB(c); err != nil {
		slog.Error("failed to inject MongoDB services", slog.String("error", err.Error()))
		os.Exit(1)
	}

	builder.WithInboundPorts()
	defer builder.Close(c)

	// Resolve dependencies from IoC
	var oracleResultRepo oracle_out.OracleResultRepository
	if err := c.Resolve(&oracleResultRepo); err != nil {
		slog.Error("failed to resolve OracleResultRepository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var streamConfigRepo oracle_out.OCRStreamConfigRepository
	if err := c.Resolve(&streamConfigRepo); err != nil {
		slog.Error("failed to resolve OCRStreamConfigRepository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var eventPublisher oracle_out.OracleEventPublisher
	if err := c.Resolve(&eventPublisher); err != nil {
		slog.Error("failed to resolve OracleEventPublisher", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var providers []oracle_out.ExternalScorePort
	if err := c.Resolve(&providers); err != nil {
		slog.Warn("no external score providers available", slog.String("error", err.Error()))
		providers = []oracle_out.ExternalScorePort{}
	}

	var importHandler oracle_in.GameImportCommandHandler
	if err := c.Resolve(&importHandler); err != nil {
		slog.Error("failed to resolve GameImportCommandHandler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Build game discovery config
	supportedGames := make([]replay_common.GameIDKey, 0, len(cfg.SupportedGames))
	for _, g := range cfg.SupportedGames {
		supportedGames = append(supportedGames, replay_common.GameIDKey(g))
	}

	discoveryConfig := oracle_services.GameDiscoveryConfig{
		PollingInterval:   cfg.PollingInterval,
		LookbackWindow:    cfg.LookbackWindow,
		MaxMatchesPerPoll: cfg.MaxMatchesPerPoll,
		SupportedGames:    supportedGames,
	}

	discoveryService := oracle_services.NewGameDiscoveryService(
		providers,
		oracleResultRepo,
		streamConfigRepo,
		eventPublisher,
		discoveryConfig,
	)

	// Health server
	health := newHealthServer()
	healthMux := http.NewServeMux()
	healthMux.Handle("/", health)
	healthSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HealthPort),
		Handler: healthMux,
	}

	// Metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: metricsMux,
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(systemContext(context.Background()))
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start health server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("health server starting", slog.Int("port", cfg.HealthPort))
		if err := healthSrv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("health server error", slog.String("error", err.Error()))
		}
	}()

	// Start metrics server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("metrics server starting", slog.Int("port", cfg.MetricsPort))
		if err := metricsSrv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("metrics server error", slog.String("error", err.Error()))
		}
	}()

	health.ready.Store(true)
	health.mu.Lock()
	health.details["status"] = "ready"
	health.details["polling_interval"] = cfg.PollingInterval.String()
	health.details["lookback_window"] = cfg.LookbackWindow.String()
	health.details["supported_games"] = fmt.Sprintf("%v", cfg.SupportedGames)
	health.details["providers"] = fmt.Sprintf("%d", len(providers))
	health.mu.Unlock()

	slog.Info("game discovery worker ready",
		slog.Int("health_port", cfg.HealthPort),
		slog.Int("metrics_port", cfg.MetricsPort),
		slog.Int("providers", len(providers)),
		slog.Duration("polling_interval", cfg.PollingInterval),
	)

	// Start discovery loop
	triggerOCR := cfg.TriggerOCR
	triggerAPIIngest := cfg.TriggerAPIIngest

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := discoveryService.RunDiscoveryLoop(ctx, func(ctx context.Context, match oracle_services.DiscoveredMatch) error {
			start := time.Now()

			// Record metrics
			discoveryMatchesFound.WithLabelValues(
				string(match.ExternalMatch.Provider),
				string(match.ExternalMatch.GameID),
				fmt.Sprintf("%t", match.IsNew),
			).Inc()
			discoveryLastPollTime.Set(float64(time.Now().Unix()))

			if !match.IsNew {
				return nil
			}

			slog.Info("importing discovered match",
				slog.String("external_match_id", match.ExternalMatch.ExternalMatchID),
				slog.String("teams", fmt.Sprintf("%s vs %s", match.ExternalMatch.TeamAName, match.ExternalMatch.TeamBName)),
				slog.String("tournament", match.ExternalMatch.TournamentName),
				slog.String("provider", string(match.ExternalMatch.Provider)),
				slog.Bool("has_vod", match.HasVOD),
				slog.Bool("has_stream", match.HasStream),
			)

			cmd := oracle_in.ImportDiscoveredMatchCommand{
				ExternalMatch:    match.ExternalMatch,
				TriggerOCR:       triggerOCR && (match.HasVOD || match.HasStream),
				TriggerAPIIngest: triggerAPIIngest,
			}

			if err := importHandler.ImportDiscoveredMatch(ctx, cmd); err != nil {
				discoveryImportsTotal.WithLabelValues(string(match.ExternalMatch.GameID), "error").Inc()
				return fmt.Errorf("import failed: %w", err)
			}

			discoveryImportsTotal.WithLabelValues(string(match.ExternalMatch.GameID), "success").Inc()
			discoveryPollDuration.WithLabelValues(string(match.ExternalMatch.GameID)).Observe(time.Since(start).Seconds())

			return nil
		})

		if err != nil && err != context.Canceled {
			slog.Error("discovery loop ended with error", slog.String("error", err.Error()))
		}
	}()

	// Wait for shutdown signal
	sig := <-sigCh
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))

	health.ready.Store(false)
	health.alive.Store(false)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = healthSrv.Shutdown(shutdownCtx)
	_ = metricsSrv.Shutdown(shutdownCtx)

	wg.Wait()
	slog.Info("game discovery worker stopped")
}

// --- Helpers ---

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envIntOrDefault(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
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

func envListOrDefault(key string, defaultVal []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := make([]string, 0)
		for _, p := range splitAndTrim(v, ",") {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultVal
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, p := range splitString(s, sep) {
		trimmed := trimSpace(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
