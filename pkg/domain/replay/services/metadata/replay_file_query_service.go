package metadata

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

type ReplayFileQueryService struct {
	common.BaseQueryService[replay_entity.ReplayFile]
}

func NewReplayFileQueryService(fileMetadataReader replay_out.ReplayFileMetadataReader) replay_in.ReplayFileReader {
	queryableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Size":          true,
		"InternalURI":   common.DENY,
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
		"Size":          true,
		"InternalURI":   common.DENY,
		"Status":        true,
		"Error":         common.DENY,
		"Header.*":      true,
		"ResourceOwner": true,
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	return &common.BaseQueryService[replay_entity.ReplayFile]{
		Reader:          fileMetadataReader.(common.Searchable[replay_entity.ReplayFile]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
