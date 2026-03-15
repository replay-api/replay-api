package analytics_usecases

import (
	"context"
	"log/slog"

	analytics_entities "github.com/replay-api/replay-api/pkg/domain/analytics/entities"
	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
)

var _ analytics_in.ViewStatisticsQueryHandler = (*GetViewStatisticsUseCase)(nil)

type GetViewStatisticsUseCase struct {
	statsReader analytics_out.ViewStatisticsReader
}

func NewGetViewStatisticsUseCase(
	statsReader analytics_out.ViewStatisticsReader,
) analytics_in.ViewStatisticsQueryHandler {
	return &GetViewStatisticsUseCase{
		statsReader: statsReader,
	}
}

func (uc *GetViewStatisticsUseCase) GetViewStatistics(ctx context.Context, query analytics_in.GetViewStatisticsQuery) (*analytics_entities.ViewStatistics, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	stats, err := uc.statsReader.GetByEntity(ctx, query.EntityID, query.EntityType)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read view statistics", "err", err, "entity_id", query.EntityID)
		return nil, err
	}

	if stats == nil {
		return analytics_entities.NewViewStatistics(query.EntityID, query.EntityType), nil
	}

	return stats, nil
}
