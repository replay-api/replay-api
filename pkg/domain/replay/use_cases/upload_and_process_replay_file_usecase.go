package use_cases

import (
	"context"
	"io"
	"log/slog"

	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
)

type UploadAndProcessReplayFileUseCase struct {
	UploadCommand       replay_in.UploadReplayFileCommand
	ProcessCommand      replay_in.ProcessReplayFileCommand
	UpdateHeaderCommand replay_in.UpdateReplayFileHeaderCommand
}

func (usecase *UploadAndProcessReplayFileUseCase) Exec(ctx context.Context, file io.Reader) (*replay_entity.Match, error) {
	replayFile, err := usecase.UploadCommand.Exec(ctx, file)
	if err != nil {
		slog.ErrorContext(ctx, "error uploading replay file", "err", err)
		return nil, err
	}

	match, err := usecase.ProcessCommand.Exec(ctx, replayFile.ID)

	if err != nil {
		slog.ErrorContext(ctx, "error processing replay file", "err", err)
		return nil, err
	}

	_, err = usecase.UpdateHeaderCommand.Exec(ctx, replayFile.ID)
	if err != nil {
		slog.ErrorContext(ctx, "UploadAndProcessReplayFileUseCase failed to update replay file HEADER", "err", err)
		return nil, err
	}

	return match, nil
}

func NewUploadAndProcessReplayFileUseCase(uploadCommand replay_in.UploadReplayFileCommand, processCommand replay_in.ProcessReplayFileCommand, updateHeaderCommand replay_in.UpdateReplayFileHeaderCommand) *UploadAndProcessReplayFileUseCase {
	return &UploadAndProcessReplayFileUseCase{
		UploadCommand:       uploadCommand,
		ProcessCommand:      processCommand,
		UpdateHeaderCommand: updateHeaderCommand,
	}
}
