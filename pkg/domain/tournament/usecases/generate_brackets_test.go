package tournament_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTournamentWithParticipants(count int, format tournament_entities.TournamentFormat, status tournament_entities.TournamentStatus) *tournament_entities.Tournament {
	participants := make([]tournament_entities.TournamentPlayer, count)
	for i := 0; i < count; i++ {
		participants[i] = tournament_entities.TournamentPlayer{
			PlayerID:     uuid.New(),
			DisplayName:  "Player" + string(rune('A'+i)),
			Seed:         i + 1,
			RegisteredAt: time.Now().UTC(),
			Status:       "registered",
		}
	}

	startTime := time.Now().UTC().Add(1 * time.Hour)
	resourceOwner := common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
	}

	return &tournament_entities.Tournament{
		BaseEntity:        common.NewUnrestrictedEntity(resourceOwner),
		Name:              "Test Tournament",
		Format:            format,
		MaxParticipants:   16,
		MinParticipants:   count,
		Status:            status,
		Participants:      participants,
		Matches:           []tournament_entities.TournamentMatch{},
		StartTime:         startTime,
		RegistrationOpen:  time.Now().UTC().Add(-24 * time.Hour),
		RegistrationClose: startTime.Add(-1 * time.Hour),
		EntryFee:          wallet_vo.NewAmount(10.0),
		Currency:          "USD",
		OrganizerID:       uuid.New(),
	}
}

func TestGenerateBrackets_Success_SingleElimination(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createTournamentWithParticipants(8, tournament_entities.TournamentFormatSingleElimination, tournament_entities.TournamentStatusReady)
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_Success_DoubleElimination(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createTournamentWithParticipants(8, tournament_entities.TournamentFormatDoubleElimination, tournament_entities.TournamentStatusReady)
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_Success_RoundRobin(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createTournamentWithParticipants(6, tournament_entities.TournamentFormatRoundRobin, tournament_entities.TournamentStatusReady)
	tournament.MinParticipants = 4
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_Success_Swiss(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createTournamentWithParticipants(8, tournament_entities.TournamentFormatSwiss, tournament_entities.TournamentStatusReady)
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()

	err := usecase.Exec(ctx, uuid.New())

	assert.Error(t, err)
}

func TestGenerateBrackets_TournamentNotFound(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournamentID := uuid.New()

	// mock tournament not found
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(nil, assert.AnError)

	err := usecase.Exec(ctx, tournamentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tournament not found")
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_WrongStatus(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	// create tournament in wrong status (registration instead of ready)
	tournament := createTournamentWithParticipants(8, tournament_entities.TournamentFormatSingleElimination, tournament_entities.TournamentStatusRegistration)
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be in ready status")
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_NotEnoughParticipants(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	// create tournament with only 2 participants but needs 8
	tournament := createTournamentWithParticipants(2, tournament_entities.TournamentFormatSingleElimination, tournament_entities.TournamentStatusReady)
	tournament.MinParticipants = 8
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	err := usecase.Exec(ctx, tournamentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough participants")
	mockTournamentRepo.AssertExpectations(t)
}

func TestGenerateBrackets_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewGenerateBracketsUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createTournamentWithParticipants(8, tournament_entities.TournamentFormatSingleElimination, tournament_entities.TournamentStatusReady)
	tournamentID := tournament.ID

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, tournamentID)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}
