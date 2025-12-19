package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	//	"golang.org/x/oauth2/jwt"

	"github.com/replay-api/replay-api/cmd/rest-api/routing"
	jobs "github.com/replay-api/replay-api/pkg/app/jobs"
	ioc "github.com/replay-api/replay-api/pkg/infra/ioc"
	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	c := builder.WithEnvFile().With(ioc.InjectMongoDB).WithSquadAPI().WithInboundPorts().Build()

	defer builder.Close(c)

	// Start WebSocket Hub
	var wsHub *websocket.WebSocketHub
	if err := c.Resolve(&wsHub); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve WebSocket hub", "error", err)
		panic(err)
	}
	go wsHub.Run(ctx)
	slog.InfoContext(ctx, "WebSocket hub started")

	// Start Prize Distribution Job
	var prizeJob *jobs.PrizeDistributionJob
	if err := c.Resolve(&prizeJob); err != nil {
		slog.ErrorContext(ctx, "Failed to resolve PrizeDistributionJob", "error", err)
		panic(err)
	}
	go prizeJob.Run(ctx)
	slog.InfoContext(ctx, "Prize distribution job started")

	router := routing.NewRouter(ctx, c)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.InfoContext(ctx, "Starting server on port "+port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown handler for Kubernetes SIGTERM
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-shutdownChan
		slog.InfoContext(ctx, "Received shutdown signal", "signal", sig.String())

		// Give Kubernetes time to update endpoints
		slog.InfoContext(ctx, "Waiting for Kubernetes endpoint update...")
		time.Sleep(5 * time.Second)

		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		slog.InfoContext(ctx, "Shutting down server gracefully...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.ErrorContext(ctx, "Server shutdown error", "error", err)
		}

		// Cancel main context to stop background jobs
		cancel()
		slog.InfoContext(ctx, "Server shutdown complete")
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.ErrorContext(ctx, "Server error", "err", err)
		os.Exit(1)
	}

}
