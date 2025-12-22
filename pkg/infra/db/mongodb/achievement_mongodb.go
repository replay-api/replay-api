package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	achievement_entities "github.com/replay-api/replay-api/pkg/domain/achievement/entities"
	achievement_out "github.com/replay-api/replay-api/pkg/domain/achievement/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	achievementCollection       = "achievements"
	playerAchievementCollection = "player_achievements"
)

// AchievementMongoDBRepository implements AchievementRepository using MongoDB
type AchievementMongoDBRepository struct {
	achievementColl       *mongo.Collection
	playerAchievementColl *mongo.Collection
}

// NewAchievementMongoDBRepository creates a new AchievementMongoDBRepository
func NewAchievementMongoDBRepository(client *mongo.Client, dbName string) achievement_out.AchievementRepository {
	return &AchievementMongoDBRepository{
		achievementColl:       client.Database(dbName).Collection(achievementCollection),
		playerAchievementColl: client.Database(dbName).Collection(playerAchievementCollection),
	}
}

// GetAllAchievements retrieves all achievements
func (r *AchievementMongoDBRepository) GetAllAchievements(ctx context.Context) ([]achievement_entities.Achievement, error) {
	cursor, err := r.achievementColl.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get all achievements", "error", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []achievement_entities.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// GetAchievementByID retrieves an achievement by ID
func (r *AchievementMongoDBRepository) GetAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.Achievement, error) {
	var achievement achievement_entities.Achievement
	err := r.achievementColl.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &achievement, nil
}

// GetAchievementByCode retrieves an achievement by code
func (r *AchievementMongoDBRepository) GetAchievementByCode(ctx context.Context, code string) (*achievement_entities.Achievement, error) {
	var achievement achievement_entities.Achievement
	err := r.achievementColl.FindOne(ctx, bson.M{"code": code}).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &achievement, nil
}

// GetAchievementsByCategory retrieves achievements by category
func (r *AchievementMongoDBRepository) GetAchievementsByCategory(ctx context.Context, category achievement_entities.AchievementCategory) ([]achievement_entities.Achievement, error) {
	cursor, err := r.achievementColl.Find(ctx, bson.M{"category": category, "is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []achievement_entities.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// CreateAchievement creates a new achievement
func (r *AchievementMongoDBRepository) CreateAchievement(ctx context.Context, achievement *achievement_entities.Achievement) error {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	
	_, err := r.achievementColl.InsertOne(ctx, achievement)
	return err
}

// UpdateAchievement updates an existing achievement
func (r *AchievementMongoDBRepository) UpdateAchievement(ctx context.Context, achievement *achievement_entities.Achievement) error {
	achievement.UpdatedAt = time.Now()
	
	_, err := r.achievementColl.UpdateOne(ctx, bson.M{"_id": achievement.ID}, bson.M{"$set": achievement})
	return err
}

// GetPlayerAchievements retrieves all achievements for a player
func (r *AchievementMongoDBRepository) GetPlayerAchievements(ctx context.Context, playerID uuid.UUID) ([]achievement_entities.PlayerAchievement, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := r.playerAchievementColl.Find(ctx, bson.M{"player_id": playerID}, opts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get player achievements", "error", err, "player_id", playerID)
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []achievement_entities.PlayerAchievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// GetPlayerAchievementByCode retrieves a player achievement by code
func (r *AchievementMongoDBRepository) GetPlayerAchievementByCode(ctx context.Context, playerID uuid.UUID, achievementCode string) (*achievement_entities.PlayerAchievement, error) {
	// First get the achievement by code
	achievement, err := r.GetAchievementByCode(ctx, achievementCode)
	if err != nil || achievement == nil {
		return nil, err
	}

	var playerAchievement achievement_entities.PlayerAchievement
	err = r.playerAchievementColl.FindOne(ctx, bson.M{
		"player_id":      playerID,
		"achievement_id": achievement.ID,
	}).Decode(&playerAchievement)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	
	return &playerAchievement, nil
}

// GetPlayerAchievementByID retrieves a player achievement by ID
func (r *AchievementMongoDBRepository) GetPlayerAchievementByID(ctx context.Context, id uuid.UUID) (*achievement_entities.PlayerAchievement, error) {
	var playerAchievement achievement_entities.PlayerAchievement
	err := r.playerAchievementColl.FindOne(ctx, bson.M{"_id": id}).Decode(&playerAchievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &playerAchievement, nil
}

// CreatePlayerAchievement creates a new player achievement
func (r *AchievementMongoDBRepository) CreatePlayerAchievement(ctx context.Context, pa *achievement_entities.PlayerAchievement) error {
	pa.CreatedAt = time.Now()
	pa.UpdatedAt = time.Now()
	
	_, err := r.playerAchievementColl.InsertOne(ctx, pa)
	return err
}

// UpdatePlayerAchievement updates an existing player achievement
func (r *AchievementMongoDBRepository) UpdatePlayerAchievement(ctx context.Context, pa *achievement_entities.PlayerAchievement) error {
	pa.UpdatedAt = time.Now()
	
	_, err := r.playerAchievementColl.UpdateOne(ctx, bson.M{"_id": pa.ID}, bson.M{"$set": pa})
	return err
}

// GetRecentUnlocks retrieves recent achievement unlocks for a player
func (r *AchievementMongoDBRepository) GetRecentUnlocks(ctx context.Context, playerID uuid.UUID, limit int) ([]achievement_entities.PlayerAchievement, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "unlocked_at", Value: -1}}).
		SetLimit(int64(limit))
	
	filter := bson.M{
		"player_id":   playerID,
		"unlocked_at": bson.M{"$ne": nil},
	}
	
	cursor, err := r.playerAchievementColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []achievement_entities.PlayerAchievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// GetUnlockedCount returns the count of unlocked achievements for a player
func (r *AchievementMongoDBRepository) GetUnlockedCount(ctx context.Context, playerID uuid.UUID) (int, error) {
	filter := bson.M{
		"player_id":   playerID,
		"unlocked_at": bson.M{"$ne": nil},
	}
	
	count, err := r.playerAchievementColl.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	
	return int(count), nil
}

