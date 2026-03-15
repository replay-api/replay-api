package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoOracleResultRepository implements the OracleResultRepository port using MongoDB
type MongoOracleResultRepository struct {
	mongodb.MongoDBRepository[oracle_entities.OracleResult]
}

// Compile-time interface satisfaction check
var _ oracle_out.OracleResultRepository = (*MongoOracleResultRepository)(nil)

// NewMongoOracleResultRepository creates a new MongoDB-backed oracle result repository
func NewMongoOracleResultRepository(mongoClient *mongo.Client, dbName string) oracle_out.OracleResultRepository {
	entityType := oracle_entities.OracleResult{}
	repo := mongodb.NewMongoDBRepository(mongoClient, dbName, entityType, "oracle_results", "OracleResult")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"BaseEntityID":    true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"MatchID":         true,
		"ExternalMatchID": true,
		"GameID":          true,
		"Status":          true,
		"ConfidenceLevel": true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
		"FinalizedAt":     true,
		"UserID":          true,
		"TenantID":        true,
		"GroupID":          true,
		"ClientID":        true,
	}, map[string]string{
		"ID":              "baseentity._id",
		"BaseEntityID":    "baseentity._id",
		"VisibilityLevel": "baseentity.visibility_level",
		"VisibilityType":  "baseentity.visibility_type",
		"MatchID":         "match_id",
		"ExternalMatchID": "external_match_id",
		"GameID":          "game_id",
		"Status":          "status",
		"ConfidenceLevel": "confidence_level",
		"CreatedAt":       "baseentity.created_at",
		"UpdatedAt":       "baseentity.updated_at",
		"FinalizedAt":     "finalized_at",
		"UserID":          "baseentity.resource_owner.user_id",
		"TenantID":        "baseentity.resource_owner.tenant_id",
		"GroupID":          "baseentity.resource_owner.group_id",
		"ClientID":        "baseentity.resource_owner.client_id",
	})

	// Create indexes
	collection := repo.Collection()
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "match_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "external_match_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "baseentity.created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "submissions.source_type", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "publications.chain_id", Value: 1},
				{Key: "publications.status", Value: 1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		slog.Warn("failed to create oracle_results indexes", slog.String("error", err.Error()))
	}

	return &MongoOracleResultRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MongoOracleResultRepository) Save(ctx context.Context, result *oracle_entities.OracleResult) error {
	if result.GetID() == uuid.Nil {
		return fmt.Errorf("oracle result ID cannot be nil")
	}

	result.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Create(ctx, result)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save oracle result",
			slog.String("oracle_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to save oracle result: %w", err)
	}

	slog.InfoContext(ctx, "oracle result saved",
		slog.String("oracle_result_id", result.ID.String()),
	)
	return nil
}

func (r *MongoOracleResultRepository) FindByID(ctx context.Context, id uuid.UUID) (*oracle_entities.OracleResult, error) {
	// OracleResult embeds shared.BaseEntity which BSON serializes as nested "baseentity"
	// (not inlined). The top-level _id is a MongoDB ObjectID, while the UUID is at
	// baseentity._id. We must query by baseentity._id instead of top-level _id.
	var result oracle_entities.OracleResult
	filter := bson.M{"baseentity._id": id}

	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("oracle result not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find oracle result by ID", "id", id, "error", err)
		return nil, err
	}

	return &result, nil
}

func (r *MongoOracleResultRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*oracle_entities.OracleResult, error) {
	var result oracle_entities.OracleResult
	filter := bson.M{"match_id": matchID}

	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("oracle result not found for match: %s", matchID)
		}
		return nil, fmt.Errorf("failed to find oracle result: %w", err)
	}

	return &result, nil
}

func (r *MongoOracleResultRepository) FindByExternalMatchID(ctx context.Context, externalMatchID string) (*oracle_entities.OracleResult, error) {
	var result oracle_entities.OracleResult
	filter := bson.M{"external_match_id": externalMatchID}

	err := r.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("oracle result not found for external match: %s", externalMatchID)
		}
		return nil, fmt.Errorf("failed to find oracle result: %w", err)
	}

	return &result, nil
}

func (r *MongoOracleResultRepository) FindByStatus(ctx context.Context, status oracle_vo.OracleStatus, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	filter := bson.M{"status": string(status)}
	results, err := r.findMany(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

func (r *MongoOracleResultRepository) FindPendingPublication(ctx context.Context) ([]*oracle_entities.OracleResult, error) {
	filter := bson.M{"status": string(oracle_vo.OracleStatusConsensusReached)}
	return r.findMany(ctx, filter, 100, 0)
}

func (r *MongoOracleResultRepository) FindPublishedBefore(ctx context.Context, before time.Time) ([]*oracle_entities.OracleResult, error) {
	filter := bson.M{
		"status": string(oracle_vo.OracleStatusPublished),
		"baseentity.updated_at": bson.M{"$lt": before},
	}
	return r.findMany(ctx, filter, 100, 0)
}

func (r *MongoOracleResultRepository) Update(ctx context.Context, result *oracle_entities.OracleResult) error {
	if result.GetID() == uuid.Nil {
		return fmt.Errorf("oracle result ID cannot be nil")
	}

	result.UpdatedAt = time.Now().UTC()

	// OracleResult embeds BaseEntity as nested "baseentity" in BSON (not inlined).
	// The top-level _id is a MongoDB ObjectID; the UUID lives at baseentity._id.
	// We must filter by baseentity._id for the update to match the correct document.
	filter := bson.M{"baseentity._id": result.GetID()}

	updateResult, err := r.MongoDBRepository.Collection().ReplaceOne(ctx, filter, result)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update oracle result",
			slog.String("oracle_result_id", result.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update oracle result: %w", err)
	}

	if updateResult.MatchedCount == 0 {
		slog.WarnContext(ctx, "oracle result not found for update",
			slog.String("oracle_result_id", result.ID.String()),
		)
		return fmt.Errorf("oracle result not found for update: %s", result.ID)
	}

	return nil
}

func (r *MongoOracleResultRepository) Count(ctx context.Context, filter oracle_out.OracleResultFilter) (int64, error) {
	bsonFilter := r.buildFilter(filter)
	return r.countDocuments(ctx, bsonFilter)
}

func (r *MongoOracleResultRepository) Search(ctx context.Context, filter oracle_out.OracleResultFilter, limit int, offset int) ([]*oracle_entities.OracleResult, int64, error) {
	bsonFilter := r.buildFilter(filter)
	results, err := r.findMany(ctx, bsonFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.countDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

// --- Internal helpers ---

func (r *MongoOracleResultRepository) buildFilter(filter oracle_out.OracleResultFilter) bson.M {
	bsonFilter := bson.M{}

	if filter.GameID != nil {
		bsonFilter["game_id"] = *filter.GameID
	}
	if filter.Status != nil {
		bsonFilter["status"] = *filter.Status
	}
	if filter.MatchID != nil {
		bsonFilter["match_id"] = *filter.MatchID
	}
	if filter.ExternalMatchID != nil {
		bsonFilter["external_match_id"] = *filter.ExternalMatchID
	}
	if filter.MinConfidence != nil {
		bsonFilter["confidence_level"] = bson.M{"$gte": *filter.MinConfidence}
	}

	return bsonFilter
}

func (r *MongoOracleResultRepository) findMany(ctx context.Context, filter bson.M, limit int, offset int) ([]*oracle_entities.OracleResult, error) {
	collection := r.MongoDBRepository.Collection()

	opts := options.Find().SetSort(bson.M{"baseentity.created_at": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find oracle results: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*oracle_entities.OracleResult
	for cursor.Next(ctx) {
		var result oracle_entities.OracleResult
		if err := cursor.Decode(&result); err != nil {
			slog.WarnContext(ctx, "failed to decode oracle result", slog.String("error", err.Error()))
			continue
		}
		results = append(results, &result)
	}

	return results, nil
}

func (r *MongoOracleResultRepository) countDocuments(ctx context.Context, filter bson.M) (int64, error) {
	collection := r.MongoDBRepository.Collection()
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count oracle results: %w", err)
	}
	return count, nil
}
