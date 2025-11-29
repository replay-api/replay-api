package matchmaking_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/replay-api/replay-api/pkg/domain/matchmaking/usecases"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
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

	// create lobby with the player using proper entity structure
	resourceOwner := common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
	lobby, _ := matchmaking_entities.NewMatchmakingLobby(
		resourceOwner,
		uuid.New(), // creator
		"cs2",
		"us-east",
		string(matchmaking_entities.TierFree),
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		10,
		true,
		false,
	)
	lobby.ID = lobbyID
	// Add the player to the lobby
	_ = lobby.AddPlayer(playerID, 1500)
	lobby.CreatedAt = time.Now().UTC()
	lobby.UpdatedAt = time.Now().UTC()

	// mock lobby retrieval
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(lobby, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock lobby update
	mockLobbyRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

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
	resourceOwner := common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
	lobby, _ := matchmaking_entities.NewMatchmakingLobby(
		resourceOwner,
		uuid.New(), // creator
		"cs2",
		"us-east",
		string(matchmaking_entities.TierFree),
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		10,
		true,
		false,
	)
	lobby.ID = lobbyID
	_ = lobby.AddPlayer(playerID, 1500)

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

	// create lobby without the player (only creator is there)
	resourceOwner := common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
	lobby, _ := matchmaking_entities.NewMatchmakingLobby(
		resourceOwner,
		uuid.New(), // creator (different from playerID)
		"cs2",
		"us-east",
		string(matchmaking_entities.TierFree),
		matchmaking_vo.DistributionRuleWinnerTakesAll,
		10,
		true,
		false,
	)
	lobby.ID = lobbyID

	// mock lobby retrieval
	mockLobbyRepo.On("FindByID", mock.Anything, lobbyID).Return(lobby, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}
