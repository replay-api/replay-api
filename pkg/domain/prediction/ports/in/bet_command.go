package prediction_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
)

// BetCommand defines write operations for bets
type BetCommand interface {
	PlaceBet(ctx context.Context, cmd PlaceBetCommand) (*prediction_entities.Bet, error)
}

// PlaceBetCommand holds the data to place a bet
type PlaceBetCommand struct {
	MarketID  uuid.UUID
	OptionKey string
	Amount    int64 // in cents
}

func (c *PlaceBetCommand) Validate() error {
	if c.MarketID == uuid.Nil {
		return errors.New("market_id is required")
	}
	if c.OptionKey == "" {
		return errors.New("option_key is required")
	}
	if c.Amount < prediction_entities.MinBetAmount {
		return errors.New("bet amount is below minimum")
	}
	if c.Amount > prediction_entities.MaxBetAmount {
		return errors.New("bet amount exceeds maximum")
	}
	return nil
}

// BetQuery defines read operations for bets
type BetQuery interface {
	GetUserBets(ctx context.Context, query GetUserBetsQuery) (*BetListResult, error)
	GetMarketBets(ctx context.Context, marketID uuid.UUID, limit, offset int) (*BetListResult, error)
	GetUserBetSummary(ctx context.Context, marketID uuid.UUID, userID uuid.UUID) (*prediction_entities.UserBetSummary, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*prediction_entities.BetLeaderboardEntry, error)
}

// GetUserBetsQuery defines query params for user bets
type GetUserBetsQuery struct {
	UserID uuid.UUID
	Status prediction_entities.BetStatus // optional filter
	Limit  int
	Offset int
}

func (q *GetUserBetsQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// BetListResult contains paginated bet results
type BetListResult struct {
	Bets       []*prediction_entities.Bet `json:"bets"`
	TotalCount int64                       `json:"total_count"`
	Limit      int                         `json:"limit"`
	Offset     int                         `json:"offset"`
}
