package replay_out

import (
	"context"
	"io"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type EventsByGameReader interface {
	GetByGameIDAndMatchID(ctx context.Context, gameID string, matchID string) ([]replay_entity.GameEvent, error)
}

type GameEventReader interface {
	shared.Searchable[replay_entity.GameEvent]
}

type MatchMetadataReader interface {
	shared.Searchable[replay_entity.Match]
}

type ReplayFileMetadataReader interface {
	shared.Searchable[replay_entity.ReplayFile]
	GetByID(ctx context.Context, replayFileID uuid.UUID) (*replay_entity.ReplayFile, error)
}

type ReplayFileContentReader interface {
	GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error)
}

type TeamReader interface {
	shared.Searchable[replay_entity.Team]
}

type PlayerMetadataReader interface {
	shared.Searchable[replay_entity.PlayerMetadata]
}

type BadgeReader interface {
	shared.Searchable[replay_entity.Badge]
}

type ShareTokenReader interface {
	shared.Searchable[replay_entity.ShareToken]
	FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error)
	FindByResourceID(ctx context.Context, resourceID uuid.UUID) ([]replay_entity.ShareToken, error)
}
