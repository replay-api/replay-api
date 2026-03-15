package analytics_usecases

import (
	"context"
	"log/slog"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
)

var _ analytics_in.MyAnalyticsQueryHandler = (*GetMyAnalyticsUseCase)(nil)

type GetMyAnalyticsUseCase struct {
	statsReader analytics_out.ViewStatisticsReader
}

func NewGetMyAnalyticsUseCase(
	statsReader analytics_out.ViewStatisticsReader,
) analytics_in.MyAnalyticsQueryHandler {
	return &GetMyAnalyticsUseCase{
		statsReader: statsReader,
	}
}

func (uc *GetMyAnalyticsUseCase) GetMyAnalytics(ctx context.Context, query analytics_in.GetMyAnalyticsQuery) ([]analytics_entities.ViewStatistics, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	stats, err := uc.statsReader.GetByOwner(ctx, query.UserID, query.EntityType)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read user analytics", "err", err, "user_id", query.UserID)
		return nil, err
	}

	return stats, nil
}
