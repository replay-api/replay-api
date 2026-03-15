package prediction_out

import (
	"context"

	"github.com/google/uuid"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
)

// MarketRepository defines persistence for prediction markets
type MarketRepository interface {
	Save(ctx context.Context, market *prediction_entities.PredictionMarket) error
	FindByID(ctx context.Context, id uuid.UUID) (*prediction_entities.PredictionMarket, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID, status string, limit, offset int) ([]*prediction_entities.PredictionMarket, int64, error)
	Update(ctx context.Context, market *prediction_entities.PredictionMarket) error
}

// BetRepository defines persistence for bets
type BetRepository interface {
	Save(ctx context.Context, bet *prediction_entities.Bet) error
	FindByID(ctx context.Context, id uuid.UUID) (*prediction_entities.Bet, error)
	FindByMarketID(ctx context.Context, marketID uuid.UUID, limit, offset int) ([]*prediction_entities.Bet, int64, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*prediction_entities.Bet, int64, error)
	FindByMarketAndUser(ctx context.Context, marketID, userID uuid.UUID) ([]*prediction_entities.Bet, error)
	FindPendingByMarketID(ctx context.Context, marketID uuid.UUID) ([]*prediction_entities.Bet, error)
	Update(ctx context.Context, bet *prediction_entities.Bet) error
	GetLeaderboard(ctx context.Context, limit int) ([]*prediction_entities.BetLeaderboardEntry, error)
}

// PredictionEventPublisher publishes prediction domain events
type PredictionEventPublisher interface {
	PublishBetPlaced(ctx context.Context, bet *prediction_entities.Bet) error
	PublishMarketResolved(ctx context.Context, market *prediction_entities.PredictionMarket) error
}