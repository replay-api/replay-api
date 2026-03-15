package analytics_usecases

import (
	"context"
	"log/slog"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

var _ analytics_in.ViewInsightsQueryHandler = (*GetViewInsightsUseCase)(nil)

type GetViewInsightsUseCase struct {
	insightReader analytics_out.ViewerInsightReader
	privacyReader analytics_out.ViewPrivacyReader
}

func NewGetViewInsightsUseCase(
	insightReader analytics_out.ViewerInsightReader,
	privacyReader analytics_out.ViewPrivacyReader,
) analytics_in.ViewInsightsQueryHandler {
	return &GetViewInsightsUseCase{
		insightReader: insightReader,
		privacyReader: privacyReader,
	}
}

func (uc *GetViewInsightsUseCase) GetViewInsights(ctx context.Context, query analytics_in.GetViewInsightsQuery) ([]analytics_entities.ViewerInsight, int64, error) {
	if err := query.Validate(); err != nil {
		return nil, 0, err
	}

	// Check if the entity owner has view insights enabled
	privacySettings, err := uc.privacyReader.GetByUserID(ctx, query.OwnerID)
	if err == nil && privacySettings != nil && !privacySettings.ShowProfileViews {
		return []analytics_entities.ViewerInsight{}, 0, nil
	}

	sortField := "last_viewed_at"
	sortDir := shared.DescendingIDKey
	if query.SortBy == "frequent" {
		sortField = "view_count"
	}

	search := shared.Search{
		SearchParams: []shared.SearchAggregation{
			{
				Params: []shared.SearchParameter{
					{
						ValueParams: []shared.SearchableValue{
							{Field: "EntityID", Values: []interface{}{query.EntityID}, Operator: shared.EqualsOperator},
							{Field: "EntityType", Values: []interface{}{query.EntityType}, Operator: shared.EqualsOperator},
							{Field: "IsAnonymous", Values: []interface{}{false}, Operator: shared.EqualsOperator},
						},
					},
				},
			},
		},
		ResultOptions: shared.SearchResultOptions{
			Skip:  query.Skip,
			Limit: query.Limit,
		},
		SortOptions: []shared.SortableField{
			{Field: sortField, Direction: sortDir},
		},
	}

	insights, total, err := uc.insightReader.GetByEntity(ctx, query.EntityID, query.EntityType, search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read view insights", "err", err, "entity_id", query.EntityID)
		return nil, 0, err
	}

	return insights, total, nil
}
