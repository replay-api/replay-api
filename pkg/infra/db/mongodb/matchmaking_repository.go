package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

const (
	matchmakingSessionsCollection = "matchmaking_sessions"
	matchmakingPoolsCollection    = "matchmaking_pools"
)

// MatchmakingSessionRepository implements MongoDB persistence for matchmaking sessions
type MatchmakingSessionRepository struct {
	mongodb.MongoDBRepository[matchmaking_entities.MatchmakingSession]
}

// NewMatchmakingSessionRepository creates a new repository instance
func NewMatchmakingSessionRepository(client *mongo.Client, dbName string) *MatchmakingSessionRepository {
	repo := mongodb.NewMongoDBRepository[matchmaking_entities.MatchmakingSession](client, dbName, matchmaking_entities.MatchmakingSession{}, matchmakingSessionsCollection, "MatchmakingSession")

	repo.InitQueryableFields(map[string]bool{
		"ID":           true,
		"PlayerID":     true,
		"SquadID":      true,
		"Status":       true,
		"PlayerMMR":    true,
		"QueuedAt":     true,
		"MatchedAt":    true,
		"MatchID":      true,
		"EstimatedWait": true,
		"ExpiresAt":    true,
		"CreatedAt":    true,
		"UpdatedAt":    true,
	}, map[string]string{
		"ID":           "_id",
		"PlayerID":     "player_id",
		"SquadID":      "squad_id",
		"Status":       "status",
		"PlayerMMR":    "player_mmr",
		"QueuedAt":     "queued_at",
		"MatchedAt":    "matched_at",
		"MatchID":      "match_id",
		"EstimatedWait": "estimated_wait_seconds",
		"ExpiresAt":    "expires_at",
		"CreatedAt":    "created_at",
		"UpdatedAt":    "updated_at",
	})

	return &MatchmakingSessionRepository{
		MongoDBRepository: *repo,
	}
}

// Save creates or updates a matchmaking session
func (r *MatchmakingSessionRepository) Save(ctx context.Context, session *matchmaking_entities.MatchmakingSession) error {
	collection := r.MongoDBRepository.Collection()

	session.UpdatedAt = time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	opts := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": session.ID}, session, opts)
	return err
}

// GetByID retrieves a session by ID
func (r *MatchmakingSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingSession, error) {
	collection := r.MongoDBRepository.Collection()

	var session matchmaking_entities.MatchmakingSession
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetByPlayerID retrieves active sessions for a player
func (r *MatchmakingSessionRepository) GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*matchmaking_entities.MatchmakingSession, error) {
	collection := r.MongoDBRepository.Collection()

	filter := bson.M{
		"player_id": playerID,
		"status": bson.M{
			"$in": []string{
				string(matchmaking_entities.StatusQueued),
				string(matchmaking_entities.StatusSearching),
			},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*matchmaking_entities.MatchmakingSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetActiveSessions retrieves all active sessions with filters
func (r *MatchmakingSessionRepository) GetActiveSessions(ctx context.Context, filters matchmaking_out.SessionFilters) ([]*matchmaking_entities.MatchmakingSession, error) {
	collection := r.MongoDBRepository.Collection()

	query := bson.M{}

	// Apply filters
	if filters.GameID != "" {
		query["preferences.game_id"] = filters.GameID
	}
	if filters.GameMode != "" {
		query["preferences.game_mode"] = filters.GameMode
	}
	if filters.Region != "" {
		query["preferences.region"] = filters.Region
	}
	if filters.Tier != nil {
		query["preferences.tier"] = string(*filters.Tier)
	}
	if filters.Status != nil {
		query["status"] = string(*filters.Status)
	} else {
		// Default to active statuses
		query["status"] = bson.M{
			"$in": []string{
				string(matchmaking_entities.StatusQueued),
				string(matchmaking_entities.StatusSearching),
			},
		}
	}
	if filters.MinMMR != nil {
		query["player_mmr"] = bson.M{"$gte": *filters.MinMMR}
	}
	if filters.MaxMMR != nil {
		if mmrFilter, ok := query["player_mmr"].(bson.M); ok {
			mmrFilter["$lte"] = *filters.MaxMMR
		} else {
			query["player_mmr"] = bson.M{"$lte": *filters.MaxMMR}
		}
	}

	// Add expiration filter
	query["expires_at"] = bson.M{"$gt": time.Now()}

	opts := options.Find().
		SetSort(bson.D{{Key: "queued_at", Value: 1}}) // Oldest first

	if filters.Limit > 0 {
		opts.SetLimit(int64(filters.Limit))
	}
	if filters.Offset > 0 {
		opts.SetSkip(int64(filters.Offset))
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*matchmaking_entities.MatchmakingSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// UpdateStatus updates the session status
func (r *MatchmakingSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status matchmaking_entities.SessionStatus) error {
	collection := r.MongoDBRepository.Collection()

	update := bson.M{
		"$set": bson.M{
			"status":     string(status),
			"updated_at": time.Now(),
		},
	}

	// Set matched_at timestamp if status is matched
	if status == matchmaking_entities.StatusMatched {
		now := time.Now()
		update["$set"].(bson.M)["matched_at"] = now
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete removes a session
func (r *MatchmakingSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	collection := r.MongoDBRepository.Collection()
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// DeleteExpired removes expired sessions
func (r *MatchmakingSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	collection := r.MongoDBRepository.Collection()

	result, err := collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lte": time.Now()},
	})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// MatchmakingPoolRepository implements MongoDB persistence for matchmaking pools
type MatchmakingPoolRepository struct {
	mongodb.MongoDBRepository[matchmaking_entities.MatchmakingPool]
}

// NewMatchmakingPoolRepository creates a new repository instance
func NewMatchmakingPoolRepository(client *mongo.Client, dbName string) *MatchmakingPoolRepository {
	repo := mongodb.NewMongoDBRepository[matchmaking_entities.MatchmakingPool](client, dbName, matchmaking_entities.MatchmakingPool{}, matchmakingPoolsCollection, "MatchmakingPool")

	repo.InitQueryableFields(map[string]bool{
		"ID":             true,
		"GameID":         true,
		"GameMode":       true,
		"Region":         true,
		"ActiveSessions": true,
		"PoolStats":      true,
		"CreatedAt":      true,
		"UpdatedAt":      true,
	}, map[string]string{
		"ID":             "_id",
		"GameID":         "game_id",
		"GameMode":       "game_mode",
		"Region":         "region",
		"ActiveSessions": "active_sessions",
		"PoolStats":      "pool_stats",
		"CreatedAt":      "created_at",
		"UpdatedAt":      "updated_at",
	})

	return &MatchmakingPoolRepository{
		MongoDBRepository: *repo,
	}
}

// Save creates or updates a pool
func (r *MatchmakingPoolRepository) Save(ctx context.Context, pool *matchmaking_entities.MatchmakingPool) error {
	collection := r.MongoDBRepository.Collection()

	pool.UpdatedAt = time.Now()
	if pool.CreatedAt.IsZero() {
		pool.CreatedAt = time.Now()
	}

	opts := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": pool.ID}, pool, opts)
	return err
}

// GetByID retrieves a pool by ID
func (r *MatchmakingPoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*matchmaking_entities.MatchmakingPool, error) {
	collection := r.MongoDBRepository.Collection()

	var pool matchmaking_entities.MatchmakingPool
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&pool)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

// GetByGameModeRegion retrieves a pool by game, mode, and region
func (r *MatchmakingPoolRepository) GetByGameModeRegion(ctx context.Context, gameID, gameMode, region string) (*matchmaking_entities.MatchmakingPool, error) {
	collection := r.MongoDBRepository.Collection()

	var pool matchmaking_entities.MatchmakingPool
	err := collection.FindOne(ctx, bson.M{
		"game_id":   gameID,
		"game_mode": gameMode,
		"region":    region,
	}).Decode(&pool)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

// UpdateStats updates pool statistics
func (r *MatchmakingPoolRepository) UpdateStats(ctx context.Context, poolID uuid.UUID, stats matchmaking_entities.PoolStatistics) error {
	collection := r.MongoDBRepository.Collection()

	update := bson.M{
		"$set": bson.M{
			"pool_stats": stats,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": poolID}, update)
	return err
}

// GetAllActive retrieves all active pools
func (r *MatchmakingPoolRepository) GetAllActive(ctx context.Context) ([]*matchmaking_entities.MatchmakingPool, error) {
	collection := r.MongoDBRepository.Collection()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pools []*matchmaking_entities.MatchmakingPool
	if err := cursor.All(ctx, &pools); err != nil {
		return nil, err
	}

	return pools, nil
}
