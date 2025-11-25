package matchmaking_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBillableOperationHandler struct {
	mock.Mock
}

func (m *MockBillableOperationHandler) Validate(ctx context.Context, cmd billing_in.BillableOperationCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBillableOperationHandler) Exec(ctx context.Context, cmd billing_in.BillableOperationCommand) (*uuid.UUID, *float64, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*uuid.UUID), args.Get(1).(*float64), args.Error(2)
}

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
	return args.Get(0).([]*matchmaking_entities.MatchmakingSession), args.Error(1)
}

func (m *MockMatchmakingSessionRepository) GetActiveSessions(ctx context.Context, filters interface{}) ([]*matchmaking_entities.MatchmakingSession, error) {
	args := m.Called(ctx, filters)
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
	return args.Get(0).([]*matchmaking_entities.MatchmakingPool), args.Error(1)
}

func TestJoinMatchmakingQueue_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)
	mockPoolRepo := new(MockMatchmakingPoolRepository)

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		mockPoolRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	playerID := uuid.New()
	cmd := matchmaking_in.JoinMatchmakingQueueCommand{
		PlayerID:   playerID,
		GameID:     "cs2",
		GameMode:   "competitive",
		Region:     "us-east",
		Tier:       matchmaking_entities.TierFree,
		PlayerMMR:  1500,
		TeamFormat: matchmaking_in.TeamFormat5v5,
		MaxPing:    50,
	}

	// mock no existing sessions
	mockSessionRepo.On("GetByPlayerID", mock.Anything, playerID).Return([]*matchmaking_entities.MatchmakingSession{}, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock pool retrieval - return existing pool
	pool := &matchmaking_entities.MatchmakingPool{
		ID:       uuid.New(),
		GameID:   "cs2",
		GameMode: "competitive",
		Region:   "us-east",
		PoolStats: matchmaking_entities.PoolStatistics{
			TotalPlayers:    10,
			AverageWaitTime: 45,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mockPoolRepo.On("GetByGameModeRegion", mock.Anything, "cs2", "competitive", "us-east").Return(pool, nil)

	// mock session save
	mockSessionRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

	session, err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, matchmaking_entities.StatusQueued, session.Status)
	mockBilling.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockPoolRepo.AssertExpectations(t)
}

func TestJoinMatchmakingQueue_Unauthenticated(t *testing.T){
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)
	mockPoolRepo := new(MockMatchmakingPoolRepository)

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		mockPoolRepo,
	)

	ctx := context.Background()

	cmd := matchmaking_in.JoinMatchmakingQueueCommand{
		PlayerID:   uuid.New(),
		GameID:     "cs2",
		GameMode:   "competitive",
		Region:     "us-east",
		Tier:       matchmaking_entities.TierFree,
		PlayerMMR:  1500,
		TeamFormat: matchmaking_in.TeamFormat5v5,
	}

	session, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, session)
}

func TestJoinMatchmakingQueue_AlreadyInQueue(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)
	mockPoolRepo := new(MockMatchmakingPoolRepository)

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		mockPoolRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	playerID := uuid.New()
	cmd := matchmaking_in.JoinMatchmakingQueueCommand{
		PlayerID:   playerID,
		GameID:     "cs2",
		GameMode:   "competitive",
		Region:     "us-east",
		Tier:       matchmaking_entities.TierFree,
		PlayerMMR:  1500,
		TeamFormat: matchmaking_in.TeamFormat5v5,
	}

	// mock existing session
	existingSession := &matchmaking_entities.MatchmakingSession{
		ID:       uuid.New(),
		PlayerID: playerID,
		Status:   matchmaking_entities.StatusQueued,
	}
	mockSessionRepo.On("GetByPlayerID", mock.Anything, playerID).Return([]*matchmaking_entities.MatchmakingSession{existingSession}, nil)

	session, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "already in queue")
}
