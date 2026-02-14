package iam_use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

type VerifyRIDUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

func NewVerifyRIDUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.VerifyRIDKeyCommand {
	return &VerifyRIDUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

func (usecase *VerifyRIDUseCase) Exec(ctx context.Context, key uuid.UUID) (shared.ResourceOwner, shared.IntendedAudienceKey, error) {
	token, err := usecase.RIDReader.FindByID(ctx, key)

	if err != nil {
		slog.ErrorContext(ctx, "error getting rid token by key", "err", err)
		return shared.ResourceOwner{}, shared.UserAudienceIDKey, err
	}

	if token == nil || token.ID == uuid.Nil {
		err = fmt.Errorf("invalid rid key")
		slog.ErrorContext(ctx, err.Error(), "key", key)
		return shared.ResourceOwner{}, shared.UserAudienceIDKey, err
	}

	if token.IsExpired() {
		slog.WarnContext(ctx, "expired rid token", "key", key, "expires_at", token.ExpiresAt)
		return shared.ResourceOwner{}, shared.UserAudienceIDKey, fmt.Errorf("expired RID token")
	}

	if token.IsRevoked() {
		slog.WarnContext(ctx, "revoked rid token", "key", key)
		return shared.ResourceOwner{}, shared.UserAudienceIDKey, fmt.Errorf("revoked RID token")
	}

	return token.ResourceOwner, token.IntendedAudience, nil
}
