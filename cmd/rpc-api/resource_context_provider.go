package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
)

type ResourceContextProvider struct {
	verifyRID iam_in.VerifyRIDKeyCommand
}

func NewResourceContextProvider(container *container.Container) *ResourceContextProvider {
	var verifyRID iam_in.VerifyRIDKeyCommand
	err := container.Resolve(&verifyRID)

	if err != nil {
		slog.Error("unable to resolve VerifyRIDKeyCommand")
	}

	return &ResourceContextProvider{
		verifyRID: verifyRID,
	}
}

func (m *ResourceContextProvider) GetVerifiedContext(rpcContext context.Context, ridToken uuid.UUID) (context.Context, error) {
	ctx := context.WithValue(rpcContext, shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, false)

	if ridToken == uuid.Nil {
		return ctx, fmt.Errorf("rid token is nil")
	}

	reso, aud, err := m.verifyRID.Exec(ctx, ridToken)
	if err != nil {
		slog.ErrorContext(ctx, "unable to verify rid", "ResourceOwnerIDHeaderKey", controllers.ResourceOwnerIDHeaderKey, "ridToken", ridToken, "err", err)
		return ctx, err
	}

	slog.InfoContext(ctx, "resource owner verified", "reso", reso, "aud", aud)

	if !reso.IsUser() {
		slog.WarnContext(ctx, "non end user resource owner", "reso", reso)
	}

	ctx = context.WithValue(ctx, shared.GroupIDKey, reso.GroupID)
	ctx = context.WithValue(ctx, shared.UserIDKey, reso.UserID)
	ctx = context.WithValue(ctx, shared.AudienceKey, aud)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)

	return ctx, nil
}
