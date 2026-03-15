package prediction_in

import (
	"context"
	"errors"

	"github.com/google/uuid"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
)

// MarketCommand defines write operations for prediction markets
type MarketCommand interface {
	CreateMarket(ctx context.Context, cmd CreateMarketCommand) (*prediction_entities.PredictionMarket, error)
	LockMarket(ctx context.Context, marketID uuid.UUID) error
	ResolveMarket(ctx context.Context, cmd ResolveMarketCommand) error
	CancelMarket(ctx context.Context, marketID uuid.UUID) error
}

// CreateMarketCommand holds the data to create a new market
type CreateMarketCommand struct {
	MatchID     uuid.UUID
	GameID      string
	BetType     prediction_entities.BetType
	Title       string
	Description string
	Options     []prediction_entities.MarketOption
}

func (c *CreateMarketCommand) Validate() error {
	if c.MatchID == uuid.Nil {
		return errors.New("match_id is required")
	}
	if c.GameID == "" {
		return errors.New("game_id is required")
	}
	if !c.BetType.IsValid() {
		return errors.New("invalid bet type")
	}
	if c.Title == "" {
		return errors.New("title is required")
	}
	if len(c.Options) < 2 {
		return errors.New("at least 2 options are required")
	}
	return nil
}

// ResolveMarketCommand holds the data to resolve a market
type ResolveMarketCommand struct {
	MarketID   uuid.UUID
	OutcomeKey string
}

func (c *ResolveMarketCommand) Validate() error {
	if c.MarketID == uuid.Nil {
		return errors.New("market_id is required")
	}
	if c.OutcomeKey == "" {
		return errors.New("outcome_key is required")
	}
	return nil
}

// MarketQuery defines read operations for prediction markets
type MarketQuery interface {
	GetMarket(ctx context.Context, marketID uuid.UUID) (*prediction_entities.PredictionMarket, error)
	ListMatchMarkets(ctx context.Context, query ListMatchMarketsQuery) (*MarketListResult, error)
}

// ListMatchMarketsQuery defines query params
type ListMatchMarketsQuery struct {
	MatchID uuid.UUID
	Status  prediction_entities.PredictionStatus // optional filter
	Limit   int
	Offset  int
}

func (q *ListMatchMarketsQuery) Validate() error {
	if q.MatchID == uuid.Nil {
		return errors.New("match_id is required")
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// MarketListResult contains paginated market results
type MarketListResult struct {
	Markets    []*prediction_entities.PredictionMarket `json:"markets"`
	TotalCount int64                                    `json:"total_count"`
	Limit      int                                      `json:"limit"`
	Offset     int                                      `json:"offset"`
}