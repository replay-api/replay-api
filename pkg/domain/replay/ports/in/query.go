package replay_in

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
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
}
