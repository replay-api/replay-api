package use_cases

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type UploadReplayFileUseCase struct {
	MetadataWriter replay_out.ReplayFileMetadataWriter
	ContentWriter  replay_out.ReplayFileContentWriter
}

func NewUploadReplayFileUseCase(metadataWriter replay_out.ReplayFileMetadataWriter, dataCommand replay_out.ReplayFileContentWriter) *UploadReplayFileUseCase {
	return &UploadReplayFileUseCase{
		MetadataWriter: metadataWriter,
		ContentWriter:  dataCommand,
	}
}

func (usecase *UploadReplayFileUseCase) Exec(ctx context.Context, reader io.Reader) (*replay_entity.ReplayFile, error) {
	file, err := io.ReadAll(reader)
	if err != nil {
		slog.ErrorContext(ctx, "error reading replay file", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "uploading replay file", "size", len(file))

	// create Metadata
	entity := replay_entity.NewReplayFile("cs", "steam", len(file), "", common.GetResourceOwner(ctx))
	replayFile, err := usecase.MetadataWriter.Create(ctx, entity)

	if err != nil {
		slog.ErrorContext(ctx, "error creating new replay metadata", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "created new replay metadata", "replayFile", replayFile)

	// Put Contents into Blob Store
	uri, err := usecase.ContentWriter.Put(ctx, replayFile.ID, bytes.NewReader(file))
	if err != nil {
		replayFile.Status = replay_entity.ReplayFileStatusFailed
		replayFile.Error = err.Error()
		_, _ = usecase.MetadataWriter.Update(ctx, replayFile)
		slog.ErrorContext(ctx, "error uploading replay data", "err", err, "replayFile", replayFile)
		return nil, err
	}

	slog.InfoContext(ctx, "uploaded replay data", "replayFile", replayFile, "uri", uri)

	// Update Metadata
	replayFile.InternalURI = uri
	replayFile.Status = replay_entity.ReplayFileStatusProcessing
	replayFile, err = usecase.MetadataWriter.Update(ctx, replayFile)

	if err != nil {
		slog.ErrorContext(ctx, "error updating uploaded replay metadata", "replayFile", replayFile, "err", err)
		return nil, err
	}

	// return updated metadata
	return replayFile, nil
}
