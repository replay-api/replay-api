package metadata

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type EventQueryService struct {
	shared.BaseQueryService[replay_entity.GameEvent]
}

func NewEventQueryService(eventReader replay_out.GameEventReader) replay_in.EventReader {
	queryableFields := map[string]bool{
		"ID":              true,
		"GameID":          true,
		"MatchID":         true,
		"Type":            true,
		"Time":            true,
		"EventData":       true,
		"PlayerStats":     shared.DENY,
		"NetworkPlayerID": true,
		"PlayerName":      true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
	}

	readableFields := map[string]bool{
		"ID":              true,
		"GameID":          true,
		"MatchID":         true,
		"Type":            true,
		"Time":            true,
		"EventData":       shared.DENY,
		"PlayerStats":     shared.DENY,
		"NetworkPlayerID": shared.DENY,
		"PlayerName":      true,
		"ResourceOwner":   shared.DENY,
		"CreatedAt":       true,
	}

	return &shared.BaseQueryService[replay_entity.GameEvent]{
		Reader:          eventReader.(shared.Searchable[replay_entity.GameEvent]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        shared.UserAudienceIDKey,
	}
}
