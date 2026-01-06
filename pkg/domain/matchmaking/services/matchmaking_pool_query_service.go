package matchmaking_services

import (
	"context"

	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// MatchmakingPoolQueryService provides domain query operations for matchmaking pools using technology-agnostic search patterns
type MatchmakingPoolQueryService struct {
	reader          shared.Searchable[matchmaking_entities.MatchmakingPool]
	queryableFields map[string]bool
}

// NewMatchmakingPoolQueryService creates a new matchmaking pool query service
func NewMatchmakingPoolQueryService(poolReader shared.Searchable[matchmaking_entities.MatchmakingPool]) *MatchmakingPoolQueryService {
	queryableFields := map[string]bool{
		"ID":          true,
		"GameID":      true,
		"GameMode":    true,
		"Region":      true,
		"IsActive":    true,
		"MinMMR":      true,
		"MaxMMR":      true,
		"PlayerCount": true,
		"CreatedAt":   true,
		"UpdatedAt":   true,
	}

	return &MatchmakingPoolQueryService{
		reader:          poolReader,
		queryableFields: queryableFields,
	}
}

// FindByGameModeRegion finds a pool by game, mode and region
// Business rule: Pools are uniquely identified by game+mode+region combination
func (s *MatchmakingPoolQueryService) FindByGameModeRegion(ctx context.Context, gameID, gameMode, region string) (*matchmaking_entities.MatchmakingPool, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			NewParam().
			WithValueParam("GameID", gameID).
			WithValueParam("GameMode", gameMode).
			WithValueParam("Region", region).
			WithValueParam("IsActive", true).
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

// FindAllActive finds all currently active pools
// Business rule: Only active pools are valid for matchmaking
func (s *MatchmakingPoolQueryService) FindAllActive(ctx context.Context, limit int) ([]*matchmaking_entities.MatchmakingPool, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			WithValueParam("IsActive", true).
			Build()).
		WithSort("PlayerCount", shared.DescendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	pools := make([]*matchmaking_entities.MatchmakingPool, len(entities))
	for i := range entities {
		pools[i] = &entities[i]
	}

	return pools, nil
}

// FindByGame finds all pools for a specific game
// Business rule: Return both active and inactive pools for analytics
func (s *MatchmakingPoolQueryService) FindByGame(ctx context.Context, gameID string, activeOnly bool) ([]*matchmaking_entities.MatchmakingPool, error) {
	aggBuilder := shared.NewSearchAggregation().
		WithValueParam("GameID", gameID)

	if activeOnly {
		aggBuilder.WithValueParam("IsActive", true)
	}

	search := shared.NewSearchBuilder().
		WithAggregation(aggBuilder.Build()).
		WithSort("Region", shared.AscendingIDKey).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	pools := make([]*matchmaking_entities.MatchmakingPool, len(entities))
	for i := range entities {
		pools[i] = &entities[i]
	}

	return pools, nil
}

// FindPoolsWithMinPlayers finds pools that have at least the specified number of players
// Business rule: Used for analytics to find busy pools
func (s *MatchmakingPoolQueryService) FindPoolsWithMinPlayers(ctx context.Context, minPlayers int, limit int) ([]*matchmaking_entities.MatchmakingPool, error) {
	search := shared.NewSearchBuilder().
		WithAggregation(shared.NewSearchAggregation().
			NewParam().
			WithValueParam("IsActive", true).
			WithValueParamWithOperator("PlayerCount", shared.GreaterThanOrEqualOperator, minPlayers).
			Build()).
		WithSort("PlayerCount", shared.DescendingIDKey).
		WithLimit(uint(limit)).
		Build()

	entities, err := s.reader.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	pools := make([]*matchmaking_entities.MatchmakingPool, len(entities))
	for i := range entities {
		pools[i] = &entities[i]
	}

	return pools, nil
}
