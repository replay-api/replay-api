package use_cases

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	e "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"
)

const CHUNK_SIZE = 10

type ProcessReplayFileUseCase struct {
	ReplayMetadataReader replay_out.ReplayFileMetadataReader
	ReplayContentReader  replay_out.ReplayFileContentReader
	ReplayMetadataWriter replay_out.ReplayFileMetadataWriter
	ReplayContentWriter  replay_out.ReplayFileContentWriter

	PlayerMetadataWriter replay_out.PlayerMetadataWriter
	MatchMetadataWriter  replay_out.MatchMetadataWriter

	Parser      replay_out.ReplayParser
	EventWriter replay_out.GameEventWriter
}

func NewProcessReplayFileUseCase(metadataReader replay_out.ReplayFileMetadataReader, contentReader replay_out.ReplayFileContentReader, metadataWriter replay_out.ReplayFileMetadataWriter, contentWriter replay_out.ReplayFileContentWriter, parser replay_out.ReplayParser, eventWriter replay_out.GameEventWriter, playerMetadataWriter replay_out.PlayerMetadataWriter, matchMetadataWriter replay_out.MatchMetadataWriter) *ProcessReplayFileUseCase {
	return &ProcessReplayFileUseCase{
		ReplayMetadataReader: metadataReader,
		ReplayContentReader:  contentReader,
		ReplayMetadataWriter: metadataWriter,
		ReplayContentWriter:  contentWriter,

		PlayerMetadataWriter: playerMetadataWriter,
		MatchMetadataWriter:  matchMetadataWriter,

		Parser:      parser,
		EventWriter: eventWriter,
	}
}

func (usecase *ProcessReplayFileUseCase) Exec(ctx context.Context, replayFileID uuid.UUID) (*e.Match, error) {
	replayFile, err := usecase.ReplayMetadataReader.GetByID(ctx, replayFileID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting replay metadata", "replayFileID", replayFileID, "err", err)
		return nil, err
	}

	// Update Metadata Status
	replayFile.Status = e.ReplayFileStatusProcessing
	replayFile, err = usecase.ReplayMetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata", "replayFile", replayFile, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "processing replay file", "replayFile", replayFile)

	match := &e.Match{
		ID:            uuid.New(),
		GameID:        replayFile.GameID,
		ReplayFileID:  replayFile.ID,
		ResourceOwner: replayFile.ResourceOwner,
		Events:        make([]*e.GameEvent, 0),
	}

	file, err := usecase.ReplayContentReader.GetByID(ctx, replayFileID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting replay file content data", "err", err)
		return nil, err
	}
	defer file.Close()

	slog.InfoContext(ctx, "parsing replay file", "Size", replayFile.Size, "replayFileID", replayFileID)

	eventsChan := make(chan *e.GameEvent, 1000)

	entitiesMap := make(map[common.ResourceType][]interface{})

	gameEvents := make([]*e.GameEvent, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range eventsChan {
			slog.InfoContext(ctx, "event", "event.Type", event.Type)
			if event.Type != common.Event_GenericGameEventID {
				match.Events = append(match.Events, event)
			}

			gameEvents = append(gameEvents, event)

			for k, v := range event.Entities {
				if entitiesMap[k] == nil {
					entitiesMap[k] = make([]interface{}, 0)
				}

				entitiesMap[k] = append(entitiesMap[k], v...)
			}
		}
	}()

	err = usecase.Parser.Parse(ctx, match.ID, file, eventsChan)

	if err != nil {
		slog.ErrorContext(ctx, "error parsing replay events", "err", err)
		return nil, err
	}

	wg.Wait()

	for resourceKey, entities := range entitiesMap {
		switch resourceKey {
		case common.ResourceTypePlayerMetadata:
			for _, rawEntity := range entities {
				slog.InfoContext(ctx, "PlayerMetadata", "entity", rawEntity)
				entity := rawEntity.(*e.PlayerMetadata)

				// TODO: getByNetworkID, refactor Create=>Upsert, Set RXN to app, Set UpdatedAt, Set USER_ID, Use BaseEntity, Se visibility to Public(?), Set NameHistory
				// TODO: distinct/mergeProps before upsert? (dups)
				err = usecase.PlayerMetadataWriter.Create(ctx, *entity)

				if err != nil {
					slog.ErrorContext(ctx, "error writing PlayerMetadata entities", "err", err)
					return nil, err
				}
			}

		case common.ResourceTypeMatch:
			for _, rawEntity := range entities {
				slog.InfoContext(ctx, "MatchMetadata", "entity", rawEntity)
				entity := rawEntity.(*e.Match)

				err = usecase.MatchMetadataWriter.Create(ctx, *entity)

				if err != nil {
					slog.ErrorContext(ctx, "error writing MatchMetadata entities", "err", err)
					return nil, err
				}
			}
		}
	}

	err = usecase.EventWriter.CreateMany(ctx, gameEvents)

	if err != nil {
		slog.ErrorContext(ctx, "error writing GameEvents", "err", err, "len(gameEvents)", len(gameEvents))
		return nil, err
	}

	// Update Metadata Status
	replayFile.Status = e.ReplayFileStatusCompleted
	replayFile, err = usecase.ReplayMetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata status to Completed", "replayFile", replayFile, "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Replay file processed", "ReplayFileID", replayFileID)

	return match, nil
}
