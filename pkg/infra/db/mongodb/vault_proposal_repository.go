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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoVaultProposalRepository implements VaultProposalRepository using MongoDB
type MongoVaultProposalRepository struct {
	mongodb.MongoDBRepository[wallet_entities.VaultProposal]
	coll *mongo.Collection
}

// NewMongoVaultProposalRepository creates a new VaultProposal MongoDB repository
func NewMongoVaultProposalRepository(mongoClient *mongo.Client, dbName string) wallet_out.VaultProposalRepository {
	entityType := wallet_entities.VaultProposal{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "vault_proposals", "VaultProposal")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"BaseEntityID":    true,
		"VaultID":         true,
		"SquadID":         true,
		"ProposerID":      true,
		"Type":            true,
		"Status":          true,
		"ExpiresAt":       true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
		"GroupID":          true,
	}, map[string]string{
		"ID":              "baseentity._id",
		"BaseEntityID":    "baseentity._id",
		"VaultID":         "vault_id",
		"SquadID":         "squad_id",
		"ProposerID":      "proposer_id",
		"Type":            "type",
		"Status":          "status",
		"ExpiresAt":       "expires_at",
		"CreatedAt":       "baseentity.created_at",
		"UpdatedAt":       "baseentity.updated_at",
		"GroupID":          "baseentity.resource_owner.group_id",
	})

	coll := mongoClient.Database(dbName).Collection("vault_proposals")
	r := &MongoVaultProposalRepository{MongoDBRepository: *repo, coll: coll}
	r.ensureIndexes(coll)
	return r
}

func (r *MongoVaultProposalRepository) ensureIndexes(coll *mongo.Collection) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "vault_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_vault_proposals_vault_status"),
		},
		{
			Keys:    bson.D{{Key: "proposer_id", Value: 1}},
			Options: options.Index().SetName("idx_vault_proposals_proposer"),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetName("idx_vault_proposals_expires"),
		},
		{
			Keys:    bson.D{{Key: "baseentity.created_at", Value: -1}},
			Options: options.Index().SetName("idx_vault_proposals_created_at"),
		},
	}
	if _, err := coll.Indexes().CreateMany(context.Background(), indexes); err != nil {
		slog.Warn("Failed to create vault_proposals indexes", "error", err)
	}
}

func (r *MongoVaultProposalRepository) Save(ctx context.Context, proposal *wallet_entities.VaultProposal) error {
	if proposal.GetID() == uuid.Nil {
		return fmt.Errorf("proposal ID cannot be nil")
	}
	proposal.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, proposal)
	if err != nil {
		return fmt.Errorf("failed to save vault proposal: %w", err)
	}
	return nil
}

func (r *MongoVaultProposalRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.VaultProposal, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoVaultProposalRepository) Update(ctx context.Context, proposal *wallet_entities.VaultProposal) error {
	proposal.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, proposal)
	if err != nil {
		return fmt.Errorf("failed to update vault proposal: %w", err)
	}
	return nil
}

func (r *MongoVaultProposalRepository) FindByVaultID(ctx context.Context, vaultID uuid.UUID, limit, offset int) ([]wallet_entities.VaultProposal, int64, error) {
	filter := bson.M{"vault_id": vaultID}
	opts := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count proposals: %w", err)
	}

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find proposals: %w", err)
	}
	defer cursor.Close(ctx)

	var proposals []wallet_entities.VaultProposal
	if err := cursor.All(ctx, &proposals); err != nil {
		return nil, 0, fmt.Errorf("failed to decode proposals: %w", err)
	}
	return proposals, total, nil
}

func (r *MongoVaultProposalRepository) FindPendingByVaultID(ctx context.Context, vaultID uuid.UUID) ([]wallet_entities.VaultProposal, error) {
	filter := bson.M{
		"vault_id": vaultID,
		"status":   wallet_entities.ProposalStatusPending,
	}
	opts := options.Find().SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending proposals: %w", err)
	}
	defer cursor.Close(ctx)

	var proposals []wallet_entities.VaultProposal
	if err := cursor.All(ctx, &proposals); err != nil {
		return nil, fmt.Errorf("failed to decode proposals: %w", err)
	}
	return proposals, nil
}

func (r *MongoVaultProposalRepository) FindExpired(ctx context.Context) ([]wallet_entities.VaultProposal, error) {
	filter := bson.M{
		"status":     wallet_entities.ProposalStatusPending,
		"expires_at": bson.M{"$lte": time.Now()},
	}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find expired proposals: %w", err)
	}
	defer cursor.Close(ctx)

	var proposals []wallet_entities.VaultProposal
	if err := cursor.All(ctx, &proposals); err != nil {
		return nil, fmt.Errorf("failed to decode expired proposals: %w", err)
	}
	return proposals, nil
}

func (r *MongoVaultProposalRepository) CountByVaultIDAndStatus(ctx context.Context, vaultID uuid.UUID, status wallet_entities.ProposalStatus) (int64, error) {
	filter := bson.M{
		"vault_id": vaultID,
		"status":   status,
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals: %w", err)
	}
	return count, nil
}
