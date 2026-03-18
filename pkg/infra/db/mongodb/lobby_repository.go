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
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "lobbies", "MatchmakingLobby")

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
	// Use GetByIDUnsafe because lobbies are shared resources accessed by multiple players
	// from different groups/tenants. Player authorization is enforced at the domain level
	// by checking PlayerSlots membership.
	return r.MongoDBRepository.GetByIDUnsafe(ctx, id)
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

	_, err := r.MongoDBRepository.UpdateUnsafe(ctx, lobby)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update lobby", "lobby_id", lobby.ID, "error", err)
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "lobby updated successfully", "lobby_id", lobby.ID)
	return nil
}

func (r *MongoLobbyRepository) FindExpiredReadyChecks(ctx context.Context) ([]*matchmaking_entities.MatchmakingLobby, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"status":          matchmaking_entities.LobbyStatusReadyCheck,
		"ready_check_end": bson.M{"$lte": now},
	}

	findOptions := options.Find().SetLimit(50)
	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find expired ready checks", "error", err)
		return nil, fmt.Errorf("failed to find expired ready checks: %w", err)
	}
	defer cursor.Close(ctx)

	lobbies := make([]*matchmaking_entities.MatchmakingLobby, 0)
	for cursor.Next(ctx) {
		var lobby matchmaking_entities.MatchmakingLobby
		if err := cursor.Decode(&lobby); err != nil {
			slog.ErrorContext(ctx, "failed to decode expired lobby", "error", err)
			continue
		}
		lobbies = append(lobbies, &lobby)
	}

	return lobbies, nil
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

// SetPlayerReadyAtomic atomically sets a player's ready status using findOneAndUpdate
// with positional array filters. Returns the updated lobby.
func (r *MongoLobbyRepository) SetPlayerReadyAtomic(ctx context.Context, lobbyID uuid.UUID, playerID uuid.UUID, isReady bool) (*matchmaking_entities.MatchmakingLobby, error) {
	filter := bson.M{
		"_id":                    lobbyID,
		"player_slots.player_id": playerID,
	}

	update := bson.M{
		"$set": bson.M{
			"player_slots.$[elem].is_ready": isReady,
			"updated_at":                    time.Now().UTC(),
		},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.player_id": playerID},
		},
	}

	opts := options.FindOneAndUpdate().
		SetArrayFilters(arrayFilters).
		SetReturnDocument(options.After)

	var lobby matchmaking_entities.MatchmakingLobby
	err := r.MongoDBRepository.Collection().FindOneAndUpdate(ctx, filter, update, opts).Decode(&lobby)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("lobby %s not found or player %s not in lobby", lobbyID, playerID)
		}
		return nil, fmt.Errorf("failed to atomically set player ready: %w", err)
	}

	slog.InfoContext(ctx, "player ready set atomically", "lobby_id", lobbyID, "player_id", playerID, "is_ready", isReady)
	return &lobby, nil
}

// TransitionStatus atomically transitions lobby status using CAS (compare-and-set).
// Returns true if the transition was applied, false if status didn't match expectedStatus.
func (r *MongoLobbyRepository) TransitionStatus(ctx context.Context, lobbyID uuid.UUID, expectedStatus, newStatus matchmaking_entities.LobbyStatus, extraUpdates map[string]interface{}) (bool, error) {
	filter := bson.M{
		"_id":    lobbyID,
		"status": string(expectedStatus),
	}

	setFields := bson.M{
		"status":     string(newStatus),
		"updated_at": time.Now().UTC(),
	}
	for k, v := range extraUpdates {
		setFields[k] = v
	}

	result, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, bson.M{"$set": setFields})
	if err != nil {
		return false, fmt.Errorf("failed to transition lobby status: %w", err)
	}

	if result.MatchedCount > 0 {
		slog.InfoContext(ctx, "lobby status transitioned", "lobby_id", lobbyID, "from", expectedStatus, "to", newStatus)
	}
	return result.MatchedCount > 0, nil
}

// Ensure MongoLobbyRepository implements LobbyRepository interface
var _ matchmaking_out.LobbyRepository = (*MongoLobbyRepository)(nil)
