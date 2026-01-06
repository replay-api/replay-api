package matchmaking_services

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	shared "github.com/resource-ownership/go-common/pkg/common"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// StubPrizePoolSearchable is a simple stub implementation for testing
type StubPrizePoolSearchable struct {
	pools []matchmaking_entities.PrizePool
}

func (s *StubPrizePoolSearchable) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PrizePool, error) {
	for _, pool := range s.pools {
		if pool.ID == id {
			return &pool, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (s *StubPrizePoolSearchable) Search(ctx context.Context, search shared.Search) ([]matchmaking_entities.PrizePool, error) {
	return s.pools, nil
}

func (s *StubPrizePoolSearchable) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	return &shared.Search{}, nil
}

func TestPrizePoolQueryService_NewPrizePoolQueryService(t *testing.T) {
	stubSearchable := &StubPrizePoolSearchable{}
	service := NewPrizePoolQueryService(stubSearchable)

	assert.NotNil(t, service)
	assert.Equal(t, "PrizePool", service.GetName())
	assert.Equal(t, "prize_pools", service.EntityType)
	assert.NotNil(t, service.QueryableFields)
	assert.NotNil(t, service.ReadableFields)
}

func TestPrizePoolQueryService_FindPendingDistributions(t *testing.T) {
	stubSearchable := &StubPrizePoolSearchable{
		pools: []matchmaking_entities.PrizePool{createTestPrizePool()},
	}
	service := NewPrizePoolQueryService(stubSearchable)

	// Set up context with resource owner
	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
	limit := 10

	result, err := service.FindPendingDistributions(ctx, limit)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPrizePoolQueryService_FindByMatchID(t *testing.T) {
	stubSearchable := &StubPrizePoolSearchable{
		pools: []matchmaking_entities.PrizePool{createTestPrizePool()},
	}
	service := NewPrizePoolQueryService(stubSearchable)

	// Set up context with resource owner
	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
	matchID := uuid.New()

	result, err := service.FindByMatchID(ctx, matchID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPrizePoolQueryService_FindByStatus(t *testing.T) {
	stubSearchable := &StubPrizePoolSearchable{
		pools: []matchmaking_entities.PrizePool{createTestPrizePool()},
	}
	service := NewPrizePoolQueryService(stubSearchable)

	// Set up context with resource owner
	ctx := context.WithValue(context.Background(), shared.TenantIDKey, replay_common.TeamPROTenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
	ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())
	status := matchmaking_entities.PrizePoolStatusInEscrow
	limit := 5

	result, err := service.FindByStatus(ctx, status, limit)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// Helper function to create a test prize pool
func createTestPrizePool() matchmaking_entities.PrizePool {
	resourceOwner := shared.NewResourceOwner(uuid.New(), uuid.New(), uuid.New(), uuid.New())
	matchID := uuid.New()
	gameID := replay_common.GameIDKey("test-game")
	region := "us-east"
	currency := wallet_vo.CurrencyUSD
	distributionRule := matchmaking_vo.DistributionRuleWinnerTakesAll

	return matchmaking_entities.PrizePool{
		BaseEntity:           shared.NewEntity(resourceOwner),
		MatchID:              matchID,
		GameID:               gameID,
		Region:               region,
		Currency:             currency,
		TotalAmount:          wallet_vo.NewAmount(100),
		PlatformContribution: wallet_vo.NewAmount(10),
		PlayerContributions:  make(map[uuid.UUID]wallet_vo.Amount),
		DistributionRule:     distributionRule,
		Status:               matchmaking_entities.PrizePoolStatusInEscrow,
	}
}