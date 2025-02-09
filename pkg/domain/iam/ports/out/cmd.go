package iam_out

import (
	"context"

	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type RIDTokenWriter interface {
	Create(ctx context.Context, rid *iam_entities.RIDToken) (*iam_entities.RIDToken, error)
}

type UserWriter interface {
	CreateMany(createCtx context.Context, events []*iam_entities.User) error
	Create(createCtx context.Context, events *iam_entities.User) (*iam_entities.User, error)
}

type GroupWriter interface {
	CreateMany(createCtx context.Context, events []*iam_entities.Group) error
	Create(createCtx context.Context, events *iam_entities.Group) (*iam_entities.Group, error)
}

type ProfileWriter interface {
	CreateMany(createCtx context.Context, events []*iam_entities.Profile) error
	Create(createCtx context.Context, events *iam_entities.Profile) (*iam_entities.Profile, error)
}

type MembershipWriter interface {
	CreateMany(createCtx context.Context, events []*iam_entities.Membership) error
	Create(createCtx context.Context, events *iam_entities.Membership) (*iam_entities.Membership, error)
}

type JwkWriter interface {
	CreateMany(createCtx context.Context, events []*iam_entities.Jwk) error
	Create(createCtx context.Context, events *iam_entities.Jwk) (*iam_entities.Jwk, error)
}
