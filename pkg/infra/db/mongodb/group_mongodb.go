package db

import (
	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type GroupRepository struct {
	mongodb.MongoDBRepository[iam_entities.Group]
}

func NewGroupRepository(client *mongo.Client, dbName string, entityType *iam_entities.Group, collectionName string) *GroupRepository {
	repo := mongodb.NewMongoDBRepository[iam_entities.Group](client, dbName, *entityType, collectionName, "Group")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"Name":          true,
		"Type":          true,
		"ParentGroupID": true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Name":          "name",
		"Type":          "type",
		"ParentGroupID": "parent_group_id",
		"ResourceOwner": "resource_owner",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &GroupRepository{
		MongoDBRepository: *repo,
	}
}
