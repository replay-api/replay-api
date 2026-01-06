package iam_in

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type CreateRIDTokenCommand interface {
	Exec(ctx context.Context, reso shared.ResourceOwner, source iam_entities.RIDSourceKey, aud shared.IntendedAudienceKey) (*iam_entities.RIDToken, error)
}

type VerifyRIDKeyCommand interface {
	Exec(ctx context.Context, key uuid.UUID) (shared.ResourceOwner, shared.IntendedAudienceKey, error)
}

// RefreshRIDTokenCommand refreshes an existing token with a new expiration
type RefreshRIDTokenCommand interface {
	Exec(ctx context.Context, tokenID uuid.UUID) (*iam_entities.RIDToken, error)
}

// RevokeRIDTokenCommand revokes/blacklists a token (logout)
type RevokeRIDTokenCommand interface {
	Exec(ctx context.Context, tokenID uuid.UUID) error
}

type OnboardOpenIDUserCommand struct {
	Source         iam_entities.RIDSourceKey `json:"rid_source" bson:"rid_source"`
	Key            string                    `json:"key" bson:"key"`
	Name           string                    `json:"name" bson:"name"`
	ProfileDetails interface{}               `json:"profile_details" bson:"profile_details"`
}

type OnboardOpenIDUserCommandHandler interface {
	Exec(ctx context.Context, cmd OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error)
}
