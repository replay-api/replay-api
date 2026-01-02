package matchmaking_usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/replay-api/replay-api/pkg/domain/matchmaking/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJoinMatchmakingQueue_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		nil, // eventPublisher - not needed for this test
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

	// mock session save
	mockSessionRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution - return proper BillableEntry
	mockEntry := &billing_entities.BillableEntry{}
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(mockEntry, nil, nil)

	session, err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, matchmaking_entities.StatusQueued, session.Status)
	mockBilling.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestJoinMatchmakingQueue_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		nil, // eventPublisher
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

	usecase := matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
		nil, // eventPublisher
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
	resourceOwner := common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		UserID:   playerID,
	}
	existingSession := &matchmaking_entities.MatchmakingSession{
		BaseEntity: common.NewEntity(resourceOwner),
		PlayerID:   playerID,
		Status:     matchmaking_entities.StatusQueued,
	}
	mockSessionRepo.On("GetByPlayerID", mock.Anything, playerID).Return([]*matchmaking_entities.MatchmakingSession{existingSession}, nil)

	session, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "already in queue")
}
