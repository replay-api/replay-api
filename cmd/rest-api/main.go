package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	//	"golang.org/x/oauth2/jwt"

	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/routing"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	ioc "github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	c := builder.WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().Build()

	defer builder.Close(c)

	config := common.Config{}
	c.Resolve(&config)

	router := routing.NewRouter(ctx, c)

	slog.InfoContext(ctx, "Starting server on port "+config.Api.Port)

	http.ListenAndServe(":"+config.Api.Port, router)

}
