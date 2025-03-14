package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type PlanRepository struct {
	MongoDBRepository[billing_entities.Plan]
}

func NewPlanRepository(client *mongo.Client, dbName string, entityType billing_entities.Plan, collectionName string) *PlanRepository {
	repo := MongoDBRepository[billing_entities.Plan]{
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
		"ID":              true,
		"Name":            true,
		"Description":     true,
		"Prices":          true,
		"OperationLimits": true,
		"IsFree":          true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
		"ResourceOwner":   true,
	}, map[string]string{
		"ID":              "baseentity._id",
		"Name":            "name",
		"Description":     "description",
		"Prices":          "prices",
		"OperationLimits": "operation_limits",
		"IsFree":          "is_free",
		"VisibilityLevel": "baseentity.visibility_level",
		"VisibilityType":  "baseentity.visibility_type",
		"ResourceOwner":   "baseentity.resource_owner",
		"TenantID":        "baseentity.resource_owner.tenant_id",
		"UserID":          "baseentity.resource_owner.user_id",
		"GroupID":         "baseentity.resource_owner.group_id",
		"ClientID":        "baseentity.resource_owner.client_id",
		"CreatedAt":       "baseentity.create_at",
		"UpdatedAt":       "baseentity.updated_at",
	})

	return &PlanRepository{
		repo,
	}
}
