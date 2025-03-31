package db

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type BillableEntryRepository struct {
	MongoDBRepository[billing_entities.BillableEntry]
}

func (repo *BillableEntryRepository) Find(ctx context.Context, filter interface{}, result interface{}) error {
	collection := repo.collection
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, result)
}

func NewBillableEntryRepository(client *mongo.Client, dbName string, entityType billing_entities.BillableEntry, collectionName string) *BillableEntryRepository {
	repo := MongoDBRepository[billing_entities.BillableEntry]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entityType),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entityType).Name(),
		QueryableFields:   make(map[string]bool),
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

func (repo *BillableEntryRepository) GetEntriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry, error) {
	var entries []billing_entities.BillableEntry
	err := repo.Find(ctx, bson.M{"subscription_id": subscriptionID}, &entries)

	if err != nil {
		return nil, err
	}

	entriesMap := make(map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry)
	for _, entry := range entries {
		entriesMap[entry.OperationID] = append(entriesMap[entry.OperationID], entry)
	}

	return entriesMap, nil
}
