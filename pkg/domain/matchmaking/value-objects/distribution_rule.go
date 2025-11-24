package matchmaking_vo

import (
	"fmt"

	"github.com/google/uuid"
	wallet_vo "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/value-objects"
)

// DistributionRule defines how prize money is distributed
type DistributionRule string

const (
	DistributionRuleWinnerTakesAll DistributionRule = "winner_takes_all"         // 100% to winner
	DistributionRuleTopThreeSplit  DistributionRule = "top_three_split_60_30_10" // 60% winner, 30% runner-up, 10% 3rd
	DistributionRulePerformanceMVP DistributionRule = "performance_mvp_70_20_10" // 70% winner, 20% runner-up, 10% MVP
)

// AllDistributionRules returns all available distribution rules
func AllDistributionRules() []DistributionRule {
	return []DistributionRule{
		DistributionRuleWinnerTakesAll,
		DistributionRuleTopThreeSplit,
		DistributionRulePerformanceMVP,
	}
}

// IsValid checks if the distribution rule is valid
func (d DistributionRule) IsValid() bool {
	switch d {
	case DistributionRuleWinnerTakesAll, DistributionRuleTopThreeSplit, DistributionRulePerformanceMVP:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (d DistributionRule) String() string {
	return string(d)
}

// DisplayName returns a user-friendly name
func (d DistributionRule) DisplayName() string {
	switch d {
	case DistributionRuleWinnerTakesAll:
		return "Winner Takes All"
	case DistributionRuleTopThreeSplit:
		return "Top 3 Split"
	case DistributionRulePerformanceMVP:
		return "Performance MVP"
	default:
		return "Unknown"
	}
}

// Description returns a description of the rule
func (d DistributionRule) Description() string {
	switch d {
	case DistributionRuleWinnerTakesAll:
		return "100% of the prize pool goes to the winner. High risk, high reward!"
	case DistributionRuleTopThreeSplit:
		return "Prize split: 60% winner, 30% runner-up, 10% third place. Competitive balance."
	case DistributionRulePerformanceMVP:
		return "70% winner, 20% runner-up, 10% MVP bonus. Skill-based rewards."
	default:
		return ""
	}
}

// Icon returns an emoji icon for the rule
func (d DistributionRule) Icon() string {
	switch d {
	case DistributionRuleWinnerTakesAll:
		return "🏆"
	case DistributionRuleTopThreeSplit:
		return "🥇"
	case DistributionRulePerformanceMVP:
		return "🎯"
	default:
		return "💰"
	}
}

// Calculate calculates the prize distribution based on the rule
func (d DistributionRule) Calculate(
	totalPrize wallet_vo.Amount,
	rankedPlayerIDs []uuid.UUID,
	mvpPlayerID *uuid.UUID,
) (*PrizeDistribution, error) {
	if totalPrize.IsNegative() || totalPrize.IsZero() {
		return nil, fmt.Errorf("total prize must be positive, got: %s", totalPrize.String())
	}

	if len(rankedPlayerIDs) < 1 {
		return nil, fmt.Errorf("must have at least one player (winner)")
	}

	distribution := &PrizeDistribution{
		Rule:  d,
		Total: totalPrize,
	}

	switch d {
	case DistributionRuleWinnerTakesAll:
		distribution.WinnerAmount = totalPrize
		distribution.WinnerPlayerID = rankedPlayerIDs[0]

	case DistributionRuleTopThreeSplit:
		if len(rankedPlayerIDs) < 2 {
			return nil, fmt.Errorf("top three split requires at least 2 players")
		}

		distribution.WinnerAmount = totalPrize.Percentage(60.0)
		distribution.RunnerUpAmount = totalPrize.Percentage(30.0)
		distribution.WinnerPlayerID = rankedPlayerIDs[0]
		distribution.RunnerUpPlayerID = rankedPlayerIDs[1]

		if len(rankedPlayerIDs) >= 3 {
			distribution.ThirdPlaceAmount = totalPrize.Percentage(10.0)
			distribution.ThirdPlacePlayerID = rankedPlayerIDs[2]
		} else {
			// If only 2 players, give 3rd place share to runner-up
			distribution.RunnerUpAmount = distribution.RunnerUpAmount.Add(totalPrize.Percentage(10.0))
		}

	case DistributionRulePerformanceMVP:
		if len(rankedPlayerIDs) < 2 {
			return nil, fmt.Errorf("performance MVP requires at least 2 players")
		}

		distribution.WinnerAmount = totalPrize.Percentage(70.0)
		distribution.RunnerUpAmount = totalPrize.Percentage(20.0)
		distribution.WinnerPlayerID = rankedPlayerIDs[0]
		distribution.RunnerUpPlayerID = rankedPlayerIDs[1]

		if mvpPlayerID != nil {
			distribution.MVPBonus = totalPrize.Percentage(10.0)
			distribution.MVPPlayerID = *mvpPlayerID
		} else {
			// If no MVP, give bonus to winner
			distribution.WinnerAmount = distribution.WinnerAmount.Add(totalPrize.Percentage(10.0))
		}

	default:
		return nil, fmt.Errorf("unsupported distribution rule: %s", d)
	}

	// Validate total equals sum of parts (allow 1 cent rounding error)
	calculatedTotal := distribution.WinnerAmount.
		Add(distribution.RunnerUpAmount).
		Add(distribution.ThirdPlaceAmount).
		Add(distribution.MVPBonus)

	diff := calculatedTotal.Subtract(totalPrize).Abs()
	if diff.Cents() > 1 {
		return nil, fmt.Errorf("distribution total %s does not match prize total %s",
			calculatedTotal.String(), totalPrize.String())
	}

	return distribution, nil
}

// PrizeDistribution represents calculated prize amounts
type PrizeDistribution struct {
	Rule               DistributionRule `json:"rule" bson:"rule"`
	Total              wallet_vo.Amount `json:"total" bson:"total"`
	WinnerAmount       wallet_vo.Amount `json:"winner_amount" bson:"winner_amount"`
	WinnerPlayerID     uuid.UUID        `json:"winner_player_id" bson:"winner_player_id"`
	RunnerUpAmount     wallet_vo.Amount `json:"runner_up_amount" bson:"runner_up_amount"`
	RunnerUpPlayerID   uuid.UUID        `json:"runner_up_player_id,omitempty" bson:"runner_up_player_id,omitempty"`
	ThirdPlaceAmount   wallet_vo.Amount `json:"third_place_amount" bson:"third_place_amount"`
	ThirdPlacePlayerID uuid.UUID        `json:"third_place_player_id,omitempty" bson:"third_place_player_id,omitempty"`
	MVPBonus           wallet_vo.Amount `json:"mvp_bonus" bson:"mvp_bonus"`
	MVPPlayerID        uuid.UUID        `json:"mvp_player_id,omitempty" bson:"mvp_player_id,omitempty"`
}

// GetPayoutForPlayer returns the payout amount for a specific player
func (pd *PrizeDistribution) GetPayoutForPlayer(playerID uuid.UUID) wallet_vo.Amount {
	total := wallet_vo.NewAmount(0)

	if pd.WinnerPlayerID == playerID {
		total = total.Add(pd.WinnerAmount)
	}
	if pd.RunnerUpPlayerID == playerID {
		total = total.Add(pd.RunnerUpAmount)
	}
	if pd.ThirdPlacePlayerID == playerID {
		total = total.Add(pd.ThirdPlaceAmount)
	}
	if pd.MVPPlayerID == playerID {
		total = total.Add(pd.MVPBonus)
	}

	return total
}

// Validate ensures the distribution is valid
func (pd *PrizeDistribution) Validate() error {
	if !pd.Rule.IsValid() {
		return fmt.Errorf("invalid distribution rule: %s", pd.Rule)
	}

	if pd.Total.IsNegative() || pd.Total.IsZero() {
		return fmt.Errorf("total must be positive")
	}

	if pd.WinnerPlayerID == uuid.Nil {
		return fmt.Errorf("winner player ID cannot be nil")
	}

	// Calculate sum
	sum := pd.WinnerAmount.
		Add(pd.RunnerUpAmount).
		Add(pd.ThirdPlaceAmount).
		Add(pd.MVPBonus)

	// Allow 1 cent rounding error
	diff := sum.Subtract(pd.Total).Abs()
	if diff.Cents() > 1 {
		return fmt.Errorf("distribution amounts sum to %s but total is %s", sum.String(), pd.Total.String())
	}

	return nil
}
