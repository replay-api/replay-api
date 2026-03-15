package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_vo "github.com/replay-api/replay-api/pkg/domain/exchange/value-objects"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoExchangeRateRepository implements exchange_out.ExchangeRateRepository for MongoDB
type MongoExchangeRateRepository struct {
	mongodb.MongoDBRepository[*exchange_entities.ExchangeRate]
}

// NewExchangeRateRepository creates a new MongoDB exchange rate repository
func NewExchangeRateRepository(mongoClient *mongo.Client, dbName string) exchange_out.ExchangeRateRepository {
	entityType := &exchange_entities.ExchangeRate{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "exchange_rates", "ExchangeRate")

	repo.InitQueryableFields(map[string]bool{
		"ID":          true,
		"Pair":        true,
		"MedianPrice": true,
		"Timestamp":   true,
		"Confidence":  true,
		"Spread":      true,
		"CreatedAt":   true,
		"UpdatedAt":   true,
	}, map[string]string{
		"ID":          "baseentity._id",
		"Pair":        "pair",
		"MedianPrice": "median_price",
		"Timestamp":   "timestamp",
		"Confidence":  "confidence",
		"Spread":      "spread",
		"CreatedAt":   "baseentity.created_at",
		"UpdatedAt":   "baseentity.updated_at",
	})

	r := &MongoExchangeRateRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes on startup
	go r.createIndexes()

	return r
}

func (r *MongoExchangeRateRepository) createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "pair", Value: 1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("idx_exchange_rates_pair_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "timestamp", Value: 1},
			},
			Options: options.Index().
				SetName("idx_exchange_rates_ttl").
				SetExpireAfterSeconds(7 * 24 * 60 * 60), // 7 days retention
		},
	}

	_, err := r.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("failed to create exchange rate indexes", "error", err)
	} else {
		slog.Info("exchange rate indexes created successfully")
	}
}

// Save creates a new exchange rate record
func (r *MongoExchangeRateRepository) Save(ctx context.Context, rate *exchange_entities.ExchangeRate) error {
	if rate.GetID() == uuid.Nil {
		return fmt.Errorf("exchange rate ID cannot be nil")
	}

	rate.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, rate)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save exchange rate", "rate_id", rate.ID, "pair", rate.Pair, "error", err)
		return fmt.Errorf("failed to save exchange rate: %w", err)
	}

	slog.InfoContext(ctx, "exchange rate saved successfully", "rate_id", rate.ID, "pair", rate.Pair, "median_price", rate.MedianPrice)
	return nil
}

// FindLatest retrieves the most recent exchange rate for a given pair
func (r *MongoExchangeRateRepository) FindLatest(ctx context.Context, pair exchange_vo.ExchangePair) (*exchange_entities.ExchangeRate, error) {
	var rate exchange_entities.ExchangeRate

	filter := bson.M{"pair": pair}
	findOptions := options.FindOne().
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	err := r.MongoDBRepository.Collection().FindOne(ctx, filter, findOptions).Decode(&rate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("exchange rate not found for pair: %s", pair)
		}
		slog.ErrorContext(ctx, "failed to find latest exchange rate", "pair", pair, "error", err)
		return nil, fmt.Errorf("failed to find latest exchange rate: %w", err)
	}

	return &rate, nil
}

// FindHistory retrieves recent exchange rates for a given pair
func (r *MongoExchangeRateRepository) FindHistory(ctx context.Context, pair exchange_vo.ExchangePair, limit int) ([]*exchange_entities.ExchangeRate, error) {
	filter := bson.M{"pair": pair}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find exchange rate history", "pair", pair, "limit", limit, "error", err)
		return nil, fmt.Errorf("failed to find exchange rate history: %w", err)
	}
	defer cursor.Close(ctx)

	var rates []*exchange_entities.ExchangeRate
	if err := cursor.All(ctx, &rates); err != nil {
		slog.ErrorContext(ctx, "failed to decode exchange rate history", "pair", pair, "error", err)
		return nil, fmt.Errorf("failed to decode exchange rate history: %w", err)
	}

	return rates, nil
}

// Ensure MongoExchangeRateRepository implements ExchangeRateRepository
var _ exchange_out.ExchangeRateRepository = (*MongoExchangeRateRepository)(nil)
