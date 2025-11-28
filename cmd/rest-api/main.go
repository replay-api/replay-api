package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

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

	c := builder.WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().WithSquadAPI().Build()

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

	http.ListenAndServe(":"+port, router)

}
