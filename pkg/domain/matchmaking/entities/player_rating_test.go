package matchmaking_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
)

// testMatchmakingResourceOwner creates a test resource owner
func testMatchmakingResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestGlicko2Constants(t *testing.T) {
	assert := assert.New(t)

	// Verify Glicko-2 defaults are properly set
	assert.Equal(1500.0, DefaultRating)
	assert.Equal(350.0, DefaultRatingDeviation)
	assert.Equal(0.06, DefaultVolatility)
	assert.Equal(173.7178, ScaleFactor)
	assert.Equal(0.5, Tau)
}

func TestRank_ValuesExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(Rank("bronze"), RankBronze)
	assert.Equal(Rank("silver"), RankSilver)
	assert.Equal(Rank("gold"), RankGold)
	assert.Equal(Rank("platinum"), RankPlatinum)
	assert.Equal(Rank("diamond"), RankDiamond)
	assert.Equal(Rank("master"), RankMaster)
	assert.Equal(Rank("grandmaster"), RankGrandmaster)
	assert.Equal(Rank("challenger"), RankChallenger)
}

// =============================================================================
// NewPlayerRating Tests
// =============================================================================

func TestNewPlayerRating_CreatesValidRating(t *testing.T) {
	assert := assert.New(t)
	playerID := uuid.New()
	rxn := testMatchmakingResourceOwner()

	rating := NewPlayerRating(playerID, replay_common.CS2_GAME_ID, rxn)

	assert.NotNil(rating)
	assert.NotEqual(uuid.Nil, rating.ID)
	assert.Equal(playerID, rating.PlayerID)
	assert.Equal(replay_common.CS2_GAME_ID, rating.GameID)
	assert.Equal(DefaultRating, rating.Rating)
	assert.Equal(DefaultRatingDeviation, rating.RatingDeviation)
	assert.Equal(DefaultVolatility, rating.Volatility)
	assert.Equal(0, rating.MatchesPlayed)
	assert.Equal(0, rating.Wins)
	assert.Equal(0, rating.Losses)
	assert.Equal(0, rating.Draws)
	assert.Equal(0, rating.WinStreak)
	assert.Equal(DefaultRating, rating.PeakRating)
	assert.Empty(rating.RatingHistory)
	assert.Equal(rxn, rating.ResourceOwner)
}

func TestNewPlayerRating_SetsTimestamps(t *testing.T) {
	assert := assert.New(t)
	before := time.Now()

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.WithinDuration(before, rating.CreatedAt, time.Second)
	assert.WithinDuration(before, rating.UpdatedAt, time.Second)
	assert.Nil(rating.LastMatchAt)
}

func TestNewPlayerRating_InitializesEmptyHistory(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.NotNil(rating.RatingHistory)
	assert.Len(rating.RatingHistory, 0)
}

// =============================================================================
// GetID Tests
// =============================================================================

func TestPlayerRating_GetID(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.Equal(rating.ID, rating.GetID())
}

// =============================================================================
// GetMMR Tests
// =============================================================================

func TestGetMMR_ReturnsRoundedInteger(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	// Default rating
	assert.Equal(1500, rating.GetMMR())

	// Test rounding
	rating.Rating = 1534.4
	assert.Equal(1534, rating.GetMMR())

	rating.Rating = 1534.5
	assert.Equal(1535, rating.GetMMR()) // rounds up

	rating.Rating = 1534.6
	assert.Equal(1535, rating.GetMMR())
}

func TestGetMMR_HandlesExtremeValues(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	rating.Rating = 0
	assert.Equal(0, rating.GetMMR())

	rating.Rating = 3000.99
	assert.Equal(3001, rating.GetMMR())
}

// =============================================================================
// GetRank Tests
// =============================================================================

func TestGetRank_ReturnsBronzeForLowRating(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1000

	assert.Equal(RankBronze, rating.GetRank())
}

func TestGetRank_ReturnsSilverAt1200(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1200

	assert.Equal(RankSilver, rating.GetRank())
}

func TestGetRank_ReturnsGoldAt1400(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1400

	assert.Equal(RankGold, rating.GetRank())
}

func TestGetRank_ReturnsPlatinumAt1600(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1600

	assert.Equal(RankPlatinum, rating.GetRank())
}

func TestGetRank_ReturnsDiamondAt1900(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1900

	assert.Equal(RankDiamond, rating.GetRank())
}

func TestGetRank_ReturnsMasterAt2200(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 2200

	assert.Equal(RankMaster, rating.GetRank())
}

func TestGetRank_ReturnsGrandmasterAt2500(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 2500

	assert.Equal(RankGrandmaster, rating.GetRank())
}

func TestGetRank_ReturnsChallengerAt2800(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 2800

	assert.Equal(RankChallenger, rating.GetRank())
}

func TestGetRank_AllBoundaries(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		rating       float64
		expectedRank Rank
	}{
		{0, RankBronze},
		{1199, RankBronze},
		{1200, RankSilver},
		{1399, RankSilver},
		{1400, RankGold},
		{1500, RankGold},      // Default rating = Gold
		{1599, RankGold},
		{1600, RankPlatinum},
		{1899, RankPlatinum},
		{1900, RankDiamond},
		{2199, RankDiamond},
		{2200, RankMaster},
		{2499, RankMaster},
		{2500, RankGrandmaster},
		{2799, RankGrandmaster},
		{2800, RankChallenger},
		{3500, RankChallenger},
	}

	for _, tc := range testCases {
		rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
		rating.Rating = tc.rating
		assert.Equal(tc.expectedRank, rating.GetRank(), "Rating %.0f should be %s", tc.rating, tc.expectedRank)
	}
}

// =============================================================================
// GetConfidence Tests
// =============================================================================

func TestGetConfidence_ReturnsLowConfidenceForNewPlayer(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	// New player has RD of 350, should have ~0% confidence
	confidence := rating.GetConfidence()
	assert.InDelta(0, confidence, 1) // ~0% confidence
}

func TestGetConfidence_ReturnsHighConfidenceForEstablishedPlayer(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.RatingDeviation = 50 // Very confident

	confidence := rating.GetConfidence()
	assert.InDelta(85.7, confidence, 1) // ~85.7% confidence
}

func TestGetConfidence_Returns100PercentAtZeroRD(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.RatingDeviation = 0

	confidence := rating.GetConfidence()
	assert.Equal(100.0, confidence)
}

func TestGetConfidence_ClampsBetween0And100(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	// Very high RD should return 0
	rating.RatingDeviation = 500
	assert.Equal(0.0, rating.GetConfidence())

	// Very low RD should return 100
	rating.RatingDeviation = -100 // Edge case
	assert.Equal(100.0, rating.GetConfidence())
}

// =============================================================================
// IsProvisional Tests
// =============================================================================

func TestIsProvisional_ReturnsTrueForNewPlayer(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.True(rating.IsProvisional())
}

func TestIsProvisional_ReturnsTrueForLessThan10Matches(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.MatchesPlayed = 9

	assert.True(rating.IsProvisional())
}

func TestIsProvisional_ReturnsFalseFor10OrMoreMatches(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.MatchesPlayed = 10

	assert.False(rating.IsProvisional())
}

// =============================================================================
// GetWinRate Tests
// =============================================================================

func TestGetWinRate_ReturnsZeroForNoMatches(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.Equal(0.0, rating.GetWinRate())
}

func TestGetWinRate_Returns100ForAllWins(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Wins = 10
	rating.Losses = 0
	rating.Draws = 0

	assert.Equal(100.0, rating.GetWinRate())
}

func TestGetWinRate_Returns0ForAllLosses(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Wins = 0
	rating.Losses = 10
	rating.Draws = 0

	assert.Equal(0.0, rating.GetWinRate())
}

func TestGetWinRate_CalculatesCorrectly(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Wins = 6
	rating.Losses = 3
	rating.Draws = 1

	// 6/(6+3+1) * 100 = 60%
	assert.Equal(60.0, rating.GetWinRate())
}

func TestGetWinRate_IncludesDrawsInTotal(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Wins = 5
	rating.Losses = 0
	rating.Draws = 5

	// 5/(5+0+5) * 100 = 50%
	assert.Equal(50.0, rating.GetWinRate())
}

// =============================================================================
// ApplyInactivityDecay Tests
// =============================================================================

func TestApplyInactivityDecay_DoesNothingForZeroDays(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	initialRD := rating.RatingDeviation

	rating.ApplyInactivityDecay(0)

	assert.Equal(initialRD, rating.RatingDeviation)
}

func TestApplyInactivityDecay_DoesNothingForNegativeDays(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	initialRD := rating.RatingDeviation

	rating.ApplyInactivityDecay(-5)

	assert.Equal(initialRD, rating.RatingDeviation)
}

func TestApplyInactivityDecay_IncreasesRDForInactivity(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.RatingDeviation = 100 // Established player

	rating.ApplyInactivityDecay(30) // 1 month inactive

	assert.Greater(rating.RatingDeviation, 100.0)
}

func TestApplyInactivityDecay_CapsAtDefaultRD(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.RatingDeviation = 100

	// The decay formula is sqrt(RD² + (25/30)² * days)
	// With RD=100 and very long inactivity, RD would grow beyond 350
	// For RD to reach 350 from 100: 350² = 100² + 0.694*days => days ≈ 162,000
	rating.ApplyInactivityDecay(200000) // Enough to exceed 350

	// RD should be capped at 350, not exceed it
	assert.Equal(DefaultRatingDeviation, rating.RatingDeviation)
}

func TestApplyInactivityDecay_UpdatesTimestamp(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.RatingDeviation = 100
	oldUpdatedAt := rating.UpdatedAt

	time.Sleep(time.Millisecond)
	rating.ApplyInactivityDecay(30)

	assert.True(rating.UpdatedAt.After(oldUpdatedAt))
}

// =============================================================================
// RatingChange Tests
// =============================================================================

func TestRatingChange_StructFields(t *testing.T) {
	assert := assert.New(t)
	matchID := uuid.New()
	now := time.Now()

	change := RatingChange{
		MatchID:        matchID,
		OldRating:      1500,
		NewRating:      1525,
		Change:         25,
		Result:         "win",
		OpponentRating: 1520,
		Timestamp:      now,
	}

	assert.Equal(matchID, change.MatchID)
	assert.Equal(1500.0, change.OldRating)
	assert.Equal(1525.0, change.NewRating)
	assert.Equal(25.0, change.Change)
	assert.Equal("win", change.Result)
	assert.Equal(1520.0, change.OpponentRating)
	assert.Equal(now, change.Timestamp)
}

// =============================================================================
// Business Scenario Tests - E-Sports Platform Specific
// =============================================================================

func TestScenario_NewPlayerPlacementMatches(t *testing.T) {
	assert := assert.New(t)

	// New player starts with provisional rating
	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	assert.True(rating.IsProvisional())
	assert.Equal(RankGold, rating.GetRank()) // Default 1500 = Gold
	assert.InDelta(0, rating.GetConfidence(), 1)

	// After 10 matches, no longer provisional
	rating.MatchesPlayed = 10
	rating.Wins = 7
	rating.Losses = 3
	rating.RatingDeviation = 200

	assert.False(rating.IsProvisional())
	assert.InDelta(42.9, rating.GetConfidence(), 1)
}

func TestScenario_HighLevelCompetitivePlayer(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 2850
	rating.RatingDeviation = 60
	rating.Volatility = 0.04
	rating.MatchesPlayed = 500
	rating.Wins = 320
	rating.Losses = 180
	rating.PeakRating = 2920

	assert.Equal(RankChallenger, rating.GetRank())
	assert.InDelta(82.9, rating.GetConfidence(), 1)
	assert.InDelta(64, rating.GetWinRate(), 0.1)
	assert.False(rating.IsProvisional())
	assert.Equal(2850, rating.GetMMR())
}

func TestScenario_InactivePlayerReturns(t *testing.T) {
	assert := assert.New(t)

	// Player was active at Diamond
	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1950
	rating.RatingDeviation = 80
	rating.MatchesPlayed = 100

	// 60 days of inactivity
	rating.ApplyInactivityDecay(60)

	// RD should increase but not exceed 350
	assert.Greater(rating.RatingDeviation, 80.0)
	assert.LessOrEqual(rating.RatingDeviation, DefaultRatingDeviation)

	// Rating itself should stay the same
	assert.Equal(1950.0, rating.Rating)
	assert.Equal(RankDiamond, rating.GetRank())
}

func TestScenario_RatingProgression(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())

	// Simulate matches and rating progression
	rating.Wins = 5
	rating.Losses = 5
	rating.MatchesPlayed = 10

	// Silver player (low skill)
	rating.Rating = 1150
	rating.RatingDeviation = 200
	assert.Equal(RankBronze, rating.GetRank())

	// Improved to Gold
	rating.Rating = 1450
	assert.Equal(RankGold, rating.GetRank())

	// Became Platinum after grinding
	rating.Rating = 1650
	rating.MatchesPlayed = 100
	rating.Wins = 60
	rating.Losses = 40
	rating.RatingDeviation = 100
	assert.Equal(RankPlatinum, rating.GetRank())
	assert.Equal(60.0, rating.GetWinRate())
}

// =============================================================================
// Glicko-2 Scale Conversion Tests (Internal Functions)
// =============================================================================

func TestToGlicko2Scale_ConvertsCorrectly(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 1500
	rating.RatingDeviation = 350

	mu, phi := rating.toGlicko2Scale()

	// At default, mu should be 0, phi should be ~2.0145
	assert.InDelta(0, mu, 0.001)
	assert.InDelta(2.0145, phi, 0.01)
}

func TestToGlicko2Scale_HigherRating(t *testing.T) {
	assert := assert.New(t)

	rating := NewPlayerRating(uuid.New(), replay_common.CS2_GAME_ID, testMatchmakingResourceOwner())
	rating.Rating = 2000 // 500 above default
	rating.RatingDeviation = 100

	mu, phi := rating.toGlicko2Scale()

	// mu = (2000-1500)/173.7178 ≈ 2.878
	assert.InDelta(2.878, mu, 0.01)
	// phi = 100/173.7178 ≈ 0.576
	assert.InDelta(0.576, phi, 0.01)
}

func TestFromGlicko2Scale_ConvertsBackCorrectly(t *testing.T) {
	assert := assert.New(t)

	// Convert mu=0, phi=2.0145 back to rating scale
	newRating, rd := fromGlicko2Scale(0, 2.0145)

	assert.InDelta(1500, newRating, 1)
	assert.InDelta(350, rd, 1)
}

func TestGFunction_ReturnsCorrectValues(t *testing.T) {
	assert := assert.New(t)

	// g(phi) decreases as phi increases
	g1 := g(0.5)
	g2 := g(1.0)
	g3 := g(2.0)

	assert.Greater(g1, g2)
	assert.Greater(g2, g3)

	// All values should be between 0 and 1
	assert.LessOrEqual(g1, 1.0)
	assert.GreaterOrEqual(g3, 0.0)
}

func TestExpectedScore_ReturnsCorrectRange(t *testing.T) {
	assert := assert.New(t)

	// Equal ratings should give ~0.5 expected score
	e := expectedScore(0, 0, 1.0)
	assert.InDelta(0.5, e, 0.01)

	// Higher mu should give higher expected score
	eHigher := expectedScore(1.0, 0, 1.0)
	assert.Greater(eHigher, 0.5)

	// Lower mu should give lower expected score
	eLower := expectedScore(-1.0, 0, 1.0)
	assert.Less(eLower, 0.5)
}

