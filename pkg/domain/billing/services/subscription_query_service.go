package billing_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

type SubscriptionQueryService struct {
	common.BaseQueryService[billing_entities.Subscription]
}

func NewSubscriptionQueryService(eventReader billing_out.SubscriptionReader) billing_in.SubscriptionReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"PlanID":        true,
		"BillingPeriod": true,
		"StartAt":       true,
		"EndAt":         true,
		"Status":        true,
		"IsFree":        true,
		"VoucherCode":   true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"PlanID":        true,
		"BillingPeriod": true,
		"StartAt":       true,
		"EndAt":         true,
		"Status":        true,
		"History":       true,
		"IsFree":        true,
		"VoucherCode":   true,
		"Args":          true,
	}

	return &common.BaseQueryService[billing_entities.Subscription]{
		Reader:          eventReader.(common.Searchable[billing_entities.Subscription]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}

func (s *SubscriptionQueryService) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*billing_entities.Subscription, error) {
	isAuthenticated := ctx.Value(common.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return nil, common.NewErrUnauthorized()
	}

	if userID != common.GetResourceOwner(ctx).UserID && !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden()
	}

	search := common.Search{
		VisibilityOptions: common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(ctx),
			IntendedAudience: ctx.Value(common.UserAudienceIDKey).(common.IntendedAudienceKey),
		},
		IncludeParams: []common.IncludeParam{
			{
				From:         "BillableEntry",
				LocalField:   "SubscriptionID",
				ForeignField: "ID",
				IsArray:      true,
			},
			{
				From:         "Plan",
				LocalField:   "PlanID",
				ForeignField: "ID",
				IsArray:      false,
			},
		},
		SearchParams: []common.SearchAggregation{
			{
				AggregationClause: common.AndAggregationClause,
				Params: []common.SearchParameter{
					{
						ValueParams: []common.SearchableValue{
							{
								Field:    "EndAt",
								Operator: common.GreaterThanOperator,
								Values:   []interface{}{time.Now()},
							},
							{
								Field:    "Status",
								Operator: common.EqualsOperator,
								Values:   []interface{}{billing_entities.SubscriptionStatusActive},
							},
						},
					},
				},
			},
		},
		SortOptions: []common.SortableField{
			{
				Field:     "StartAt",
				Direction: common.DescendingIDKey,
			},
		},
		ResultOptions: common.SearchResultOptions{
			Skip:  0,
			Limit: 1,
		},
	}

	subs, err := s.Reader.Search(ctx, search)

	if err != nil {
		slog.ErrorContext(ctx, "error querying subscription and usage", "err", err)
		return nil, err
	}

	if len(subs) == 0 {
		return nil, fmt.Errorf("no active subscription found for user %v", userID)
	}

	sub := subs[0]

	return &sub, nil
}

func (s *SubscriptionQueryService) CheckOperationAvailable(ctx context.Context, userID uuid.UUID, operationID billing_entities.BillableOperationKey) error {
	sub, err := s.GetSubscriptionByUserID(ctx, userID)

	if err != nil {
		return err
	}

	ops := sub.GetFeatures()

	for _, op := range ops {
		if op == operationID {
			if !sub.Available(operationID) {
				return fmt.Errorf("operation %s not available for user %v", operationID, userID)
			}
			return nil
		}
	}

	return fmt.Errorf("operation %s not found for user %v", operationID, userID)
}

func (s *SubscriptionQueryService) GetAvailabilityAndUsage(ctx context.Context, userID uuid.UUID) (map[billing_entities.BillableOperationKey]float64, map[billing_entities.BillableOperationKey]float64, error) {
	sub, err := s.GetSubscriptionByUserID(ctx, userID)

	ops := sub.GetFeatures()

	if err != nil {
		return nil, nil, err
	}

	usage, limits := make(map[billing_entities.BillableOperationKey]float64), make(map[billing_entities.BillableOperationKey]float64)

	for _, op := range ops {
		usage[op], limits[op] = sub.GetUsageAndLimits(op)
	}

	return usage, limits, nil
}
