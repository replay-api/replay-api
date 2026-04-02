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
				{Key: "baseentity.resource_owner.user_id", Value: 1},
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
		{
			Collection: "player_profiles",
			Name:       "idx_profiles_slug_uri_unique",
			Keys: bson.D{
				{Key: "slug_uri", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},
		{
			Collection: "player_profiles",
			Name:       "idx_profiles_nickname_unique",
			Keys: bson.D{
				{Key: "nickname", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
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
		{
			Collection: "squads",
			Name:       "idx_squads_slug_uri_unique",
			Keys: bson.D{
				{Key: "slug_uri", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},
		{
			Collection: "squads",
			Name:       "idx_squads_name_unique",
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},

		// Replay Files Indexes - Optimized for searchable framework with resource ownership
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_game_status",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_visibility",
			Keys: bson.D{
				{Key: "visibility_type", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_resource_owner_user",
			Keys: bson.D{
				{Key: "resource_owner.user_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_resource_owner_group",
			Keys: bson.D{
				{Key: "resource_owner.group_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_resource_owner_client",
			Keys: bson.D{
				{Key: "resource_owner.client_id", Value: 1},
				{Key: "visibility_type", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_network_id",
			Keys: bson.D{
				{Key: "network_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		// Content hash for replay deduplication - sparse to handle null values
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_content_hash",
			Keys: bson.D{
				{Key: "content_hash", Value: 1},
			},
			Options: options.Index().
				SetSparse(true), // Don't index documents without content_hash
		},
		// Original replay reference for deduplication tracking
		{
			Collection: "replay_files",
			Name:       "idx_replay_files_original_replay_id",
			Keys: bson.D{
				{Key: "original_replay_id", Value: 1},
			},
			Options: options.Index().
				SetSparse(true), // Don't index documents without original_replay_id
		},

		// Game Events Indexes - Optimized for highlights and event queries
		{
			Collection: "game_events",
			Name:       "idx_game_events_game_type",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "event_type", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "game_events",
			Name:       "idx_game_events_replay_file",
			Keys: bson.D{
				{Key: "replay_file_id", Value: 1},
				{Key: "event_type", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "game_events",
			Name:       "idx_game_events_match",
			Keys: bson.D{
				{Key: "match_id", Value: 1},
				{Key: "round_number", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "game_events",
			Name:       "idx_game_events_player",
			Keys: bson.D{
				{Key: "player_id", Value: 1},
				{Key: "event_type", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "game_events",
			Name:       "idx_game_events_visibility",
			Keys: bson.D{
				{Key: "visibility_type", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "game_events",
			Name:       "idx_game_events_resource_owner_user",
			Keys: bson.D{
				{Key: "resource_owner.user_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},

		// Match Metadata Indexes - Optimized for match queries
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_game_status",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_replay_file",
			Keys: bson.D{
				{Key: "replay_file_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_network",
			Keys: bson.D{
				{Key: "network_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_visibility",
			Keys: bson.D{
				{Key: "visibility_type", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_resource_owner_user",
			Keys: bson.D{
				{Key: "resource_owner.user_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		// Index for map filtering
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_map_name",
			Keys: bson.D{
				{Key: "map_name", Value: 1},
				{Key: "game_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		// Index for played_at date range queries
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_played_at",
			Keys: bson.D{
				{Key: "played_at", Value: -1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index(),
		},
		// Index for slug-based match reconciliation (unique, sparse)
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_slug",
			Keys: bson.D{
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},
		// Index for external_match_id lookups (unique to prevent duplicate imports)
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_external_match_id",
			Keys: bson.D{
				{Key: "external_match_id", Value: 1},
			},
			Options: options.Index().SetSparse(true).SetUnique(true),
		},
		// Index for source-based filtering
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_source",
			Keys: bson.D{
				{Key: "source", Value: 1},
			},
			Options: options.Index(),
		},
		// Index for matches needing manual review (conflict detection)
		{
			Collection: "match_metadata",
			Name:       "idx_match_metadata_needs_review",
			Keys: bson.D{
				{Key: "needs_review", Value: 1},
			},
			Options: options.Index().SetSparse(true),
		},

		// Player Metadata Indexes - For player profile lookups
		{
			Collection: "player_metadata",
			Name:       "idx_player_metadata_network_user",
			Keys: bson.D{
				{Key: "network_id", Value: 1},
				{Key: "network_user_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "player_metadata",
			Name:       "idx_player_metadata_game",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},

		// RID Tokens Indexes - For authentication/session management
		{
			Collection: "rid_tokens",
			Name:       "idx_rid_tokens_expires",
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetExpireAfterSeconds(0), // TTL index
		},
		{
			Collection: "rid_tokens",
			Name:       "idx_rid_tokens_user",
			Keys: bson.D{
				{Key: "resource_owner.user_id", Value: 1},
			},
			Options: options.Index(),
		},

		// Match Comments Indexes
		{
			Collection: "match_comments",
			Name:       "idx_comments_match_created",
			Keys: bson.D{
				{Key: "match_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_comments",
			Name:       "idx_comments_author",
			Keys: bson.D{
				{Key: "author.id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_comments",
			Name:       "idx_comments_parent",
			Keys: bson.D{
				{Key: "parent_id", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "match_comments",
			Name:       "idx_comments_status",
			Keys: bson.D{
				{Key: "match_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},

		// Direct Messages Indexes
		{
			Collection: "direct_messages",
			Name:       "idx_dm_conversation_created",
			Keys: bson.D{
				{Key: "conversation_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "direct_messages",
			Name:       "idx_dm_sender",
			Keys: bson.D{
				{Key: "sender_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "direct_messages",
			Name:       "idx_dm_recipient_unread",
			Keys: bson.D{
				{Key: "recipient_id", Value: 1},
				{Key: "read_at", Value: 1},
			},
			Options: options.Index(),
		},

		// Team Messages Indexes
		{
			Collection: "team_messages",
			Name:       "idx_team_msg_team_channel_created",
			Keys: bson.D{
				{Key: "team_id", Value: 1},
				{Key: "channel", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "team_messages",
			Name:       "idx_team_msg_sender",
			Keys: bson.D{
				{Key: "sender_id", Value: 1},
			},
			Options: options.Index(),
		},

		// Prediction Markets Indexes
		{
			Collection: "prediction_markets",
			Name:       "idx_markets_match_status_created",
			Keys: bson.D{
				{Key: "match_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "prediction_markets",
			Name:       "idx_markets_game_id",
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "prediction_markets",
			Name:       "idx_markets_status",
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
			Options: options.Index(),
		},

		// Bets Indexes
		{
			Collection: "bets",
			Name:       "idx_bets_market_created",
			Keys: bson.D{
				{Key: "market_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "bets",
			Name:       "idx_bets_user_status",
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "bets",
			Name:       "idx_bets_market_user",
			Keys: bson.D{
				{Key: "market_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "bets",
			Name:       "idx_bets_market_status",
			Keys: bson.D{
				{Key: "market_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index(),
		},

		// Ledger Entries Indexes (wallet financial audit trail)
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_idempotency_key",
			Keys: bson.D{
				{Key: "idempotency_key", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetSparse(true), // Prevents duplicate financial operations
		},
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_account_created",
			Keys: bson.D{
				{Key: "account_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_transaction_id",
			Keys: bson.D{
				{Key: "transaction_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_account_currency",
			Keys: bson.D{
				{Key: "account_id", Value: 1},
				{Key: "currency", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_created_at",
			Keys: bson.D{
				{Key: "created_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Collection: "ledger_entries",
			Name:       "idx_ledger_entries_source_ip_created",
			Keys: bson.D{
				{Key: "metadata.source_ip", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetSparse(true), // Fraud tracking — only entries with source_ip
		},

		// Idempotent Operations Indexes (deduplication with auto-cleanup)
		{
			Collection: "idempotent_operations",
			Name:       "idx_idempotent_operations_key",
			Keys: bson.D{
				{Key: "key", Value: 1},
			},
			Options: options.Index().
				SetUnique(true),
		},
		{
			Collection: "idempotent_operations",
			Name:       "idx_idempotent_operations_expires_at",
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetExpireAfterSeconds(0), // TTL index — auto-cleanup expired operations
		},

		// Ledger Accounts Indexes
		{
			Collection: "ledger_accounts",
			Name:       "idx_ledger_accounts_wallet_id",
			Keys: bson.D{
				{Key: "wallet_id", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Collection: "ledger_accounts",
			Name:       "idx_ledger_accounts_type_currency",
			Keys: bson.D{
				{Key: "account_type", Value: 1},
				{Key: "currency", Value: 1},
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
