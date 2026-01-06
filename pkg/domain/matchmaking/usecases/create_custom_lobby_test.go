package matchmaking_usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_usecases "github.com/replay-api/replay-api/pkg/domain/matchmaking/usecases"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCustomLobby_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	creatorID := uuid.New()
	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:        creatorID,
		GameID:           "cs2",
		Region:           "us-east",
		Tier:             string(matchmaking_entities.TierPremium),
		DistributionRule: matchmaking_vo.DistributionRuleWinnerTakesAll,
		MaxPlayers:       10,
		AutoFill:         true,
		InviteOnly:       false,
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock lobby save
	mockLobbyRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution - return proper BillableEntry
	mockEntry := &billing_entities.BillableEntry{}
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(mockEntry, nil, nil)

	lobby, err := usecase.CreateLobby(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, lobby)
	assert.Equal(t, creatorID, lobby.CreatorID)
	assert.Equal(t, "cs2", lobby.GameID)
	assert.Equal(t, 10, lobby.MaxPlayers)
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}

func TestCreateCustomLobby_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()

	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:  uuid.New(),
		GameID:     "cs2",
		MaxPlayers: 10,
	}

	lobby, err := usecase.CreateLobby(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, lobby)
}

func TestCreateCustomLobby_InvalidMaxPlayers(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	// test max players too low
	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:  uuid.New(),
		GameID:     "cs2",
		MaxPlayers: 1,
	}

	lobby, err := usecase.CreateLobby(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "max players must be between 2 and 10")

	// test max players too high
	cmd.MaxPlayers = 11

	lobby, err = usecase.CreateLobby(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "max players must be between 2 and 10")
}

func TestCreateCustomLobby_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:  uuid.New(),
		GameID:     "cs2",
		MaxPlayers: 10,
	}

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	lobby, err := usecase.CreateLobby(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	mockBilling.AssertExpectations(t)
}

func TestCreateCustomLobby_SaveFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:        uuid.New(),
		GameID:           "cs2",
		Region:           "us-east",
		Tier:             string(matchmaking_entities.TierFree),
		DistributionRule: matchmaking_vo.DistributionRuleWinnerTakesAll,
		MaxPlayers:       10,
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock lobby save failure
	mockLobbyRepo.On("Save", mock.Anything, mock.Anything).Return(assert.AnError)

	lobby, err := usecase.CreateLobby(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, lobby)
	assert.Contains(t, err.Error(), "failed to create lobby")
	mockBilling.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}
