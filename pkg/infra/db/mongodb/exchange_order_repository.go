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

// MongoExchangeOrderRepository implements exchange_out.OrderRepository for MongoDB
type MongoExchangeOrderRepository struct {
	mongodb.MongoDBRepository[*exchange_entities.Order]
}

// NewExchangeOrderRepository creates a new MongoDB exchange order repository
func NewExchangeOrderRepository(mongoClient *mongo.Client, dbName string) exchange_out.OrderRepository {
	entityType := &exchange_entities.Order{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "exchange_orders", "ExchangeOrder")

	repo.InitQueryableFields(map[string]bool{
		"ID":                    true,
		"UserID":                true,
		"WalletID":              true,
		"Side":                  true,
		"Pair":                  true,
		"Status":                true,
		"ExchangeProvider":      true,
		"IdempotencyKey":        true,
		"RequestedAmountUSD":    true,
		"ExecutedAmountBTC":     true,
		"ExecutedPriceUSD":      true,
		"FeePercent":            true,
		"FeeAmountUSD":          true,
		"NetAmountUSD":          true,
		"CreatedAt":             true,
		"UpdatedAt":             true,
	}, map[string]string{
		"ID":                    "baseentity._id",
		"UserID":                "user_id",
		"WalletID":              "wallet_id",
		"Side":                  "side",
		"Pair":                  "pair",
		"Status":                "status",
		"ExchangeProvider":      "exchange_provider",
		"IdempotencyKey":        "idempotency_key",
		"RequestedAmountUSD":    "requested_amount_usd",
		"ExecutedAmountBTC":     "executed_amount_btc",
		"ExecutedPriceUSD":      "executed_price_usd",
		"FeePercent":            "fee_percent",
		"FeeAmountUSD":          "fee_amount_usd",
		"NetAmountUSD":          "net_amount_usd",
		"CreatedAt":             "baseentity.created_at",
		"UpdatedAt":             "baseentity.updated_at",
	})

	r := &MongoExchangeOrderRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes on startup
	go r.createIndexes()

	return r
}

func (r *MongoExchangeOrderRepository) createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "baseentity.created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_exchange_orders_user_created"),
		},
		{
			Keys: bson.D{
				{Key: "idempotency_key", Value: 1},
			},
			Options: options.Index().SetName("idx_exchange_orders_idempotency_key").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_exchange_orders_status"),
		},
	}

	_, err := r.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("failed to create exchange order indexes", "error", err)
	} else {
		slog.Info("exchange order indexes created successfully")
	}
}

// Save creates a new exchange order record
func (r *MongoExchangeOrderRepository) Save(ctx context.Context, order *exchange_entities.Order) error {
	if order.GetID() == uuid.Nil {
		return fmt.Errorf("order ID cannot be nil")
	}

	order.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, order)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save exchange order", "order_id", order.ID, "error", err)
		return fmt.Errorf("failed to save exchange order: %w", err)
	}

	slog.InfoContext(ctx, "exchange order saved successfully", "order_id", order.ID, "side", order.Side, "status", order.Status)
	return nil
}

// FindByID retrieves an exchange order by its ID
func (r *MongoExchangeOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*exchange_entities.Order, error) {
	var order exchange_entities.Order

	filter := bson.M{"baseentity._id": id}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("exchange order not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find exchange order by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find exchange order: %w", err)
	}

	return &order, nil
}

// FindByUserID retrieves exchange orders for a user with pagination
func (r *MongoExchangeOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*exchange_entities.Order, int, error) {
	filter := bson.M{"user_id": userID}

	// Count total documents matching the filter
	total, err := r.MongoDBRepository.Collection().CountDocuments(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count exchange orders by user ID", "user_id", userID, "error", err)
		return nil, 0, fmt.Errorf("failed to count exchange orders: %w", err)
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}})

	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find exchange orders by user ID", "user_id", userID, "error", err)
		return nil, 0, fmt.Errorf("failed to find exchange orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*exchange_entities.Order
	if err := cursor.All(ctx, &orders); err != nil {
		slog.ErrorContext(ctx, "failed to decode exchange orders", "user_id", userID, "error", err)
		return nil, 0, fmt.Errorf("failed to decode exchange orders: %w", err)
	}

	return orders, int(total), nil
}

// FindByIdempotencyKey retrieves an exchange order by its idempotency key
func (r *MongoExchangeOrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*exchange_entities.Order, error) {
	var order exchange_entities.Order

	filter := bson.M{"idempotency_key": key}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("exchange order not found for idempotency key: %s", key)
		}
		slog.ErrorContext(ctx, "failed to find exchange order by idempotency key", "idempotency_key", key, "error", err)
		return nil, fmt.Errorf("failed to find exchange order: %w", err)
	}

	return &order, nil
}

// Update replaces an existing exchange order record
func (r *MongoExchangeOrderRepository) Update(ctx context.Context, order *exchange_entities.Order) error {
	if order.GetID() == uuid.Nil {
		return fmt.Errorf("order ID cannot be nil")
	}

	order.UpdatedAt = time.Now().UTC()

	filter := bson.M{"baseentity._id": order.ID}
	result, err := r.MongoDBRepository.Collection().ReplaceOne(ctx, filter, order)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update exchange order", "order_id", order.ID, "error", err)
		return fmt.Errorf("failed to update exchange order: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("exchange order not found: %s", order.ID)
	}

	slog.InfoContext(ctx, "exchange order updated successfully", "order_id", order.ID, "status", order.Status)
	return nil
}

// FindActiveByUserID retrieves all non-terminal orders for a user
func (r *MongoExchangeOrderRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*exchange_entities.Order, error) {
	filter := bson.M{
		"user_id": userID,
		"status": bson.M{
			"$nin": []exchange_vo.OrderStatus{
				exchange_vo.OrderStatusCompleted,
				exchange_vo.OrderStatusFailed,
				exchange_vo.OrderStatusCancelled,
				exchange_vo.OrderStatusRefunded,
			},
		},
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}})

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find active exchange orders by user ID", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to find active exchange orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*exchange_entities.Order
	if err := cursor.All(ctx, &orders); err != nil {
		slog.ErrorContext(ctx, "failed to decode active exchange orders", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to decode active exchange orders: %w", err)
	}

	return orders, nil
}

// CountByUserIDSince counts orders for a user since a given time
func (r *MongoExchangeOrderRepository) CountByUserIDSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	filter := bson.M{
		"user_id": userID,
		"baseentity.created_at": bson.M{
			"$gte": since,
		},
	}

	count, err := r.MongoDBRepository.Collection().CountDocuments(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count exchange orders by user ID since", "user_id", userID, "since", since, "error", err)
		return 0, fmt.Errorf("failed to count exchange orders: %w", err)
	}

	return count, nil
}

// Ensure MongoExchangeOrderRepository implements OrderRepository
var _ exchange_out.OrderRepository = (*MongoExchangeOrderRepository)(nil)
