package replay_out

import (
	"context"
	"io"

	"github.com/google/uuid"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type ReplayParser interface {
	Parse(ctx context.Context, match uuid.UUID, content io.Reader, eventsChan chan replay_entity.GameEvent) error
}

type GameEventWriter interface {
	CreateMany(createCtx context.Context, events []replay_entity.GameEvent) error
}

type MatchMetadataWriter interface {
	// CreateMany(createCtx context.Context, matches []replay_entity.Match) error
	CreateMany(createCtx context.Context, matches []interface{}) error
}

type PlayerMetadataWriter interface {
	// CreateMany(createCtx context.Context, players []replay_entity.Player) error
	CreateMany(createCtx context.Context, players []interface{}) error
}

type ReplayFileMetadataWriter interface {
	Create(createCtx context.Context, replayFile replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
	Update(createCtx context.Context, replayFile replay_entity.ReplayFile) (*replay_entity.ReplayFile, error)
}

type ReplayFileContentWriter interface {
	Put(createCtx context.Context, replayFileID uuid.UUID, reader io.ReadSeeker) (string, error)
}
