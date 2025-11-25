package matchmaking_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLeaveLobby_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewLeaveLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	lobbyID := uuid.New()
	playerID := uuid.New()
	cmd := matchmaking_in.LeaveLobbyCommand{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	}

	// create lobby with the player
	lobby := &matchmaking_entities.MatchmakingLobby{
		ID:               lobbyID,
		CreatorID:        uuid.New(),
		GameID:           "cs2",
		Region:           "us-east",
		Tier:             matchmaking_entities.TierFree,
		DistributionRule: matchmaking_entities.DistributionRuleRandom,
		MaxPlayers:       10,
		AutoFill:         true,
		InviteOnly:       false,
		Players: []matchmaking_entities.LobbyPlayer{
			{
				PlayerID: playerID,
				MMR:      1500,
				IsReady:  false,
				JoinedAt: time.Now().UTC(),
			},
		},
		Status:    matchmaking_entities.LobbyStatusWaiting,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// mock lobby retrieval
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(lobby, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock lobby update
	mockLobbyRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

	err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}

func TestLeaveLobby_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewLeaveLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()

	cmd := matchmaking_in.LeaveLobbyCommand{
		LobbyID:  uuid.New(),
		PlayerID: uuid.New(),
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
}

func TestLeaveLobby_LobbyNotFound(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewLeaveLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	lobbyID := uuid.New()
	cmd := matchmaking_in.LeaveLobbyCommand{
		LobbyID:  lobbyID,
		PlayerID: uuid.New(),
	}

	// mock lobby not found
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(nil, assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby not found")
	mockLobbyRepo.AssertExpectations(t)
}

func TestLeaveLobby_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewLeaveLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	lobbyID := uuid.New()
	playerID := uuid.New()
	cmd := matchmaking_in.LeaveLobbyCommand{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	}

	// create lobby with the player
	lobby := &matchmaking_entities.MatchmakingLobby{
		ID:         lobbyID,
		MaxPlayers: 10,
		Players: []matchmaking_entities.LobbyPlayer{
			{
				PlayerID: playerID,
				MMR:      1500,
			},
		},
		Status: matchmaking_entities.LobbyStatusWaiting,
	}

	// mock lobby retrieval
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(lobby, nil)

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}

func TestLeaveLobby_PlayerNotInLobby(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewLeaveLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	lobbyID := uuid.New()
	playerID := uuid.New()
	cmd := matchmaking_in.LeaveLobbyCommand{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	}

	// create lobby without the player
	lobby := &matchmaking_entities.MatchmakingLobby{
		ID:         lobbyID,
		MaxPlayers: 10,
		Players:    []matchmaking_entities.LobbyPlayer{}, // empty
		Status:     matchmaking_entities.LobbyStatusWaiting,
	}

	// mock lobby retrieval
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(lobby, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}
