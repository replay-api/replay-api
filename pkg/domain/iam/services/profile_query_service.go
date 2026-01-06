package iam_query_services

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
)

type ProfileQueryService struct {
	shared.BaseQueryService[iam_entities.Profile]
}

func NewProfileQueryService(profileReader shared.Searchable[iam_entities.Profile]) *ProfileQueryService {
	queryableFields := map[string]bool{
		"ID":                  true,
		"RIDSource":           shared.ALLOW,
		"SourceKey":           shared.ALLOW,
		"Details.RealName":    shared.ALLOW,
		"Details.realname":    shared.ALLOW,
		"Details.given_name":  shared.ALLOW,
		"Details.family_name": shared.ALLOW,
		"ResourceOwner":       shared.ALLOW,
		"CreatedAt":           true,
		"UpdatedAt":           true,
	}

	readableFields := map[string]bool{
		"ID":                  true,
		"RIDSource":           true,
		"SourceKey":           true,
		"Details":             shared.DENY,
		"Details.RealName":    shared.ALLOW,
		"Details.realname":    shared.ALLOW,
		"Details.given_name":  shared.ALLOW,
		"Details.family_name": shared.ALLOW,
		"ResourceOwner":       true,
		"Details.email":       shared.DENY,
		"CreatedAt":           true,
		"UpdatedAt":           true,
	}

	return &ProfileQueryService{
		BaseQueryService: shared.BaseQueryService[iam_entities.Profile]{
			Reader:          profileReader,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        shared.UserAudienceIDKey,
		},
	}
}
