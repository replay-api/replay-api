package tournament_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	common "github.com/replay-api/replay-api/pkg/domain"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	tournamentID := uuid.New()

	// create tournament with 8 participants
	participants := make([]tournament_entities.TournamentParticipant, 8)
	for i := 0; i < 8; i++ {
		participants[i] = tournament_entities.TournamentParticipant{
			PlayerID:    uuid.New(),
			DisplayName: "Player" + string(rune(i)),
			Seed:        i + 1,
			RegisteredAt: time.Now().UTC(),
		}
	}

	startTime := time.Now().UTC().Add(1 * time.Hour)
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Name:            "Test Tournament",
		Format:          tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants: 16,
		MinParticipants: 8,
		Status:          tournament_entities.TournamentStatusReady,
		Participants:    participants,
		Matches:         []tournament_entities.TournamentMatch{},
		StartTime:       startTime,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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

	tournamentID := uuid.New()

	// create tournament with 8 participants
	participants := make([]tournament_entities.TournamentParticipant, 8)
	for i := 0; i < 8; i++ {
		participants[i] = tournament_entities.TournamentParticipant{
			PlayerID:     uuid.New(),
			DisplayName:  "Player" + string(rune(i)),
			Seed:         i + 1,
			RegisteredAt: time.Now().UTC(),
		}
	}

	startTime := time.Now().UTC().Add(1 * time.Hour)
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Format:          tournament_entities.TournamentFormatDoubleElimination,
		MaxParticipants: 16,
		MinParticipants: 8,
		Status:          tournament_entities.TournamentStatusReady,
		Participants:    participants,
		StartTime:       startTime,
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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

	tournamentID := uuid.New()

	// create tournament with 6 participants (for round robin)
	participants := make([]tournament_entities.TournamentParticipant, 6)
	for i := 0; i < 6; i++ {
		participants[i] = tournament_entities.TournamentParticipant{
			PlayerID:     uuid.New(),
			DisplayName:  "Player" + string(rune(i)),
			Seed:         i + 1,
			RegisteredAt: time.Now().UTC(),
		}
	}

	startTime := time.Now().UTC().Add(1 * time.Hour)
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Format:          tournament_entities.TournamentFormatRoundRobin,
		MaxParticipants: 16,
		MinParticipants: 4,
		Status:          tournament_entities.TournamentStatusReady,
		Participants:    participants,
		StartTime:       startTime,
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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

	tournamentID := uuid.New()

	// create tournament with 8 participants
	participants := make([]tournament_entities.TournamentParticipant, 8)
	for i := 0; i < 8; i++ {
		participants[i] = tournament_entities.TournamentParticipant{
			PlayerID:     uuid.New(),
			DisplayName:  "Player" + string(rune(i)),
			Seed:         i + 1,
			RegisteredAt: time.Now().UTC(),
		}
	}

	startTime := time.Now().UTC().Add(1 * time.Hour)
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Format:          tournament_entities.TournamentFormatSwiss,
		MaxParticipants: 16,
		MinParticipants: 8,
		Status:          tournament_entities.TournamentStatusReady,
		Participants:    participants,
		StartTime:       startTime,
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	entryID := uuid.New()
	amount := 1.0
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(&entryID, &amount, nil)

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

	tournamentID := uuid.New()

	// create tournament in wrong status
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Status:          tournament_entities.TournamentStatusRegistering, // wrong status
		MinParticipants: 8,
		Participants: []tournament_entities.TournamentParticipant{
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
		},
	}

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

	tournamentID := uuid.New()

	// create tournament with not enough participants
	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Status:          tournament_entities.TournamentStatusReady,
		MinParticipants: 8,
		Participants: []tournament_entities.TournamentParticipant{
			{PlayerID: uuid.New()},
			{PlayerID: uuid.New()},
		}, // only 2, need 8
	}

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

	tournamentID := uuid.New()

	// create valid tournament
	participants := make([]tournament_entities.TournamentParticipant, 8)
	for i := 0; i < 8; i++ {
		participants[i] = tournament_entities.TournamentParticipant{
			PlayerID: uuid.New(),
			Seed:     i + 1,
		}
	}

	tournament := &tournament_entities.Tournament{
		ID:              tournamentID,
		Format:          tournament_entities.TournamentFormatSingleElimination,
		Status:          tournament_entities.TournamentStatusReady,
		MinParticipants: 8,
		Participants:    participants,
		StartTime:       time.Now().UTC().Add(1 * time.Hour),
		EntryFee:        decimal.NewFromFloat(10.0),
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, tournamentID)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}
