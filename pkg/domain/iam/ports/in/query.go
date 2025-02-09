package iam_in

import (
	"context"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_dtos "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type ProfileReader interface {
	common.Searchable[iam_entities.Profile]
}

type MembershipReader interface {
	common.Searchable[iam_entities.Membership]
	ListMemberGroups(ctx context.Context, s *common.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error)
}

type WellKnownReader interface {
	GetOpenConfiguration(ctx context.Context) (iam_dtos.OpenConfigurationDTO, error)
	GetOpenConfigurationJwks(ctx context.Context) (iam_dtos.OpenConfigurationJwksDTO, error)
}
