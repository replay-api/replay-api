package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoWalletRepository struct {
	mongodb.MongoDBRepository[wallet_entities.UserWallet]
}

func NewMongoWalletRepository(mongoClient *mongo.Client, dbName string) wallet_out.WalletRepository {
	entityType := wallet_entities.UserWallet{}
	repo := mongodb.NewMongoDBRepository[wallet_entities.UserWallet](mongoClient, dbName, entityType, "wallets", "UserWallet")

	// Note: DO NOT include both "ResourceOwner" and its sub-fields (UserID, TenantID, etc.)
	// in queryable fields - this causes MongoDB projection path collisions.
	// The resource_owner data is populated through the individual field projections.
	repo.InitQueryableFields(map[string]bool{
		"ID":                  true,
		"BaseEntityID":        true, // Explicitly include baseentity._id
		"VisibilityLevel":     true,
		"VisibilityType":      true,
		"EVMAddress":          true,
		"Balances":            true,
		"PendingTransactions": true,
		"TotalDeposited":      true,
		"TotalWithdrawn":      true,
		"TotalPrizesWon":      true,
		"DailyPrizeWinnings":  true,
		"LastPrizeWinDate":    true,
		"IsLocked":            true,
		"LockReason":          true,
		"CreatedAt":           true,
		"UpdatedAt":           true,
		// Resource ownership fields for RLS queries
		"UserID":   true,
		"TenantID": true,
		"GroupID":  true,
		"ClientID": true,
	}, map[string]string{
		"ID":                  "baseentity._id",
		"BaseEntityID":        "baseentity._id",
		"VisibilityLevel":     "baseentity.visibility_level",
		"VisibilityType":      "baseentity.visibility_type",
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
		"CreatedAt":           "baseentity.created_at",
		"UpdatedAt":           "baseentity.updated_at",
		// Resource ownership field mappings for queries
		"UserID":   "baseentity.resource_owner.user_id",
		"TenantID": "baseentity.resource_owner.tenant_id",
		"GroupID":  "baseentity.resource_owner.group_id",
		"ClientID": "baseentity.resource_owner.client_id",
	})

	return &MongoWalletRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MongoWalletRepository) Save(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	if wallet.GetID() == uuid.Nil {
		return fmt.Errorf("wallet ID cannot be nil")
	}

	wallet.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, wallet)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save wallet", "wallet_id", wallet.ID, "error", err)
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	slog.InfoContext(ctx, "wallet saved successfully", "wallet_id", wallet.ID, "evm_address", wallet.EVMAddress.String())
	return nil
}

func (r *MongoWalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.UserWallet, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*wallet_entities.UserWallet, error) {
	var wallet wallet_entities.UserWallet

	// Use the correct field path that matches BaseEntity BSON structure
	filter := bson.M{"baseentity.resource_owner.user_id": userID}
	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&wallet)
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
	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found for EVM address: %s", evmAddress)
		}
		slog.ErrorContext(ctx, "failed to find wallet by EVM address", "evm_address", evmAddress, "error", err)
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return &wallet, nil
}

func (r *MongoWalletRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	filter := bson.M{"baseentity.resource_owner.user_id": userID}
	var result bson.M
	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		slog.ErrorContext(ctx, "failed to check wallet existence by user ID", "user_id", userID, "error", err)
		return false, fmt.Errorf("failed to check wallet existence: %w", err)
	}
	return true, nil
}

func (r *MongoWalletRepository) Update(ctx context.Context, wallet *wallet_entities.UserWallet) error {
	if wallet.GetID() == uuid.Nil {
		return fmt.Errorf("wallet ID cannot be nil")
	}

	wallet.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Update(ctx, wallet)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update wallet", "wallet_id", wallet.ID, "error", err)
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	slog.InfoContext(ctx, "wallet updated successfully", "wallet_id", wallet.ID, "evm_address", wallet.EVMAddress.String())
	return nil
}

func (r *MongoWalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.MongoDBRepository.DeleteOneWithRLS(ctx, filter)
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
