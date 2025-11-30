// Package db provides MongoDB repository implementations
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const paymentCollectionName = "payments"

// PaymentMongoDBRepository implements PaymentRepository for MongoDB
type PaymentMongoDBRepository struct {
	collection *mongo.Collection
}

// NewPaymentMongoDBRepository creates a new payment repository
func NewPaymentMongoDBRepository(client *mongo.Client, dbName string) payment_out.PaymentRepository {
	collection := client.Database(dbName).Collection(paymentCollectionName)

	// Create indexes
	ctx := context.Background()

	// Index on provider_payment_id for webhook lookups
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "provider_payment_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		slog.Error("Failed to create provider_payment_id index", "err", err)
	}

	// Index on idempotency_key for duplicate detection
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "idempotency_key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		slog.Error("Failed to create idempotency_key index", "err", err)
	}

	// Index on user_id for user payment history
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		slog.Error("Failed to create user_id index", "err", err)
	}

	// Index on wallet_id for wallet payment history
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "wallet_id", Value: 1}},
	})
	if err != nil {
		slog.Error("Failed to create wallet_id index", "err", err)
	}

	// Index on status and created_at for finding pending payments
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "created_at", Value: 1},
		},
	})
	if err != nil {
		slog.Error("Failed to create status+created_at index", "err", err)
	}

	return &PaymentMongoDBRepository{
		collection: collection,
	}
}

// Save creates a new payment record
func (r *PaymentMongoDBRepository) Save(ctx context.Context, payment *payment_entities.Payment) error {
	_, err := r.collection.InsertOne(ctx, payment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("payment already exists with idempotency key: %s", payment.IdempotencyKey)
		}
		return fmt.Errorf("failed to save payment: %w", err)
	}

	slog.InfoContext(ctx, "Payment saved",
		"payment_id", payment.ID,
		"provider", payment.Provider,
		"amount", payment.Amount)

	return nil
}

// FindByID retrieves a payment by its ID
func (r *PaymentMongoDBRepository) FindByID(ctx context.Context, id uuid.UUID) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindByProviderPaymentID retrieves a payment by the provider's payment ID
func (r *PaymentMongoDBRepository) FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	err := r.collection.FindOne(ctx, bson.M{"provider_payment_id": providerPaymentID}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found for provider ID: %s", providerPaymentID)
		}
		return nil, fmt.Errorf("failed to find payment by provider ID: %w", err)
	}

	return &payment, nil
}

// FindByIdempotencyKey retrieves a payment by idempotency key
func (r *PaymentMongoDBRepository) FindByIdempotencyKey(ctx context.Context, key string) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	err := r.collection.FindOne(ctx, bson.M{"idempotency_key": key}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found for idempotency key: %s", key)
		}
		return nil, fmt.Errorf("failed to find payment by idempotency key: %w", err)
	}

	return &payment, nil
}

// FindByUserID retrieves all payments for a user
func (r *PaymentMongoDBRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	filter := bson.M{"user_id": userID}
	r.applyFilters(filter, filters)

	return r.findMany(ctx, filter, filters)
}

// FindByWalletID retrieves all payments for a wallet
func (r *PaymentMongoDBRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	filter := bson.M{"wallet_id": walletID}
	r.applyFilters(filter, filters)

	return r.findMany(ctx, filter, filters)
}

// Update updates an existing payment record
func (r *PaymentMongoDBRepository) Update(ctx context.Context, payment *payment_entities.Payment) error {
	payment.UpdatedAt = time.Now().UTC()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": payment.ID}, payment)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	slog.InfoContext(ctx, "Payment updated",
		"payment_id", payment.ID,
		"status", payment.Status)

	return nil
}

// GetPendingPayments retrieves all pending payments older than specified duration
func (r *PaymentMongoDBRepository) GetPendingPayments(ctx context.Context, olderThanSeconds int) ([]*payment_entities.Payment, error) {
	cutoffTime := time.Now().UTC().Add(-time.Duration(olderThanSeconds) * time.Second)

	filter := bson.M{
		"status": bson.M{"$in": []payment_entities.PaymentStatus{
			payment_entities.PaymentStatusPending,
			payment_entities.PaymentStatusProcessing,
		}},
		"created_at": bson.M{"$lt": cutoffTime},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*payment_entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("failed to decode pending payments: %w", err)
	}

	return payments, nil
}

// applyFilters applies additional filters to the query
func (r *PaymentMongoDBRepository) applyFilters(filter bson.M, filters payment_out.PaymentFilters) {
	if filters.Provider != nil {
		filter["provider"] = *filters.Provider
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.Type != nil {
		filter["type"] = *filters.Type
	}
}

// findMany executes a find query with filters and returns multiple payments
func (r *PaymentMongoDBRepository) findMany(ctx context.Context, filter bson.M, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filters.Limit > 0 {
		opts.SetLimit(int64(filters.Limit))
	} else {
		opts.SetLimit(50) // Default limit
	}

	if filters.Offset > 0 {
		opts.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*payment_entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	return payments, nil
}

// Ensure PaymentMongoDBRepository implements PaymentRepository
var _ payment_out.PaymentRepository = (*PaymentMongoDBRepository)(nil)
