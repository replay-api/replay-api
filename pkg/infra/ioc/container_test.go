//go:build integration

// Package ioc_test contains integration tests for the IoC container.
// These tests require a running MongoDB instance and should only run
// in environments with database access (e.g., local dev or integration CI job).
package ioc_test

import (
	"context"
	"os"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/replay-api/replay-api/pkg/domain/steam/ports/out"
	ioc "github.com/replay-api/replay-api/pkg/infra/ioc"

	shared "github.com/resource-ownership/go-common/pkg/common"
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
		instance := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().WithSquadAPI().Build()
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

	groupID := uuid.New()
	userID := uuid.New()

	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)

	steamUser := &steam_entity.SteamUser{ID: userID,
		VHash: "4ef1c47e874ec4425c5786cddadd9adfc908a530ada95a602742f49c32430185",
		Steam: steam_entity.Steam{
			ID: "76561198169377459",
		},
		ResourceOwner: shared.GetResourceOwner(ctx),
	}
	err = command.Validate(ctx, steamUser)

	if err != nil {
		t.Fatalf("failed to validate OnboardSteamUserCommand: %v", err)
	}

	steamUser, token, err := command.Exec(ctx, steamUser)

	if err != nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: %v", err)
	}

	if token == nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token is nil")
	}

	if token.ID == uuid.Nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ID is nil")
	}

	if token.ResourceOwner.UserID == uuid.Nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ResourceOwner UserID is nil")
	}

	if token.ResourceOwner.GroupID == uuid.Nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ResourceOwner GroupID is nil")
	}

	if token.ResourceOwner.ClientID == uuid.Nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ResourceOwner ClientID is nil")
	}

	if token.ResourceOwner.TenantID == uuid.Nil {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ResourceOwner TenantID is nil")
	}

	if token.CreatedAt.IsZero() {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token CreatedAt is zero")
	}

	if token.IntendedAudience == shared.UserAudienceIDKey && token.ResourceOwner.UserID != steamUser.ResourceOwner.UserID {
		t.Fatalf("failed to execute OnboardSteamUserCommand: token ResourceOwner UserID does not match steamUser ResourceOwner UserID")
	}

}

func TestResolverSteamUserReader(t *testing.T) {
	container := getContainer()

	var writer steam_out.SteamUserWriter
	err := container.Resolve(&writer)
	if err != nil {
		t.Fatalf("failed to resolve SteamUserWriter: %v", err)
	}

	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())

	steamCommunityDetails := steam_entity.Steam{
		ID: "1",
	}

	reso := shared.GetResourceOwner(ctx)

	user := &steam_entity.SteamUser{
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

	s := shared.NewSearchByID(ctx, user.ID, shared.ClientApplicationAudienceIDKey)

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

	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())

	user := &steam_entity.SteamUser{
		ID:    shared.GetResourceOwner(ctx).UserID,
		Steam: steam_entity.Steam{ID: "1"},
	}

	_, err = writer.Create(ctx, user)

	if err != nil {
		t.Fatalf("failed to create SteamUserWriter: %v", err)
	}
}
