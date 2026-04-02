package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	oracle_ocr "github.com/replay-api/replay-api/pkg/infra/adapters/oracle/ocr"
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

type OCRWorkerConfig struct {
	HealthPort             int
	MetricsPort            int
	StreamQuality          string
	PaddleOCRScriptPath    string
	PaddleOCRPythonPath    string
	PaddleOCRUseGPU        bool
	YouTubeAPIKey          string
	DefaultCaptureInterval int
	MongoURI               string
	MongoDatabase          string
}

func NewOCRWorkerConfigFromEnv() *OCRWorkerConfig {
	return &OCRWorkerConfig{
		HealthPort:             envIntOrDefault("HEALTH_PORT", 8082),
		MetricsPort:            envIntOrDefault("METRICS_PORT", 9091),
		StreamQuality:          envOrDefault("STREAM_QUALITY", "720p,720p60,best"),
		PaddleOCRScriptPath:    envOrDefault("PADDLEOCR_SCRIPT_PATH", "/app/scripts/paddleocr_wrapper.py"),
		PaddleOCRPythonPath:    envOrDefault("PADDLEOCR_PYTHON_PATH", "python3"),
		PaddleOCRUseGPU:        os.Getenv("PADDLEOCR_USE_GPU") == "true",
		YouTubeAPIKey:          os.Getenv("YOUTUBE_API_KEY"),
		DefaultCaptureInterval: envIntOrDefault("DEFAULT_CAPTURE_INTERVAL", 10),
		MongoURI:               envOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:          envOrDefault("MONGODB_DATABASE", "replay_api"),
	}
}

// --- Prometheus Metrics ---

var (
	ocrFramesCaptured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ocr_worker_frames_captured_total",
			Help: "Total number of frames captured from streams",
		},
		[]string{"stream_url", "status"},
	)
	ocrScoresDetected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ocr_worker_scores_detected_total",
			Help: "Total number of scores detected via OCR",
		},
		[]string{"game_id"},
	)
	ocrActiveStreams = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ocr_worker_active_streams",
			Help: "Number of streams currently being monitored",
		},
	)
	ocrProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ocr_worker_processing_duration_seconds",
			Help:    "Duration of OCR processing operations",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(
		ocrFramesCaptured,
		ocrScoresDetected,
		ocrActiveStreams,
		ocrProcessingDuration,
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

// --- Stream Management API ---

type streamAPI struct {
	mu          sync.Mutex
	monitors    map[string]context.CancelFunc // streamURL -> cancel
	monitorsMu  sync.RWMutex
	newMonitorFn func(config oracle_entities.OCRStreamConfig) (*oracle_services.StreamMonitor, oracle_services.StreamMonitorConfig)
}

type startStreamRequest struct {
	StreamURL       string `json:"stream_url"`
	GameID          string `json:"game_id"`
	ExternalMatchID string `json:"external_match_id"`
	TeamAHint       string `json:"team_a_hint,omitempty"`
	TeamBHint       string `json:"team_b_hint,omitempty"`
	IntervalSeconds int    `json:"interval_seconds,omitempty"`
	RegionX         int    `json:"region_x,omitempty"`
	RegionY         int    `json:"region_y,omitempty"`
	RegionW         int    `json:"region_w,omitempty"`
	RegionH         int    `json:"region_h,omitempty"`
}

func (api *streamAPI) handleStart(w http.ResponseWriter, r *http.Request) {
	var req startStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.StreamURL == "" || req.GameID == "" || req.ExternalMatchID == "" {
		http.Error(w, "stream_url, game_id, and external_match_id are required", http.StatusBadRequest)
		return
	}

	api.monitorsMu.RLock()
	if _, exists := api.monitors[req.StreamURL]; exists {
		api.monitorsMu.RUnlock()
		http.Error(w, "stream already being monitored", http.StatusConflict)
		return
	}
	api.monitorsMu.RUnlock()

	interval := req.IntervalSeconds
	if interval <= 0 {
		interval = 10
	}

	config := oracle_entities.OCRStreamConfig{
		StreamURL:              req.StreamURL,
		GameID:                 replay_common.GameIDKey(req.GameID),
		ExternalMatchID:        req.ExternalMatchID,
		CaptureIntervalSeconds: interval,
		TeamAHint:              req.TeamAHint,
		TeamBHint:              req.TeamBHint,
	}

	if req.RegionW > 0 && req.RegionH > 0 {
		config.ScoreboardRegion = &oracle_entities.ScoreboardRegion{
			X: req.RegionX, Y: req.RegionY,
			Width: req.RegionW, Height: req.RegionH,
		}
	} else if req.GameID == "cs2" || req.GameID == "csgo" {
		config.ScoreboardRegion = &oracle_entities.ScoreboardRegion{
			X: 350, Y: 0, Width: 530, Height: 80,
		}
	}

	if api.newMonitorFn == nil {
		http.Error(w, "monitor factory not initialized", http.StatusServiceUnavailable)
		return
	}

	monitor, monitorCfg := api.newMonitorFn(config)
	monitorCtx, monitorCancel := context.WithCancel(r.Context())

	api.monitorsMu.Lock()
	api.monitors[req.StreamURL] = monitorCancel
	api.monitorsMu.Unlock()
	ocrActiveStreams.Inc()

	go func() {
		defer func() {
			api.monitorsMu.Lock()
			delete(api.monitors, req.StreamURL)
			api.monitorsMu.Unlock()
			ocrActiveStreams.Dec()
		}()

		if err := monitor.MonitorStream(monitorCtx, monitorCfg); err != nil && err != context.Canceled {
			slog.Warn("stream monitor ended with error",
				slog.String("stream_url", req.StreamURL),
				slog.String("error", err.Error()),
			)
		}
	}()

	slog.Info("started stream monitor via API",
		slog.String("stream_url", req.StreamURL),
		slog.String("game_id", req.GameID),
	)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "started",
		"stream_url": req.StreamURL,
	})
}

func (api *streamAPI) handleStop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StreamURL string `json:"stream_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	api.monitorsMu.Lock()
	cancel, exists := api.monitors[req.StreamURL]
	if exists {
		cancel()
		delete(api.monitors, req.StreamURL)
		ocrActiveStreams.Dec()
	}
	api.monitorsMu.Unlock()

	if !exists {
		http.Error(w, "stream not found", http.StatusNotFound)
		return
	}

	slog.Info("stopped stream monitor", slog.String("stream_url", req.StreamURL))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (api *streamAPI) handleList(w http.ResponseWriter, r *http.Request) {
	api.monitorsMu.RLock()
	defer api.monitorsMu.RUnlock()

	streams := make([]string, 0, len(api.monitors))
	for url := range api.monitors {
		streams = append(streams, url)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_streams": streams,
		"count":          len(streams),
	})
}

// --- Main ---

func main() {
	// Structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("OCR worker starting...")

	cfg := NewOCRWorkerConfigFromEnv()

	// IoC container for oracle dependencies
	builder := ioc.NewContainerBuilder()
	c := builder.WithEnvFile().WithEventPublisher().Build()

	if err := ioc.InjectMongoDB(c); err != nil {
		slog.Error("failed to inject MongoDB services", slog.String("error", err.Error()))
		os.Exit(1)
	}

	builder.WithInboundPorts()
	defer builder.Close(c)

	// Resolve oracle command handler from IoC
	var commandHandler oracle_in.OracleCommandHandler
	if err := c.Resolve(&commandHandler); err != nil {
		slog.Error("failed to resolve oracle command handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create OCR infrastructure
	streamCapture := oracle_ocr.NewStreamlinkCapture(cfg.StreamQuality)
	vodCapture := oracle_ocr.NewVodCapture()
	ocrEngine := oracle_ocr.NewPaddleOCRAdapter(cfg.PaddleOCRPythonPath, cfg.PaddleOCRScriptPath, cfg.PaddleOCRUseGPU)
	scoreParser := oracle_services.NewOCRScoreParser()

	// Team resolver (optional — needs MongoDB)
	var teamResolver oracle_out.TeamResolverPort
	if err := c.Resolve(&teamResolver); err != nil || teamResolver == nil {
		slog.Info("team resolver: will use deterministic IDs (MongoDB resolver not wired)")
		teamResolver = nil
	} else {
		slog.Info("team resolver: MongoDB-backed resolver initialized")
	}

	// YouTube client (optional — for stream discovery)
	var ytClient *oracle_ocr.YouTubeClient
	if cfg.YouTubeAPIKey != "" {
		ytClient = oracle_ocr.NewYouTubeClient(cfg.YouTubeAPIKey)
		slog.Info("youtube client initialized")
	} else {
		slog.Warn("YOUTUBE_API_KEY not set, stream discovery disabled")
	}

	// Health server
	health := newHealthServer()
	healthMux := http.NewServeMux()
	healthMux.Handle("/", health)
	healthServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HealthPort),
		Handler: healthMux,
	}

	// Metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: metricsMux,
	}

	// Stream management API on health port
	api := &streamAPI{
		monitors: make(map[string]context.CancelFunc),
	}

	// Wire the newMonitorFn to create real stream monitors
	api.newMonitorFn = func(config oracle_entities.OCRStreamConfig) (*oracle_services.StreamMonitor, oracle_services.StreamMonitorConfig) {
		// Use VodCapture for YouTube VODs, StreamlinkCapture for live streams
		isVOD := config.StreamPlatform == oracle_entities.StreamPlatformYouTube
		var capture oracle_out.StreamCapturePort
		if isVOD {
			capture = vodCapture
		} else {
			capture = streamCapture
		}

		monitor := oracle_services.NewStreamMonitor(capture, ocrEngine, teamResolver, scoreParser, commandHandler)

		var region *oracle_out.Region
		if config.ScoreboardRegion != nil {
			region = &oracle_out.Region{
				X:      config.ScoreboardRegion.X,
				Y:      config.ScoreboardRegion.Y,
				Width:  config.ScoreboardRegion.Width,
				Height: config.ScoreboardRegion.Height,
			}
		}

		monitorConfig := oracle_services.StreamMonitorConfig{
			StreamURL:              config.StreamURL,
			GameID:                 config.GameID,
			ExternalMatchID:        config.ExternalMatchID,
			CaptureIntervalSeconds: config.CaptureIntervalSeconds,
			ScoreboardRegion:       region,
			TeamAHint:              config.TeamAHint,
			TeamBHint:              config.TeamBHint,
			IsVOD:                  isVOD,
		}

		return monitor, monitorConfig
	}

	healthMux.HandleFunc("/api/streams/start", api.handleStart)
	healthMux.HandleFunc("/api/streams/stop", api.handleStop)
	healthMux.HandleFunc("/api/streams", api.handleList)

	// YouTube stream search endpoint
	if ytClient != nil {
		healthMux.HandleFunc("/api/streams/search", func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("q")
			if query == "" {
				query = "CS2 tournament live"
			}
			streams, err := ytClient.SearchLiveStreams(r.Context(), query, 10)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to search live streams", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(streams)
		})
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
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("health server error", slog.String("error", err.Error()))
		}
	}()

	// Start metrics server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("metrics server starting", slog.Int("port", cfg.MetricsPort))
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("metrics server error", slog.String("error", err.Error()))
		}
	}()

	health.ready.Store(true)
	health.mu.Lock()
	health.details["status"] = "ready"
	health.details["ocr_engine"] = "paddleocr"
	health.details["stream_capture"] = "streamlink+ffmpeg"
	health.mu.Unlock()

	slog.Info("OCR worker ready",
		slog.Int("health_port", cfg.HealthPort),
		slog.Int("metrics_port", cfg.MetricsPort),
	)

	// Resolve OCR stream config repository for polling pending VODs
	var streamConfigRepo oracle_out.OCRStreamConfigRepository
	if err := c.Resolve(&streamConfigRepo); err != nil {
		slog.Warn("OCRStreamConfigRepository not available, VOD polling disabled", slog.String("error", err.Error()))
	}

	// Start VOD processing loop — polls for pending OCR stream configs
	if streamConfigRepo != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pollInterval := time.Duration(cfg.DefaultCaptureInterval*6) * time.Second
			if pollInterval < 30*time.Second {
				pollInterval = 30 * time.Second
			}
			slog.Info("VOD processing loop starting", slog.Duration("poll_interval", pollInterval))

			ticker := time.NewTicker(pollInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					slog.Info("VOD processing loop stopped")
					return
				case <-ticker.C:
					pendingConfigs, err := streamConfigRepo.FindPending(ctx, 5)
					if err != nil {
						slog.Warn("failed to fetch pending OCR configs", slog.String("error", err.Error()))
						continue
					}

					for _, config := range pendingConfigs {
						api.monitorsMu.RLock()
						_, alreadyRunning := api.monitors[config.StreamURL]
						api.monitorsMu.RUnlock()
						if alreadyRunning {
							continue
						}

						slog.Info("starting VOD processing from pending config",
							slog.String("stream_url", config.StreamURL),
							slog.String("external_match_id", config.ExternalMatchID),
							slog.String("game_id", string(config.GameID)),
						)

						// Mark as active
						if err := config.Activate(); err != nil {
							slog.Warn("failed to activate config", slog.String("error", err.Error()))
							continue
						}
						if err := streamConfigRepo.Update(ctx, config); err != nil {
							slog.Warn("failed to update config status", slog.String("error", err.Error()))
						}

						monitor, monitorCfg := api.newMonitorFn(*config)

						monitorCtx, monitorCancel := context.WithCancel(ctx)
						api.monitorsMu.Lock()
						api.monitors[config.StreamURL] = monitorCancel
						api.monitorsMu.Unlock()
						ocrActiveStreams.Inc()

						streamURL := config.StreamURL
						wg.Add(1)
						go func() {
							defer wg.Done()
							defer func() {
								api.monitorsMu.Lock()
								delete(api.monitors, streamURL)
								api.monitorsMu.Unlock()
								ocrActiveStreams.Dec()
							}()

							if err := monitor.MonitorStream(monitorCtx, monitorCfg); err != nil && err != context.Canceled {
								slog.Warn("stream monitor ended with error",
									slog.String("stream_url", streamURL),
									slog.String("error", err.Error()),
								)
							}

							// Mark as completed
							cfg, err := streamConfigRepo.FindByExternalMatchID(monitorCtx, monitorCfg.ExternalMatchID)
							if err == nil && cfg != nil {
								cfg.Complete()
								_ = streamConfigRepo.Update(context.Background(), cfg)
							}
						}()
					}
				}
			}
		}()
	}

	_ = ytClient

	// Wait for shutdown signal
	sig := <-sigCh
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))

	health.ready.Store(false)
	health.alive.Store(false)

	// Cancel all stream monitors
	api.monitorsMu.Lock()
	for url, cancelFn := range api.monitors {
		cancelFn()
		slog.Info("stopped stream monitor", slog.String("stream_url", url))
	}
	api.monitorsMu.Unlock()

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = healthServer.Shutdown(shutdownCtx)
	_ = metricsServer.Shutdown(shutdownCtx)

	wg.Wait()
	slog.Info("OCR worker stopped")
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

func envListOrDefault(key string, defaultVal []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return defaultVal
}
