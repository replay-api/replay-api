package db

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexDefinition represents a MongoDB index
type IndexDefinition struct {
	Collection string
	Name       string
	Keys       bson.D
	Options    *options.IndexOptions
}

// GetAllIndexes returns all index definitions for the system
func GetAllIndexes() []IndexDefinition {
	return []IndexDefinition{
		// Matchmaking Sessions Indexes
		{
			Collection: "matchmaking_sessions",
			Name:       "idx_sessions_status_created",
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "matchmaking_sessions",
			Name:       "idx_sessions_player_status",
			Keys: bson.D{
				{Key: "player_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "matchmaking_sessions",
			Name:       "idx_sessions_game_mode_region_tier",
			Keys: bson.D{
				{Key: "preferences.game_id", Value: 1},
				{Key: "preferences.game_mode", Value: 1},
				{Key: "preferences.region", Value: 1},
				{Key: "preferences.tier", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "matchmaking_sessions",
			Name:       "idx_sessions_expires_at",
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetExpireAfterSeconds(0), // TTL index - documents expire at expires_at time
		},

		// Matchmaking Pools Indexes
		{
			Collection: "matchmaking_pools",
			Name:       "idx_pools_game_mode_region",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "game_mode", Value: 1},
				{Key: "region", Value: 1},
			},
			Options: options.Index().
				SetUnique(true), // Only one pool per game/mode/region
		},
		{
			Collection: "matchmaking_pools",
			Name:       "idx_pools_is_active",
			Keys: bson.D{
				{Key: "is_active", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index(),
		},

		// Prize Pools Indexes
		{
			Collection: "prize_pools",
			Name:       "idx_prize_pools_status_escrow",
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "escrow_end_time", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "prize_pools",
			Name:       "idx_prize_pools_match_id",
			Keys: bson.D{
				{Key: "match_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true), // One prize pool per match
		},
		{
			Collection: "prize_pools",
			Name:       "idx_prize_pools_game_region",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "region", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},

		// Lobbies Indexes
		{
			Collection: "lobbies",
			Name:       "idx_lobbies_status",
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "lobbies",
			Name:       "idx_lobbies_creator",
			Keys: bson.D{
				{Key: "creator_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "lobbies",
			Name:       "idx_lobbies_game_region_tier",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "region", Value: 1},
				{Key: "tier", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index(),
		},

		// Wallets Indexes
		{
			Collection: "wallets",
			Name:       "idx_wallets_user_id",
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true), // One wallet per user
		},
		{
			Collection: "wallets",
			Name:       "idx_wallets_evm_address",
			Keys: bson.D{
				{Key: "evm_address.address", Value: 1},
			},
			Options: options.Index().
				SetSparse(true), // Not all wallets have EVM addresses
		},
		{
			Collection: "wallets",
			Name:       "idx_wallets_transactions_created",
			Keys: bson.D{
				{Key: "transactions.created_at", Value: -1},
			},
			Options: options.Index(),
		},

		// Tournaments Indexes
		{
			Collection: "tournaments",
			Name:       "idx_tournaments_status_start",
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "start_time", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "tournaments",
			Name:       "idx_tournaments_game_region",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "region", Value: 1},
				{Key: "start_time", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "tournaments",
			Name:       "idx_tournaments_organizer",
			Keys: bson.D{
				{Key: "organizer_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "tournaments",
			Name:       "idx_tournaments_participants",
			Keys: bson.D{
				{Key: "participants.player_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "tournaments",
			Name:       "idx_tournaments_registration",
			Keys: bson.D{
				{Key: "registration_open", Value: 1},
				{Key: "registration_close", Value: 1},
			},
			Options: options.Index(),
		},

		// Player Profiles Indexes
		{
			Collection: "player_profiles",
			Name:       "idx_profiles_user_id",
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "player_profiles",
			Name:       "idx_profiles_steam_id",
			Keys: bson.D{
				{Key: "steam_id", Value: 1},
			},
			Options: options.Index().
				SetSparse(true),
		},
		{
			Collection: "player_profiles",
			Name:       "idx_profiles_display_name",
			Keys: bson.D{
				{Key: "display_name", Value: 1},
			},
			Options: options.Index(),
		},

		// Squads Indexes
		{
			Collection: "squads",
			Name:       "idx_squads_leader",
			Keys: bson.D{
				{Key: "leader_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "squads",
			Name:       "idx_squads_members",
			Keys: bson.D{
				{Key: "members.player_id", Value: 1},
			},
			Options: options.Index(),
		},
	}
}

// CreateIndexes creates all indexes for the database
func CreateIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	db := client.Database(dbName)
	indexes := GetAllIndexes()

	slog.InfoContext(ctx, "Creating MongoDB indexes", "total_indexes", len(indexes))

	successCount := 0
	errorCount := 0

	for _, idx := range indexes {
		collection := db.Collection(idx.Collection)

		model := mongo.IndexModel{
			Keys:    idx.Keys,
			Options: idx.Options.SetName(idx.Name),
		}

		indexName, err := collection.Indexes().CreateOne(ctx, model)
		if err != nil {
			// Check if it's a "duplicate key" error (index already exists)
			if mongo.IsDuplicateKeyError(err) {
				slog.WarnContext(ctx, "Index already exists",
					"collection", idx.Collection,
					"index", idx.Name)
				successCount++
				continue
			}

			slog.ErrorContext(ctx, "Failed to create index",
				"collection", idx.Collection,
				"index", idx.Name,
				"error", err)
			errorCount++
			continue
		}

		slog.InfoContext(ctx, "Created index",
			"collection", idx.Collection,
			"index", indexName)
		successCount++
	}

	slog.InfoContext(ctx, "Index creation complete",
		"success", successCount,
		"errors", errorCount,
		"total", len(indexes))

	if errorCount > 0 {
		return fmt.Errorf("failed to create %d indexes", errorCount)
	}

	return nil
}

// DropAllIndexes drops all custom indexes (keeps _id index)
func DropAllIndexes(ctx context.Context, client *mongo.Client, dbName string) error {
	db := client.Database(dbName)
	indexes := GetAllIndexes()

	slog.InfoContext(ctx, "Dropping MongoDB indexes", "total_indexes", len(indexes))

	successCount := 0
	errorCount := 0

	for _, idx := range indexes {
		collection := db.Collection(idx.Collection)

		_, err := collection.Indexes().DropOne(ctx, idx.Name)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to drop index",
				"collection", idx.Collection,
				"index", idx.Name,
				"error", err)
			errorCount++
			continue
		}

		slog.InfoContext(ctx, "Dropped index",
			"collection", idx.Collection,
			"index", idx.Name)
		successCount++
	}

	slog.InfoContext(ctx, "Index drop complete",
		"success", successCount,
		"errors", errorCount,
		"total", len(indexes))

	if errorCount > 0 {
		return fmt.Errorf("failed to drop %d indexes", errorCount)
	}

	return nil
}

// ListIndexes lists all indexes in a collection
func ListIndexes(ctx context.Context, client *mongo.Client, dbName, collectionName string) ([]bson.M, error) {
	collection := client.Database(dbName).Collection(collectionName)
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, fmt.Errorf("failed to decode indexes: %w", err)
	}

	return indexes, nil
}
