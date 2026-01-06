package matchmaking_services

import (
	"context"
	"time"

	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

type PrizePoolQueryService struct {
	shared.BaseQueryService[matchmaking_entities.PrizePool]
}

// Helper functions for common search patterns
func NewPrizePoolStatusSearch(status matchmaking_entities.PrizePoolStatus) *shared.SearchAggregationBuilder {
	return shared.NewSearchAggregation().WithValueParam("Status", status)
}

func NewPrizePoolMatchIDSearch(matchID interface{}) *shared.SearchAggregationBuilder {
	return shared.NewSearchAggregation().WithValueParam("MatchID", matchID)
}

func NewPendingDistributionsSearch() *shared.SearchAggregationBuilder {
	now := time.Now()
	return shared.NewSearchAggregation().
		NewParam().
		WithValueParam("Status", matchmaking_entities.PrizePoolStatusInEscrow).
		WithDateParam("EscrowEndTime", nil, &now)
}

func NewPendingDistributionsResultOptions(limit int) *shared.SearchResultOptionsBuilder {
	return shared.NewSearchResultOptionsBuilder().WithLimit(uint(limit))
}

func NewPrizePoolByStatusResultOptions(status matchmaking_entities.PrizePoolStatus, limit int) *shared.SearchResultOptionsBuilder {
	return shared.NewSearchResultOptionsBuilder().WithLimit(uint(limit))
}

func NewPrizePoolByMatchIDResultOptions() *shared.SearchResultOptionsBuilder {
	return shared.NewSearchResultOptionsBuilder().WithLimit(1)
}

// PrizePoolSearchBuilder provides domain-specific fluent API for building prize pool searches
type PrizePoolSearchBuilder struct {
	builder *shared.SearchBuilder
}

func NewPrizePoolSearch() *PrizePoolSearchBuilder {
	return &PrizePoolSearchBuilder{
		builder: shared.NewSearchBuilder(),
	}
}

func (b *PrizePoolSearchBuilder) WithPendingDistributions() *PrizePoolSearchBuilder {
	now := time.Now()
	aggregation := shared.NewSearchAggregation().
		NewParam().
		WithValueParam("Status", matchmaking_entities.PrizePoolStatusInEscrow).
		WithDateParam("EscrowEndTime", nil, &now).
		Build()
	
	b.builder.WithAggregation(aggregation)
	return b
}

func (b *PrizePoolSearchBuilder) WithStatus(status matchmaking_entities.PrizePoolStatus) *PrizePoolSearchBuilder {
	aggregation := shared.NewSearchAggregation().
		WithValueParam("Status", status).
		Build()
	
	b.builder.WithAggregation(aggregation)
	return b
}

func (b *PrizePoolSearchBuilder) WithMatchID(matchID interface{}) *PrizePoolSearchBuilder {
	aggregation := shared.NewSearchAggregation().
		WithValueParam("MatchID", matchID).
		Build()
	
	b.builder.WithAggregation(aggregation)
	return b
}

func (b *PrizePoolSearchBuilder) WithLimit(limit int) *PrizePoolSearchBuilder {
	b.builder.WithLimit(uint(limit))
	return b
}

func (b *PrizePoolSearchBuilder) WithSortByCreatedAt(direction shared.SortDirectionKey) *PrizePoolSearchBuilder {
	b.builder.WithSort("CreatedAt", direction)
	return b
}

func (b *PrizePoolSearchBuilder) WithSortByEscrowEndTime(direction shared.SortDirectionKey) *PrizePoolSearchBuilder {
	b.builder.WithSort("EscrowEndTime", direction)
	return b
}

func (b *PrizePoolSearchBuilder) Build() shared.Search {
	return b.builder.Build()
}

func NewPrizePoolQueryService(prizePoolReader shared.Searchable[matchmaking_entities.PrizePool]) *PrizePoolQueryService {
	queryableFields := map[string]bool{
		"ID":                   true,
		"MatchID":              true,
		"GameID":               true,
		"Region":               true,
		"Currency":             true,
		"TotalAmount":          true,
		"PlatformContribution": true,
		"DistributionRule":     true,
		"Status":               true,
		"LockedAt":             true,
		"DistributedAt":        true,
		"MVPPlayerID":          true,
		"EscrowEndTime":        true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
		"ResourceOwner":        true,
	}

	readableFields := map[string]bool{
		"ID":                   true,
		"MatchID":              true,
		"GameID":               true,
		"Region":               true,
		"Currency":             true,
		"TotalAmount":          true,
		"PlatformContribution": true,
		"PlayerContributions":  true,
		"DistributionRule":     true,
		"Status":               true,
		"LockedAt":             true,
		"DistributedAt":        true,
		"Winners":              true,
		"MVPPlayerID":          true,
		"EscrowEndTime":        true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
		"ResourceOwner":        true,
	}

	defaultSearchFields := []string{"MatchID", "GameID"}
	sortableFields := []string{"CreatedAt", "UpdatedAt", "TotalAmount", "EscrowEndTime"}
	filterableFields := []string{"Status", "GameID", "Region", "Currency", "DistributionRule"}

	return &PrizePoolQueryService{
		BaseQueryService: shared.BaseQueryService[matchmaking_entities.PrizePool]{
			Reader:              prizePoolReader,
			QueryableFields:     queryableFields,
			ReadableFields:      readableFields,
			DefaultSearchFields: defaultSearchFields,
			SortableFields:      sortableFields,
			FilterableFields:    filterableFields,
			MaxPageSize:         100,
			Audience:            shared.UserAudienceIDKey,
			EntityType:          "prize_pools",
		},
	}
}

// FindPendingDistributions finds prize pools that are ready for distribution
// Business rule: Prize pools in escrow status with expired escrow period
func (s *PrizePoolQueryService) FindPendingDistributions(ctx context.Context, limit int) ([]*matchmaking_entities.PrizePool, error) {
	search := NewPrizePoolSearch().
		WithPendingDistributions().
		WithLimit(limit).
		WithSortByEscrowEndTime(shared.AscendingIDKey).
		Build()

	entities, err := s.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	pools := make([]*matchmaking_entities.PrizePool, len(entities))
	for i := range entities {
		pools[i] = &entities[i]
	}

	return pools, nil
}

// FindByMatchID finds a prize pool by match ID
func (s *PrizePoolQueryService) FindByMatchID(ctx context.Context, matchID interface{}) (*matchmaking_entities.PrizePool, error) {
	search := NewPrizePoolSearch().
		WithMatchID(matchID).
		WithLimit(1).
		Build()

	entities, err := s.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, nil
	}

	return &entities[0], nil
}

// FindByStatus finds prize pools by status
func (s *PrizePoolQueryService) FindByStatus(ctx context.Context, status matchmaking_entities.PrizePoolStatus, limit int) ([]*matchmaking_entities.PrizePool, error) {
	search := NewPrizePoolSearch().
		WithStatus(status).
		WithLimit(limit).
		WithSortByCreatedAt(shared.DescendingIDKey).
		Build()

	entities, err := s.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	pools := make([]*matchmaking_entities.PrizePool, len(entities))
	for i := range entities {
		pools[i] = &entities[i]
	}

	return pools, nil
}