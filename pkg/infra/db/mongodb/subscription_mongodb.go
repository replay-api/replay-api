package db

import (
	"context"
	"log/slog"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type SubscriptionRepository struct {
	MongoDBRepository[billing_entities.Subscription]
}

func NewSubscriptionRepository(client *mongo.Client, dbName string, entityType billing_entities.Subscription, collectionName string) *SubscriptionRepository {
	repo := MongoDBRepository[billing_entities.Subscription]{
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
		"PlanID":          true,
		"BillingPeriod":   true,
		"Status":          true,
		"StartDate":       true,
		"EndDate":         true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   true,
		"CreatedAt":       true,
		"UpdatedAt":       true,
		"History":         true,
		"IsFree":          true,
	}, map[string]string{
		"ID":              "baseentity._id",
		"PlanID":          "plan_id",
		"BillingPeriod":   "billing_period",
		"Status":          "status",
		"StartDate":       "start_at",
		"EndDate":         "end_at",
		"VisibilityLevel": "baseentity.visibility_level",
		"VisibilityType":  "baseentity.visibility_type",
		"ResourceOwner":   "baseentity.resource_owner",
		"TenantID":        "baseentity.resource_owner.tenant_id",
		"UserID":          "baseentity.resource_owner.user_id",
		"GroupID":         "baseentity.resource_owner.group_id",
		"ClientID":        "baseentity.resource_owner.client_id",
		"CreatedAt":       "baseentity.create_at",
		"UpdatedAt":       "baseentity.updated_at",
		"History":         "history",
		"IsFree":          "is_free",
	})

	return &SubscriptionRepository{
		repo,
	}
}

func (r *SubscriptionRepository) Search(ctx context.Context, s common.Search) ([]billing_entities.Subscription, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error querying player profile entity", "err", err)
		return nil, err
	}

	subscriptions := make([]billing_entities.Subscription, 0)

	for cursor.Next(ctx) {
		var p billing_entities.Subscription
		err := cursor.Decode(&p)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding player profile entity", "err", err)
			return nil, err
		}

		subscriptions = append(subscriptions, p)
	}

	return subscriptions, nil
}
