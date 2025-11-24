package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoWalletRepository struct {
	*MongoDBRepository[*wallet_entities.UserWallet]
}

func NewMongoWalletRepository(mongoClient *mongo.Client, dbName string) wallet_out.WalletRepository {
	mappingCache := make(map[string]CacheItem)
	entityModel := reflect.TypeOf(wallet_entities.UserWallet{})
	repo := &MongoWalletRepository{
		MongoDBRepository: &MongoDBRepository[*wallet_entities.UserWallet]{
			mongoClient:       mongoClient,
			dbName:            dbName,
			mappingCache:      mappingCache,
			entityModel:       entityModel,
			collectionName:    "wallets",
			entityName:        "UserWallet",
			bsonFieldMappings: make(map[string]string),
			queryableFields:   make(map[string]bool),
		},
	}

	// Define BSON field mappings
	bsonFieldMappings := map[string]string{
		"ID":                  "_id",
		"EVMAddress":          "evm_address",
		"Balances":            "balances",
		"PendingTransactions": "pending_transactions",
		"TotalDeposited":      "total_deposited",
		"TotalWithdrawn":      "total_withdrawn",
		"TotalPrizesWon":      "total_prizes_won",
		"DailyPrizeWinnings":  "daily_prize_winnings",
		"LastPrizeWinDate":    "last_prize_win_date",
		"IsLocked":            "is_locked",
		"LockReason":          "lock_reason",
		"CreatedAt":           "created_at",
		"UpdatedAt":           "updated_at",
		"ResourceOwner":       "resource_owner",
	}

	// Define queryable fields for search operations
	queryableFields := map[string]bool{
		"EVMAddress":         true,
		"TotalDeposited":     true,
		"TotalWithdrawn":     true,
		"TotalPrizesWon":     true,
		"DailyPrizeWinnings": true,
		"LastPrizeWinDate":   true,
		"IsLocked":           true,
		"CreatedAt":          true,
		"UpdatedAt":          true,
	}

	repo.InitQueryableFields(queryableFields, bsonFieldMappings)

	return repo
}

func (r *MongoWalletRepository) Save(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	if wallet.GetID() == uuid.Nil {
		return fmt.Errorf("wallet ID cannot be nil")
	}

	wallet.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, wallet)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save wallet", "wallet_id", wallet.ID, "error", err)
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	slog.InfoContext(ctx, "wallet saved successfully", "wallet_id", wallet.ID, "evm_address", wallet.EVMAddress.String())
	return nil
}

func (r *MongoWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
	var wallet wallet_entities.UserWallet

	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find wallet by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return &wallet, nil
}

func (r *MongoWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	var wallet wallet_entities.UserWallet

	filter := bson.M{"resource_owner.user_id": userID}
	err := r.collection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found for user: %s", userID)
		}
		slog.ErrorContext(ctx, "failed to find wallet by user ID", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return &wallet, nil
}

func (r *MongoWalletRepository) FindByEVMAddress(ctx context.Context, evmAddress string) (*wallet_entities.UserWallet, error) {
	var wallet wallet_entities.UserWallet

	filter := bson.M{"evm_address": evmAddress}
	err := r.collection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found for EVM address: %s", evmAddress)
		}
		slog.ErrorContext(ctx, "failed to find wallet by EVM address", "evm_address", evmAddress, "error", err)
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return &wallet, nil
}

func (r *MongoWalletRepository) Update(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	if wallet.GetID() == uuid.Nil {
		return fmt.Errorf("wallet ID cannot be nil")
	}

	wallet.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": wallet.ID}
	update := bson.M{"$set": wallet}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update wallet", "wallet_id", wallet.ID, "error", err)
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("wallet not found for update: %s", wallet.ID)
	}

	slog.InfoContext(ctx, "wallet updated successfully", "wallet_id", wallet.ID, "evm_address", wallet.EVMAddress.String())
	return nil
}

func (r *MongoWalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete wallet", "id", id, "error", err)
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("wallet not found for deletion: %s", id)
	}

	slog.InfoContext(ctx, "wallet deleted successfully", "wallet_id", id)
	return nil
}

// Ensure MongoWalletRepository implements WalletRepository interface
var _ wallet_out.WalletRepository = (*MongoWalletRepository)(nil)
