package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/psavelis/team-pro/replay-api/pkg/domain/matchmaking/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLobbyRepository struct {
	*MongoDBRepository[*matchmaking_entities.MatchmakingLobby]
}

func NewMongoLobbyRepository(mongoClient *mongo.Client, dbName string) matchmaking_out.LobbyRepository {
	mappingCache := make(map[string]CacheItem)
	entityModel := reflect.TypeOf(matchmaking_entities.MatchmakingLobby{})
	repo := &MongoLobbyRepository{
		MongoDBRepository: &MongoDBRepository[*matchmaking_entities.MatchmakingLobby]{
			mongoClient:       mongoClient,
			dbName:            dbName,
			mappingCache:      mappingCache,
			entityModel:       entityModel,
			collectionName:    "lobbies",
			entityName:        "MatchmakingLobby",
			bsonFieldMappings: make(map[string]string),
			queryableFields:   make(map[string]bool),
		},
	}

	// Define BSON field mappings
	bsonFieldMappings := map[string]string{
		"ID":               "_id",
		"CreatorID":        "creator_id",
		"GameID":           "game_id",
		"Region":           "region",
		"Tier":             "tier",
		"DistributionRule": "distribution_rule",
		"MaxPlayers":       "max_players",
		"PlayerSlots":      "player_slots",
		"Status":           "status",
		"ReadyCheckStart":  "ready_check_start",
		"ReadyCheckEnd":    "ready_check_end",
		"MatchID":          "match_id",
		"CancelReason":     "cancel_reason",
		"AutoFill":         "auto_fill",
		"ReadyTimeout":     "ready_timeout",
		"InviteOnly":       "invite_only",
		"CreatedAt":        "created_at",
		"UpdatedAt":        "updated_at",
		"ResourceOwner":    "resource_owner",
	}

	// Define queryable fields for search operations
	queryableFields := map[string]bool{
		"CreatorID":  true,
		"GameID":     true,
		"Region":     true,
		"Tier":       true,
		"MaxPlayers": true,
		"Status":     true,
		"MatchID":    true,
		"AutoFill":   true,
		"InviteOnly": true,
		"CreatedAt":  true,
		"UpdatedAt":  true,
	}

	repo.InitQueryableFields(queryableFields, bsonFieldMappings)

	return repo
}

func (r *MongoLobbyRepository) Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	if lobby.GetID() == uuid.Nil {
		return fmt.Errorf("lobby ID cannot be nil")
	}

	lobby.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save lobby", "lobby_id", lobby.ID, "error", err)
		return fmt.Errorf("failed to save lobby: %w", err)
	}

	slog.InfoContext(ctx, "lobby saved successfully", "lobby_id", lobby.ID)
	return nil
}

func (r *MongoLobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error) {
	var lobby matchmaking_entities.MatchmakingLobby

	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&lobby)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("lobby not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find lobby by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find lobby: %w", err)
	}

	return &lobby, nil
}

func (r *MongoLobbyRepository) FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error) {
	filter := bson.M{"creator_id": creatorID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find lobbies by creator ID", "creator_id", creatorID, "error", err)
		return nil, fmt.Errorf("failed to find lobbies: %w", err)
	}
	defer cursor.Close(ctx)

	lobbies := make([]*matchmaking_entities.MatchmakingLobby, 0)
	for cursor.Next(ctx) {
		var lobby matchmaking_entities.MatchmakingLobby
		if err := cursor.Decode(&lobby); err != nil {
			slog.ErrorContext(ctx, "failed to decode lobby", "error", err)
			continue
		}
		lobbies = append(lobbies, &lobby)
	}

	return lobbies, nil
}

func (r *MongoLobbyRepository) FindOpenLobbies(ctx context.Context, gameID, region, tier string, limit int) ([]*matchmaking_entities.MatchmakingLobby, error) {
	filter := bson.M{
		"status":  matchmaking_entities.LobbyStatusOpen,
		"game_id": gameID,
		"region":  region,
		"tier":    tier,
	}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find open lobbies", "error", err)
		return nil, fmt.Errorf("failed to find open lobbies: %w", err)
	}
	defer cursor.Close(ctx)

	lobbies := make([]*matchmaking_entities.MatchmakingLobby, 0)
	for cursor.Next(ctx) {
		var lobby matchmaking_entities.MatchmakingLobby
		if err := cursor.Decode(&lobby); err != nil {
			slog.ErrorContext(ctx, "failed to decode lobby", "error", err)
			continue
		}

		// Additional check: lobby has available slots
		if lobby.GetPlayerCount() < lobby.MaxPlayers {
			lobbies = append(lobbies, &lobby)
		}
	}

	slog.InfoContext(ctx, "found open lobbies", "count", len(lobbies), "game_id", gameID, "region", region, "tier", tier)
	return lobbies, nil
}

func (r *MongoLobbyRepository) Update(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	if lobby.GetID() == uuid.Nil {
		return fmt.Errorf("lobby ID cannot be nil")
	}

	lobby.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": lobby.ID}
	update := bson.M{"$set": lobby}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update lobby", "lobby_id", lobby.ID, "error", err)
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("lobby not found for update: %s", lobby.ID)
	}

	slog.InfoContext(ctx, "lobby updated successfully", "lobby_id", lobby.ID)
	return nil
}

func (r *MongoLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete lobby", "id", id, "error", err)
		return fmt.Errorf("failed to delete lobby: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("lobby not found for deletion: %s", id)
	}

	slog.InfoContext(ctx, "lobby deleted successfully", "lobby_id", id)
	return nil
}

// Ensure MongoLobbyRepository implements LobbyRepository interface
var _ matchmaking_out.LobbyRepository = (*MongoLobbyRepository)(nil)
