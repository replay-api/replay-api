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
