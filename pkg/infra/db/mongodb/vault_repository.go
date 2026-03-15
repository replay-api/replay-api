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

// MongoTeamVaultRepository implements TeamVaultRepository using MongoDB
type MongoTeamVaultRepository struct {
	mongodb.MongoDBRepository[wallet_entities.TeamVault]
}

// NewMongoTeamVaultRepository creates a new TeamVault MongoDB repository
func NewMongoTeamVaultRepository(mongoClient *mongo.Client, dbName string) wallet_out.TeamVaultRepository {
	entityType := wallet_entities.TeamVault{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "team_vaults", "TeamVault")

	repo.InitQueryableFields(map[string]bool{
		"ID":               true,
		"BaseEntityID":     true,
		"VisibilityLevel":  true,
		"VisibilityType":   true,
		"SquadID":          true,
		"Name":             true,
		"Balances":         true,
		"IsLocked":         true,
		"CreatedAt":        true,
		"UpdatedAt":        true,
		"UserID":           true,
		"TenantID":         true,
		"GroupID":           true,
		"ClientID":         true,
	}, map[string]string{
		"ID":               "baseentity._id",
		"BaseEntityID":     "baseentity._id",
		"VisibilityLevel":  "baseentity.visibility_level",
		"VisibilityType":   "baseentity.visibility_type",
		"SquadID":          "squad_id",
		"Name":             "name",
		"Balances":         "balances",
		"IsLocked":         "is_locked",
		"CreatedAt":        "baseentity.created_at",
		"UpdatedAt":        "baseentity.updated_at",
		"UserID":           "baseentity.resource_owner.user_id",
		"TenantID":         "baseentity.resource_owner.tenant_id",
		"GroupID":           "baseentity.resource_owner.group_id",
		"ClientID":         "baseentity.resource_owner.client_id",
	})

	r := &MongoTeamVaultRepository{MongoDBRepository: *repo}
	r.ensureIndexes(mongoClient, dbName)
	return r
}

func (r *MongoTeamVaultRepository) ensureIndexes(client *mongo.Client, dbName string) {
	coll := client.Database(dbName).Collection("team_vaults")
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "squad_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_team_vaults_squad_id"),
		},
		{
			Keys:    bson.D{{Key: "baseentity.resource_owner.group_id", Value: 1}},
			Options: options.Index().SetName("idx_team_vaults_group_id"),
		},
		{
			Keys:    bson.D{{Key: "baseentity.created_at", Value: -1}},
			Options: options.Index().SetName("idx_team_vaults_created_at"),
		},
	}
	if _, err := coll.Indexes().CreateMany(context.Background(), indexes); err != nil {
		slog.Warn("Failed to create team_vaults indexes", "error", err)
	}
}

func (r *MongoTeamVaultRepository) Save(ctx context.Context, vault *wallet_entities.TeamVault) error {
	if vault.GetID() == uuid.Nil {
		return fmt.Errorf("vault ID cannot be nil")
	}
	vault.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, vault)
	if err != nil {
		return fmt.Errorf("failed to save team vault: %w", err)
	}
	return nil
}

func (r *MongoTeamVaultRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.TeamVault, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoTeamVaultRepository) Update(ctx context.Context, vault *wallet_entities.TeamVault) error {
	vault.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, vault)
	if err != nil {
		return fmt.Errorf("failed to update team vault: %w", err)
	}
	return nil
}

func (r *MongoTeamVaultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("team vaults cannot be deleted, only locked")
}

func (r *MongoTeamVaultRepository) FindBySquadID(ctx context.Context, squadID uuid.UUID) (*wallet_entities.TeamVault, error) {
	var vault wallet_entities.TeamVault
	filter := bson.M{"squad_id": squadID}
	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&vault)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("vault not found for squad: %s", squadID)
		}
		return nil, fmt.Errorf("failed to find vault: %w", err)
	}
	return &vault, nil
}

func (r *MongoTeamVaultRepository) ExistsBySquadID(ctx context.Context, squadID uuid.UUID) (bool, error) {
	filter := bson.M{"squad_id": squadID}
	count, err := r.MongoDBRepository.Collection().CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check vault existence: %w", err)
	}
	return count > 0, nil
}
