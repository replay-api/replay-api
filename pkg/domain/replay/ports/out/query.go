package replay_out

import (
	"context"
	"io"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
)

type EventsByGameReader interface {
	GetByGameIDAndMatchID(ctx context.Context, gameID string, matchID string) ([]replay_entity.GameEvent, error)
}

type GameEventReader interface {
	common.Searchable[replay_entity.GameEvent]
}

type MatchMetadataReader interface {
	common.Searchable[replay_entity.Match]
}

type ReplayFileMetadataReader interface {
	common.Searchable[replay_entity.ReplayFile]
	GetByID(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error)
}

type ReplayFileContentReader interface {
	GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error)
}

type TeamReader interface {
	common.Searchable[replay_entity.Team]
}

type PlayerMetadataReader interface {
	common.Searchable[replay_entity.Player]
}

type BadgeReader interface {
	common.Searchable[replay_entity.Badge]
}
