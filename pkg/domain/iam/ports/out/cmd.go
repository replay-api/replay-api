package iam_out

import (
	"context"

	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type RIDTokenWriter interface {
	Create(ctx context.Context, rid iam_entity.RIDToken) (*iam_entity.RIDToken, error)
}
