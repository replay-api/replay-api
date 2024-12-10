package db

import (
	"context"
	"log"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type EventsRepository struct {
	MongoDBRepository[replay_entity.GameEvent]
}

func NewEventsRepository(client *mongo.Client, dbName string, entityType *replay_entity.GameEvent, collectionName string) *EventsRepository {
	repo := MongoDBRepository[replay_entity.GameEvent]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		queryableFields:   make(map[string]bool),
	}

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
		repo,
	}
}

// func (r *EventsRepository) Search(ctx context.Context, s common.Search) ([]replay_entity.GameEvent, error) {
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
	collection := r.mongoClient.Database(r.dbName).Collection(r.collectionName)

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
