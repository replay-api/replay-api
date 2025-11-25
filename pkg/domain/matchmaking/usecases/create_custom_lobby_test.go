package matchmaking_usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/out"
	matchmaking_usecases "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockLobbyRepository) Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	args := m.Called(ctx, lobby)
	return args.Error(0)
}

func (m *MockLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLobbyRepository) FindActiveLobbies(ctx context.Context, filters matchmaking_out.LobbyFilters) ([]*matchmaking_entities.MatchmakingLobby, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*matchmaking_entities.MatchmakingLobby), args.Error(1)
}

func TestCreateCustomLobby_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockLobbyRepo := new(MockLobbyRepository)

	usecase := matchmaking_usecases.NewCreateCustomLobbyUseCase(
		mockBilling,
		mockLobbyRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	creatorID := uuid.New()
	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:        creatorID,
		GameID:           "cs2",
		Region:           "us-east",
		Tier:             matchmaking_entities.TierPremium,
		DistributionRule: matchmaking_entities.DistributionRuleRandom,
		MaxPlayers:       10,
		AutoFill:         true,
		InviteOnly:       false,
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock lobby save
	mockLobbyRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	cmd := matchmaking_in.CreateLobbyCommand{
		CreatorID:  uuid.New(),
		GameID:     "cs2",
		MaxPlayers: 10,
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
