package db

import (
	"context"
	"log"
	"log/slog"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type EventsRepository struct {
	mongodb.MongoDBRepository[replay_entity.GameEvent]
}

func NewEventsRepository(client *mongo.Client, dbName string, entityType *replay_entity.GameEvent, collectionName string) *EventsRepository {
	repo := mongodb.NewMongoDBRepository[replay_entity.GameEvent](client, dbName, *entityType, collectionName, "GameEvent")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"GameID":          true,
		"MatchID":         true,
		"Type":            true,
		"Time":            true,
		"EventData":       true,
		"PlayerStats":     true,
		"NetworkPlayerID": true,
		"PlayerName":      true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
	}, map[string]string{
		"ID":              "_id",
		"GameID":          "game_id",
		"MatchID":         "match_id",
		"Type":            "type",
		"Time":            "event_time",
		"EventData":       "event_data",
		"PlayerStats":     "player_stats",
		"NetworkPlayerID": "network_player_id",
		"PlayerName":      "player_name",
		"ResourceOwner":   "resource_owner",
		"CreatedAt":       "created_at",
	})

	return &EventsRepository{
		MongoDBRepository: *repo,
	}
}

// func (r *EventsRepository) Search(ctx context.Context, s shared.Search) ([]replay_entity.GameEvent, error) {
// 	cursor, err := r.Query(ctx, s)
// 	if cursor != nil {
// 		defer cursor.Close(ctx)
// 	}

// 	if err != nil {
// 		slog.ErrorContext(ctx, "error querying game events", "err", err)
// 		return nil, err
// 	}

// 	gameEvents := make([]replay_entity.GameEvent, 0)

// 	for cursor.Next(ctx) {
// 		var replayFile replay_entity.GameEvent
// 		err := cursor.Decode(&replayFile)

// 		if err != nil {
// 			slog.ErrorContext(ctx, "error decoding game event", "err", err)
// 			return nil, err
// 		}

// 		gameEvents = append(gameEvents, replayFile)
// 	}

// 	return gameEvents, nil
// }

// func (r *EventsRepository) CreateMany(createCtx context.Context, events []replay_entity.GameEvent) error {
// 	collection := r.mongoClient.Database("replay").Collection("game_events")

// 	toInsert := make([]interface{}, len(events))

// 	for i := range events {
// 		toInsert[i] = events[i]
// 	}

// 	_, err := collection.InsertMany(createCtx, toInsert)
// 	if err != nil {
// 		slog.ErrorContext(createCtx, err.Error())
// 		return err
// 	}

// 	return nil
// }

func (r *EventsRepository) GetByGameIDAndMatchID(queryCtx context.Context, gameID string, matchID string) ([]replay_entity.GameEvent, error) {
	collection := r.MongoDBRepository.Collection()

	query := bson.D{
		{Key: "game_id", Value: gameID},
		{Key: "match_id", Value: matchID},
	}

	cur, err := collection.Find(queryCtx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(queryCtx)

	res := []replay_entity.GameEvent{}
	for cur.Next(queryCtx) {
		var event *replay_entity.GameEvent
		err := cur.Decode(&event)
		if err != nil {
			log.Fatal(err)
		}

		res = append(res, *event)
	}

	if err := cur.Err(); err != nil {
		slog.ErrorContext(queryCtx, err.Error())
		return nil, err
	}

	return res, nil
}

// GetMatchEvents returns events for a specific match with pagination and optional filtering
func (r *EventsRepository) GetMatchEvents(ctx context.Context, gameID string, matchID uuid.UUID, limit, offset int, eventType string) ([]replay_entity.GameEvent, error) {
	collection := r.MongoDBRepository.Collection()

	// Build query with proper UUID type
	filter := bson.D{
		{Key: "game_id", Value: gameID},
		{Key: "match_id", Value: matchID},
	}

	if eventType != "" {
		filter = append(filter, bson.E{Key: "type", Value: eventType})
	}

	// Set up find options with pagination and sorting
	opts := options.Find().
		SetSort(bson.D{{Key: "tick_id", Value: 1}}).
		SetSkip(int64(offset))

	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		slog.ErrorContext(ctx, "error querying match events", "err", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var events []replay_entity.GameEvent
	for cur.Next(ctx) {
		var event replay_entity.GameEvent
		if err := cur.Decode(&event); err != nil {
			slog.ErrorContext(ctx, "error decoding game event", "err", err)
			return nil, err
		}
		events = append(events, event)
	}

	if err := cur.Err(); err != nil {
		slog.ErrorContext(ctx, "cursor error", "err", err)
		return nil, err
	}

	return events, nil
}

// GetMatchEventsWithCount returns events for a specific match with total count for pagination
// Supports multiple event types for efficient filtering (e.g., ["kill", "clutchstart", "clutchend"])
func (r *EventsRepository) GetMatchEventsWithCount(ctx context.Context, gameID string, matchID uuid.UUID, limit, offset int, eventTypes []string) ([]replay_entity.GameEvent, int64, error) {
	collection := r.MongoDBRepository.Collection()

	// Build query with proper UUID type
	filter := bson.D{
		{Key: "game_id", Value: gameID},
		{Key: "match_id", Value: matchID},
	}

	// Support multiple event types with $in operator for scalable filtering
	if len(eventTypes) == 1 {
		filter = append(filter, bson.E{Key: "type", Value: eventTypes[0]})
	} else if len(eventTypes) > 1 {
		filter = append(filter, bson.E{Key: "type", Value: bson.D{{Key: "$in", Value: eventTypes}}})
	}

	// Get total count first (for pagination metadata)
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error counting match events", "err", err)
		return nil, 0, err
	}

	// Set up find options with pagination and sorting
	opts := options.Find().
		SetSort(bson.D{{Key: "tick_id", Value: 1}}).
		SetSkip(int64(offset))

	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		slog.ErrorContext(ctx, "error querying match events", "err", err)
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var events []replay_entity.GameEvent
	for cur.Next(ctx) {
		var event replay_entity.GameEvent
		if err := cur.Decode(&event); err != nil {
			slog.ErrorContext(ctx, "error decoding game event", "err", err)
			return nil, 0, err
		}
		events = append(events, event)
	}

	if err := cur.Err(); err != nil {
		slog.ErrorContext(ctx, "cursor error", "err", err)
		return nil, 0, err
	}

	return events, totalCount, nil
}
