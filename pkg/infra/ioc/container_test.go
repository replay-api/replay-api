package ioc_test

import (
	"context"
	"os"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	steam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"
	ioc "github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

var (
	c *container.Container
)

func getContainer() *container.Container {
	os.Setenv("DEV_ENV", "test")
	os.Setenv("MONGO_URI", "mongodb://127.0.0.1:37019/replay")
	os.Setenv("MONGO_DB_NAME", "replay")
	os.Setenv("STEAM_VHASH_SOURCE", "82DA0F0D0135FEA0F5DDF6F96528B48A")

	if c == nil {
		instance := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().Build()
		return &instance
	}

	return c
}

func TestResolveOnboardSteamUserCommand(t *testing.T) {
	container := getContainer()
	var command steam_in.OnboardSteamUserCommand
	err := container.Resolve(&command)
	if err != nil {
		t.Fatalf("failed to resolve OnboardSteamUserCommand: %v", err)
	}

	ctx := context.WithValue(context.Background(), common.TenantIDKey, common.TeamPROTenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

	steamUser := steam_entity.SteamUser{ID: uuid.New(),
		Steam: steam_entity.Steam{
			ID: "8d2f508f233cea95b70b14ac0b7b9ae58d01ec3029ab8a08dac96234bc5f5746",
		}}
	err = command.Validate(ctx, steamUser)

	if err != nil {
		t.Fatalf("failed to validate OnboardSteamUserCommand: %v", err)
	}

	_, err = command.Exec(ctx, steamUser)

	if err != nil {
		t.Fatalf("failed to validate OnboardSteamUserCommand: %v", err)
	}
}

func TestResolverSteamUserReader(t *testing.T) {
	container := getContainer()

	var writer steam_out.SteamUserWriter
	err := container.Resolve(&writer)
	if err != nil {
		t.Fatalf("failed to resolve SteamUserWriter: %v", err)
	}

	ctx := context.WithValue(context.Background(), common.TenantIDKey, common.TeamPROTenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

	steamCommunityDetails := steam_entity.Steam{
		ID: "1",
	}

	reso := common.GetResourceOwner(ctx)

	user := steam_entity.SteamUser{
		ID:            reso.UserID,
		Steam:         steamCommunityDetails,
		ResourceOwner: reso,
	}

	_, err = writer.Create(ctx, user)

	if err != nil {
		t.Fatalf("failed to create SteamUserWriter: %v", err)
	}

	var reader steam_out.SteamUserReader
	err = container.Resolve(&reader)
	if err != nil {
		t.Fatalf("failed to resolve SteamUserReader: %v", err)
	}

	s := common.NewSearchByID(ctx, user.ID, common.ClientApplicationAudienceIDKey)

	steamUser, err := reader.Search(ctx, s)

	if err != nil {
		t.Fatalf("failed to search SteamUserReader: %v", err)
	}

	if len(steamUser) == 0 {
		t.Fatalf("failed to search SteamUserReader: no results")
	}

	if err != nil {
		t.Fatalf("failed to search SteamUserReader: %v", err)
	}
}

func TestResolverSteamUserWriter(t *testing.T) {
	container := getContainer()
	var writer steam_out.SteamUserWriter
	err := container.Resolve(&writer)
	if err != nil {
		t.Fatalf("failed to resolve SteamUserWriter: %v", err)
	}

	ctx := context.WithValue(context.Background(), common.TenantIDKey, common.TeamPROTenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

	user := steam_entity.SteamUser{
		ID:    common.GetResourceOwner(ctx).UserID,
		Steam: steam_entity.Steam{ID: "1"},
	}

	_, err = writer.Create(ctx, user)

	if err != nil {
		t.Fatalf("failed to create SteamUserWriter: %v", err)
	}
}
