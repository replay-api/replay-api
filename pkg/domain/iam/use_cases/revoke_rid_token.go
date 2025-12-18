package iam_use_cases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

// RevokeRIDTokenUseCase handles token revocation (logout)
type RevokeRIDTokenUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

// NewRevokeRIDTokenUseCase creates a new instance of RevokeRIDTokenUseCase
func NewRevokeRIDTokenUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.RevokeRIDTokenCommand {
	return &RevokeRIDTokenUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

// Exec revokes a token (logout operation)
func (usecase *RevokeRIDTokenUseCase) Exec(ctx context.Context, tokenID uuid.UUID) error {
	if tokenID == uuid.Nil {
		slog.ErrorContext(ctx, "invalid token ID provided for revocation")
		return ErrInvalidTokenID
	}

	// Get existing token to verify ownership
	existingToken, err := usecase.RIDReader.FindByID(ctx, tokenID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find token for revocation", "tokenID", tokenID, "err", err)
		return ErrTokenNotFound
	}

	if existingToken == nil {
		slog.WarnContext(ctx, "token not found for revocation", "tokenID", tokenID)
		// Return success anyway - token doesn't exist so it's effectively revoked
		return nil
	}

	// Verify the authenticated user owns this token
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID != uuid.Nil && resourceOwner.UserID != existingToken.ResourceOwner.UserID {
		slog.WarnContext(ctx, "user attempted to revoke token they don't own",
			"requestUserID", resourceOwner.UserID,
			"tokenOwnerID", existingToken.ResourceOwner.UserID,
		)
		return ErrUnauthorized
	}

	// Revoke the token
	if err := usecase.RIDWriter.Revoke(ctx, tokenID.String()); err != nil {
		slog.ErrorContext(ctx, "failed to revoke token", "tokenID", tokenID, "err", err)
		return err
	}

	slog.InfoContext(ctx, "token revoked successfully",
		"tokenID", tokenID,
		"userID", existingToken.ResourceOwner.UserID,
	)

	return nil
}

