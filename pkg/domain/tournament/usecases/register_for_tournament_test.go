package tournament_usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createRegistrationTournament(status tournament_entities.TournamentStatus) *tournament_entities.Tournament {
	startTime := time.Now().UTC().Add(24 * time.Hour)
	registrationOpen := time.Now().UTC().Add(-1 * time.Hour)
	registrationClose := startTime.Add(-1 * time.Hour)

	resourceOwner := common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
	}

	return &tournament_entities.Tournament{
		BaseEntity:        common.NewUnrestrictedEntity(resourceOwner),
		Name:              "Test Tournament",
		Format:            tournament_entities.TournamentFormatSingleElimination,
		MaxParticipants:   16,
		MinParticipants:   8,
		EntryFee:          wallet_vo.NewAmount(10.0),
		Currency:          "USD",
		StartTime:         startTime,
		RegistrationOpen:  registrationOpen,
		RegistrationClose: registrationClose,
		Status:            status,
		Participants:      []tournament_entities.TournamentPlayer{},
		OrganizerID:       uuid.New(),
	}
}

func TestRegisterForTournament_Success(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewRegisterForTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createRegistrationTournament(tournament_entities.TournamentStatusRegistration)
	tournamentID := tournament.ID
	playerID := uuid.New()

	cmd := tournament_in.RegisterPlayerCommand{
		TournamentID: tournamentID,
		PlayerID:     playerID,
		DisplayName:  "Player123",
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// mock billing execution
	mockBilling.On("Exec", mock.Anything, mock.Anything).Return(nil, nil, nil)

	err := usecase.Exec(ctx, cmd)

	assert.NoError(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestRegisterForTournament_Unauthenticated(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewRegisterForTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()

	cmd := tournament_in.RegisterPlayerCommand{
		TournamentID: uuid.New(),
		PlayerID:     uuid.New(),
		DisplayName:  "Player123",
	}

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
}

func TestRegisterForTournament_TournamentNotFound(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewRegisterForTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournamentID := uuid.New()

	cmd := tournament_in.RegisterPlayerCommand{
		TournamentID: tournamentID,
		PlayerID:     uuid.New(),
		DisplayName:  "Player123",
	}

	// mock tournament not found
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(nil, assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tournament not found")
	mockTournamentRepo.AssertExpectations(t)
}

func TestRegisterForTournament_BillingValidationFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewRegisterForTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createRegistrationTournament(tournament_entities.TournamentStatusRegistration)
	tournamentID := tournament.ID

	cmd := tournament_in.RegisterPlayerCommand{
		TournamentID: tournamentID,
		PlayerID:     uuid.New(),
		DisplayName:  "Player123",
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation failure
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}

func TestRegisterForTournament_UpdateFails(t *testing.T) {
	mockBilling := new(MockBillableOperationHandler)
	mockTournamentRepo := new(MockTournamentRepository)

	usecase := tournament_usecases.NewRegisterForTournamentUseCase(
		mockBilling,
		mockTournamentRepo,
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, common.AuthenticatedKey, true)
	userID := uuid.New()
	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())

	tournament := createRegistrationTournament(tournament_entities.TournamentStatusRegistration)
	tournamentID := tournament.ID

	cmd := tournament_in.RegisterPlayerCommand{
		TournamentID: tournamentID,
		PlayerID:     uuid.New(),
		DisplayName:  "Player123",
	}

	// mock tournament retrieval
	mockTournamentRepo.On("FindByID", mock.Anything, tournamentID).Return(tournament, nil)

	// mock billing validation
	mockBilling.On("Validate", mock.Anything, mock.Anything).Return(nil)

	// mock tournament update failure
	mockTournamentRepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	err := usecase.Exec(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register for tournament")
	mockBilling.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}
