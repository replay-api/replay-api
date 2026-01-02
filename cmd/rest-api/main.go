package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	//	"golang.org/x/oauth2/jwt"

	"github.com/replay-api/replay-api/cmd/rest-api/routing"
	jobs "github.com/replay-api/replay-api/pkg/app/jobs"
	ioc "github.com/replay-api/replay-api/pkg/infra/ioc"
	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	// Build container with env and event publisher
	c := builder.WithEnvFile().WithEventPublisher().Build()

	// Inject MongoDB services (includes squad services)
	if err := ioc.InjectMongoDB(c); err != nil {
		slog.Error("Failed to inject MongoDB services", "error", err)
		panic(err)
	}

	// Register Squad API services (repositories, writers, command handlers)
	// This must be called AFTER InjectMongoDB since squad services depend on MongoDB
	builder.WithSquadAPI()

	// Register inbound ports (services/use cases that depend on MongoDB repositories)
	// This must be called AFTER InjectMongoDB and WithSquadAPI since inbound ports depend on outbound ports (repos)
	builder.WithInboundPorts()

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

	if err := server.ListenAndServe(); err != nil {
		slog.ErrorContext(ctx, "Server error", "err", err)
	}

}
