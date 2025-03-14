package db

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type BillableEntryRepository struct {
	MongoDBRepository[billing_entities.BillableEntry]
}

func NewBillableEntryRepository(client *mongo.Client, dbName string, entityType billing_entities.BillableEntry, collectionName string) *BillableEntryRepository {
	repo := MongoDBRepository[billing_entities.BillableEntry]{
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
		"ID":             true,
		"CreatedAt":      true,
		"UpdatedAt":      true,
		"ResourceOwner":  true,
		"OperationID":    true,
		"PlanID":         true,
		"Amount":         true,
		"PayableID":      true,
		"SubscriptionID": true,
	}, map[string]string{
		"OperationID":     "operation_id",
		"PlanID":          "plan_id",
		"Amount":          "amount",
		"PayableID":       "payable_id",
		"SubscriptionID":  "subscription_id",
		"ID":              "baseentity._id",
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

	return &BillableEntryRepository{
		repo,
	}
}
