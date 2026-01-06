package use_cases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	fps_events "github.com/replay-api/replay-common/pkg/replay/events/game/fps"
	shared "github.com/resource-ownership/go-common/pkg/common"
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

	params := []shared.SearchableValue{
		{
			Field: "Type",
			Values: []interface{}{
				fps_events.Event_MatchStartID,
			},
		},
	}

	resultOptions := shared.SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	s := shared.NewSearchByValues(ctx, params, resultOptions, shared.UserAudienceIDKey)

	events, err := usecase.EventReader.Search(ctx, s)

	if err != nil {
		slog.ErrorContext(ctx, "error searching for replay file header event (fps_events.Event_MatchStartID)", "err", err, "Type", fps_events.Event_MatchStartID, "events", events)
		return nil, err
	}

	if len(events) > 1 {
		slog.WarnContext(ctx, "replay file has more than one match", "events", events, "replayFile", file)
	}

	if len(events) == 0 {
		slog.ErrorContext(ctx, "replay file header Event not found", "Type", fps_events.Event_MatchStartID, "s", s, "events", events)
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
