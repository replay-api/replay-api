package replay_in

import (
	"context"
	"io"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
)

type EventReader interface {
	shared.Searchable[replay_entity.GameEvent]
}

type MatchReader interface {
	shared.Searchable[replay_entity.Match]
}

type ReplayFileReader interface {
	shared.Searchable[replay_entity.ReplayFile]
}

type ReplayContentReader interface {
	GetByID(ctx context.Context, replayFileID uuid.UUID) (io.ReadSeekCloser, error)
}

type PlayerMetadataReader interface {
	shared.Searchable[replay_entity.PlayerMetadata]
}

type TeamReader interface {
	shared.Searchable[replay_entity.Team]
}

type RoundReader interface {
	shared.Searchable[replay_entity.Round]
}

type BadgeReader interface {
	shared.Searchable[replay_entity.Badge]
}

type ShareTokenReader interface {
	shared.Searchable[replay_entity.ShareToken]
	FindByToken(ctx context.Context, tokenID uuid.UUID) (*replay_entity.ShareToken, error)
}
