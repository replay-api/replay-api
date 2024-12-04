package metadata

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"
)

type MatchQueryService struct {
	common.BaseQueryService[replay_entity.Match]
}

func NewMatchQueryService(matchReader replay_out.MatchMetadataReader) replay_in.MatchReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Status":        true,
		"Error":         common.DENY,
		"Header.*":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	readableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Status":        true,
		"Error":         common.DENY,
		"Header.*":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &common.BaseQueryService[replay_entity.Match]{
		Reader:          matchReader.(common.Searchable[replay_entity.Match]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
