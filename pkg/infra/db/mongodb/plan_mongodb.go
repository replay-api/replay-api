package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type PlanRepository struct {
	mongodb.MongoDBRepository[billing_entities.Plan]
}

func NewPlanRepository(client *mongo.Client, dbName string, entityType billing_entities.Plan, collectionName string) *PlanRepository {
	repo := mongodb.NewMongoDBRepository[billing_entities.Plan](client, dbName, entityType, collectionName, "Plan")

	repo.InitQueryableFields(map[string]bool{
		"ID":                   true,
		"Name":                 true,
		"Description":          true,
		"Kind":                 true,
		"CustomerType":         true,
		"Prices":               true,
		"OperationLimits":      true,
		"IsFree":               true,
		"IsAvailable":          true,
		"IsLegacy":             true,
		"IsActive":             true,
		"DisplayPriorityScore": true,
		"Regions":              true,
		"Languages":            true,
		"EffectiveDate":        true,
		"ExpirationDate":       true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
		"ResourceOwner":        true,
	}, map[string]string{
		"ID":                   "baseentity._id",
		"Name":                 "name",
		"Description":          "description",
		"Kind":                 "kind",
		"CustomerType":         "customer_type",
		"Prices":               "prices",
		"OperationLimits":      "operation_limits",
		"IsFree":               "is_free",
		"IsAvailable":          "is_available",
		"IsLegacy":             "is_legacy",
		"IsActive":             "is_active",
		"DisplayPriorityScore": "display_priority_score",
		"Regions":              "regions",
		"Languages":            "languages",
		"EffectiveDate":        "effective_date",
		"ExpirationDate":       "expiration_date",
		"VisibilityLevel":      "baseentity.visibility_level",
		"VisibilityType":       "baseentity.visibility_type",
		"ResourceOwner":        "baseentity.resource_owner",
		"TenantID":             "baseentity.resource_owner.tenant_id",
		"UserID":               "baseentity.resource_owner.user_id",
		"GroupID":              "baseentity.resource_owner.group_id",
		"ClientID":             "baseentity.resource_owner.client_id",
		"CreatedAt":            "baseentity.create_at",
		"UpdatedAt":            "baseentity.updated_at",
	})

	return &PlanRepository{
		MongoDBRepository: *repo,
	}
}

func (repo *PlanRepository) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	return repo.MongoDBRepository.FindOneWithRLS(ctx, filter).Decode(result)
}

func (repo *PlanRepository) GetDefaultFreePlan(ctx context.Context) (*billing_entities.Plan, error) {
	var plan billing_entities.Plan
	opts := options.FindOne().SetSort(bson.D{
		{Key: "display_priority_score", Value: -1},
		{Key: "baseentity.created_at", Value: -1},
	})
	err := repo.MongoDBRepository.FindOneWithRLS(ctx, bson.M{
		"is_free":      true,
		"is_active":    true,
		"is_legacy":    false,
		"is_available": true,
		"effective_date": bson.M{
			"$lte": time.Now(),
		},
		"$or": []bson.M{
			{"expiration_date": bson.M{"$gte": time.Now()}},
			{"expiration_date": bson.M{"$eq": nil}},
		},
	}, opts).Decode(&plan)

	if err != nil {
		slog.Error("Failed to retrieve default free plan.", "err", err)
		return nil, err
	}

	slog.Info("Successfully retrieved default free plan.", "plan", plan)

	return &plan, nil
}

// GetByID retrieves a plan by its ID
func (repo *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Plan, error) {
	return repo.MongoDBRepository.GetByID(ctx, id)
}

// GetAvailablePlans retrieves all available plans
func (repo *PlanRepository) GetAvailablePlans(ctx context.Context) ([]*billing_entities.Plan, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "display_priority_score", Value: -1},
		{Key: "kind", Value: 1},
	})

	cursor, err := repo.MongoDBRepository.FindWithRLS(ctx, bson.M{
		"is_active":    true,
		"is_legacy":    false,
		"is_available": true,
		"effective_date": bson.M{
			"$lte": time.Now(),
		},
		"$or": []bson.M{
			{"expiration_date": bson.M{"$gte": time.Now()}},
			{"expiration_date": bson.M{"$eq": nil}},
		},
	}, opts)

	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve available plans", "err", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []*billing_entities.Plan
	for cursor.Next(ctx) {
		var plan billing_entities.Plan
		if err := cursor.Decode(&plan); err != nil {
			slog.ErrorContext(ctx, "Failed to decode plan", "err", err)
			continue
		}
		plans = append(plans, &plan)
	}

	return plans, nil
}
