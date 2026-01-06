package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PlayerRatingMongoDBRepository implements PlayerRatingRepository for MongoDB
type PlayerRatingMongoDBRepository struct {
	collection *mongo.Collection
}

// NewPlayerRatingMongoDBRepository creates a new PlayerRatingMongoDBRepository
func NewPlayerRatingMongoDBRepository(db *mongo.Database) matchmaking_out.PlayerRatingRepository {
	collection := db.Collection("player_ratings")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "player_id", Value: 1}, {Key: "game_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "game_id", Value: 1}, {Key: "rating", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "player_id", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create player_ratings indexes", "error", err)
	}

	return &PlayerRatingMongoDBRepository{
		collection: collection,
	}
}

// Save creates a new player rating
func (r *PlayerRatingMongoDBRepository) Save(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, rating)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save player rating", "player_id", rating.PlayerID, "error", err)
		return fmt.Errorf("failed to save player rating: %w", err)
	}

	return nil
}

// Update updates an existing player rating
func (r *PlayerRatingMongoDBRepository) Update(ctx context.Context, rating *matchmaking_entities.PlayerRating) error {
	rating.UpdatedAt = time.Now()

	filter := bson.M{"_id": rating.ID}
	update := bson.M{"$set": rating}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update player rating", "id", rating.ID, "error", err)
		return fmt.Errorf("failed to update player rating: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("player rating not found: %s", rating.ID)
	}

	return nil
}

// FindByPlayerAndGame retrieves a player's rating for a specific game
func (r *PlayerRatingMongoDBRepository) FindByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID replay_common.GameIDKey) (*matchmaking_entities.PlayerRating, error) {
	filter := bson.M{
		"player_id": playerID,
		"game_id":   gameID,
	}

	var rating matchmaking_entities.PlayerRating
	err := r.collection.FindOne(ctx, filter).Decode(&rating)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found, return nil without error
		}
		slog.ErrorContext(ctx, "Failed to find player rating", "player_id", playerID, "game_id", gameID, "error", err)
		return nil, fmt.Errorf("failed to find player rating: %w", err)
	}

	return &rating, nil
}

// GetByID retrieves a rating by ID
func (r *PlayerRatingMongoDBRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.PlayerRating, error) {
	filter := bson.M{"_id": id}

	var rating matchmaking_entities.PlayerRating
	err := r.collection.FindOne(ctx, filter).Decode(&rating)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		slog.ErrorContext(ctx, "Failed to get player rating by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get player rating: %w", err)
	}

	return &rating, nil
}

// GetTopPlayers retrieves top players by rating for leaderboard
func (r *PlayerRatingMongoDBRepository) GetTopPlayers(ctx context.Context, gameID replay_common.GameIDKey, limit int) ([]*matchmaking_entities.PlayerRating, error) {
	filter := bson.M{
		"game_id":        gameID,
		"matches_played": bson.M{"$gte": 10}, // Minimum 10 matches to be on leaderboard
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "rating", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get top players", "game_id", gameID, "error", err)
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*matchmaking_entities.PlayerRating
	if err := cursor.All(ctx, &ratings); err != nil {
		slog.ErrorContext(ctx, "Failed to decode top players", "error", err)
		return nil, fmt.Errorf("failed to decode top players: %w", err)
	}

	return ratings, nil
}

// GetRankDistribution returns the count of players in each rank
func (r *PlayerRatingMongoDBRepository) GetRankDistribution(ctx context.Context, gameID replay_common.GameIDKey) (map[matchmaking_entities.Rank]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"game_id": gameID, "matches_played": bson.M{"$gte": 1}}}},
		{{Key: "$project", Value: bson.M{
			"rank": bson.M{
				"$switch": bson.M{
					"branches": []bson.M{
						{"case": bson.M{"$gte": []interface{}{"$rating", 2800}}, "then": "challenger"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 2500}}, "then": "grandmaster"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 2200}}, "then": "master"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 1900}}, "then": "diamond"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 1600}}, "then": "platinum"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 1400}}, "then": "gold"},
						{"case": bson.M{"$gte": []interface{}{"$rating", 1200}}, "then": "silver"},
					},
					"default": "bronze",
				},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$rank",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get rank distribution", "game_id", gameID, "error", err)
		return nil, fmt.Errorf("failed to get rank distribution: %w", err)
	}
	defer cursor.Close(ctx)

	distribution := make(map[matchmaking_entities.Rank]int)
	for cursor.Next(ctx) {
		var result struct {
			Rank  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		distribution[matchmaking_entities.Rank(result.Rank)] = result.Count
	}

	return distribution, nil
}

// Delete removes a player rating
func (r *PlayerRatingMongoDBRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete player rating", "id", id, "error", err)
		return fmt.Errorf("failed to delete player rating: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("player rating not found: %s", id)
	}

	return nil
}

// Ensure PlayerRatingMongoDBRepository implements PlayerRatingRepository
var _ matchmaking_out.PlayerRatingRepository = (*PlayerRatingMongoDBRepository)(nil)

