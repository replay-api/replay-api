package db

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SquadRepository struct {
	mongodb.MongoDBRepository[squad_entities.Squad]
}

func NewSquadRepository(client *mongo.Client, dbName string, entityType squad_entities.Squad, collectionName string) *SquadRepository {
	repo := mongodb.NewMongoDBRepository[squad_entities.Squad](client, dbName, entityType, collectionName, "Squad")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"GameID":          true,
		"Name":            true,
		"Symbol":          true,
		"SlugURI":         true,
		"Description":     true,
		"Membership":      true,
		"LogoURI":         true,
		"BannerURI":       true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":                 "baseentity._id",
		"GameID":             "game_id",
		"Name":               "name",
		"Symbol":             "symbol",
		"SlugURI":            "slug_uri",
		"Description":        "description",
		"Membership":         "membership",
		"Membership.Type":    "membership.type",
		"Membership.Roles":   "membership.roles",
		"Membership.Status":  "membership.status",
		"Membership.History": "membership.history",
		"LogoURI":            "logo_uri",
		"BannerURI":          "banner_uri",
		"VisibilityLevel":    "baseentity.visibility_level",
		"VisibilityType":     "baseentity.visibility_type",
		"ResourceOwner":      "baseentity.resource_owner",
		"TenantID":           "baseentity.resource_owner.tenant_id",
		"UserID":             "baseentity.resource_owner.user_id",
		"GroupID":            "baseentity.resource_owner.group_id",
		"ClientID":           "baseentity.resource_owner.client_id",
		"CreatedAt":          "baseentity.created_at",
		"UpdatedAt":          "baseentity.updated_at",
	})

	return &SquadRepository{
		MongoDBRepository: *repo,
	}
}

func (r *SquadRepository) Search(ctx context.Context, s shared.Search) ([]squad_entities.Squad, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying squad entity", "err", err)
		return nil, err
	}

	squads := make([]squad_entities.Squad, 0)

	for cursor.Next(ctx) {
		var p squad_entities.Squad
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding squad entity", "err", err)
			return nil, err
		}

		squads = append(squads, p)
	}

	return squads, nil
}

func (r *SquadRepository) Update(ctx context.Context, squad *squad_entities.Squad) (*squad_entities.Squad, error) {
	filter := map[string]interface{}{"_id": squad.ID}
	update := map[string]interface{}{"$set": squad}

	_, err := r.MongoDBRepository.Collection().UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "error updating squad", "err", err, "squad_id", squad.ID)
		return nil, err
	}

	return squad, nil
}

func (r *SquadRepository) Delete(ctx context.Context, squadID uuid.UUID) error {
	filter := map[string]interface{}{"_id": squadID}

	result, err := r.MongoDBRepository.Collection().DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting squad", "err", err, "squad_id", squadID)
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
