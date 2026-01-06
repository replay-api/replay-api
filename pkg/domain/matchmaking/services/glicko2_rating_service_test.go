package matchmaking_services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRatingRepository is a mock implementation for testing
type MockPlayerRatingRepository struct {
	mock.Mock
	ratings map[string]*matchmaking_entities.PlayerRating
}

func NewMockPlayerRatingRepository() *MockPlayerRatingRepository {
	return &MockPlayerRatingRepository{
		ratings: make(map[string]*matchmaking_entities.PlayerRating),
	}
}

func (m *MockPlayerRatingRepository) Save(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	args := m.Called(ctx, rating)
	key := rating.PlayerID.String() + ":" + string(rating.GameID)
	m.ratings[key] = rating
	return args.Error(0)
}

func (m *MockPlayerRatingRepository) Update(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	args := m.Called(ctx, rating)
	key := rating.PlayerID.String() + ":" + string(rating.GameID)
	m.ratings[key] = rating
	return args.Error(0)
}

func (m *MockPlayerRatingRepository) FindByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID replay_common.GameIDKey) (*matchmaking_entities.PlayerRating, error) {
	args := m.Called(ctx, playerID, gameID)
	key := playerID.String() + ":" + string(gameID)
	if rating, ok := m.ratings[key]; ok {
		return rating, args.Error(1)
	}
	if args.Get(0) != nil {
		return args.Get(0).(*matchmaking_entities.PlayerRating), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlayerRatingRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PlayerRating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*matchmaking_entities.PlayerRating), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlayerRatingRepository) GetTopPlayers(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error) {
	args := m.Called(ctx, gameID, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]*matchmaking_entities.PlayerRating), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlayerRatingRepository) GetRankDistribution(ctx context.Context, gameID replay_common.GameIDKey) (map[matchmaking_entities.Rank]int, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) != nil {
		return args.Get(0).(map[matchmaking_entities.Rank]int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlayerRatingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestContext() context.Context {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()
	ctx = context.WithValue(ctx, shared.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.GroupIDKey, groupID)
	return ctx
}

func TestGlicko2RatingService_GetPlayerRating_NewPlayer(t *testing.T) {
	mockRepo := NewMockPlayerRatingRepository()
	service := matchmaking_services.NewGlicko2RatingService(mockRepo)
	ctx := setupTestContext()

	playerID := uuid.New()
	gameID := replay_common.CS2.ID

	// Return nil = player not found
	mockRepo.On("FindByPlayerAndGame", ctx, playerID, gameID).Return(nil, nil)
	mockRepo.On("Save", ctx, mock.AnythingOfType("*matchmaking_entities.PlayerRating")).Return(nil)

	rating, err := service.GetPlayerRating(ctx, playerID, gameID)

	assert.NoError(t, err)
	assert.NotNil(t, rating)
	assert.Equal(t, playerID, rating.PlayerID)
	assert.Equal(t, gameID, rating.GameID)
	assert.Equal(t, matchmaking_entities.DefaultRating, rating.Rating)
	assert.Equal(t, matchmaking_entities.DefaultRatingDeviation, rating.RatingDeviation)
	assert.Equal(t, matchmaking_entities.DefaultVolatility, rating.Volatility)
	assert.True(t, rating.IsProvisional())
	mockRepo.AssertExpectations(t)
}

func TestGlicko2RatingService_GetPlayerRating_ExistingPlayer(t *testing.T) {
	mockRepo := NewMockPlayerRatingRepository()
	service := matchmaking_services.NewGlicko2RatingService(mockRepo)
	ctx := setupTestContext()

	playerID := uuid.New()
	gameID := replay_common.CS2.ID
	
	existingRating := &matchmaking_entities.PlayerRating{
		ID:              uuid.New(),
		PlayerID:        playerID,
		GameID:          gameID,
		Rating:          1800.0,
		RatingDeviation: 100.0,
		Volatility:      0.05,
		MatchesPlayed:   50,
		Wins:            30,
		Losses:          20,
	}

	mockRepo.On("FindByPlayerAndGame", ctx, playerID, gameID).Return(existingRating, nil)

	rating, err := service.GetPlayerRating(ctx, playerID, gameID)

	assert.NoError(t, err)
	assert.NotNil(t, rating)
	assert.Equal(t, 1800.0, rating.Rating)
	assert.Equal(t, 50, rating.MatchesPlayed)
	assert.False(t, rating.IsProvisional())
	mockRepo.AssertExpectations(t)
}

func TestGlicko2RatingService_UpdateRatingsAfterMatch(t *testing.T) {
	mockRepo := NewMockPlayerRatingRepository()
	service := matchmaking_services.NewGlicko2RatingService(mockRepo)
	ctx := setupTestContext()

	winner1 := uuid.New()
	winner2 := uuid.New()
	loser1 := uuid.New()
	loser2 := uuid.New()
	gameID := replay_common.CS2.ID

	// Create equal ratings for all players
	for _, playerID := range []uuid.UUID{winner1, winner2, loser1, loser2} {
		rating := matchmaking_entities.NewPlayerRating(playerID, gameID, shared.GetResourceOwner(ctx))
		rating.Rating = 1500.0
		rating.RatingDeviation = 100.0
		mockRepo.ratings[playerID.String()+":"+string(gameID)] = rating
	}

	mockRepo.On("FindByPlayerAndGame", ctx, mock.Anything, gameID).Return(nil, nil)
	mockRepo.On("Save", ctx, mock.Anything).Return(nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(nil)

	cmd := matchmaking_in.UpdateRatingsCommand{
		MatchID:         uuid.New(),
		GameID:          gameID,
		WinnerPlayerIDs: []uuid.UUID{winner1, winner2},
		LoserPlayerIDs:  []uuid.UUID{loser1, loser2},
	}

	err := service.UpdateRatingsAfterMatch(ctx, cmd)
	assert.NoError(t, err)

	// Verify winners gained rating
	winner1Rating := mockRepo.ratings[winner1.String()+":"+string(gameID)]
	assert.Greater(t, winner1Rating.Rating, 1500.0, "Winner should gain rating")
	assert.Equal(t, 1, winner1Rating.Wins)
	assert.Equal(t, 1, winner1Rating.MatchesPlayed)
	assert.Equal(t, 1, winner1Rating.WinStreak)

	// Verify losers lost rating
	loser1Rating := mockRepo.ratings[loser1.String()+":"+string(gameID)]
	assert.Less(t, loser1Rating.Rating, 1500.0, "Loser should lose rating")
	assert.Equal(t, 1, loser1Rating.Losses)
	assert.Equal(t, 0, loser1Rating.WinStreak)
}

func TestPlayerRating_Ranks(t *testing.T) {
	testCases := []struct {
		rating       float64
		expectedRank matchmaking_entities.Rank
	}{
		{2850, matchmaking_entities.RankChallenger},
		{2600, matchmaking_entities.RankGrandmaster},
		{2300, matchmaking_entities.RankMaster},
		{2000, matchmaking_entities.RankDiamond},
		{1700, matchmaking_entities.RankPlatinum},
		{1500, matchmaking_entities.RankGold},
		{1300, matchmaking_entities.RankSilver},
		{1000, matchmaking_entities.RankBronze},
		{500, matchmaking_entities.RankBronze},
	}

	for _, tc := range testCases {
		rating := &matchmaking_entities.PlayerRating{Rating: tc.rating}
		assert.Equal(t, tc.expectedRank, rating.GetRank(), "Rating %.0f should be %s", tc.rating, tc.expectedRank)
	}
}

func TestPlayerRating_GetMMR(t *testing.T) {
	rating := &matchmaking_entities.PlayerRating{Rating: 1523.7}
	assert.Equal(t, 1524, rating.GetMMR())

	rating.Rating = 1499.2
	assert.Equal(t, 1499, rating.GetMMR())
}

func TestPlayerRating_WinRate(t *testing.T) {
	rating := &matchmaking_entities.PlayerRating{
		Wins:   7,
		Losses: 3,
		Draws:  0,
	}
	assert.InDelta(t, 70.0, rating.GetWinRate(), 0.1)

	rating.Wins = 0
	rating.Losses = 0
	assert.Equal(t, 0.0, rating.GetWinRate())
}

func TestPlayerRating_Confidence(t *testing.T) {
	rating := &matchmaking_entities.PlayerRating{
		RatingDeviation: 50.0, // Very confident
	}
	assert.InDelta(t, 85.7, rating.GetConfidence(), 1.0)

	rating.RatingDeviation = 350.0 // New player, not confident
	assert.InDelta(t, 0.0, rating.GetConfidence(), 1.0)
}

func TestPlayerRating_InactivityDecay(t *testing.T) {
	rating := &matchmaking_entities.PlayerRating{
		RatingDeviation: 100.0,
	}

	// After 30 days of inactivity, RD should increase by ~25
	rating.ApplyInactivityDecay(30)
	assert.InDelta(t, 104.2, rating.RatingDeviation, 5.0)

	// After 90 days, RD should be higher but capped at 350
	rating.RatingDeviation = 100.0
	rating.ApplyInactivityDecay(365)
	assert.LessOrEqual(t, rating.RatingDeviation, matchmaking_entities.DefaultRatingDeviation)
}

func TestPlayerRating_IsProvisional(t *testing.T) {
	rating := &matchmaking_entities.PlayerRating{MatchesPlayed: 5}
	assert.True(t, rating.IsProvisional())

	rating.MatchesPlayed = 10
	assert.False(t, rating.IsProvisional())

	rating.MatchesPlayed = 15
	assert.False(t, rating.IsProvisional())
}

func TestGlicko2RatingService_GetLeaderboard(t *testing.T) {
	mockRepo := NewMockPlayerRatingRepository()
	service := matchmaking_services.NewGlicko2RatingService(mockRepo)
	ctx := context.Background()

	gameID := replay_common.CS2.ID
	topPlayers := []*matchmaking_entities.PlayerRating{
		{PlayerID: uuid.New(), Rating: 2800, MatchesPlayed: 100},
		{PlayerID: uuid.New(), Rating: 2600, MatchesPlayed: 80},
		{PlayerID: uuid.New(), Rating: 2400, MatchesPlayed: 60},
	}

	mockRepo.On("GetTopPlayers", ctx, gameID, 10).Return(topPlayers, nil)

	leaderboard, err := service.GetLeaderboard(ctx, gameID, 10)

	assert.NoError(t, err)
	assert.Len(t, leaderboard, 3)
	assert.Equal(t, 2800.0, leaderboard[0].Rating)
	assert.Equal(t, matchmaking_entities.RankChallenger, leaderboard[0].GetRank())
	mockRepo.AssertExpectations(t)
}

func TestPlayerRating_PeakRating(t *testing.T) {
	ctx := setupTestContext()
	playerID := uuid.New()
	gameID := replay_common.CS2.ID

	rating := matchmaking_entities.NewPlayerRating(playerID, gameID, shared.GetResourceOwner(ctx))
	assert.Equal(t, matchmaking_entities.DefaultRating, rating.PeakRating)

	// Simulate rating increase
	rating.Rating = 1700
	if rating.Rating > rating.PeakRating {
		rating.PeakRating = rating.Rating
	}
	assert.Equal(t, 1700.0, rating.PeakRating)

	// Rating drops but peak stays
	rating.Rating = 1600
	assert.Equal(t, 1700.0, rating.PeakRating)
}

func TestPlayerRating_RatingHistory(t *testing.T) {
	ctx := setupTestContext()
	playerID := uuid.New()
	gameID := replay_common.CS2.ID

	rating := matchmaking_entities.NewPlayerRating(playerID, gameID, shared.GetResourceOwner(ctx))
	assert.Empty(t, rating.RatingHistory)

	// Add rating changes
	for i := 0; i < 55; i++ {
		rating.RatingHistory = append(rating.RatingHistory, matchmaking_entities.RatingChange{
			MatchID:   uuid.New(),
			OldRating: 1500,
			NewRating: 1510,
			Change:    10,
			Result:    "win",
			Timestamp: time.Now(),
		})
	}

	// Trim to last 50
	if len(rating.RatingHistory) > 50 {
		rating.RatingHistory = rating.RatingHistory[len(rating.RatingHistory)-50:]
	}

	assert.Len(t, rating.RatingHistory, 50)
}

