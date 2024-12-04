package state

import (
	cs2 "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

// GameSenseState represents the player's current game sense level
type GameSenseState string

const (
	GameSenseLow      GameSenseState = "Low"
	GameSenseMedium   GameSenseState = "Medium"
	GameSenseHigh     GameSenseState = "High"
	GameSenseVeryHigh GameSenseState = "VeryHigh"
)

// CS2GameSenseStats tracks game sense-related statistics
type CS2GameSenseStats struct {
	TotalDecisions            int     // Total number of decisions made (e.g., peeks, rotations, utility usage)
	PositiveDecisions         int     // Number of decisions considered "good" based on outcome or expert analysis
	NegativeDecisions         int     // Number of decisions considered "bad"
	ClutchDecisions           int     // Number of decisions made in clutch situations
	SuccessfulClutchDecisions int     // Number of clutch decisions that led to positive outcomes
	DecisionAccuracy          float64 // Percentage of good decisions
	ClutchDecisionAccuracy    float64 // Percentage of successful clutch decisions
	// Add more metrics as needed (e.g., reaction time, information usage)
}

// CS2GameSenseContext holds the state and stats for a player's game sense
type CS2GameSenseContext struct {
	Player *cs2.Player
	State  GameSenseState
	Stats  *CS2GameSenseStats
}

func NewCS2GameSenseContext(player *cs2.Player) *CS2GameSenseContext {
	return &CS2GameSenseContext{
		Player: player,
		State:  GameSenseMedium, // Start with a neutral assumption
		Stats: &CS2GameSenseStats{
			TotalDecisions:            0,
			PositiveDecisions:         0,
			NegativeDecisions:         0,
			ClutchDecisions:           0,
			SuccessfulClutchDecisions: 0,
			DecisionAccuracy:          0.0,
			ClutchDecisionAccuracy:    0.0,
		},
	}
}

// UpdateState determines the current GameSenseState based on Stats
func (c *CS2GameSenseContext) UpdateState() {
	// Simple example (you'll need to refine this logic)
	accuracy := c.Stats.DecisionAccuracy
	if accuracy < 0.3 {
		c.State = GameSenseLow
	} else if accuracy < 0.6 {
		c.State = GameSenseMedium
	} else if accuracy < 0.8 {
		c.State = GameSenseHigh
	} else {
		c.State = GameSenseVeryHigh
	}
}

func (c *CS2GameSenseContext) RegisterDecision(isPositiveDecision bool, isClutch bool) {
	c.Stats.TotalDecisions++
	if isPositiveDecision {
		c.Stats.PositiveDecisions++
	} else {
		c.Stats.NegativeDecisions++
	}

	// aqui é importante pois dá pra mensurar melhor a tomada de decisão em situações de clutch
	if isClutch {
		c.Stats.ClutchDecisions++
		if isPositiveDecision {
			c.Stats.SuccessfulClutchDecisions++
		}
	}
}

// type CS2GameSenseContext struct {
// 	Player        *cs2.Player
// 	State         GameSenseState
// 	Stats         *CS2GameSenseStats
// 	lastFlashTime time.Time // Track last flash time for timing decisions
// }

// func NewCS2GameSenseContext(player *cs2.Player) *CS2GameSenseContext {
// 	return &CS2GameSenseContext{
// 		Player:        player,
// 		State:         GameSenseMedium,
// 		Stats:         &CS2GameSenseStats{},
// 		lastFlashTime: time.Time{}, // Initialize to zero time
// 	}
// }

// // UpdateState determines the current GameSenseState based on Stats
// func (c *CS2GameSenseContext) UpdateState() {
// 	// Weigh different aspects of game sense based on their importance
// 	weightedScore := 0.4*c.Stats.DecisionAccuracy +
// 		0.3*c.Stats.ClutchDecisionAccuracy +
// 		0.2*(1-c.Stats.DeathByBadPositioningRatio) +
// 		0.1*c.Stats.UtilityUsageEfficiency

// 	// Classify the player's game sense level
// 	switch {
// 	case weightedScore < 0.3:
// 		c.State = GameSenseLow
// 	case weightedScore < 0.6:
// 		c.State = GameSenseMedium
// 	case weightedScore < 0.8:
// 		c.State = GameSenseHigh
// 	default:
// 		c.State = GameSenseVeryHigh
// 	}
// }

// func (c *CS2GameSenseContext) RegisterKill(e event.Kill, p *cs2.Parser) {
// 	isHeadshot := e.HitGroup == cs2.HitGroupHead
// 	isEntryFrag := len(p.GameState().Team(e.Killer.Team).MembersAlive()) == 5
// 	wasOpponentFlashed := p.CurrentTime().Sub(c.lastFlashTime) < 5*time.Second // Check if Opponent was flashed recently

// 	isBonusFrag := isHeadshot || isEntryFrag || wasOpponentFlashed
// 	c.Stats.TotalFrags++
// 	if isBonusFrag {
// 		c.Stats.BonusFrags++
// 	}

// 	c.RegisterDecision(isBonusFrag, false) // Not a clutch situation
// }

// func (c *CS2GameSenseContext) RegisterElimination(event, p *cs2.Parser) {
// 	// Check for deaths due to bad positioning (you'll need to define criteria for this)
// 	// ...

// 	c.Stats.TotalEliminations++
// 	if isFaultElimination {
// 		c.Stats.EliminatedByBadPositioning++
// 	}
// }

// func (c *CS2GameSenseContext) RegisterFlash(e event..) {
// 	c.lastFlashTime = e.Time

// }
