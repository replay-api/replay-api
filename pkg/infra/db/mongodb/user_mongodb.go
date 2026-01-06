package db

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type UserRepository struct {
	mongodb.MongoDBRepository[iam_entities.User]
}

func NewUserRepository(client *mongo.Client, dbName string, entityType *iam_entities.User, collectionName string) *UserRepository {
	repo := mongodb.NewMongoDBRepository[iam_entities.User](client, dbName, *entityType, collectionName, "User")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"Name":          true,
		"Type":          true,
		"ParentUserID":  true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Name":          "name",
		"Type":          "type",
		"ParentUserID":  "parent_group_id",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &UserRepository{
		MongoDBRepository: *repo,
	}
}

func (r *UserRepository) Search(ctx context.Context, s shared.Search) ([]iam_entities.User, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying user entity", "err", err)
		return nil, err
	}

	users := make([]iam_entities.User, 0)

	for cursor.Next(ctx) {
		var p iam_entities.User
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding user entity", "err", err)
			return nil, err
		}

		users = append(users, p)
	}

	return users, nil
}
