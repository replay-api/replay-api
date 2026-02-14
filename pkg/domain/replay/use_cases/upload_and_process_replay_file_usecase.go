package use_cases

import (
	"context"
	"io"
	"log/slog"

	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

type UploadAndProcessReplayFileUseCase struct {
	UploadCommand       replay_in.UploadReplayFileCommand
	ProcessCommand      replay_in.ProcessReplayFileCommand
	UpdateHeaderCommand replay_in.UpdateReplayFileHeaderCommand
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
	// go func() {
	// 	ctx := context.Background() // New context for background processing
	// 	match, err := usecase.ProcessCommand.Exec(ctx, replayFile.ID)
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "error processing replay file", "err", err)
	// 		return
	// 	}

	// 	_, err = usecase.UpdateHeaderCommand.Exec(ctx, replayFile.ID)
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "UploadAndProcessReplayFileUseCase failed to update replay file HEADER", "err", err)
	// 		return
	// 	}

	// 	slog.InfoContext(ctx, "completed processing replay file", "match", match.ID)
	// }()

	return replayFile, nil
}

func NewUploadAndProcessReplayFileUseCase(uploadCommand replay_in.UploadReplayFileCommand, processCommand replay_in.ProcessReplayFileCommand, updateHeaderCommand replay_in.UpdateReplayFileHeaderCommand) *UploadAndProcessReplayFileUseCase {
	return &UploadAndProcessReplayFileUseCase{
		UploadCommand:       uploadCommand,
		ProcessCommand:      processCommand,
		UpdateHeaderCommand: updateHeaderCommand,
	}
}
