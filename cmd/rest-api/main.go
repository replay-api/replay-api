package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	//	"golang.org/x/oauth2/jwt"

	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/routing"
	ioc "github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	c := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().Build()

	router := routing.NewRouter(ctx, c)

	slog.InfoContext(ctx, "Starting server on port 4991")

	http.ListenAndServe(":4991", router)
}