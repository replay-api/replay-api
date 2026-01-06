package db

import (
	"go.mongodb.org/mongo-driver/mongo"

	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
)

type MembershipRepository struct {
	mongodb.MongoDBRepository[iam_entities.Membership]
}

func NewMembershipRepository(client *mongo.Client, dbName string, entityType *iam_entities.Membership, collectionName string) *MembershipRepository {
	repo := mongodb.NewMongoDBRepository[iam_entities.Membership](client, dbName, *entityType, collectionName, "Membership")

	repo.InitQueryableFields(map[string]bool{
		"ID":            true,
		"Type":          true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}, map[string]string{
		"ID":            "_id",
		"Type":          "type",
		"ResourceOwner": "resource_owner",
		"UserID":        "resource_owner.user_id",
		"CreatedAt":     "created_at",
		"UpdatedAt":     "updated_at",
	})

	return &MembershipRepository{
		MongoDBRepository: *repo,
	}
}
