package iam_in

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type ProfileReader interface {
	shared.Searchable[iam_entities.Profile]
}

type MembershipReader interface {
	shared.Searchable[iam_entities.Membership]
	ListMemberGroups(ctx context.Context, s *shared.Search) (map[uuid.UUID]iam_dtos.GroupMembershipDTO, error)
}
