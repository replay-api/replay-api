package matchmaking_vo_test

import (
	"testing"

	"github.com/google/uuid"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributionRule_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		rule     matchmaking_vo.DistributionRule
		expected bool
	}{
		{
			name:     "winner takes all is valid",
			rule:     matchmaking_vo.DistributionRuleWinnerTakesAll,
			expected: true,
		},
		{
			name:     "top three split is valid",
			rule:     matchmaking_vo.DistributionRuleTopThreeSplit,
			expected: true,
		},
		{
			name:     "performance mvp is valid",
			rule:     matchmaking_vo.DistributionRulePerformanceMVP,
			expected: true,
		},
		{
			name:     "empty string is invalid",
			rule:     matchmaking_vo.DistributionRule(""),
			expected: false,
		},
		{
			name:     "unknown rule is invalid",
			rule:     matchmaking_vo.DistributionRule("unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rule.IsValid())
		})
	}
}

func TestDistributionRule_String(t *testing.T) {
	assert.Equal(t, "winner_takes_all", matchmaking_vo.DistributionRuleWinnerTakesAll.String())
	assert.Equal(t, "top_three_split_60_30_10", matchmaking_vo.DistributionRuleTopThreeSplit.String())
	assert.Equal(t, "performance_mvp_70_20_10", matchmaking_vo.DistributionRulePerformanceMVP.String())
}

func TestDistributionRule_DisplayName(t *testing.T) {
	assert.Equal(t, "Winner Takes All", matchmaking_vo.DistributionRuleWinnerTakesAll.DisplayName())
	assert.Equal(t, "Top 3 Split", matchmaking_vo.DistributionRuleTopThreeSplit.DisplayName())
	assert.Equal(t, "Performance MVP", matchmaking_vo.DistributionRulePerformanceMVP.DisplayName())
	assert.Equal(t, "Unknown", matchmaking_vo.DistributionRule("invalid").DisplayName())
}

func TestDistributionRule_Description(t *testing.T) {
	assert.Contains(t, matchmaking_vo.DistributionRuleWinnerTakesAll.Description(), "100%")
	assert.Contains(t, matchmaking_vo.DistributionRuleTopThreeSplit.Description(), "60%")
	assert.Contains(t, matchmaking_vo.DistributionRulePerformanceMVP.Description(), "70%")
	assert.Empty(t, matchmaking_vo.DistributionRule("invalid").Description())
}

func TestDistributionRule_Icon(t *testing.T) {
	assert.NotEmpty(t, matchmaking_vo.DistributionRuleWinnerTakesAll.Icon())
	assert.NotEmpty(t, matchmaking_vo.DistributionRuleTopThreeSplit.Icon())
	assert.NotEmpty(t, matchmaking_vo.DistributionRulePerformanceMVP.Icon())
	assert.NotEmpty(t, matchmaking_vo.DistributionRule("invalid").Icon())
}

func TestAllDistributionRules(t *testing.T) {
	rules := matchmaking_vo.AllDistributionRules()

	assert.Len(t, rules, 3)
	assert.Contains(t, rules, matchmaking_vo.DistributionRuleWinnerTakesAll)
	assert.Contains(t, rules, matchmaking_vo.DistributionRuleTopThreeSplit)
	assert.Contains(t, rules, matchmaking_vo.DistributionRulePerformanceMVP)
}

func TestDistributionRule_Calculate_WinnerTakesAll(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleWinnerTakesAll
	totalPrize := wallet_vo.NewAmount(100.00)
	winnerID := uuid.New()
	rankedPlayers := []uuid.UUID{winnerID}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)
	assert.Equal(t, rule, distribution.Rule)
	assert.Equal(t, totalPrize.Cents(), distribution.Total.Cents())
	assert.Equal(t, totalPrize.Cents(), distribution.WinnerAmount.Cents())
	assert.Equal(t, winnerID, distribution.WinnerPlayerID)
	assert.Equal(t, int64(0), distribution.RunnerUpAmount.Cents())
	assert.Equal(t, int64(0), distribution.ThirdPlaceAmount.Cents())
	assert.Equal(t, int64(0), distribution.MVPBonus.Cents())
}

func TestDistributionRule_Calculate_TopThreeSplit_ThreePlayers(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleTopThreeSplit
	totalPrize := wallet_vo.NewAmount(100.00)
	player1 := uuid.New()
	player2 := uuid.New()
	player3 := uuid.New()
	rankedPlayers := []uuid.UUID{player1, player2, player3}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)
	assert.Equal(t, rule, distribution.Rule)

	// 60% to winner = $60.00
	assert.Equal(t, int64(6000), distribution.WinnerAmount.Cents())
	assert.Equal(t, player1, distribution.WinnerPlayerID)

	// 30% to runner-up = $30.00
	assert.Equal(t, int64(3000), distribution.RunnerUpAmount.Cents())
	assert.Equal(t, player2, distribution.RunnerUpPlayerID)

	// 10% to third = $10.00
	assert.Equal(t, int64(1000), distribution.ThirdPlaceAmount.Cents())
	assert.Equal(t, player3, distribution.ThirdPlacePlayerID)
}

func TestDistributionRule_Calculate_TopThreeSplit_TwoPlayers(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleTopThreeSplit
	totalPrize := wallet_vo.NewAmount(100.00)
	player1 := uuid.New()
	player2 := uuid.New()
	rankedPlayers := []uuid.UUID{player1, player2}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)

	// 60% to winner = $60.00
	assert.Equal(t, int64(6000), distribution.WinnerAmount.Cents())

	// 30% + 10% to runner-up = $40.00 (third place share goes to runner-up)
	assert.Equal(t, int64(4000), distribution.RunnerUpAmount.Cents())

	// No third place
	assert.Equal(t, int64(0), distribution.ThirdPlaceAmount.Cents())
}

func TestDistributionRule_Calculate_TopThreeSplit_OnlyOnePlayer(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleTopThreeSplit
	totalPrize := wallet_vo.NewAmount(100.00)
	rankedPlayers := []uuid.UUID{uuid.New()}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 players")
}

func TestDistributionRule_Calculate_PerformanceMVP_WithMVP(t *testing.T) {
	rule := matchmaking_vo.DistributionRulePerformanceMVP
	totalPrize := wallet_vo.NewAmount(100.00)
	player1 := uuid.New()
	player2 := uuid.New()
	mvp := player2 // MVP is runner-up
	rankedPlayers := []uuid.UUID{player1, player2}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, &mvp)

	require.NoError(t, err)

	// 70% to winner = $70.00
	assert.Equal(t, int64(7000), distribution.WinnerAmount.Cents())
	assert.Equal(t, player1, distribution.WinnerPlayerID)

	// 20% to runner-up = $20.00
	assert.Equal(t, int64(2000), distribution.RunnerUpAmount.Cents())
	assert.Equal(t, player2, distribution.RunnerUpPlayerID)

	// 10% MVP bonus = $10.00
	assert.Equal(t, int64(1000), distribution.MVPBonus.Cents())
	assert.Equal(t, mvp, distribution.MVPPlayerID)
}

func TestDistributionRule_Calculate_PerformanceMVP_NoMVP(t *testing.T) {
	rule := matchmaking_vo.DistributionRulePerformanceMVP
	totalPrize := wallet_vo.NewAmount(100.00)
	player1 := uuid.New()
	player2 := uuid.New()
	rankedPlayers := []uuid.UUID{player1, player2}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)

	// 70% + 10% to winner = $80.00 (MVP bonus goes to winner when no MVP)
	assert.Equal(t, int64(8000), distribution.WinnerAmount.Cents())

	// 20% to runner-up = $20.00
	assert.Equal(t, int64(2000), distribution.RunnerUpAmount.Cents())

	// No MVP bonus
	assert.Equal(t, int64(0), distribution.MVPBonus.Cents())
}

func TestDistributionRule_Calculate_PerformanceMVP_OnlyOnePlayer(t *testing.T) {
	rule := matchmaking_vo.DistributionRulePerformanceMVP
	totalPrize := wallet_vo.NewAmount(100.00)
	rankedPlayers := []uuid.UUID{uuid.New()}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 players")
}

func TestDistributionRule_Calculate_ZeroPrize(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleWinnerTakesAll
	totalPrize := wallet_vo.NewAmount(0)
	rankedPlayers := []uuid.UUID{uuid.New()}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestDistributionRule_Calculate_NegativePrize(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleWinnerTakesAll
	totalPrize := wallet_vo.NewAmount(-100)
	rankedPlayers := []uuid.UUID{uuid.New()}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestDistributionRule_Calculate_NoPlayers(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleWinnerTakesAll
	totalPrize := wallet_vo.NewAmount(100)
	rankedPlayers := []uuid.UUID{}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one player")
}

func TestDistributionRule_Calculate_InvalidRule(t *testing.T) {
	rule := matchmaking_vo.DistributionRule("invalid")
	totalPrize := wallet_vo.NewAmount(100)
	rankedPlayers := []uuid.UUID{uuid.New()}

	_, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported distribution rule")
}

func TestPrizeDistribution_GetPayoutForPlayer(t *testing.T) {
	player1 := uuid.New()
	player2 := uuid.New()
	player3 := uuid.New()
	mvp := player2

	distribution := &matchmaking_vo.PrizeDistribution{
		Rule:               matchmaking_vo.DistributionRuleTopThreeSplit,
		Total:              wallet_vo.NewAmount(100),
		WinnerAmount:       wallet_vo.NewAmount(60),
		WinnerPlayerID:     player1,
		RunnerUpAmount:     wallet_vo.NewAmount(30),
		RunnerUpPlayerID:   player2,
		ThirdPlaceAmount:   wallet_vo.NewAmount(10),
		ThirdPlacePlayerID: player3,
		MVPBonus:           wallet_vo.NewAmount(5),
		MVPPlayerID:        mvp,
	}

	// Winner gets $60
	assert.Equal(t, int64(6000), distribution.GetPayoutForPlayer(player1).Cents())

	// Runner-up gets $30 + $5 MVP = $35
	assert.Equal(t, int64(3500), distribution.GetPayoutForPlayer(player2).Cents())

	// Third gets $10
	assert.Equal(t, int64(1000), distribution.GetPayoutForPlayer(player3).Cents())

	// Unknown player gets $0
	assert.Equal(t, int64(0), distribution.GetPayoutForPlayer(uuid.New()).Cents())
}

func TestPrizeDistribution_GetPayoutForPlayer_WinnerIsMVP(t *testing.T) {
	player1 := uuid.New()
	player2 := uuid.New()

	distribution := &matchmaking_vo.PrizeDistribution{
		Rule:             matchmaking_vo.DistributionRulePerformanceMVP,
		Total:            wallet_vo.NewAmount(100),
		WinnerAmount:     wallet_vo.NewAmount(70),
		WinnerPlayerID:   player1,
		RunnerUpAmount:   wallet_vo.NewAmount(20),
		RunnerUpPlayerID: player2,
		MVPBonus:         wallet_vo.NewAmount(10),
		MVPPlayerID:      player1, // Winner is also MVP
	}

	// Winner gets $70 + $10 MVP = $80
	assert.Equal(t, int64(8000), distribution.GetPayoutForPlayer(player1).Cents())

	// Runner-up gets $20
	assert.Equal(t, int64(2000), distribution.GetPayoutForPlayer(player2).Cents())
}

func TestPrizeDistribution_Validate(t *testing.T) {
	tests := []struct {
		name        string
		distribution *matchmaking_vo.PrizeDistribution
		expectError string
	}{
		{
			name: "valid distribution",
			distribution: &matchmaking_vo.PrizeDistribution{
				Rule:           matchmaking_vo.DistributionRuleWinnerTakesAll,
				Total:          wallet_vo.NewAmount(100),
				WinnerAmount:   wallet_vo.NewAmount(100),
				WinnerPlayerID: uuid.New(),
			},
		},
		{
			name: "invalid rule",
			distribution: &matchmaking_vo.PrizeDistribution{
				Rule:           matchmaking_vo.DistributionRule("invalid"),
				Total:          wallet_vo.NewAmount(100),
				WinnerAmount:   wallet_vo.NewAmount(100),
				WinnerPlayerID: uuid.New(),
			},
			expectError: "invalid distribution rule",
		},
		{
			name: "zero total",
			distribution: &matchmaking_vo.PrizeDistribution{
				Rule:           matchmaking_vo.DistributionRuleWinnerTakesAll,
				Total:          wallet_vo.NewAmount(0),
				WinnerAmount:   wallet_vo.NewAmount(0),
				WinnerPlayerID: uuid.New(),
			},
			expectError: "total must be positive",
		},
		{
			name: "nil winner player ID",
			distribution: &matchmaking_vo.PrizeDistribution{
				Rule:           matchmaking_vo.DistributionRuleWinnerTakesAll,
				Total:          wallet_vo.NewAmount(100),
				WinnerAmount:   wallet_vo.NewAmount(100),
				WinnerPlayerID: uuid.Nil,
			},
			expectError: "winner player ID cannot be nil",
		},
		{
			name: "amounts don't sum to total",
			distribution: &matchmaking_vo.PrizeDistribution{
				Rule:           matchmaking_vo.DistributionRuleTopThreeSplit,
				Total:          wallet_vo.NewAmount(100),
				WinnerAmount:   wallet_vo.NewAmount(50), // Should be 60
				RunnerUpAmount: wallet_vo.NewAmount(30),
				WinnerPlayerID: uuid.New(),
			},
			expectError: "distribution amounts sum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.distribution.Validate()
			if tt.expectError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			}
		})
	}
}

func TestPrizeDistribution_Validate_AllowsOneCentRoundingError(t *testing.T) {
	distribution := &matchmaking_vo.PrizeDistribution{
		Rule:           matchmaking_vo.DistributionRuleWinnerTakesAll,
		Total:          wallet_vo.NewAmount(100),
		WinnerAmount:   wallet_vo.NewAmountFromCents(10001), // 1 cent over
		WinnerPlayerID: uuid.New(),
	}

	// Should not error - allows 1 cent rounding error
	err := distribution.Validate()
	assert.NoError(t, err)
}

func TestDistributionRule_Calculate_LargePrizePool(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleTopThreeSplit
	totalPrize := wallet_vo.NewAmount(1000000.00) // $1M prize pool
	player1 := uuid.New()
	player2 := uuid.New()
	player3 := uuid.New()
	rankedPlayers := []uuid.UUID{player1, player2, player3}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)

	// Verify amounts
	assert.Equal(t, int64(60000000), distribution.WinnerAmount.Cents())    // $600,000
	assert.Equal(t, int64(30000000), distribution.RunnerUpAmount.Cents())  // $300,000
	assert.Equal(t, int64(10000000), distribution.ThirdPlaceAmount.Cents()) // $100,000

	// Verify validation passes
	err = distribution.Validate()
	assert.NoError(t, err)
}

func TestDistributionRule_Calculate_SmallPrizePool(t *testing.T) {
	rule := matchmaking_vo.DistributionRuleTopThreeSplit
	totalPrize := wallet_vo.NewAmount(1.00) // $1 prize pool
	player1 := uuid.New()
	player2 := uuid.New()
	player3 := uuid.New()
	rankedPlayers := []uuid.UUID{player1, player2, player3}

	distribution, err := rule.Calculate(totalPrize, rankedPlayers, nil)

	require.NoError(t, err)

	// 60 cents to winner
	assert.Equal(t, int64(60), distribution.WinnerAmount.Cents())
	// 30 cents to runner-up
	assert.Equal(t, int64(30), distribution.RunnerUpAmount.Cents())
	// 10 cents to third
	assert.Equal(t, int64(10), distribution.ThirdPlaceAmount.Cents())
}
