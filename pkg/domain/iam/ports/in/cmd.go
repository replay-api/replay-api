package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type CreateRIDTokenCommand interface {
	Exec(ctx context.Context, reso common.ResourceOwner, source iam_entities.RIDSourceKey, aud common.IntendedAudienceKey) (*iam_entities.RIDToken, error)
}

type VerifyRIDKeyCommand interface {
	Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, common.IntendedAudienceKey, error)
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

type SetRIDTokenProfileCommand struct {
	RIDTokenID uuid.UUID   `json:"rid_token_id" bson:"rid_token_id"`
	Profiles   []uuid.UUID `json:"profiles" bson:"profiles"`
}

type SetRIDTokenProfileCommandHandler interface {
	Exec(ctx context.Context, cmd SetRIDTokenProfileCommand) (*iam_entities.RIDToken, iam_dtos.ProfilesDTO, error)
}
