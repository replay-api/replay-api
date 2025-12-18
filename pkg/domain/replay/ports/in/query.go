package replay_in

import (
	"context"
	"io"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type EventReader interface {
	common.Searchable[replay_entity.GameEvent]
}

type MatchReader interface {
	common.Searchable[replay_entity.Match]
}

type ReplayFileReader interface {
	common.Searchable[replay_entity.ReplayFile]
}

type ReplayContentReader interface {
	GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error)
}

type PlayerMetadataReader interface {
	common.Searchable[replay_entity.PlayerMetadata]
}

type TeamReader interface {
	common.Searchable[replay_entity.Team]
}

type RoundReader interface {
	common.Searchable[replay_entity.Round]
}

type BadgeReader interface {
	common.Searchable[replay_entity.Badge]
}

type ShareTokenReader interface {
	common.Searchable[replay_entity.ShareToken]
	FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error)
}
