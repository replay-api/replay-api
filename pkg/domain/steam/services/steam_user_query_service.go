package steam_query_services

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	steam_entities "github.com/replay-api/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
)

type SteamUserQueryService struct {
	shared.BaseQueryService[steam_entities.SteamUser]
}

func NewSteamUserQueryService(eventReader shared.Searchable[steam_entities.SteamUser]) steam_in.SteamUserReader {
	queryableFields := map[string]bool{
		"ID":                true,
		"VHash":             shared.DENY,
		"Steam.*":           true,
		"Steam.realname":    true,
		"Steam.personaname": true,
	}

	readableFields := map[string]bool{
		"ID":                true,
		"VHash":             shared.DENY,
		"Steam.*":           true,
		"Steam.realname":    true,
		"Steam.personaname": true,
	}

	return &shared.BaseQueryService[steam_entities.SteamUser]{
		Reader:          eventReader,
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        shared.UserAudienceIDKey,
	}
}
