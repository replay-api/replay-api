package db

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

type ReplayFileContentRepository struct {
	client *mongo.Client
	bucket *gridfs.Bucket
}

func NewReplayFileContentRepository(client *mongo.Client) *ReplayFileContentRepository {
	db := client.Database("replay")
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName("replay_file_content"))

	if err != nil {
		slog.Warn("error creating GridFS Bucket", "err", err)
	}

	return &ReplayFileContentRepository{
		client: client,
		bucket: bucket,
	}
}

func (r *ReplayFileContentRepository) GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error) {
	fileName := replayFileID.String() + ".dem"
	file, err := os.Create("/app/replay_files/" + fileName)
	ioWriteCloser := io.WriteCloser(file)
	file.Chmod(0500)

	if err != nil {
		slog.ErrorContext(ctx, "error creating file", "err", err)
		return nil, err
	}

	length, err := r.bucket.DownloadToStreamByName(fileName, ioWriteCloser)
	ioWriteCloser.Close()
	if err != nil {
		slog.ErrorContext(ctx, "error downloading file", "length", length, "err", err)
		return nil, err
	}

	file.Truncate(0)

	file, err = os.Open("/app/replay_files/" + fileName)
	if err != nil {
		slog.ErrorContext(ctx, "error opening file", "err", err)
		return nil, err
	}

	slog.InfoContext(ctx, "ReplayFileContentRepository.GetByID: successfully downloaded file", "fileName", fileName)

	// _, err = file.Seek(0, 0)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "error seeking to start of file", err)
	// }

	return file, nil
}

func (r *ReplayFileContentRepository) Put(ctx context.Context, replayFileID uuid.UUID, file io.ReadSeeker) (string, error) {
	fileName := replayFileID.String() + ".dem"
	_, err := file.Seek(0, 0)
	if err != nil {
		slog.ErrorContext(ctx, "error seeking to start of file", "err", err)
	}

	_, err = r.bucket.UploadFromStream(fileName, file)
	if err != nil {
		slog.ErrorContext(ctx, "error uploading file", "err", err)
		return "", err
	}

	slog.InfoContext(ctx, "ReplayFileContentRepository.Put: successfully uploaded file", "fileName", fileName)

	return fileName, nil

}
