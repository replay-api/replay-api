package matchmaking_services

import (
	"context"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// MatchmakingSessionQueryService provides domain query operations for matchmaking sessions using technology-agnostic search patterns
type MatchmakingSessionQueryService struct {
	reader          shared.Searchable[matchmaking_entities.MatchmakingSession]
	queryableFields map[string]bool
}

// NewMatchmakingSessionQueryService creates a new matchmaking session query service
func NewMatchmakingSessionQueryService(sessionReader shared.Searchable[matchmaking_entities.MatchmakingSession]) *MatchmakingSessionQueryService {
	queryableFields := map[string]bool{
		"ID":            true,
		"PlayerID":      true,
		"SquadID":       true,
		"Status":        true,
		"PlayerMMR":     true,
		"QueuedAt":      true,
		"MatchedAt":     true,
		"MatchID":       true,
		"EstimatedWait": true,
		"ExpiresAt":     true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
		// Preferences fields
		"GameID":   true,
		"GameMode": true,
		"Region":   true,
		"Tier":     true,
	}

	return &MatchmakingSessionQueryService{
		reader:          sessionReader,
		queryableFields: queryableFields,
	}
}

// GetByID returns a single session by its ID
func (s *MatchmakingSessionQueryService) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingSession, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("ID", id).
			Build()).
		WithLimit(1).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, nil
	}

	return &entities[0], nil
}

// FindByPlayerID finds active matchmaking sessions for a specific player
// Business rule: Only return sessions in queued or searching status
func (s *MatchmakingSessionQueryService) FindByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			NewParam().
			WithValueParam("PlayerID", playerID).
			WithValueParam("Status", matchmaking_entities.StatusQueued, matchmaking_entities.StatusSearching).
			Build()).
		WithSort("QueuedAt", shared.AscendingIDKey).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	sessions := make([]*matchmaking_entities.MatchmakingSession, len(entities))
	for i := range entities {
		sessions[i] = &entities[i]
	}

	return sessions, nil
}

// FindActiveSessions finds matchmaking sessions based on complex filters
// Business rule: Sessions must not be expired and match the specified criteria
func (s *MatchmakingSessionQueryService) FindActiveSessions(ctx context.Context, gameID, gameMode, region string, tier *matchmaking_entities.MatchmakingTier, status *matchmaking_entities.SessionStatus, minMMR, maxMMR *int, limit, offset int) ([]*matchmaking_entities.MatchmakingSession, error) {
	// Build aggregation with filters
	aggBuilder := shared.NewSearchAggregation().NewParam()

	// Add filters
	if gameID != "" {
		aggBuilder.WithValueParam("GameID", gameID)
	}
	if gameMode != "" {
		aggBuilder.WithValueParam("GameMode", gameMode)
	}
	if region != "" {
		aggBuilder.WithValueParam("Region", region)
	}
	if tier != nil {
		aggBuilder.WithValueParam("Tier", *tier)
	}
	if status != nil {
		aggBuilder.WithValueParam("Status", *status)
	} else {
		// Default to active statuses
		aggBuilder.WithValueParam("Status", matchmaking_entities.StatusQueued, matchmaking_entities.StatusSearching)
	}

	// Add MMR range filters
	if minMMR != nil {
		aggBuilder.WithValueParamWithOperator("PlayerMMR", shared.GreaterThanOrEqualOperator, *minMMR)
	}
	if maxMMR != nil {
		aggBuilder.WithValueParamWithOperator("PlayerMMR", shared.LessThanOrEqualOperator, *maxMMR)
	}

	// Add expiration filter - sessions must not be expired
	now := time.Now()
	aggBuilder.WithDateParam("ExpiresAt", &now, nil) // ExpiresAt > now

	builder := shared.NewSearchBuilder().
		WithAggregation(aggBuilder.Build()).
		WithSort("QueuedAt", shared.AscendingIDKey)

	if limit > 0 {
		builder.WithLimit(uint(limit))
	}
	if offset > 0 {
		builder.WithSkip(uint(offset))
	}

	search := builder.Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	sessions := make([]*matchmaking_entities.MatchmakingSession, len(entities))
	for i := range entities {
		sessions[i] = &entities[i]
	}

	return sessions, nil
}

// FindExpiredSessions finds sessions that have expired
// Business rule: Sessions with ExpiresAt in the past
func (s *MatchmakingSessionQueryService) FindExpiredSessions(ctx context.Context, limit int) ([]*matchmaking_entities.MatchmakingSession, error) {
	now := time.Now()
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			NewParam().
			WithDateParam("ExpiresAt", nil, &now). // ExpiresAt <= now
			Build()).
		WithSort("ExpiresAt", shared.AscendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	sessions := make([]*matchmaking_entities.MatchmakingSession, len(entities))
	for i := range entities {
		sessions[i] = &entities[i]
	}

	return sessions, nil
}
