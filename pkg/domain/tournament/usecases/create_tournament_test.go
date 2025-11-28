package tournament_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
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

func (m *MockTournamentRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID, filters tournament_out.TournamentFilters) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, organizerID, filters)
	return args.Get(0).([]*tournament_entities.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) FindActive(ctx context.Context, filters tournament_out.TournamentFilters) ([]*tournament_entities.Tournament, error) {
	args := m.Called(ctx, filters)
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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	organizerID := uuid.New()
	startTime := time.Now().UTC().Add(24 * time.Hour)
	registrationOpen := time.Now().UTC()
	registrationClose := startTime.Add(-1 * time.Hour)

	cmd := tournament_in.CreateTournamentCommand{
		ResourceOwner: common.ResourceOwner{
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
		EntryFee:         decimal.NewFromFloat(10.0),
		Currency:         "USD",
		StartTime:        startTime,
		RegistrationOpen: registrationOpen,
		RegistrationClose: registrationClose,
		Rules:            "Standard CS2 competitive rules",
		OrganizerID:      organizerID,
	}

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament save
	mockTournamentRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	cmd := tournament_in.CreateTournamentCommand{
		Name:            "Test Tournament",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
		EntryFee:        decimal.NewFromFloat(10.0),
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
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	startTime := time.Now().UTC().Add(24 * time.Hour)
	registrationOpen := time.Now().UTC()
	registrationClose := startTime.Add(-1 * time.Hour)

	cmd := tournament_in.CreateTournamentCommand{
		ResourceOwner: common.ResourceOwner{
			UserID:   userID,
			TenantID: uuid.New(),
		},
		Name:              "Test Tournament",
		Format:            tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:   16,
		MinParticipants:   8,
		EntryFee:          decimal.NewFromFloat(10.0),
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
