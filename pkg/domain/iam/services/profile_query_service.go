package iam_query_services

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type ProfileQueryService struct {
	common.BaseQueryService[iam_entities.Profile]
}

func NewProfileQueryService(profileReader common.Searchable[iam_entities.Profile]) *ProfileQueryService {
	queryableFields := map[string]bool{
		"ID":                  true,
		"RIDSource":           common.ALLOW,
		"SourceKey":           common.ALLOW,
		"Details.RealName":    common.ALLOW,
		"Details.realname":    common.ALLOW,
		"Details.given_name":  common.ALLOW,
		"Details.family_name": common.ALLOW,
		"ResourceOwner":       common.ALLOW,
		"CreatedAt":           true,
		"UpdatedAt":           true,
	}

	readableFields := map[string]bool{
		"ID":                  true,
		"RIDSource":           true,
		"SourceKey":           true,
		"Details":             common.DENY,
		"Details.RealName":    common.ALLOW,
		"Details.realname":    common.ALLOW,
		"Details.given_name":  common.ALLOW,
		"Details.family_name": common.ALLOW,
		"ResourceOwner":       true,
		"Details.email":       common.DENY,
		"CreatedAt":           true,
		"UpdatedAt":           true,
	}

	return &ProfileQueryService{
		BaseQueryService: common.BaseQueryService[iam_entities.Profile]{
			Reader:          profileReader,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.UserAudienceIDKey,
		},
	}
}
