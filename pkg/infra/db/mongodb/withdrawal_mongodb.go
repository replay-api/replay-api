package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const withdrawalCollection = "withdrawals"

// WithdrawalMongoDBRepository implements WithdrawalRepository using MongoDB
type WithdrawalMongoDBRepository struct {
	mongodb.MongoDBRepository[*billing_entities.Withdrawal]
}

// NewWithdrawalMongoDBRepository creates a new WithdrawalMongoDBRepository
func NewWithdrawalMongoDBRepository(client *mongo.Client, dbName string) billing_out.WithdrawalRepository {
	repo := mongodb.NewMongoDBRepository[*billing_entities.Withdrawal](client, dbName, &billing_entities.Withdrawal{}, withdrawalCollection, "Withdrawal")

	repo.InitQueryableFields(map[string]bool{
		"ID":                true,
		"UserID":            true,
		"WalletID":          true,
		"Amount":            true,
		"Currency":          true,
		"Method":            true,
		"Status":            true,
		"Fee":               true,
		"NetAmount":         true,
		"ExternalReference": true,
		"ProviderReference": true,
		"RejectionReason":   true,
		"Notes":             true,
		"ReviewedBy":        true,
		"ReviewedAt":        true,
		"ProcessedAt":       true,
		"CompletedAt":       true,
		"CreatedAt":         true,
		"UpdatedAt":         true,
	}, map[string]string{
		"ID":                "_id",
		"UserID":            "user_id",
		"WalletID":          "wallet_id",
		"Amount":            "amount",
		"Currency":          "currency",
		"Method":            "method",
		"Status":            "status",
		"Fee":               "fee",
		"NetAmount":         "net_amount",
		"ExternalReference": "external_reference",
		"ProviderReference": "provider_reference",
		"RejectionReason":   "rejection_reason",
		"Notes":             "notes",
		"ReviewedBy":        "reviewed_by",
		"ReviewedAt":        "reviewed_at",
		"ProcessedAt":       "processed_at",
		"CompletedAt":       "completed_at",
		"CreatedAt":         "created_at",
		"UpdatedAt":         "updated_at",
	})

	return &WithdrawalMongoDBRepository{
		MongoDBRepository: *repo,
	}
}

// Create stores a new withdrawal
func (r *WithdrawalMongoDBRepository) Create(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error) {
	withdrawal.CreatedAt = time.Now()
	withdrawal.UpdatedAt = time.Now()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, withdrawal)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create withdrawal", "error", err)
		return nil, err
	}

	return withdrawal, nil
}

// Update updates an existing withdrawal
func (r *WithdrawalMongoDBRepository) Update(ctx context.Context, withdrawal *billing_entities.Withdrawal) (*billing_entities.Withdrawal, error) {
	withdrawal.UpdatedAt = time.Now()

	filter := bson.M{"_id": withdrawal.ID}
	update := bson.M{"$set": withdrawal}

	_, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update withdrawal", "error", err, "id", withdrawal.ID)
		return nil, err
	}

	return withdrawal, nil
}

// GetByID retrieves a withdrawal by ID
func (r *WithdrawalMongoDBRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Withdrawal, error) {
	var withdrawal billing_entities.Withdrawal

	err := r.MongoDBRepository.Collection().FindOne(ctx, bson.M{"_id": id}).Decode(&withdrawal)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		slog.ErrorContext(ctx, "Failed to get withdrawal by ID", "error", err, "id", id)
		return nil, err
	}

	return &withdrawal, nil
}

// GetByUserID retrieves withdrawals for a user
func (r *WithdrawalMongoDBRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get withdrawals by user ID", "error", err, "user_id", userID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var withdrawals []billing_entities.Withdrawal
	if err := cursor.All(ctx, &withdrawals); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

// GetByStatus retrieves withdrawals by status
func (r *WithdrawalMongoDBRepository) GetByStatus(ctx context.Context, status billing_entities.WithdrawalStatus, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get withdrawals by status", "error", err, "status", status)
		return nil, err
	}
	defer cursor.Close(ctx)

	var withdrawals []billing_entities.Withdrawal
	if err := cursor.All(ctx, &withdrawals); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

// GetPending retrieves all pending withdrawals
func (r *WithdrawalMongoDBRepository) GetPending(ctx context.Context, limit int, offset int) ([]billing_entities.Withdrawal, error) {
	return r.GetByStatus(ctx, billing_entities.WithdrawalStatusPending, limit, offset)
}

