package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type CreateRIDTokenCommand interface {
	Exec(ctx context.Context, reso common.ResourceOwner, source iam_entities.RIDSourceKey, aud common.IntendedAudienceKey) (*iam_entities.RIDToken, error)
}

type VerifyRIDKeyCommand interface {
	Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, error)
}

type OnboardOpenIDUserCommand struct {
	Source iam_entities.RIDSourceKey `json:"rid_source" bson:"rid_source"`
	Key    string                    `json:"key" bson:"key"`
	Name   string                    `json:"name" bson:"name"`
}

type OnboardOpenIDUserCommandHandler interface {
	Exec(ctx context.Context, cmd OnboardOpenIDUserCommand) (*iam_entities.Profile, *iam_entities.RIDToken, error)
}
