package db

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type PlayerProfileRepository struct {
	mongodb.MongoDBRepository[squad_entities.PlayerProfile]
}

func NewPlayerProfileRepository(client *mongo.Client, dbName string, entityType squad_entities.PlayerProfile, collectionName string) *PlayerProfileRepository {
	repo := mongodb.NewMongoDBRepository[squad_entities.PlayerProfile](client, dbName, entityType, collectionName, "PlayerProfile")

	repo.InitQueryableFields(map[string]bool{
		"ID":              true,
		"GameID":          true,
		"Nickname":        true,
		"SlugURI":         true,
		"Avatar":          true,
		"Roles":           true,
		"Description":     true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}, map[string]string{
		"ID":              "baseentity._id", // TODO: review ; (opcional: aqui a exceção se tornou a regra. deixar default o que está na annotation da prop.) talvez seja melhor refletir os tipos nas anotacoes json/bson
		"GameID":          "game_id",
		"Nickname":        "nickname",
		"SlugURI":         "slug_uri",
		"Avatar":          "avatar",
		"Roles":           "roles",
		"Description":     "description",
		"VisibilityLevel": "baseentity.visibility_level",
		"VisibilityType":  "baseentity.visibility_type",
		"ResourceOwner":   "baseentity.resource_owner", // TODO: principalmente resource ownership, que é padronizado.
		"TenantID":        "baseentity.resource_owner.tenant_id",
		"UserID":          "baseentity.resource_owner.user_id",
		"GroupID":         "baseentity.resource_owner.group_id",
		"ClientID":        "baseentity.resource_owner.client_id",
		"CreatedAt":       "baseentity.create_at",
		"UpdatedAt":       "baseentity.updated_at",
	})

	return &PlayerProfileRepository{
		MongoDBRepository: *repo,
	}
}

func (r *PlayerProfileRepository) Search(ctx context.Context, s shared.Search) ([]squad_entities.PlayerProfile, error) {
	return r.MongoDBRepository.Search(ctx, s)
}

func (r *PlayerProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*squad_entities.PlayerProfile, error) {
	return r.MongoDBRepository.GetByID(ctx, id)
}

func (r *PlayerProfileRepository) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	return r.MongoDBRepository.Compile(ctx, searchParams, resultOptions)
}

func (r *PlayerProfileRepository) Create(ctx context.Context, profile *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error) {
	return r.MongoDBRepository.Create(ctx, profile)
}

func (r *PlayerProfileRepository) CreateMany(ctx context.Context, profiles []*squad_entities.PlayerProfile) error {
	return r.MongoDBRepository.CreateMany(ctx, profiles)
}

func (r *PlayerProfileRepository) Update(ctx context.Context, profile *squad_entities.PlayerProfile) (*squad_entities.PlayerProfile, error) {
	filter := map[string]interface{}{"baseentity._id": profile.ID}
	update := map[string]interface{}{"$set": profile}

	_, err := r.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "error updating player profile", "err", err, "profile_id", profile.ID)
		return nil, err
	}

	return profile, nil
}

func (r *PlayerProfileRepository) Delete(ctx context.Context, profileID uuid.UUID) error {
	filter := map[string]interface{}{"baseentity._id": profileID}

	result, err := r.DeleteOne(ctx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting player profile", "err", err, "profile_id", profileID)
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
