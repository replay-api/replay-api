package prediction_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// PredictionStatus represents the lifecycle of a prediction
type PredictionStatus string

const (
	PredictionStatusOpen     PredictionStatus = "open"     // accepting bets
	PredictionStatusLocked   PredictionStatus = "locked"   // match started, no more bets
	PredictionStatusResolved PredictionStatus = "resolved" // outcome determined
	PredictionStatusCancelled PredictionStatus = "cancelled" // match cancelled, refund
	PredictionStatusVoided   PredictionStatus = "voided"   // voided by admin
)

// BetType represents the kind of prediction
type BetType string

const (
	BetTypeMatchWinner BetType = "match_winner" // which team wins the match
	BetTypeMapScore    BetType = "map_score"    // exact map score (e.g. 16-10)
	BetTypeTotalRounds BetType = "total_rounds" // over/under total rounds
	BetTypeFirstBlood  BetType = "first_blood"  // which team gets first kill
	BetTypeRoundWinner BetType = "round_winner" // winner of specific round
)

// PredictionMarket represents a single betting market for a match
type PredictionMarket struct {
	shared.BaseEntity `bson:",inline"`

	MatchID     uuid.UUID          `json:"match_id" bson:"match_id"`
	GameID      string             `json:"game_id" bson:"game_id"`
	BetType     BetType            `json:"bet_type" bson:"bet_type"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Options     []MarketOption     `json:"options" bson:"options"`
	Status      PredictionStatus   `json:"status" bson:"status"`
	Outcome     *string            `json:"outcome,omitempty" bson:"outcome,omitempty"` // winning option key
	LockedAt    *time.Time         `json:"locked_at,omitempty" bson:"locked_at,omitempty"`
	ResolvedAt  *time.Time         `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	TotalPool   int64              `json:"total_pool" bson:"total_pool"` // total coins wagered (in cents)
	BetCount    int                `json:"bet_count" bson:"bet_count"`
}

// MarketOption represents one outcome option within a market
type MarketOption struct {
	Key         string  `json:"key" bson:"key"`                 // e.g. "team_a", "over_22.5"
	Label       string  `json:"label" bson:"label"`             // human-readable label
	Odds        float64 `json:"odds" bson:"odds"`               // decimal odds (e.g. 1.85)
	TotalStaked int64   `json:"total_staked" bson:"total_staked"` // total coins staked on this option
	BetCount    int     `json:"bet_count" bson:"bet_count"`
}

// NewPredictionMarket creates a new prediction market
func NewPredictionMarket(
	resourceOwner shared.ResourceOwner,
	matchID uuid.UUID,
	gameID string,
	betType BetType,
	title string,
	description string,
	options []MarketOption,
) (*PredictionMarket, error) {
	if matchID == uuid.Nil {
		return nil, fmt.Errorf("match_id is required")
	}
	if gameID == "" {
		return nil, fmt.Errorf("game_id is required")
	}
	if !betType.IsValid() {
		return nil, fmt.Errorf("invalid bet type: %s", betType)
	}
	if len(options) < 2 {
		return nil, fmt.Errorf("at least 2 options are required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	return &PredictionMarket{
		BaseEntity:  shared.NewEntity(resourceOwner),
		MatchID:     matchID,
		GameID:      gameID,
		BetType:     betType,
		Title:       title,
		Description: description,
		Options:     options,
		Status:      PredictionStatusOpen,
		TotalPool:   0,
		BetCount:    0,
	}, nil
}

// Lock prevents new bets (typically when match starts)
func (m *PredictionMarket) Lock() error {
	if m.Status != PredictionStatusOpen {
		return fmt.Errorf("market is not open (status: %s)", m.Status)
	}
	now := time.Now()
	m.Status = PredictionStatusLocked
	m.LockedAt = &now
	m.UpdatedAt = now
	return nil
}

// Resolve sets the winning outcome and finalises the market
func (m *PredictionMarket) Resolve(outcomeKey string) error {
	if m.Status != PredictionStatusLocked {
		return fmt.Errorf("market must be locked before resolving (status: %s)", m.Status)
	}
	// Validate outcome is a valid option
	found := false
	for _, opt := range m.Options {
		if opt.Key == outcomeKey {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("outcome key %q not found in market options", outcomeKey)
	}

	now := time.Now()
	m.Status = PredictionStatusResolved
	m.Outcome = &outcomeKey
	m.ResolvedAt = &now
	m.UpdatedAt = now
	return nil
}

// Cancel marks the market as cancelled (full refund)
func (m *PredictionMarket) Cancel() error {
	if m.Status == PredictionStatusResolved {
		return fmt.Errorf("cannot cancel a resolved market")
	}
	m.Status = PredictionStatusCancelled
	m.UpdatedAt = time.Now()
	return nil
}

// AddStake increases the pool for an option
func (m *PredictionMarket) AddStake(optionKey string, amount int64) error {
	if m.Status != PredictionStatusOpen {
		return fmt.Errorf("market is not open for betting")
	}
	for i := range m.Options {
		if m.Options[i].Key == optionKey {
			m.Options[i].TotalStaked += amount
			m.Options[i].BetCount++
			m.TotalPool += amount
			m.BetCount++
			m.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("option key %q not found", optionKey)
}

// IsValid checks if the bet type is valid
func (bt BetType) IsValid() bool {
	switch bt {
	case BetTypeMatchWinner, BetTypeMapScore, BetTypeTotalRounds, BetTypeFirstBlood, BetTypeRoundWinner:
		return true
	}
	return false
}

// Payout calculates the payout for a winning bet
func (m *PredictionMarket) Payout(optionKey string, stakeAmount int64) int64 {
	for _, opt := range m.Options {
		if opt.Key == optionKey {
			if opt.Odds <= 0 {
				return stakeAmount // safety: return stake if odds invalid
			}
			return int64(float64(stakeAmount) * opt.Odds)
		}
	}
	return 0
}
