package iam_query_services

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
)

type MembershipQueryService struct {
	common.BaseQueryService[iam_entities.Membership]
}

func NewwMembershipQueryService(MembershipReader common.Searchable[iam_entities.Membership]) *MembershipQueryService {
	queryableFields := map[string]bool{
		"ID":            true,
		"Type":          common.ALLOW,
		"ResourceOwner": common.ALLOW,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"Type":          true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &MembershipQueryService{
		BaseQueryService: common.BaseQueryService[iam_entities.Membership]{
			Reader:          MembershipReader,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.UserAudienceIDKey,
		},
	}
}
