package matchmaking_usecases_test

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	"github.com/stretchr/testify/mock"
)

// MockBillableOperationHandler implements billing_in.BillableOperationCommandHandler
type MockBillableOperationHandler struct {
	mock.Mock
}

func (m *MockBillableOperationHandler) Exec(ctx context.Context, command billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	args := m.Called(ctx, command)
	var entry *billing_entities.BillableEntry
	var sub *billing_entities.Subscription
	if args.Get(0) != nil {
		entry = args.Get(0).(*billing_entities.BillableEntry)
	}
	if args.Get(1) != nil {
		sub = args.Get(1).(*billing_entities.Subscription)
	}
	return entry, sub, args.Error(2)
}

func (m *MockBillableOperationHandler) Validate(ctx context.Context, command billing_in.BillableOperationCommand) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

// MockLobbyRepository implements matchmaking_out.LobbyRepository
type MockLobbyRepository struct {
	mock.Mock
}

func (m *MockLobbyRepository) Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	args := m.Called(ctx, lobby)
	return args.Error(0)
}

func (m *MockLobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.MatchmakingLobby), args.Error(1)
}

func (m *MockLobbyRepository) FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error) {
	args := m.Called(ctx, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*matchmaking_entities.MatchmakingLobby), args.Error(1)
}

func (m *MockLobbyRepository) FindOpenLobbies(ctx context.Context, gameID, region, tier string, limit int) ([]*matchmaking_entities.MatchmakingLobby, error) {
	args := m.Called(ctx, gameID, region, tier, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*matchmaking_entities.MatchmakingLobby), args.Error(1)
}

func (m *MockLobbyRepository) Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	args := m.Called(ctx, lobby)
	return args.Error(0)
}

func (m *MockLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockMatchmakingSessionRepository implements matchmaking_out.MatchmakingSessionRepository
type MockMatchmakingSessionRepository struct {
	mock.Mock
}

func (m *MockMatchmakingSessionRepository) Save(ctx context.Context, session *matchmaking_entities.MatchmakingSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockMatchmakingSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.MatchmakingSession), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return []*matchmaking_entities.MatchmakingSession{}, args.Error(1)
	}
	return args.Get(0).([]*matchmaking_entities.MatchmakingSession), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) GetActiveSessions(ctx context.Context, filters matchmaking_out.SessionFilters) ([]*matchmaking_entities.MatchmakingSession, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return []*matchmaking_entities.MatchmakingSession{}, args.Error(1)
	}
	return args.Get(0).([]*matchmaking_entities.MatchmakingSession), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status matchmaking_entities.SessionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockMatchmakingSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchmakingSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) Search(ctx context.Context, s shared.Search) ([]matchmaking_entities.MatchmakingSession, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return []matchmaking_entities.MatchmakingSession{}, args.Error(1)
	}
	return args.Get(0).([]matchmaking_entities.MatchmakingSession), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}

// MockMatchmakingPoolRepository implements matchmaking_out.MatchmakingPoolRepository
type MockMatchmakingPoolRepository struct {
	mock.Mock
}

func (m *MockMatchmakingPoolRepository) Save(ctx context.Context, pool *matchmaking_entities.MatchmakingPool) error {
	args := m.Called(ctx, pool)
	return args.Error(0)
}

func (m *MockMatchmakingPoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingPool, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.MatchmakingPool), args.Error(1)
}

func (m *MockMatchmakingPoolRepository) GetByGameModeRegion(ctx context.Context, gameID, gameMode, region string) (*matchmaking_entities.MatchmakingPool, error) {
	args := m.Called(ctx, gameID, gameMode, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.MatchmakingPool), args.Error(1)
}

func (m *MockMatchmakingPoolRepository) UpdateStats(ctx context.Context, poolID uuid.UUID, stats matchmaking_entities.PoolStatistics) error {
	args := m.Called(ctx, poolID, stats)
	return args.Error(0)
}

func (m *MockMatchmakingPoolRepository) GetAllActive(ctx context.Context) ([]*matchmaking_entities.MatchmakingPool, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []*matchmaking_entities.MatchmakingPool{}, args.Error(1)
	}
	return args.Get(0).([]*matchmaking_entities.MatchmakingPool), args.Error(1)
}

func (m *MockMatchmakingPoolRepository) Search(ctx context.Context, s shared.Search) ([]matchmaking_entities.MatchmakingPool, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return []matchmaking_entities.MatchmakingPool{}, args.Error(1)
	}
	return args.Get(0).([]matchmaking_entities.MatchmakingPool), args.Error(1)
}

func (m *MockMatchmakingPoolRepository) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	args := m.Called(ctx, searchParams, resultOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.Search), args.Error(1)
}
