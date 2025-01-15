package steam_query_services

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	steam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"
	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"
)

type SteamUserQueryService struct {
	common.BaseQueryService[steam_entities.SteamUser]
}

func NewSteamUserQueryService(eventReader steam_out.SteamUserReader) steam_in.SteamUserReader {
	queryableFields := map[string]bool{
		"ID":                true,
		"VHash":             common.DENY,
		"Steam.*":           true,
		"Steam.realname":    true,
		"Steam.personaname": true,
	}

	readableFields := map[string]bool{
		"ID":                true,
		"VHash":             common.DENY,
		"Steam.*":           true,
		"Steam.realname":    true,
		"Steam.personaname": true,
	}

	return &common.BaseQueryService[steam_entities.SteamUser]{
		Reader:          eventReader.(common.Searchable[steam_entities.SteamUser]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
