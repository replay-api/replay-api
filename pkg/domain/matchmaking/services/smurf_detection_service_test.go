package matchmaking_services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MOCK REPOSITORIES
// =============================================================================

type MockSmurfProfileRepository struct {
	mock.Mock
}

func (m *MockSmurfProfileRepository) Create(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockSmurfProfileRepository) Update(ctx context.Context, profile *matchmaking_entities.SmurfProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockSmurfProfileRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.SmurfProfile, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.SmurfProfile), args.Error(1)
}

func (m *MockSmurfProfileRepository) GetFlaggedProfiles(ctx context.Context, limit, offset int) ([]matchmaking_entities.SmurfProfile, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]matchmaking_entities.SmurfProfile), args.Error(1)
}

type MockMatchStatsRepository struct {
	mock.Mock
}

func (m *MockMatchStatsRepository) GetPlayerMatchStats(ctx context.Context, playerID uuid.UUID, limit int) ([]MatchStatsSummary, error) {
	args := m.Called(ctx, playerID, limit)
	return args.Get(0).([]MatchStatsSummary), args.Error(1)
}

func (m *MockMatchStatsRepository) GetPlayerMatchCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	args := m.Called(ctx, playerID)
	return args.Int(0), args.Error(1)
}

type MockPlayerRatingRepository struct {
	mock.Mock
}

func (m *MockPlayerRatingRepository) GetPlayerRating(ctx context.Context, playerID uuid.UUID) (*matchmaking_entities.PlayerRating, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*matchmaking_entities.PlayerRating), args.Error(1)
}

func (m *MockPlayerRatingRepository) GetRatingHistory(ctx context.Context, playerID uuid.UUID, limit int) ([]RatingSnapshot, error) {
	args := m.Called(ctx, playerID, limit)
	return args.Get(0).([]RatingSnapshot), args.Error(1)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func testResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}
}

func testContext() context.Context {
	ro := testResourceOwner()
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.TenantIDKey, ro.TenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, ro.ClientID)
	ctx = context.WithValue(ctx, common.UserIDKey, ro.UserID)
	return ctx
}

func testThresholds() matchmaking_entities.SmurfDetectionThresholds {
	return matchmaking_entities.SmurfDetectionThresholds{
		NewAccountMatchThreshold:        50,
		NewAccountHighWinRateThreshold:  0.75,
		SuspiciousWinRateThreshold:      0.70,
		RapidRatingGainThreshold:        30.0,
		HighHeadshotRateThreshold:       0.45,
	}
}

// generateMatchStats creates match statistics for testing
func generateMatchStats(count int, winRate float64, avgKDA float64, headshotRate float64) []MatchStatsSummary {
	stats := make([]MatchStatsSummary, count)
	for i := 0; i < count; i++ {
		won := float64(i)/float64(count) < winRate
		
		// Calculate kills/deaths for target KDA
		kills := int(avgKDA * 10)
		deaths := 10
		
		// Calculate headshots for target rate
		totalShots := 100
		headshots := int(float64(totalShots) * headshotRate)
		
		stats[i] = MatchStatsSummary{
			MatchID:       uuid.New(),
			PlayedAt:      time.Now().Add(-time.Duration(count-i) * time.Hour),
			Won:           won,
			Kills:         kills,
			Deaths:        deaths,
			Assists:       5,
			Headshots:     headshots,
			TotalShots:    totalShots,
			UtilityDamage: 30.0,
			FlashAssists:  1,
			EntryFrags:    1,
			ClutchWins:    0,
			RatingChange:  10.0,
		}
	}
	return stats
}

func generateRatingHistory(count int, startRating, endRating float64) []RatingSnapshot {
	history := make([]RatingSnapshot, count)
	ratingStep := (endRating - startRating) / float64(count-1)
	
	for i := 0; i < count; i++ {
		history[i] = RatingSnapshot{
			Rating:    startRating + float64(count-1-i)*ratingStep,
			Deviation: 50.0,
			Timestamp: time.Now().Add(-time.Duration(count-i) * 24 * time.Hour),
		}
	}
	return history
}

// =============================================================================
// UNIT TESTS
// =============================================================================

// TestNewSmurfDetectionService verifies service creation
func TestNewSmurfDetectionService(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	assert.NotNil(t, service)
	assert.Equal(t, thresholds, service.thresholds)
}

// TestAnalyzePlayer_NewAccountHighSkill detects high skill on new accounts
// Business context: Identifies potential smurfs by detecting new accounts that 
// immediately perform at expert level - a key indicator of experienced players
// creating fresh accounts to dominate lower-ranked matches.
func TestAnalyzePlayer_NewAccountHighSkill(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	// Setup: New account with 90% win rate - suspicious
	matchStats := generateMatchStats(15, 0.90, 2.5, 0.45)
	ratingHistory := generateRatingHistory(10, 1000, 1500)
	currentRating := &matchmaking_entities.PlayerRating{Rating: 1500}

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(15, nil)
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(currentRating, nil)
	// Note: Service creates new profile with ID, so Update is called even for new profiles
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Greater(t, len(profile.Indicators), 0, "Should detect smurf indicators")
	
	// Verify new account high skill indicator was added
	hasNewAccountIndicator := false
	for _, ind := range profile.Indicators {
		if ind.Type == matchmaking_entities.IndicatorNewAccountHighSkill {
			hasNewAccountIndicator = true
			break
		}
	}
	assert.True(t, hasNewAccountIndicator, "Should detect new account high skill")
}

// TestAnalyzePlayer_AbnormalWinRate detects sustained high win rates
// Business context: Sustained win rates above 70% across 20+ matches are
// statistically improbable for a player at their correct skill level,
// indicating likely smurf behavior.
func TestAnalyzePlayer_AbnormalWinRate(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	// Setup: 75% win rate over 30 matches
	matchStats := generateMatchStats(30, 0.75, 2.0, 0.40)
	ratingHistory := generateRatingHistory(20, 1200, 1800)

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(60, nil) // Not a new account
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(&matchmaking_entities.PlayerRating{Rating: 1800}, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)

	// Verify abnormal win rate indicator
	hasAbnormalWinRate := false
	for _, ind := range profile.Indicators {
		if ind.Type == matchmaking_entities.IndicatorAbnormalWinRate {
			hasAbnormalWinRate = true
			assert.GreaterOrEqual(t, ind.Weight, 2.0, "High win rate should have significant weight")
			break
		}
	}
	assert.True(t, hasAbnormalWinRate, "Should detect abnormal win rate")
}

// TestAnalyzePlayer_RapidRatingProgression detects fast skill gains
// Business context: When a player gains rating much faster than the expected
// learning curve, it suggests they already possess the skills and are
// artificially starting from a lower rank.
func TestAnalyzePlayer_RapidRatingProgression(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	matchStats := generateMatchStats(20, 0.65, 1.8, 0.35)
	
	// Rating history showing 50+ rating gain per day (threshold is 30)
	ratingHistory := []RatingSnapshot{
		{Rating: 1800, Deviation: 50, Timestamp: time.Now()},
		{Rating: 1700, Deviation: 55, Timestamp: time.Now().Add(-24 * time.Hour)},
		{Rating: 1500, Deviation: 60, Timestamp: time.Now().Add(-48 * time.Hour)},
		{Rating: 1300, Deviation: 65, Timestamp: time.Now().Add(-72 * time.Hour)},
		{Rating: 1100, Deviation: 70, Timestamp: time.Now().Add(-96 * time.Hour)},
		{Rating: 1000, Deviation: 75, Timestamp: time.Now().Add(-120 * time.Hour)}, // 5 days ago
	}

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(100, nil)
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(&matchmaking_entities.PlayerRating{Rating: 1800}, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)

	// Verify rapid skill progression indicator
	hasRapidProgression := false
	for _, ind := range profile.Indicators {
		if ind.Type == matchmaking_entities.IndicatorRapidSkillProgression {
			hasRapidProgression = true
			break
		}
	}
	assert.True(t, hasRapidProgression, "Should detect rapid skill progression")
}

// TestAnalyzePlayer_HighHeadshotRate detects professional-level aim
// Business context: Headshot rates above 45% are typically only achieved by
// highly skilled or professional players. A new account with such accuracy
// strongly suggests an experienced player on a secondary account.
func TestAnalyzePlayer_HighHeadshotRate(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	// High headshot rate (55%)
	matchStats := generateMatchStats(20, 0.55, 2.0, 0.55)
	ratingHistory := generateRatingHistory(10, 1000, 1400)

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(100, nil)
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(&matchmaking_entities.PlayerRating{Rating: 1400}, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)

	// Verify headshot rate anomaly indicator
	hasHeadshotAnomaly := false
	for _, ind := range profile.Indicators {
		if ind.Type == matchmaking_entities.IndicatorHeadshotRateAnomaly {
			hasHeadshotAnomaly = true
			assert.Greater(t, ind.Confidence, 0.0, "Should have positive confidence")
			break
		}
	}
	assert.True(t, hasHeadshotAnomaly, "Should detect headshot rate anomaly")
}

// TestAnalyzePlayer_LegitimatePlayer ensures no false positives for normal players
// Business context: The system must accurately distinguish between legitimate
// improving players and smurfs. This test ensures average performance patterns
// don't trigger false smurf alerts, which would harm player trust.
func TestAnalyzePlayer_LegitimatePlayer(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	// Normal player: 50% win rate, average KDA, average headshot rate
	matchStats := generateMatchStats(50, 0.50, 1.2, 0.30)
	ratingHistory := generateRatingHistory(30, 1200, 1300) // Slow, gradual improvement

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(200, nil) // Established account
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(&matchmaking_entities.PlayerRating{Rating: 1300}, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	
	// Legitimate player should have few or no indicators
	assert.LessOrEqual(t, len(profile.Indicators), 1, "Legitimate player should have minimal indicators")
	
	// If any indicators, they should have low weight
	for _, ind := range profile.Indicators {
		assert.Less(t, ind.Weight, 2.0, "Legitimate player indicators should have low weight")
	}
}

// TestAnalyzePlayer_ExistingProfile updates existing smurf profiles
// Business context: Continuous monitoring is essential - the system must 
// update existing profiles rather than creating duplicates, enabling
// tracking of player behavior over time.
func TestAnalyzePlayer_ExistingProfile(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()
	
	existingProfile := matchmaking_entities.NewSmurfProfile(playerID, testResourceOwner())
	existingProfile.Status = matchmaking_entities.SmurfStatusSuspicious

	matchStats := generateMatchStats(20, 0.60, 1.5, 0.35)
	ratingHistory := generateRatingHistory(15, 1100, 1300)

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(existingProfile, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(100, nil)
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(&matchmaking_entities.PlayerRating{Rating: 1300}, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, existingProfile.BaseEntity.ID, profile.BaseEntity.ID, "Should update existing profile, not create new")
	
	smurfRepo.AssertCalled(t, "Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile"))
}

// TestAnalyzePlayer_InsufficientData handles new players with few matches
// Business context: Players with very few matches cannot be accurately assessed.
// The system should gracefully handle this case without triggering false positives.
func TestAnalyzePlayer_InsufficientData(t *testing.T) {
	smurfRepo := &MockSmurfProfileRepository{}
	matchRepo := &MockMatchStatsRepository{}
	ratingRepo := &MockPlayerRatingRepository{}
	thresholds := testThresholds()

	service := NewSmurfDetectionService(smurfRepo, matchRepo, ratingRepo, thresholds)

	ctx := testContext()
	playerID := uuid.New()

	// Very few matches - insufficient data for analysis
	matchStats := generateMatchStats(3, 1.0, 3.0, 0.60) // Perfect stats but only 3 games
	ratingHistory := []RatingSnapshot{} // No rating history

	smurfRepo.On("GetByPlayerID", ctx, playerID).Return(nil, nil)
	matchRepo.On("GetPlayerMatchStats", ctx, playerID, 100).Return(matchStats, nil)
	matchRepo.On("GetPlayerMatchCount", ctx, playerID).Return(3, nil)
	ratingRepo.On("GetRatingHistory", ctx, playerID, 50).Return(ratingHistory, nil)
	ratingRepo.On("GetPlayerRating", ctx, playerID).Return(nil, nil)
	smurfRepo.On("Update", ctx, mock.AnythingOfType("*matchmaking_entities.SmurfProfile")).Return(nil)

	profile, err := service.AnalyzePlayer(ctx, playerID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, 3, profile.MatchesAnalyzed)
	// Most indicators require 10+ matches, so few should trigger
	assert.LessOrEqual(t, len(profile.Indicators), 2, "Should have limited indicators with insufficient data")
}

// =============================================================================
// HELPER FUNCTION TESTS
// =============================================================================

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5.0}, 5.0},
		{"multiple", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
		{"decimals", []float64{1.5, 2.5}, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMean(tt.values)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5.0}, 0}, // Single value has 0 variance
		{"same values", []float64{3.0, 3.0, 3.0}, 0},
		{"varied", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean := calculateMean(tt.values)
			got := calculateVariance(tt.values, mean)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		threshold float64
		maxValue  float64
		want      float64
	}{
		{"below threshold", 0.3, 0.5, 1.0, 0},
		{"at threshold", 0.5, 0.5, 1.0, 0},
		{"at max", 1.0, 0.5, 1.0, 1.0},
		{"halfway", 0.75, 0.5, 1.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateConfidence(tt.value, tt.threshold, tt.maxValue)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestMin(t *testing.T) {
	assert.Equal(t, 5, min(5, 10))
	assert.Equal(t, 5, min(10, 5))
	assert.Equal(t, 5, min(5, 5))
}

// =============================================================================
// PERFORMANCE ANALYSIS TESTS
// =============================================================================

// TestEstimateSkillFromStats verifies skill estimation algorithm
func TestEstimateSkillFromStats(t *testing.T) {
	service := NewSmurfDetectionService(nil, nil, nil, testThresholds())

	tests := []struct {
		name           string
		stats          []MatchStatsSummary
		expectedRange  [2]float64 // min and max expected rating
	}{
		{
			name:          "empty stats",
			stats:         []MatchStatsSummary{},
			expectedRange: [2]float64{1000, 1000},
		},
		{
			name:          "average player",
			stats:         generateMatchStats(20, 0.50, 1.0, 0.30),
			expectedRange: [2]float64{1000, 1200},
		},
		{
			name:          "skilled player",
			stats:         generateMatchStats(20, 0.70, 2.5, 0.50),
			expectedRange: [2]float64{1400, 1800},
		},
		{
			name:          "pro-level player",
			stats:         generateMatchStats(20, 0.85, 3.5, 0.65),
			expectedRange: [2]float64{1800, 2500},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating := service.estimateSkillFromStats(tt.stats)
			assert.GreaterOrEqual(t, rating, tt.expectedRange[0], "Rating should be >= min expected")
			assert.LessOrEqual(t, rating, tt.expectedRange[1], "Rating should be <= max expected")
		})
	}
}

// TestGetRankFromRating verifies rank tier assignment
func TestGetRankFromRating(t *testing.T) {
	service := NewSmurfDetectionService(nil, nil, nil, testThresholds())

	tests := []struct {
		rating float64
		rank   string
	}{
		{500, "Silver"},
		{700, "GN4"},
		{1000, "MG1"},
		{1300, "MG2"},
		{1600, "DMG"},
		{1900, "LEM"},
		{2200, "Supreme"},
		{2500, "Global Elite"},
	}

	for _, tt := range tests {
		t.Run(tt.rank, func(t *testing.T) {
			got := service.getRankFromRating(tt.rating)
			assert.Equal(t, tt.rank, got)
		})
	}
}

