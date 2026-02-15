package metadata

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type EventQueryService struct {
	*shared.BaseQueryService[replay_entity.GameEvent]
	eventRepo replay_out.GameEventReader
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

	baseService := &shared.BaseQueryService[replay_entity.GameEvent]{
		Reader:          eventReader.(shared.Searchable[replay_entity.GameEvent]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        shared.UserAudienceIDKey,
	}

	return &EventQueryService{
		BaseQueryService: baseService,
		eventRepo:        eventReader,
	}
}

// GetMatchEvents retrieves events for a specific match without RLS restrictions
func (s *EventQueryService) GetMatchEvents(ctx context.Context, gameID string, matchID uuid.UUID, limit, offset int, eventType string) ([]replay_entity.GameEvent, error) {
	return s.eventRepo.GetMatchEvents(ctx, gameID, matchID, limit, offset, eventType)
}

// GetMatchEventsWithCount retrieves events with total count for pagination
// Supports multiple event types for efficient server-side filtering
func (s *EventQueryService) GetMatchEventsWithCount(ctx context.Context, gameID string, matchID uuid.UUID, limit, offset int, eventTypes []string) ([]replay_entity.GameEvent, int64, error) {
	return s.eventRepo.GetMatchEventsWithCount(ctx, gameID, matchID, limit, offset, eventTypes)
}
