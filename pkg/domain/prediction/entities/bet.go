package prediction_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// BetStatus represents the lifecycle of an individual bet
type BetStatus string

const (
	BetStatusPending  BetStatus = "pending"  // placed but market not resolved
	BetStatusWon      BetStatus = "won"      // bet won
	BetStatusLost     BetStatus = "lost"     // bet lost
	BetStatusRefunded BetStatus = "refunded" // market cancelled/voided
)

const (
	MinBetAmount int64 = 100    // $1.00 (in cents)
	MaxBetAmount int64 = 10000  // $100.00 (in cents)
)

// Bet represents an individual user's bet on a market option
type Bet struct {
	shared.BaseEntity `bson:",inline"`

	MarketID    uuid.UUID `json:"market_id" bson:"market_id"`
	MatchID     uuid.UUID `json:"match_id" bson:"match_id"`
	UserID      uuid.UUID `json:"user_id" bson:"user_id"`
	OptionKey   string    `json:"option_key" bson:"option_key"`
	Amount      int64     `json:"amount" bson:"amount"`         // wager in cents
	OddsAtPlace float64   `json:"odds_at_place" bson:"odds_at_place"` // odds when bet was placed
	Status      BetStatus `json:"status" bson:"status"`
	Payout      int64     `json:"payout" bson:"payout"`         // 0 until resolved
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
}

// NewBet creates a new bet
func NewBet(
	resourceOwner shared.ResourceOwner,
	marketID uuid.UUID,
	matchID uuid.UUID,
	userID uuid.UUID,
	optionKey string,
	amount int64,
	oddsAtPlace float64,
) (*Bet, error) {
	if marketID == uuid.Nil {
		return nil, fmt.Errorf("market_id is required")
	}
	if matchID == uuid.Nil {
		return nil, fmt.Errorf("match_id is required")
	}
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	if optionKey == "" {
		return nil, fmt.Errorf("option_key is required")
	}
	if amount < MinBetAmount {
		return nil, fmt.Errorf("minimum bet is $%.2f", float64(MinBetAmount)/100)
	}
	if amount > MaxBetAmount {
		return nil, fmt.Errorf("maximum bet is $%.2f", float64(MaxBetAmount)/100)
	}
	if oddsAtPlace <= 0 {
		return nil, fmt.Errorf("odds must be positive")
	}

	return &Bet{
		BaseEntity:  shared.NewEntity(resourceOwner),
		MarketID:    marketID,
		MatchID:     matchID,
		UserID:      userID,
		OptionKey:   optionKey,
		Amount:      amount,
		OddsAtPlace: oddsAtPlace,
		Status:      BetStatusPending,
		Payout:      0,
	}, nil
}

// Resolve marks the bet as won or lost and sets payout
func (b *Bet) Resolve(winningOptionKey string) {
	now := time.Now()
	b.ResolvedAt = &now
	b.UpdatedAt = now

	if b.OptionKey == winningOptionKey {
		b.Status = BetStatusWon
		b.Payout = int64(float64(b.Amount) * b.OddsAtPlace)
	} else {
		b.Status = BetStatusLost
		b.Payout = 0
	}
}

// Refund marks the bet as refunded
func (b *Bet) Refund() {
	now := time.Now()
	b.Status = BetStatusRefunded
	b.Payout = b.Amount // return stake
	b.ResolvedAt = &now
	b.UpdatedAt = now
}

// UserBetSummary provides a summary of a user's betting on a market
type UserBetSummary struct {
	MarketID    uuid.UUID `json:"market_id" bson:"market_id"`
	UserID      uuid.UUID `json:"user_id" bson:"user_id"`
	TotalStaked int64     `json:"total_staked" bson:"total_staked"`
	TotalPayout int64     `json:"total_payout" bson:"total_payout"`
	BetCount    int       `json:"bet_count" bson:"bet_count"`
	Bets        []*Bet    `json:"bets" bson:"bets"`
}

// BetLeaderboardEntry represents a user's ranking on the prediction leaderboard
type BetLeaderboardEntry struct {
	UserID      uuid.UUID `json:"user_id" bson:"user_id"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty" bson:"avatar_url,omitempty"`
	TotalBets   int       `json:"total_bets" bson:"total_bets"`
	WinCount    int       `json:"win_count" bson:"win_count"`
	WinRate     float64   `json:"win_rate" bson:"win_rate"`
	TotalProfit int64     `json:"total_profit" bson:"total_profit"` // net profit in cents
}
