package replay_out

import (
	"context"
	"io"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type ReplayParser interface {
	Parse(ctx context.Context, match uuid.UUID, content io.Reader, eventsChan chan *replay_entity.GameEvent) error
}

type GameEventWriter interface {
	CreateMany(createCtx context.Context, events []*replay_entity.GameEvent) error
	Create(createCtx context.Context, events *replay_entity.GameEvent) (*replay_entity.GameEvent, error)
}

type MatchMetadataWriter interface {
	Create(createCtx context.Context, match replay_entity.Match) error
	CreateMany(createCtx context.Context, matches []replay_entity.Match) error
}

type PlayerMetadataWriter interface {
	Create(createCtx context.Context, player replay_entity.PlayerMetadata) error
	CreateMany(createCtx context.Context, players []replay_entity.PlayerMetadata) error
}

type ReplayFileMetadataWriter interface {
	Create(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
	Update(createCtx context.Context, replayFile *replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
}

type ReplayFileContentWriter interface {
	Put(createCtx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error)
}

type ShareTokenWriter interface {
	Create(ctx context.Context, token *replay_entity.ShareToken) error
	Update(ctx context.Context, token *replay_entity.ShareToken) error
	Delete(ctx context.Context, tokenID uuid.UUID) error
}

// ReplayEventPublisher publishes replay-related events to message broker (Kafka)
type ReplayEventPublisher interface {
	// PublishReplayUploaded publishes an event when a replay file is uploaded
	PublishReplayUploaded(ctx context.Context, replayFile *replay_entity.ReplayFile) error
	// PublishReplayProcessing publishes progress events during replay processing
	PublishReplayProcessing(ctx context.Context, replayFileID uuid.UUID, stage string, progress int, eventCount int, playerCount int) error
	// PublishReplayCompleted publishes an event when replay processing completes successfully
	PublishReplayCompleted(ctx context.Context, replayFile *replay_entity.ReplayFile, matchID uuid.UUID, eventCount int, playerCount int, processingDurationMs int64) error
	// PublishReplayFailed publishes an event when replay processing fails
	PublishReplayFailed(ctx context.Context, replayFile *replay_entity.ReplayFile, stage string, errorType string, errorMessage string, retryable bool, retryCount int) error
}
