package matchmaking_usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaveMatchmakingQueue_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	sessionID := uuid.New()
	playerID := uuid.New()
	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: sessionID,
		PlayerID:  playerID,
	}

	// mock session retrieval
	session := &matchmaking_entities.MatchmakingSession{
		ID:       sessionID,
		PlayerID: playerID,
		Status:   matchmaking_entities.StatusQueued,
	}
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock status update
	mockSessionRepo.On("UpdateStatus", mock.Anything, sessionID, matchmaking_entities.StatusCancelled).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

	err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestLeaveMatchmakingQueue_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()

	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: uuid.New(),
		PlayerID:  uuid.New(),
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
}

func TestLeaveMatchmakingQueue_SessionNotFound(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	sessionID := uuid.New()
	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: sessionID,
		PlayerID:  uuid.New(),
	}

	// mock session not found
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	mockSessionRepo.AssertExpectations(t)
}

func TestLeaveMatchmakingQueue_PlayerDoesNotOwnSession(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	sessionID := uuid.New()
	ownerID := uuid.New()
	differentPlayerID := uuid.New()

	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: sessionID,
		PlayerID:  differentPlayerID,
	}

	// mock session owned by different player
	session := &matchmaking_entities.MatchmakingSession{
		ID:       sessionID,
		PlayerID: ownerID, // different player
		Status:   matchmaking_entities.StatusQueued,
	}
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not own")
	mockSessionRepo.AssertExpectations(t)
}

func TestLeaveMatchmakingQueue_CannotLeaveFromStatus(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	sessionID := uuid.New()
	playerID := uuid.New()

	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: sessionID,
		PlayerID:  playerID,
	}

	// mock session already completed
	session := &matchmaking_entities.MatchmakingSession{
		ID:       sessionID,
		PlayerID: playerID,
		Status:   matchmaking_entities.StatusCompleted,
	}
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot leave queue from status")
	mockSessionRepo.AssertExpectations(t)
}

func TestLeaveMatchmakingQueue_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockSessionRepo := new(MockMatchmakingSessionRepository)

	usecase := matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(
		mockBilling,
		mockSessionRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	sessionID := uuid.New()
	playerID := uuid.New()

	cmd := matchmaking_in.LeaveMatchmakingQueueCommand{
		SessionID: sessionID,
		PlayerID:  playerID,
	}

	// mock session retrieval
	session := &matchmaking_entities.MatchmakingSession{
		ID:       sessionID,
		PlayerID: playerID,
		Status:   matchmaking_entities.StatusQueued,
	}
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}
