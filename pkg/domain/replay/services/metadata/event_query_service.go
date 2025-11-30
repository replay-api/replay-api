package metadata

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type EventQueryService struct {
	common.BaseQueryService[replay_entity.GameEvent]
}

func NewEventQueryService(eventReader replay_out.GameEventReader) replay_in.EventReader {
	queryableFields := map[string]bool{
		"ID":              true,
		"GameID":          true,
		"MatchID":         true,
		"Type":            true,
		"Time":            true,
		"EventData":       true,
		"PlayerStats":     common.DENY,
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
		"EventData":       common.DENY,
		"PlayerStats":     common.DENY,
		"NetworkPlayerID": common.DENY,
		"PlayerName":      true,
		"ResourceOwner":   common.DENY,
		"CreatedAt":       true,
	}

	return &common.BaseQueryService[replay_entity.GameEvent]{
		Reader:          eventReader.(common.Searchable[replay_entity.GameEvent]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
