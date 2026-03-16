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

// MongoInventoryItemRepository implements InventoryItemRepository using MongoDB
type MongoInventoryItemRepository struct {
	mongodb.MongoDBRepository[wallet_entities.InventoryItem]
	coll *mongo.Collection
}

// NewMongoInventoryItemRepository creates a new InventoryItem MongoDB repository
func NewMongoInventoryItemRepository(mongoClient *mongo.Client, dbName string) wallet_out.InventoryItemRepository {
	repo := mongodb.NewMongoDBRepositoryForType[wallet_entities.InventoryItem](mongoClient, dbName, "inventory_items", "InventoryItem")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"BaseEntityID":  true,
		"OwnerType":     true,
		"OwnerID":       true,
		"ItemType":      true,
		"Rarity":        true,
		"Status":        true,
		"Name":          true,
		"GameID":        true,
		"NftContract":   true,
		"NftTokenID":    true,
		"CreatedAt":     true,
		"GroupID":       true,
	}, map[string]string{
		"ID":            "baseentity._id",
		"BaseEntityID":  "baseentity._id",
		"OwnerType":     "owner_type",
		"OwnerID":       "owner_id",
		"ItemType":      "item_type",
		"Rarity":        "rarity",
		"Status":        "status",
		"Name":          "name",
		"GameID":        "game_id",
		"NftContract":   "nft_metadata.contract_address",
		"NftTokenID":    "nft_metadata.token_id",
		"CreatedAt":     "baseentity.created_at",
		"GroupID":       "baseentity.resource_owner.group_id",
	})

	coll := mongoClient.Database(dbName).Collection("inventory_items")
	r := &MongoInventoryItemRepository{MongoDBRepository: *repo, coll: coll}
	r.ensureIndexes(coll)
	return r
}

func (r *MongoInventoryItemRepository) ensureIndexes(coll *mongo.Collection) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "owner_type", Value: 1}, {Key: "owner_id", Value: 1}},
			Options: options.Index().SetName("idx_inventory_items_owner"),
		},
		{
			Keys:    bson.D{{Key: "nft_metadata.contract_address", Value: 1}, {Key: "nft_metadata.token_id", Value: 1}},
			Options: options.Index().SetName("idx_inventory_items_nft").SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "game_id", Value: 1}},
			Options: options.Index().SetName("idx_inventory_items_game").SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "item_type", Value: 1}, {Key: "rarity", Value: 1}},
			Options: options.Index().SetName("idx_inventory_items_type_rarity"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_inventory_items_status"),
		},
	}
	if _, err := coll.Indexes().CreateMany(context.Background(), indexes); err != nil {
		slog.Warn("Failed to create inventory_items indexes", "error", err)
	}
}

func (r *MongoInventoryItemRepository) Save(ctx context.Context, item *wallet_entities.InventoryItem) error {
	if item.GetID() == uuid.Nil {
		return fmt.Errorf("inventory item ID cannot be nil")
	}
	item.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to save inventory item: %w", err)
	}
	return nil
}

func (r *MongoInventoryItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*wallet_entities.InventoryItem, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoInventoryItemRepository) Update(ctx context.Context, item *wallet_entities.InventoryItem) error {
	item.UpdatedAt = time.Now().UTC()
	_, err := r.MongoDBRepository.Update(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}
	return nil
}

func (r *MongoInventoryItemRepository) FindByOwner(ctx context.Context, ownerType wallet_entities.InventoryOwnerType, ownerID uuid.UUID, limit, offset int) ([]*wallet_entities.InventoryItem, int64, error) {
	filter := bson.M{
		"owner_type": ownerType,
		"owner_id":   ownerID,
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "baseentity.created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find inventory items: %w", err)
	}
	defer cursor.Close(ctx)

	var items []*wallet_entities.InventoryItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to decode inventory items: %w", err)
	}
	return items, total, nil
}

func (r *MongoInventoryItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"baseentity._id": id}
	result, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("inventory item not found: %s", id)
	}
	return nil
}

func (r *MongoInventoryItemRepository) FindByNFTContract(ctx context.Context, chainID int, contractAddress string, tokenID string) (*wallet_entities.InventoryItem, error) {
	filter := bson.M{
		"nft_metadata.chain_id":          chainID,
		"nft_metadata.contract_address": contractAddress,
		"nft_metadata.token_id":         tokenID,
	}

	var item wallet_entities.InventoryItem
	err := r.coll.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find NFT item: %w", err)
	}
	return &item, nil
}

func (r *MongoInventoryItemRepository) TransferOwnership(ctx context.Context, itemID uuid.UUID, newOwnerType wallet_entities.InventoryOwnerType, newOwnerID uuid.UUID) error {
	filter := bson.M{"baseentity._id": itemID}
	update := bson.M{
		"$set": bson.M{
			"owner_type":           newOwnerType,
			"owner_id":             newOwnerID,
			"baseentity.updated_at": time.Now().UTC(),
		},
	}
	result, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to transfer inventory item ownership: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("inventory item not found: %s", itemID)
	}
	return nil
}

func (r *MongoInventoryItemRepository) CountByOwner(ctx context.Context, ownerType wallet_entities.InventoryOwnerType, ownerID uuid.UUID) (int64, error) {
	filter := bson.M{
		"owner_type": ownerType,
		"owner_id":   ownerID,
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count inventory items: %w", err)
	}
	return count, nil
}
