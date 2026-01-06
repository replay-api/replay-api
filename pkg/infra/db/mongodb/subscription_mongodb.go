package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/resource-ownership/go-mongodb/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type SubscriptionRepository struct {
	mongodb.MongoDBRepository[billing_entities.Subscription]
}

func NewSubscriptionRepository(client *mongo.Client, dbName string, entityType billing_entities.Subscription, collectionName string) *SubscriptionRepository {
	repo := mongodb.NewMongoDBRepository[billing_entities.Subscription](client, dbName, entityType, collectionName, "Subscription")

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
		"VoucherCode":     true,
		"Args":            true,
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
		"VoucherCode":     "voucher_code",
		"Args":            "args",
	})

	return &SubscriptionRepository{
		MongoDBRepository: *repo,
	}
}

func (r *SubscriptionRepository) Search(ctx context.Context, s shared.Search) ([]billing_entities.Subscription, error) {
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

func (r *SubscriptionRepository) GetCurrentSubscription(ctx context.Context, rxn shared.ResourceOwner) (*billing_entities.Subscription, error) {
	search := shared.NewSearchByValues(ctx, []shared.SearchableValue{
		{
			Field:    "Status",
			Values:   []interface{}{billing_entities.SubscriptionStatusActive},
			Operator: shared.EqualsOperator,
		},
		{
			Field:    "StartDate",
			Values:   []interface{}{time.Now()},
			Operator: shared.LessThanOperator,
		},
		{
			Field:    "EndDate",
			Values:   []interface{}{time.Now()},
			Operator: shared.GreaterThanOperator,
		},
		{
			Field: "UserID",
			Values: []interface{}{
				rxn.UserID,
			},
			Operator: shared.EqualsOperator,
		},
	}, shared.NewSearchResultOptions(0, 1), shared.ClientApplicationAudienceIDKey)

	subscriptions, err := r.Search(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error searching for current subscription", "err", err)
		return nil, err
	}

	if len(subscriptions) == 0 {
		return nil, nil
	}

	return &subscriptions[0], nil
}

// Update updates an existing subscription
func (r *SubscriptionRepository) Update(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	subscription.UpdatedAt = time.Now()

	updated, err := r.MongoDBRepository.Update(ctx, subscription)
	if err != nil {
		slog.ErrorContext(ctx, "error updating subscription", "id", subscription.ID, "err", err)
		return nil, err
	}

	return updated, nil
}

// Cancel cancels a subscription
func (r *SubscriptionRepository) Cancel(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error) {
	now := time.Now()
	subscription.Status = billing_entities.SubscriptionStatusCanceled
	subscription.UpdatedAt = now
	subscription.History = append(subscription.History, billing_entities.SubscriptionHistory{
		Date:   now,
		Status: billing_entities.SubscriptionStatusCanceled,
		Reason: "User requested cancellation",
	})

	updated, err := r.MongoDBRepository.Update(ctx, subscription)
	if err != nil {
		slog.ErrorContext(ctx, "error cancelling subscription", "id", subscription.ID, "err", err)
		return nil, err
	}

	return updated, nil
}
