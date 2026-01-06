package challenge_usecases

import (
	"context"
	"testing"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_in "github.com/replay-api/replay-api/pkg/domain/challenge/ports/in"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockChallengeRepository is a mock implementation of ChallengeRepository
type MockChallengeRepository struct {
	mock.Mock
}

func (m *MockChallengeRepository) Save(ctx context.Context, challenge *challenge_entities.Challenge) error {
	args := m.Called(ctx, challenge)
	return args.Error(0)
}

func (m *MockChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*challenge_entities.Challenge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*challenge_entities.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID, search *shared.Search) ([]*challenge_entities.Challenge, error) {
	args := m.Called(ctx, matchID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challenge_entities.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) GetByChallengerID(ctx context.Context, challengerID uuid.UUID, search *shared.Search) ([]*challenge_entities.Challenge, error) {
	args := m.Called(ctx, challengerID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challenge_entities.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) Search(ctx context.Context, criteria challenge_out.ChallengeCriteria) ([]*challenge_entities.Challenge, int64, error) {
	args := m.Called(ctx, criteria)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*challenge_entities.Challenge), args.Get(1).(int64), args.Error(2)
}

func (m *MockChallengeRepository) GetPending(ctx context.Context, priority *challenge_entities.ChallengePriority, gameID *string, limit int) ([]*challenge_entities.Challenge, error) {
	args := m.Called(ctx, priority, gameID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challenge_entities.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) GetExpired(ctx context.Context, limit int) ([]*challenge_entities.Challenge, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*challenge_entities.Challenge), args.Error(1)
}

func (m *MockChallengeRepository) CountByStatus(ctx context.Context, matchID *uuid.UUID) (map[challenge_entities.ChallengeStatus]int64, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[challenge_entities.ChallengeStatus]int64), args.Error(1)
}

func (m *MockChallengeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChallengeRepository) DeleteByMatchID(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

func createAuthenticatedContext(userID uuid.UUID) context.Context {
	tenantID := uuid.New()
	clientID := uuid.New()
	groupID := uuid.New()
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, shared.ClientIDKey, clientID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	return ctx
}

func TestCreateChallengeUseCase_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	matchID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      matchID,
		ChallengerID: userID,
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "Game Bug Found",
		Description:  "Player clipped through wall at round 15",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	mockRepo.On("GetByMatchID", ctx, matchID, mock.Anything).Return([]*challenge_entities.Challenge{}, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*challenge_entities.Challenge")).Return(nil)

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, matchID, challenge.MatchID)
	assert.Equal(t, userID, challenge.ChallengerID)
	assert.Equal(t, "cs2", challenge.GameID)
	assert.Equal(t, challenge_entities.ChallengeTypeBugReport, challenge.Type)
	assert.Equal(t, challenge_entities.ChallengeStatusPending, challenge.Status)
	mockRepo.AssertExpectations(t)
}

func TestCreateChallengeUseCase_Unauthenticated_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	// Context without user ID but needs tenant to not panic
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), shared.TenantIDKey, tenantID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      uuid.New(),
		ChallengerID: uuid.New(),
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "Bug",
		Description:  "Description",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, challenge)
}

func TestCreateChallengeUseCase_UserMismatch_ReturnsForbidden(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	differentUserID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      uuid.New(),
		ChallengerID: differentUserID, // Different from authenticated user
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "Bug",
		Description:  "Description",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, challenge)
}

func TestCreateChallengeUseCase_MissingMatchID_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      uuid.Nil, // Missing
		ChallengerID: userID,
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "Bug",
		Description:  "Description",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "match ID is required")
}

func TestCreateChallengeUseCase_MissingTitle_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      uuid.New(),
		ChallengerID: userID,
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "", // Missing
		Description:  "Description",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "title is required")
}

func TestCreateChallengeUseCase_DuplicatePendingChallenge_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	matchID := uuid.New()
	ctx := createAuthenticatedContext(userID)

	// Existing pending challenge from same user
	existingChallenge := &challenge_entities.Challenge{
		ChallengerID: userID,
		Status:       challenge_entities.ChallengeStatusPending,
	}

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:      matchID,
		ChallengerID: userID,
		GameID:       "cs2",
		Type:         challenge_entities.ChallengeTypeBugReport,
		Title:        "Another Bug",
		Description:  "Another description",
		Priority:     challenge_entities.ChallengePriorityNormal,
	}

	mockRepo.On("GetByMatchID", ctx, matchID, mock.Anything).Return([]*challenge_entities.Challenge{existingChallenge}, nil)

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, challenge)
	assert.Contains(t, err.Error(), "already have a pending challenge")
	mockRepo.AssertExpectations(t)
}

func TestCreateChallengeUseCase_WithOptionalFields_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockChallengeRepository)
	useCase := NewCreateChallengeUseCase(mockRepo)

	userID := uuid.New()
	matchID := uuid.New()
	teamID := uuid.New()
	tournamentID := uuid.New()
	roundNumber := 15
	ctx := createAuthenticatedContext(userID)

	cmd := challenge_in.CreateChallengeCommand{
		MatchID:          matchID,
		ChallengerID:     userID,
		ChallengerTeamID: &teamID,
		TournamentID:     &tournamentID,
		RoundNumber:      &roundNumber,
		GameID:           "cs2",
		Type:             challenge_entities.ChallengeTypeVAR,
		Title:            "VAR Request",
		Description:      "Request video review of round 15",
		Priority:         challenge_entities.ChallengePriorityHigh,
	}

	mockRepo.On("GetByMatchID", ctx, matchID, mock.Anything).Return([]*challenge_entities.Challenge{}, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*challenge_entities.Challenge")).Return(nil)

	// Act
	challenge, err := useCase.Exec(ctx, cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, &teamID, challenge.ChallengerTeamID)
	assert.Equal(t, &tournamentID, challenge.TournamentID)
	assert.Equal(t, &roundNumber, challenge.RoundNumber)
	assert.Equal(t, challenge_entities.ChallengeTypeVAR, challenge.Type)
	assert.Equal(t, challenge_entities.ChallengePriorityHigh, challenge.Priority)
	mockRepo.AssertExpectations(t)
}
