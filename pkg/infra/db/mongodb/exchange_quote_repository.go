package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	exchange_entities "github.com/replay-api/replay-api/pkg/domain/exchange/entities"
	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoExchangeQuoteRepository implements exchange_out.QuoteRepository for MongoDB
type MongoExchangeQuoteRepository struct {
	mongodb.MongoDBRepository[*exchange_entities.Quote]
}

// NewExchangeQuoteRepository creates a new MongoDB exchange quote repository
func NewExchangeQuoteRepository(mongoClient *mongo.Client, dbName string) exchange_out.QuoteRepository {
	entityType := &exchange_entities.Quote{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "exchange_quotes", "ExchangeQuote")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"UserID":          true,
		"Side":            true,
		"Pair":            true,
		"BTCPriceUSD":     true,
		"AmountUSD":       true,
		"BTCAmount":       true,
		"FeePercent":      true,
		"FeeAmount":       true,
		"ExpiresAt":       true,
		"Consumed":        true,
		"PriceSource":     true,
		"PriceConfidence": true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":              "baseentity._id",
		"UserID":          "user_id",
		"Side":            "side",
		"Pair":            "pair",
		"BTCPriceUSD":     "btc_price_usd",
		"AmountUSD":       "amount_usd",
		"BTCAmount":       "btc_amount",
		"FeePercent":      "fee_percent",
		"FeeAmount":       "fee_amount",
		"ExpiresAt":       "expires_at",
		"Consumed":        "consumed",
		"PriceSource":     "price_source",
		"PriceConfidence": "price_confidence",
		"CreatedAt":       "baseentity.created_at",
		"UpdatedAt":       "baseentity.updated_at",
	})

	r := &MongoExchangeQuoteRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes on startup
	go r.createIndexes()

	return r
}

func (r *MongoExchangeQuoteRepository) createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetName("idx_exchange_quotes_ttl").
				SetExpireAfterSeconds(0), // TTL index: auto-delete when expires_at is reached
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "baseentity.created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_exchange_quotes_user_created"),
		},
	}

	_, err := r.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("failed to create exchange quote indexes", "error", err)
	} else {
		slog.Info("exchange quote indexes created successfully")
	}
}

// Save creates a new exchange quote record
func (r *MongoExchangeQuoteRepository) Save(ctx context.Context, quote *exchange_entities.Quote) error {
	if quote.GetID() == uuid.Nil {
		return fmt.Errorf("quote ID cannot be nil")
	}

	quote.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, quote)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save exchange quote", "quote_id", quote.ID, "error", err)
		return fmt.Errorf("failed to save exchange quote: %w", err)
	}

	slog.InfoContext(ctx, "exchange quote saved successfully", "quote_id", quote.ID, "side", quote.Side, "expires_at", quote.ExpiresAt)
	return nil
}

// FindByID retrieves an exchange quote by its ID
func (r *MongoExchangeQuoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*exchange_entities.Quote, error) {
	var quote exchange_entities.Quote

	filter := bson.M{"baseentity._id": id}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&quote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("exchange quote not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find exchange quote by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find exchange quote: %w", err)
	}

	return &quote, nil
}

// MarkConsumed marks an exchange quote as consumed (used for an order)
func (r *MongoExchangeQuoteRepository) MarkConsumed(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"baseentity._id": id}
	update := bson.M{
		"$set": bson.M{
			"consumed":            true,
			"baseentity.updated_at": time.Now().UTC(),
		},
	}

	result, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to mark exchange quote as consumed", "quote_id", id, "error", err)
		return fmt.Errorf("failed to mark exchange quote as consumed: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("exchange quote not found: %s", id)
	}

	slog.InfoContext(ctx, "exchange quote marked as consumed", "quote_id", id)
	return nil
}

// Ensure MongoExchangeQuoteRepository implements QuoteRepository
var _ exchange_out.QuoteRepository = (*MongoExchangeQuoteRepository)(nil)
