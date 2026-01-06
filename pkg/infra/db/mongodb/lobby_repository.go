package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLobbyRepository struct {
	mongodb.MongoDBRepository[matchmaking_entities.MatchmakingLobby]
}

func NewMongoLobbyRepository(mongoClient *mongo.Client, dbName string) matchmaking_out.LobbyRepository {
	entityType := matchmaking_entities.MatchmakingLobby{}
	repo := mongodb.NewMongoDBRepository[matchmaking_entities.MatchmakingLobby](mongoClient, dbName, entityType, "lobbies", "MatchmakingLobby")

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

	return &MongoLobbyRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MongoLobbyRepository) Save(ctx context.Context, lobby *matchmaking_entities.MatchmakingLobby) error {
	if lobby.GetID() == uuid.Nil {
		return fmt.Errorf("lobby ID cannot be nil")
	}

	lobby.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Create(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save lobby", "lobby_id", lobby.ID, "error", err)
		return fmt.Errorf("failed to save lobby: %w", err)
	}

	slog.InfoContext(ctx, "lobby saved successfully", "lobby_id", lobby.ID)
	return nil
}

func (r *MongoLobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingLobby, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *MongoLobbyRepository) FindByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*matchmaking_entities.MatchmakingLobby, error) {
	filter := bson.M{"creator_id": creatorID}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter)
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

	cursor, err := r.MongoDBRepository.FindWithRLS(ctx, filter, findOptions)
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

	_, err := r.MongoDBRepository.Update(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update lobby", "lobby_id", lobby.ID, "error", err)
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "lobby updated successfully", "lobby_id", lobby.ID)
	return nil
}

func (r *MongoLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.MongoDBRepository.DeleteOne(ctx, filter)
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
