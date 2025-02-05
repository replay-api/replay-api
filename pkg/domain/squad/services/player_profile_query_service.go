package squad_services

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
)

type PlayerProfileQueryService struct {
	common.BaseQueryService[squad_entities.PlayerProfile]
}

func NewPlayerProfileQueryService(eventReader squad_out.PlayerProfileReader) squad_in.PlayerProfileReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"Nickname":      true,
		"Avatar":        true,
		"Roles":         true,
		"Description":   true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"Nickname":      true,
		"Avatar":        true,
		"Roles":         true,
		"Description":   true,
		"ResourceOwner": common.DENY,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &common.BaseQueryService[squad_entities.PlayerProfile]{
		Reader:          eventReader.(common.Searchable[squad_entities.PlayerProfile]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
