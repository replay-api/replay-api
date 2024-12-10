package use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"
)

type UpdateReplayFileHeaderUseCase struct {
	EventReader    replay_out.GameEventReader
	MetadataReader replay_out.ReplayFileMetadataReader
	MetadataWriter replay_out.ReplayFileMetadataWriter
}

func NewUpdateReplayFileHeaderUseCase(eventReader replay_out.GameEventReader,
	metadataReader replay_out.ReplayFileMetadataReader,
	metadataWriter replay_out.ReplayFileMetadataWriter) *UpdateReplayFileHeaderUseCase {
	return &UpdateReplayFileHeaderUseCase{
		EventReader:    eventReader,
		MetadataReader: metadataReader,
		MetadataWriter: metadataWriter,
	}
}

func (usecase *UpdateReplayFileHeaderUseCase) Exec(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error) {
	file, err := usecase.MetadataReader.GetByID(ctx, replayFileID)
	if err != nil || file == nil {
		slog.ErrorContext(ctx, "error updating file header: not found", "replayFileID", replayFileID, "err", err, "returnedFile", file)
		return nil, err
	}

	params := []common.SearchableValue{
		{
			Field: "Type",
			Values: []interface{}{
				common.Event_MatchStartID,
			},
		},
	}

	resultOptions := common.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	s := common.NewSearchByValues(ctx, params, resultOptions, common.UserAudienceIDKey)

	events, err := usecase.EventReader.Search(ctx, s)

	if err != nil {
		slog.ErrorContext(ctx, "error searching for replay file header event (common.Event_MatchStartID)", "err", err, "Type", common.Event_MatchStartID, "events", events)
		return nil, err
	}

	if len(events) > 1 {
		slog.WarnContext(ctx, "replay file has more than one match", "events", events, "replayFile", file)
	}

	if len(events) == 0 {
		slog.ErrorContext(ctx, "replay file header Event not found", "Type", common.Event_MatchStartID, "s", s, "events", events)
		return nil, err
	}

	slog.InfoContext(ctx, fmt.Sprintf("@@@ UpdateReplayFileHeaderUseCase @@@: len(events): %v", len(events)))

	file.Header = events[0].Payload

	file, err = usecase.MetadataWriter.Update(ctx, file)

	if err != nil {
		slog.ErrorContext(ctx, "failed to update replay file header", "err", err, "replayFile", file)
		return nil, err
	}

	return file, nil
}
