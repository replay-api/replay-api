package iam_use_cases

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenExpired      = errors.New("token has expired")
	ErrTokenRevoked      = errors.New("token has been revoked")
	ErrUnauthorized      = errors.New("unauthorized: user does not own this token")
	ErrInvalidTokenID    = errors.New("invalid token ID")
)

// RefreshRIDTokenUseCase handles token refresh operations
type RefreshRIDTokenUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

// NewRefreshRIDTokenUseCase creates a new instance of RefreshRIDTokenUseCase
func NewRefreshRIDTokenUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.RefreshRIDTokenCommand {
	return &RefreshRIDTokenUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

// Exec refreshes an existing token by creating a new one with extended expiration
func (usecase *RefreshRIDTokenUseCase) Exec(ctx context.Context, tokenID uuid.UUID) (*iam_entity.RIDToken, error) {
	if tokenID == uuid.Nil {
		slog.ErrorContext(ctx, "invalid token ID provided for refresh")
		return nil, ErrInvalidTokenID
	}

	// Get existing token
	existingToken, err := usecase.RIDReader.FindByID(ctx, tokenID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find token for refresh", "tokenID", tokenID, "err", err)
		return nil, ErrTokenNotFound
	}

	if existingToken == nil {
		slog.ErrorContext(ctx, "token not found", "tokenID", tokenID)
		return nil, ErrTokenNotFound
	}

	// Check if token is expired (allow refresh within a grace period of 7 days after expiry)
	gracePeriod := 7 * 24 * time.Hour
	if existingToken.ExpiresAt.Add(gracePeriod).Before(time.Now()) {
		slog.WarnContext(ctx, "token expired beyond grace period", "tokenID", tokenID, "expiresAt", existingToken.ExpiresAt)
		return nil, ErrTokenExpired
	}

	// Verify the authenticated user owns this token
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID != uuid.Nil && resourceOwner.UserID != existingToken.ResourceOwner.UserID {
		slog.WarnContext(ctx, "user attempted to refresh token they don't own",
			"requestUserID", resourceOwner.UserID,
			"tokenOwnerID", existingToken.ResourceOwner.UserID,
		)
		return nil, ErrUnauthorized
	}

	// Create new token with extended expiration (1 hour for access tokens)
	newExpiration := time.Now().Add(1 * time.Hour)

	newToken := &iam_entity.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           existingToken.Source,
		ResourceOwner:    existingToken.ResourceOwner,
		IntendedAudience: existingToken.IntendedAudience,
		GrantType:        "refresh_token",
		ExpiresAt:        newExpiration,
		CreatedAt:        time.Now(),
	}

	createdToken, err := usecase.RIDWriter.Create(ctx, newToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create refreshed token", "err", err)
		return nil, err
	}

	// Optionally revoke the old token (security best practice)
	// This prevents token reuse attacks
	if err := usecase.RIDWriter.Revoke(ctx, tokenID.String()); err != nil {
		// Log but don't fail the refresh - the new token is already created
		slog.WarnContext(ctx, "failed to revoke old token after refresh", "tokenID", tokenID, "err", err)
	}

	slog.InfoContext(ctx, "token refreshed successfully",
		"oldTokenID", tokenID,
		"newTokenID", createdToken.ID,
		"userID", existingToken.ResourceOwner.UserID,
	)

	return createdToken, nil
}

