package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_out "github.com/replay-api/replay-api/pkg/domain/prediction/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionPredictionMarkets = "prediction_markets"
	CollectionBets              = "bets"
)

// --- Market Repository ---

type MarketMongoRepository struct {
	collection *mongo.Collection
}

func NewMarketMongoRepository(db *mongo.Database) prediction_out.MarketRepository {
	return &MarketMongoRepository{
		collection: db.Collection(CollectionPredictionMarkets),
	}
}

func (r *MarketMongoRepository) Save(ctx context.Context, market *prediction_entities.PredictionMarket) error {
	_, err := r.collection.InsertOne(ctx, market)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to insert prediction market", "error", err)
		return fmt.Errorf("failed to save prediction market: %w", err)
	}
	return nil
}

func (r *MarketMongoRepository) FindByID(ctx context.Context, id uuid.UUID) (*prediction_entities.PredictionMarket, error) {
	var market prediction_entities.PredictionMarket
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&market)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("prediction market not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find prediction market: %w", err)
	}
	return &market, nil
}

func (r *MarketMongoRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID, status string, limit, offset int) ([]*prediction_entities.PredictionMarket, int64, error) {
	filter := bson.M{"match_id": matchID}
	if status != "" {
		filter["status"] = status
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count prediction markets: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find prediction markets: %w", err)
	}
	defer cursor.Close(ctx)

	var markets []*prediction_entities.PredictionMarket
	if err := cursor.All(ctx, &markets); err != nil {
		return nil, 0, fmt.Errorf("failed to decode prediction markets: %w", err)
	}

	return markets, total, nil
}

func (r *MarketMongoRepository) Update(ctx context.Context, market *prediction_entities.PredictionMarket) error {
	market.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": market.ID}, market)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update prediction market", "error", err, "market_id", market.ID)
		return fmt.Errorf("failed to update prediction market: %w", err)
	}
	return nil
}

// --- Bet Repository ---

type BetMongoRepository struct {
	collection *mongo.Collection
}

func NewBetMongoRepository(db *mongo.Database) prediction_out.BetRepository {
	return &BetMongoRepository{
		collection: db.Collection(CollectionBets),
	}
}

func (r *BetMongoRepository) Save(ctx context.Context, bet *prediction_entities.Bet) error {
	_, err := r.collection.InsertOne(ctx, bet)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to insert bet", "error", err)
		return fmt.Errorf("failed to save bet: %w", err)
	}
	return nil
}

func (r *BetMongoRepository) FindByID(ctx context.Context, id uuid.UUID) (*prediction_entities.Bet, error) {
	var bet prediction_entities.Bet
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&bet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("bet not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find bet: %w", err)
	}
	return &bet, nil
}

func (r *BetMongoRepository) FindByMarketID(ctx context.Context, marketID uuid.UUID, limit, offset int) ([]*prediction_entities.Bet, int64, error) {
	filter := bson.M{"market_id": marketID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bets: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find bets: %w", err)
	}
	defer cursor.Close(ctx)

	var bets []*prediction_entities.Bet
	if err := cursor.All(ctx, &bets); err != nil {
		return nil, 0, fmt.Errorf("failed to decode bets: %w", err)
	}

	return bets, total, nil
}

func (r *BetMongoRepository) FindByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*prediction_entities.Bet, int64, error) {
	filter := bson.M{"user_id": userID}
	if status != "" {
		filter["status"] = status
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user bets: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find user bets: %w", err)
	}
	defer cursor.Close(ctx)

	var bets []*prediction_entities.Bet
	if err := cursor.All(ctx, &bets); err != nil {
		return nil, 0, fmt.Errorf("failed to decode user bets: %w", err)
	}

	return bets, total, nil
}

func (r *BetMongoRepository) FindByMarketAndUser(ctx context.Context, marketID, userID uuid.UUID) ([]*prediction_entities.Bet, error) {
	filter := bson.M{
		"market_id": marketID,
		"user_id":   userID,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find user bets for market: %w", err)
	}
	defer cursor.Close(ctx)

	var bets []*prediction_entities.Bet
	if err := cursor.All(ctx, &bets); err != nil {
		return nil, fmt.Errorf("failed to decode user bets: %w", err)
	}

	return bets, nil
}

func (r *BetMongoRepository) FindPendingByMarketID(ctx context.Context, marketID uuid.UUID) ([]*prediction_entities.Bet, error) {
	filter := bson.M{
		"market_id": marketID,
		"status":    string(prediction_entities.BetStatusPending),
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending bets: %w", err)
	}
	defer cursor.Close(ctx)

	var bets []*prediction_entities.Bet
	if err := cursor.All(ctx, &bets); err != nil {
		return nil, fmt.Errorf("failed to decode pending bets: %w", err)
	}

	return bets, nil
}

func (r *BetMongoRepository) Update(ctx context.Context, bet *prediction_entities.Bet) error {
	bet.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": bet.ID}, bet)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update bet", "error", err, "bet_id", bet.ID)
		return fmt.Errorf("failed to update bet: %w", err)
	}
	return nil
}

func (r *BetMongoRepository) GetLeaderboard(ctx context.Context, limit int) ([]*prediction_entities.BetLeaderboardEntry, error) {
	pipeline := mongo.Pipeline{
		// Only resolved bets
		{{Key: "$match", Value: bson.M{"status": bson.M{"$in": []string{"won", "lost"}}}}},
		// Group by user
		{{Key: "$group", Value: bson.M{
			"_id":          "$user_id",
			"total_bets":   bson.M{"$sum": 1},
			"win_count":    bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$status", "won"}}, 1, 0}}},
			"total_profit": bson.M{"$sum": bson.M{"$subtract": bson.A{"$payout", "$amount"}}},
		}}},
		// Calculate win rate
		{{Key: "$addFields", Value: bson.M{
			"win_rate": bson.M{"$cond": bson.A{
				bson.M{"$gt": bson.A{"$total_bets", 0}},
				bson.M{"$divide": bson.A{"$win_count", "$total_bets"}},
				0,
			}},
		}}},
		// Sort by profit
		{{Key: "$sort", Value: bson.D{{Key: "total_profit", Value: -1}}}},
		// Limit
		{{Key: "$limit", Value: limit}},
		// Project
		{{Key: "$project", Value: bson.M{
			"user_id":      "$_id",
			"total_bets":   1,
			"win_count":    1,
			"win_rate":     1,
			"total_profit": 1,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*prediction_entities.BetLeaderboardEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode leaderboard: %w", err)
	}

	return entries, nil
}
