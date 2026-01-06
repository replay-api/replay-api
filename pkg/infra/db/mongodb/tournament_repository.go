package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoTournamentRepository struct {
	mongodb.MongoDBRepository[tournament_entities.Tournament]
}

func NewMongoTournamentRepository(mongoClient *mongo.Client, dbName string) tournament_out.TournamentRepository {
	entityType := tournament_entities.Tournament{}
	repo := mongodb.NewMongoDBRepository[tournament_entities.Tournament](mongoClient, dbName, entityType, "tournaments", "Tournament")

	repo.InitQueryableFields(map[string]bool{
		"ID":                true,
		"Name":              true,
		"Description":       true,
		"GameID":            true,
		"GameMode":          true,
		"Region":            true,
		"Format":            true,
		"MaxParticipants":   true,
		"MinParticipants":   true,
		"Status":            true,
		"StartTime":         true,
		"EndTime":           true,
		"RegistrationOpen":  true,
		"RegistrationClose": true,
		"OrganizerID":       true,
		"CreatedAt":         true,
		"UpdatedAt":         true,
	}, map[string]string{
		"ID":                "_id",
		"Name":              "name",
		"Description":       "description",
		"GameID":            "game_id",
		"GameMode":          "game_mode",
		"Region":            "region",
		"Format":            "format",
		"MaxParticipants":   "max_participants",
		"MinParticipants":   "min_participants",
		"Status":            "status",
		"StartTime":         "start_time",
		"EndTime":           "end_time",
		"RegistrationOpen":  "registration_open",
		"RegistrationClose": "registration_close",
		"OrganizerID":       "organizer_id",
		"CreatedAt":         "created_at",
		"UpdatedAt":         "updated_at",
	})

	return &MongoTournamentRepository{
		MongoDBRepository: *repo,
	}
}

func (r *MongoTournamentRepository) Save(ctx context.Context, tournament *tournament_entities.Tournament) error {
	if tournament.GetID() == uuid.Nil {
		return fmt.Errorf("tournament ID cannot be nil")
	}

	tournament.UpdatedAt = time.Now().UTC()

	_, err := r.MongoDBRepository.Collection().InsertOne(ctx, tournament)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save tournament", "tournament_id", tournament.ID, "error", err)
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament saved successfully", "tournament_id", tournament.ID, "name", tournament.Name)
	return nil
}

func (r *MongoTournamentRepository) FindByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	var tournament tournament_entities.Tournament

	filter := bson.M{"_id": id}
	err := r.MongoDBRepository.Collection().FindOne(ctx, filter).Decode(&tournament)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("tournament not found: %s", id)
		}
		slog.ErrorContext(ctx, "failed to find tournament by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find tournament: %w", err)
	}

	return &tournament, nil
}

// GetByID implements the Searchable interface
func (r *MongoTournamentRepository) GetByID(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error) {
	return r.FindByID(ctx, id)
}

func (r *MongoTournamentRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error) {
	filter := bson.M{"organizer_id": organizerID}
	findOptions := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tournaments by organizer", "organizer_id", organizerID, "error", err)
		return nil, fmt.Errorf("failed to find tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	tournaments := make([]*tournament_entities.Tournament, 0)
	for cursor.Next(ctx) {
		var tournament tournament_entities.Tournament
		if err := cursor.Decode(&tournament); err != nil {
			slog.ErrorContext(ctx, "failed to decode tournament", "error", err)
			continue
		}
		tournaments = append(tournaments, &tournament)
	}

	return tournaments, nil
}

func (r *MongoTournamentRepository) FindByGameAndRegion(ctx context.Context, gameID, region string, statusFilter []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error) {
	filter := bson.M{
		"game_id": gameID,
	}

	if region != "" {
		filter["region"] = region
	}

	if len(statusFilter) > 0 {
		filter["status"] = bson.M{"$in": statusFilter}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tournaments", "game_id", gameID, "region", region, "error", err)
		return nil, fmt.Errorf("failed to find tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	tournaments := make([]*tournament_entities.Tournament, 0)
	for cursor.Next(ctx) {
		var tournament tournament_entities.Tournament
		if err := cursor.Decode(&tournament); err != nil {
			slog.ErrorContext(ctx, "failed to decode tournament", "error", err)
			continue
		}
		tournaments = append(tournaments, &tournament)
	}

	slog.InfoContext(ctx, "found tournaments", "game_id", gameID, "region", region, "count", len(tournaments))
	return tournaments, nil
}

func (r *MongoTournamentRepository) FindUpcoming(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error) {
	now := time.Now().UTC()

	filter := bson.M{
		"game_id": gameID,
		"status": bson.M{"$in": []tournament_entities.TournamentStatus{
			tournament_entities.TournamentStatusRegistration,
			tournament_entities.TournamentStatusReady,
		}},
		"start_time": bson.M{"$gt": now},
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find upcoming tournaments", "game_id", gameID, "error", err)
		return nil, fmt.Errorf("failed to find upcoming tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	tournaments := make([]*tournament_entities.Tournament, 0)
	for cursor.Next(ctx) {
		var tournament tournament_entities.Tournament
		if err := cursor.Decode(&tournament); err != nil {
			slog.ErrorContext(ctx, "failed to decode tournament", "error", err)
			continue
		}
		tournaments = append(tournaments, &tournament)
	}

	slog.InfoContext(ctx, "found upcoming tournaments", "game_id", gameID, "count", len(tournaments))
	return tournaments, nil
}

func (r *MongoTournamentRepository) FindInProgress(ctx context.Context, limit int) ([]*tournament_entities.Tournament, error) {
	filter := bson.M{
		"status": tournament_entities.TournamentStatusInProgress,
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find in-progress tournaments", "error", err)
		return nil, fmt.Errorf("failed to find in-progress tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	tournaments := make([]*tournament_entities.Tournament, 0)
	for cursor.Next(ctx) {
		var tournament tournament_entities.Tournament
		if err := cursor.Decode(&tournament); err != nil {
			slog.ErrorContext(ctx, "failed to decode tournament", "error", err)
			continue
		}
		tournaments = append(tournaments, &tournament)
	}

	slog.InfoContext(ctx, "found in-progress tournaments", "count", len(tournaments))
	return tournaments, nil
}

func (r *MongoTournamentRepository) Update(ctx context.Context, tournament *tournament_entities.Tournament) error {
	if tournament.GetID() == uuid.Nil {
		return fmt.Errorf("tournament ID cannot be nil")
	}

	tournament.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": tournament.ID}
	update := bson.M{"$set": tournament}

	result, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update tournament", "tournament_id", tournament.ID, "error", err)
		return fmt.Errorf("failed to update tournament: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("tournament not found for update: %s", tournament.ID)
	}

	slog.InfoContext(ctx, "tournament updated successfully", "tournament_id", tournament.ID, "status", tournament.Status)
	return nil
}

func (r *MongoTournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := r.MongoDBRepository.Collection().DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete tournament", "id", id, "error", err)
		return fmt.Errorf("failed to delete tournament: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("tournament not found for deletion: %s", id)
	}

	slog.InfoContext(ctx, "tournament deleted successfully", "tournament_id", id)
	return nil
}

func (r *MongoTournamentRepository) FindPlayerTournaments(ctx context.Context, playerID uuid.UUID, statusFilter []tournament_entities.TournamentStatus) ([]*tournament_entities.Tournament, error) {
	filter := bson.M{
		"participants.player_id": playerID,
	}

	if len(statusFilter) > 0 {
		filter["status"] = bson.M{"$in": statusFilter}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})

	cursor, err := r.MongoDBRepository.Collection().Find(ctx, filter, findOptions)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find player tournaments", "player_id", playerID, "error", err)
		return nil, fmt.Errorf("failed to find player tournaments: %w", err)
	}
	defer cursor.Close(ctx)

	tournaments := make([]*tournament_entities.Tournament, 0)
	for cursor.Next(ctx) {
		var tournament tournament_entities.Tournament
		if err := cursor.Decode(&tournament); err != nil {
			slog.ErrorContext(ctx, "failed to decode tournament", "error", err)
			continue
		}
		tournaments = append(tournaments, &tournament)
	}

	slog.InfoContext(ctx, "found player tournaments", "player_id", playerID, "count", len(tournaments))
	return tournaments, nil
}

// Ensure MongoTournamentRepository implements TournamentRepository interface
var _ tournament_out.TournamentRepository = (*MongoTournamentRepository)(nil)
