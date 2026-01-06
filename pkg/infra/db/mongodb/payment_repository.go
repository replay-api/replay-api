package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoPaymentRepository implements PaymentRepository for MongoDB
type MongoPaymentRepository struct {
	mongodb.MongoDBRepository[*payment_entities.Payment]
}

// NewPaymentRepository creates a new MongoDB payment repository
func NewPaymentRepository(mongoClient *mongo.Client, dbName string) payment_out.PaymentRepository {
	entityType := &payment_entities.Payment{}
	repo := mongodb.NewMongoDBRepository[*payment_entities.Payment](mongoClient, dbName, entityType, "payments", "Payment")

	repo.InitQueryableFields(map[string]bool{
		"ID":                       true,
		"PayableID":                true,
		"Reference":                true,
		"Amount":                   true,
		"Currency":                 true,
		"Option":                   true,
		"Status":                   true,
		"Provider":                 true,
		"PaymentProviderReference": true,
		"Description":              true,
		"CreatedAt":                true,
		"UpdatedAt":                true,
	}, map[string]string{
		"ID":                       "_id",
		"PayableID":                "payable_id",
		"Reference":                "reference",
		"Amount":                   "amount",
		"Currency":                 "currency",
		"Option":                   "option",
		"Status":                   "status",
		"Provider":                 "provider",
		"PaymentProviderReference": "payment_provider_reference",
		"Description":              "description",
		"CreatedAt":                "created_at",
		"UpdatedAt":                "updated_at",
	})

	mongoPaymentRepo := &MongoPaymentRepository{
		MongoDBRepository: *repo,
	}

	// Create indexes on startup
	go mongoPaymentRepo.createIndexes()

	return mongoPaymentRepo
}

func (r *MongoPaymentRepository) createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "wallet_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "provider_payment_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "idempotency_key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	}

	_, err := r.MongoDBRepository.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("failed to create payment indexes", "error", err)
	} else {
		slog.Info("payment indexes created successfully")
	}
}

// Save creates a new payment record
func (r *MongoPaymentRepository) Save(ctx context.Context, payment *payment_entities.Payment) error {
	if payment.ID == uuid.Nil {
		return fmt.Errorf("payment ID cannot be nil")
	}

	payment.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, payment)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save payment", "payment_id", payment.ID, "error", err)
		return fmt.Errorf("failed to save payment: %w", err)
	}

	slog.InfoContext(ctx, "payment saved successfully", "payment_id", payment.ID, "status", payment.Status)
	return nil
}

// FindByID retrieves a payment by its ID
func (r *MongoPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	filter := bson.M{"_id": id}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find payment by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindByProviderPaymentID retrieves a payment by the provider's payment ID
func (r *MongoPaymentRepository) FindByProviderPaymentID(ctx context.Context, providerPaymentID string) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	filter := bson.M{"provider_payment_id": providerPaymentID}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found for provider payment ID: %s", providerPaymentID)
		}
		slog.ErrorContext(ctx, "failed to find payment by provider payment ID", "provider_payment_id", providerPaymentID, "error", err)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindByIdempotencyKey retrieves a payment by idempotency key
func (r *MongoPaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*payment_entities.Payment, error) {
	var payment payment_entities.Payment

	filter := bson.M{"idempotency_key": key}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found for idempotency key: %s", key)
		}
		slog.ErrorContext(ctx, "failed to find payment by idempotency key", "idempotency_key", key, "error", err)
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindByUserID retrieves all payments for a user
func (r *MongoPaymentRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	filter := bson.M{"user_id": userID}

	// Apply optional filters
	if filters.Provider != nil {
		filter["provider"] = *filters.Provider
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.Type != nil {
		filter["type"] = *filters.Type
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filters.Limit > 0 {
		findOptions.SetLimit(int64(filters.Limit))
	}
	if filters.Offset > 0 {
		findOptions.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find payments by user ID", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*payment_entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		slog.ErrorContext(ctx, "failed to decode payments", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	return payments, nil
}

// FindByWalletID retrieves all payments for a wallet
func (r *MongoPaymentRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID, filters payment_out.PaymentFilters) ([]*payment_entities.Payment, error) {
	filter := bson.M{"wallet_id": walletID}

	// Apply optional filters
	if filters.Provider != nil {
		filter["provider"] = *filters.Provider
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.Type != nil {
		filter["type"] = *filters.Type
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filters.Limit > 0 {
		findOptions.SetLimit(int64(filters.Limit))
	}
	if filters.Offset > 0 {
		findOptions.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find payments by wallet ID", "wallet_id", walletID, "error", err)
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*payment_entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		slog.ErrorContext(ctx, "failed to decode payments", "wallet_id", walletID, "error", err)
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	return payments, nil
}

// Update updates an existing payment record
func (r *MongoPaymentRepository) Update(ctx context.Context, payment *payment_entities.Payment) error {
	payment.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": payment.ID}
	result, err := r.MongoDBRepository.Collection().ReplaceOne(ctx, filter, payment)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update payment", "payment_id", payment.ID, "error", err)
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	slog.InfoContext(ctx, "payment updated successfully", "payment_id", payment.ID, "status", payment.Status)
	return nil
}

// GetPendingPayments retrieves all pending payments older than specified duration
func (r *MongoPaymentRepository) GetPendingPayments(ctx context.Context, olderThanSeconds int) ([]*payment_entities.Payment, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(olderThanSeconds) * time.Second)

	filter := bson.M{
		"status": bson.M{
			"$in": []payment_entities.PaymentStatus{
				payment_entities.PaymentStatusPending,
				payment_entities.PaymentStatusProcessing,
			},
		},
		"created_at": bson.M{"$lt": cutoff},
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pending payments", "older_than_seconds", olderThanSeconds, "error", err)
		return nil, fmt.Errorf("failed to find pending payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*payment_entities.Payment
	if err := cursor.All(ctx, &payments); err != nil {
		slog.ErrorContext(ctx, "failed to decode pending payments", "error", err)
		return nil, fmt.Errorf("failed to decode pending payments: %w", err)
	}

	return payments, nil
}

// Ensure MongoPaymentRepository implements PaymentRepository
var _ payment_out.PaymentRepository = (*MongoPaymentRepository)(nil)
