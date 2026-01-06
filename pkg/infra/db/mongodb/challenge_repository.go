package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	shared "github.com/resource-ownership/go-common/pkg/common"
	challenge_entities "github.com/replay-api/replay-api/pkg/domain/challenge/entities"
	challenge_out "github.com/replay-api/replay-api/pkg/domain/challenge/ports/out"
)

const challengeCollectionName = "challenges"

// ChallengeRepository implements MongoDB persistence for challenges
type ChallengeRepository struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewChallengeRepository creates a new challenge repository
func NewChallengeRepository(client *mongo.Client, dbName string) challenge_out.ChallengeRepository {
	return &ChallengeRepository{
		client: client,
		db:     client.Database(dbName),
	}
}

// Save persists a challenge (create or update)
func (r *ChallengeRepository) Save(ctx context.Context, challenge *challenge_entities.Challenge) error {
	collection := r.db.Collection(challengeCollectionName)

	challenge.UpdatedAt = time.Now().UTC()
	if challenge.CreatedAt.IsZero() {
		challenge.CreatedAt = time.Now().UTC()
	}

	opts := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(ctx, bson.M{"_id": challenge.ID}, challenge, opts)
	return err
}

// GetByID retrieves a challenge by its ID
func (r *ChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*challenge_entities.Challenge, error) {
	collection := r.db.Collection(challengeCollectionName)

	var challenge challenge_entities.Challenge
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&challenge)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

// GetByMatchID retrieves all challenges for a match
func (r *ChallengeRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID, search *shared.Search) ([]*challenge_entities.Challenge, error) {
	collection := r.db.Collection(challengeCollectionName)

	query := bson.M{"match_id": matchID}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if search != nil && search.ResultOptions.Limit > 0 {
		limit := search.ResultOptions.Limit
		if limit > 9223372036854775807 { // int64 max
			limit = 9223372036854775807
		}
		opts.SetLimit(int64(limit))
		if search.ResultOptions.Skip > 0 {
			skip := search.ResultOptions.Skip
			if skip > 9223372036854775807 { // int64 max
				skip = 9223372036854775807
			}
			opts.SetSkip(int64(skip))
		}
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []*challenge_entities.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, nil
}

// GetByChallengerID retrieves challenges submitted by a player
func (r *ChallengeRepository) GetByChallengerID(ctx context.Context, challengerID uuid.UUID, search *shared.Search) ([]*challenge_entities.Challenge, error) {
	collection := r.db.Collection(challengeCollectionName)

	query := bson.M{"challenger_id": challengerID}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if search != nil && search.ResultOptions.Limit > 0 {
		limit := search.ResultOptions.Limit
		if limit > 9223372036854775807 { // int64 max
			limit = 9223372036854775807
		}
		opts.SetLimit(int64(limit))
		if search.ResultOptions.Skip > 0 {
			skip := search.ResultOptions.Skip
			if skip > 9223372036854775807 { // int64 max
				skip = 9223372036854775807
			}
			opts.SetSkip(int64(skip))
		}
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []*challenge_entities.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, nil
}

// Search searches challenges based on criteria
func (r *ChallengeRepository) Search(ctx context.Context, criteria challenge_out.ChallengeCriteria) ([]*challenge_entities.Challenge, int64, error) {
	collection := r.db.Collection(challengeCollectionName)

	query := bson.M{}

	// Apply filters
	if criteria.MatchID != nil {
		query["match_id"] = *criteria.MatchID
	}
	if criteria.ChallengerID != nil {
		query["challenger_id"] = *criteria.ChallengerID
	}
	if criteria.GameID != nil {
		query["game_id"] = *criteria.GameID
	}
	if criteria.TournamentID != nil {
		query["tournament_id"] = *criteria.TournamentID
	}
	if criteria.LobbyID != nil {
		query["lobby_id"] = *criteria.LobbyID
	}
	if len(criteria.Types) > 0 {
		query["type"] = bson.M{"$in": criteria.Types}
	}
	if len(criteria.Statuses) > 0 {
		query["status"] = bson.M{"$in": criteria.Statuses}
	}
	if len(criteria.Priorities) > 0 {
		query["priority"] = bson.M{"$in": criteria.Priorities}
	}

	// Exclude expired unless explicitly included
	if !criteria.IncludeExpired {
		query["$or"] = []bson.M{
			{"expires_at": nil},
			{"expires_at": bson.M{"$gt": time.Now().UTC()}},
		}
	}

	// Resource owner filter
	if criteria.ResourceOwner != nil {
		query["resource_owner.tenant_id"] = criteria.ResourceOwner.TenantID
	}

	// Count total
	total, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Build find options
	opts := options.Find().SetSort(bson.D{
		{Key: "priority", Value: -1}, // Critical first
		{Key: "created_at", Value: -1},
	})

	if criteria.Search != nil && criteria.Search.ResultOptions.Limit > 0 {
		limit := criteria.Search.ResultOptions.Limit
		if limit > 9223372036854775807 { // int64 max
			limit = 9223372036854775807
		}
		opts.SetLimit(int64(limit))
		if criteria.Search.ResultOptions.Skip > 0 {
			skip := criteria.Search.ResultOptions.Skip
			if skip > 9223372036854775807 { // int64 max
				skip = 9223372036854775807
			}
			opts.SetSkip(int64(skip))
		}
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var challenges []*challenge_entities.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, 0, err
	}

	return challenges, total, nil
}

// GetPending retrieves pending challenges requiring review
func (r *ChallengeRepository) GetPending(ctx context.Context, priority *challenge_entities.ChallengePriority, gameID *string, limit int) ([]*challenge_entities.Challenge, error) {
	collection := r.db.Collection(challengeCollectionName)

	query := bson.M{
		"status": bson.M{
			"$in": []challenge_entities.ChallengeStatus{
				challenge_entities.ChallengeStatusPending,
				challenge_entities.ChallengeStatusVotePending,
			},
		},
		"$or": []bson.M{
			{"expires_at": nil},
			{"expires_at": bson.M{"$gt": time.Now().UTC()}},
		},
	}

	if priority != nil {
		query["priority"] = *priority
	}
	if gameID != nil {
		query["game_id"] = *gameID
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "priority", Value: -1},
			{Key: "created_at", Value: 1},
		}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []*challenge_entities.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, nil
}

// GetExpired retrieves expired challenges
func (r *ChallengeRepository) GetExpired(ctx context.Context, limit int) ([]*challenge_entities.Challenge, error) {
	collection := r.db.Collection(challengeCollectionName)

	query := bson.M{
		"status": bson.M{
			"$nin": []challenge_entities.ChallengeStatus{
				challenge_entities.ChallengeStatusResolved,
				challenge_entities.ChallengeStatusRejected,
				challenge_entities.ChallengeStatusCancelled,
				challenge_entities.ChallengeStatusExpired,
			},
		},
		"expires_at": bson.M{"$lt": time.Now().UTC()},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "expires_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var challenges []*challenge_entities.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, nil
}

// CountByStatus counts challenges grouped by status
func (r *ChallengeRepository) CountByStatus(ctx context.Context, matchID *uuid.UUID) (map[challenge_entities.ChallengeStatus]int64, error) {
	collection := r.db.Collection(challengeCollectionName)

	matchStage := bson.M{}
	if matchID != nil {
		matchStage["match_id"] = *matchID
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[challenge_entities.ChallengeStatus]int64)
	for cursor.Next(ctx) {
		var item struct {
			ID    challenge_entities.ChallengeStatus `bson:"_id"`
			Count int64                               `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// Delete deletes a challenge by ID
func (r *ChallengeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	collection := r.db.Collection(challengeCollectionName)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// DeleteByMatchID deletes all challenges for a match
func (r *ChallengeRepository) DeleteByMatchID(ctx context.Context, matchID uuid.UUID) error {
	collection := r.db.Collection(challengeCollectionName)
	_, err := collection.DeleteMany(ctx, bson.M{"match_id": matchID})
	return err
}

// CreateIndexes creates necessary indexes for the challenges collection
func (r *ChallengeRepository) CreateIndexes(ctx context.Context) error {
	collection := r.db.Collection(challengeCollectionName)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "match_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "challenger_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "game_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "priority", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "priority", Value: -1},
				{Key: "created_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "match_id", Value: 1},
				{Key: "challenger_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

