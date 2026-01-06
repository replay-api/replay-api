package iam_use_cases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entity "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

// CreateRIDTokenUseCase handles the creation of RID (Resource ID) tokens for authenticated users.
//
// This use case creates authentication tokens that identify users/clients and their permissions.
// Tokens are used for:
//   - User session management (UserAudienceIDKey)
//   - Client application authentication (ClientApplicationAudienceIDKey)
//
// Token Properties:
//   - ID: Unique token identifier (UUID v4)
//   - Key: Token key for verification (UUID v4)
//   - Source: Auth provider (Steam/Google/Email)
//   - ExpiresAt: 1 hour from creation
//   - GrantType: "authorization_code" for users, "client_credentials" for applications
//
// Dependencies:
//   - RIDWriter: Persists tokens to storage
//   - RIDReader: Validates token uniqueness (reserved for future use)
type CreateRIDTokenUseCase struct {
	RIDWriter iam_out.RIDTokenWriter
	RIDReader iam_out.RIDTokenReader
}

func NewCreateRIDTokenUseCase(rIDWriter iam_out.RIDTokenWriter, rIDReader iam_out.RIDTokenReader) iam_in.CreateRIDTokenCommand {
	return &CreateRIDTokenUseCase{
		RIDWriter: rIDWriter,
		RIDReader: rIDReader,
	}
}

// Exec creates a new RID token for the given resource owner.
//
// Parameters:
//   - ctx: Context for logging and cancellation
//   - reso: ResourceOwner containing UserID, GroupID, TenantID
//   - source: Authentication source (Steam/Google/Email)
//   - aud: Intended audience (User or Client Application)
//
// Returns:
//   - *RIDToken: Created token with ID, Key, and expiration
//   - error: Database errors or validation failures
func (usecase *CreateRIDTokenUseCase) Exec(ctx context.Context, reso shared.ResourceOwner, source iam_entity.RIDSourceKey, aud shared.IntendedAudienceKey) (*iam_entity.RIDToken, error) {
	duration, _ := time.ParseDuration("1h")
	expiresAt := time.Now().Add(duration)

	// Note: Token uniqueness is ensured by UUID generation; existence check reserved for future rate limiting

	var grantType string
	switch aud {
	case shared.UserAudienceIDKey:
		grantType = "authorization_code"
	case shared.ClientApplicationAudienceIDKey:
		grantType = "client_credentials"
	}

	token, err := usecase.RIDWriter.Create(ctx, &iam_entity.RIDToken{
		ID:               uuid.New(),
		Key:              uuid.New(),
		Source:           source,
		ResourceOwner:    reso,
		IntendedAudience: aud,
		GrantType:        grantType,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	})

	if err != nil {
		slog.ErrorContext(ctx, "unable to create rid token", "err", err)
		return nil, err
	}

	return token, nil
}
