package iam_in

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type ProfileReader interface {
	common.Searchable[iam_entities.Profile]
}

type MembershipReader interface {
	common.Searchable[iam_entities.Membership]
}
