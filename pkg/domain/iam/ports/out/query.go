package iam_out

import (
	"context"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type RIDTokenReader interface {
	Search(ctx context.Context, s common.Search) ([]iam_entity.RIDToken, error)
}

type ProfileReader interface {
	Search(ctx context.Context, s common.Search) ([]iam_entity.Profile, error)
}

type UserReader interface {
	Search(ctx context.Context, s common.Search) ([]iam_entity.User, error)
}

// type RIDTokenReader interface {
// 	common.Searchable[iam_entity.RIDToken]
// }
