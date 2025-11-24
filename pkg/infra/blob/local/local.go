package blob

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type LocalFileAdapter struct{}

func NewLocalFileAdapter() *LocalFileAdapter {
	return &LocalFileAdapter{}
}

func (adp *LocalFileAdapter) Put(ctx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error) {
	_, err := reader.Seek(0, 0)
	if err != nil {
		slog.ErrorContext(ctx, "error seeking to start of file", "error", err)
	}

	path := "/app/replay_files/" + replayFileID.String() + ".dem"
	fileBytes := []byte{}
	_, err = reader.Read(fileBytes)
	if err != nil {
		slog.ErrorContext(ctx, "error reading replay file", "error", err)
	}

	file, err := os.Create(path)
	if err != nil {
		slog.ErrorContext(ctx, "error writing replay file", "error", err)
	}

	_, err = file.Write(fileBytes)
	if err != nil {
		slog.ErrorContext(ctx, "error writing replay file", "error", err)
	}

	slog.InfoContext(ctx, "Local.Put: successfully wrote replay file", "path", path)

	return path, nil
}

func (adapter *LocalFileAdapter) GetByID(ctx context.Context, replayFileID uuid.UUID) (*os.File, error) {
	path := "/app/replay_files/" + replayFileID.String() + ".dem"
	file, err := os.Open(path)
	if err != nil {
		slog.ErrorContext(ctx, "Local.GetByID: error reading replay file", "error", err)
		return nil, err
	}

	return file, nil
}
