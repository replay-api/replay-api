package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type CreateRIDTokenCommand interface {
	Exec(ctx context.Context, reso common.ResourceOwner, source iam_entity.RIDSourceKey, aud common.IntendedAudienceKey) (*iam_entity.RIDToken, error)
}

type VerifyRIDKeyCommand interface {
	Exec(ctx context.Context, key uuid.UUID) (common.ResourceOwner, error)
}
