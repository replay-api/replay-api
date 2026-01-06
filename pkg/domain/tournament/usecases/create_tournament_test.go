package tournament_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
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

func (m *MockBillableOperationHandler) Exec(ctx context.Context, cmd billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	args := m.Called(ctx, cmd)
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

type MockTournamentRepository struct {
	mock.Mock
}

func (m *MockTournamentRepository) Save(ctx context.Context, tournament *tournament_entities.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) Update(ctx context.Context, tournament *tournament_entities.Tournament) error {
	args := m.Called(ctx, tournament)
	return args.Error(0)
}

func (m *MockTournamentRepository) FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, organizerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindByGameAndRegion(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, region, status, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, gameID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, playerID, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateTournament_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewCreateTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	organizerID := uuid.New()
	startTime := time.Now().UTC().Add(24 * time.Hour)
	registrationOpen := time.Now().UTC()
	registrationClose := startTime.Add(-1 * time.Hour)

	cmd := tournament_in.CreateTournamentCommand{
		ResourceOwner: shared.ResourceOwner{
			UserID:   userID,
			TenantID: uuid.New(),
		},
		Name:             "CS2 Championship",
		Description:      "Competitive CS2 tournament",
		GameID:           "cs2",
		GameMode:         "competitive",
		Region:           "us-east",
		Format:           tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:  16,
		MinParticipants:  8,
		EntryFee:         wallet_vo.NewAmount(10.0),
		Currency:         "USD",
		StartTime:        startTime,
		RegistrationOpen: registrationOpen,
		RegistrationClose: registrationClose,
		Rules: tournament_entities.TournamentRules{
			BestOf:          3,
			CheckInRequired: true,
		},
		OrganizerID: organizerID,
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament save
	mockTournamentRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	tournament, err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, tournament)
	assert.Equal(t, "CS2 Championship", tournament.Name)
	assert.Equal(t, tournament_entities.TournamentFormatSingleElimination, tournament.Format)
	assert.Equal(t, 16, tournament.MaxParticipants)
	assert.Equal(t, organizerID, tournament.OrganizerID)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestCreateTournament_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewCreateTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()

	cmd := tournament_in.CreateTournamentCommand{
		Name:            "Test Tournament",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
	}

	tournament, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, tournament)
}

func TestCreateTournament_InvalidFormat(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewCreateTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	cmd := tournament_in.CreateTournamentCommand{
		Name:            "Test Tournament",
		Format:          "invalid-format",
		MaxParticipants: 16,
	}

	tournament, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, tournament)
	assert.Contains(t, err.Error(), "invalid tournament format")
}

func TestCreateTournament_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewCreateTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	cmd := tournament_in.CreateTournamentCommand{
		Name:            "Test Tournament",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
		EntryFee:        wallet_vo.NewAmount(10.0),
	}

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	tournament, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, tournament)
	mockBilling.AssertExpectations(t)
}

func TestCreateTournament_SaveFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewCreateTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.TenantIDKey, uuid.New())

	startTime := time.Now().UTC().Add(24 * time.Hour)
	registrationOpen := time.Now().UTC()
	registrationClose := startTime.Add(-1 * time.Hour)

	cmd := tournament_in.CreateTournamentCommand{
		ResourceOwner: shared.ResourceOwner{
			UserID:   userID,
			TenantID: uuid.New(),
		},
		Name:              "Test Tournament",
		GameID:            "cs2",
		GameMode:          "competitive",
		Region:            "us-east",
		Format:            tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:   16,
		MinParticipants:   8,
		EntryFee:          wallet_vo.NewAmount(10.0),
		Currency:          "USD",
		StartTime:         startTime,
		RegistrationOpen:  registrationOpen,
		RegistrationClose: registrationClose,
		OrganizerID:       uuid.New(),
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament save failure
	mockTournamentRepo.On("Save", mock.Anything, mock.Anything).Return(assert.AnError)

	tournament, err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, tournament)
	assert.Contains(t, err.Error(), "failed to create tournament")
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}
