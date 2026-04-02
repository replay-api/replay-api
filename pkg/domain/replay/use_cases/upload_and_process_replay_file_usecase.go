package use_cases

import (
	"context"
	"io"
	"log/slog"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type UploadAndProcessReplayFileUseCase struct {
	UploadCommand       replay_in.UploadReplayFileCommand
	ProcessCommand      replay_in.ProcessReplayFileCommand
	UpdateHeaderCommand replay_in.UpdateReplayFileHeaderCommand
	MetadataWriter      replay_out.ReplayFileMetadataWriter
}

func (usecase *UploadAndProcessReplayFileUseCase) Exec(ctx context.Context, file io.Reader) (*replay_entity.ReplayFile, error) {
	return usecase.ExecWithOptions(ctx, file, nil)
}

// ExecWithOptions uploads and processes a replay file with optional metadata
func (usecase *UploadAndProcessReplayFileUseCase) ExecWithOptions(ctx context.Context, file io.Reader, opts *replay_entity.ReplayFileOptions) (*replay_entity.ReplayFile, error) {
	var replayFile *replay_entity.ReplayFile
	var err error

	// Check if upload command supports options
	if opts != nil {
		if cmdWithOpts, ok := usecase.UploadCommand.(replay_in.UploadReplayFileWithOptionsCommand); ok {
			replayFile, err = cmdWithOpts.ExecWithOptions(ctx, file, opts)
		} else {
			// Fall back to regular upload
			replayFile, err = usecase.UploadCommand.Exec(ctx, file)
		}
	} else {
		replayFile, err = usecase.UploadCommand.Exec(ctx, file)
	}

	if err != nil {
		slog.ErrorContext(ctx, "error uploading replay file", "err", err)
		return nil, err
	}

	// Start processing asynchronously
	go func() {
		// Create a new background context with resource owner info from the request
		bgCtx := context.Background()
		bgCtx = context.WithValue(bgCtx, shared.TenantIDKey, ctx.Value(shared.TenantIDKey))
		bgCtx = context.WithValue(bgCtx, shared.ClientIDKey, ctx.Value(shared.ClientIDKey))
		bgCtx = context.WithValue(bgCtx, shared.GroupIDKey, ctx.Value(shared.GroupIDKey))
		bgCtx = context.WithValue(bgCtx, shared.UserIDKey, ctx.Value(shared.UserIDKey))
		bgCtx = context.WithValue(bgCtx, shared.AuthenticatedKey, ctx.Value(shared.AuthenticatedKey))

		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in replay processing goroutine", "replayFileID", replayFile.ID, "panic", r)
				usecase.markFailed(bgCtx, replayFile, "panic during processing")
			}
		}()

		match, err := usecase.ProcessCommand.Exec(bgCtx, replayFile.ID)
		if err != nil {
			slog.ErrorContext(bgCtx, "error processing replay file", "replayFileID", replayFile.ID, "err", err)
			usecase.markFailed(bgCtx, replayFile, err.Error())
			return
		}

		_, err = usecase.UpdateHeaderCommand.Exec(bgCtx, replayFile.ID)
		if err != nil {
			slog.ErrorContext(bgCtx, "UploadAndProcessReplayFileUseCase failed to update replay file HEADER", "replayFileID", replayFile.ID, "err", err)
			// Don't mark failed — processing succeeded, header update is non-critical
			return
		}

		slog.InfoContext(bgCtx, "completed processing replay file", "matchID", match.ID, "replayFileID", replayFile.ID)
	}()

	return replayFile, nil
}

// markFailed sets the replay file status to Failed with error details.
func (usecase *UploadAndProcessReplayFileUseCase) markFailed(ctx context.Context, replayFile *replay_entity.ReplayFile, errMsg string) {
	if usecase.MetadataWriter == nil {
		slog.ErrorContext(ctx, "cannot mark replay as failed — MetadataWriter is nil", "replayFileID", replayFile.ID)
		return
	}
	replayFile.Status = replay_entity.ReplayFileStatusFailed
	replayFile.Error = errMsg
	if _, err := usecase.MetadataWriter.Update(ctx, replayFile); err != nil {
		slog.ErrorContext(ctx, "failed to update replay status to Failed", "replayFileID", replayFile.ID, "err", err)
	}
}

func NewUploadAndProcessReplayFileUseCase(uploadCommand replay_in.UploadReplayFileCommand, processCommand replay_in.ProcessReplayFileCommand, updateHeaderCommand replay_in.UpdateReplayFileHeaderCommand, metadataWriter replay_out.ReplayFileMetadataWriter) *UploadAndProcessReplayFileUseCase {
	return &UploadAndProcessReplayFileUseCase{
		UploadCommand:       uploadCommand,
		ProcessCommand:      processCommand,
		UpdateHeaderCommand: updateHeaderCommand,
		MetadataWriter:      metadataWriter,
	}
}
