package db

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
)

type SquadRepository struct {
	MongoDBRepository[squad_entities.Squad]
}

func NewSquadRepository(client *mongo.Client, dbName string, entityType squad_entities.Squad, collectionName string) *SquadRepository {
	repo := MongoDBRepository[squad_entities.Squad]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		bsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		queryableFields:   make(map[string]bool),
		collection:        client.Database(dbName).Collection(collectionName),
	}

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"GroupID":       true,
		"GameID":        true,
		"FullName":      true,
		"ShortName":     true,
		"Description":   true,
		"Profiles":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":                 "_id",
		"GroupID":            "group_id",
		"GameID":             "game_id",
		"FullName":           "full_name",
		"CurrentDisplayName": "short_name",
		"Symbol":             "symbol",
		"Description":        "description",
		"Profiles":           "profiles",
		"ResourceOwner":      "resource_owner",
		"TenantID":           "resource_owner.tenant_id",
		"UserID":             "resource_owner.user_id",
		// "GroupID":            "resource_owner.group_id",
		"ClientID":  "resource_owner.client_id",
		"CreatedAt": "create_at",
		"UpdatedAt": "updated_at",
	})

	return &SquadRepository{
		repo,
	}
}

func (r *SquadRepository) Search(ctx context.Context, s common.Search) ([]squad_entities.Squad, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player entity", "err", err)
		return nil, err
	}

	players := make([]squad_entities.Squad, 0)

	for cursor.Next(ctx) {
		var p squad_entities.Squad
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player entity", "err", err)
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}

func (r *SquadRepository) Update(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error) {
	filter := map[string]interface{}{"_id": squad.ID}
	update := map[string]interface{}{"$set": squad}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "error updating squad", "err", err, "squad_id", squad.ID)
		return nil, err
	}

	return squad, nil
}

func (r *SquadRepository) Delete(ctx context.Context, squadID uuid.UUID) error {
	filter := map[string]interface{}{"_id": squadID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting squad", "err", err, "squad_id", squadID)
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
