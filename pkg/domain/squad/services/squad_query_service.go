package squad_services

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

type SquadQueryService struct {
	common.BaseQueryService[squad_entities.Squad]
}

func NewSquadQueryService(eventReader squad_out.SquadReader) squad_in.SquadReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"GroupID":       true,
		"GameID":        true,
		"FullName":      true,
		"ShortName":     true,
		"Symbol":        true,
		"Description":   true,
		"Profiles.*":    true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"GroupID":       true,
		"GameID":        true,
		"FullName":      true,
		"ShortName":     true,
		"Symbol":        true,
		"Description":   true,
		"Profiles.*":    true,
		"ResourceOwner": common.DENY,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &common.BaseQueryService[squad_entities.Squad]{
		Reader:          eventReader.(common.Searchable[squad_entities.Squad]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
