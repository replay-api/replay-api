package metadata

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// TeamQueryService provides query capabilities for Team entities
// Teams are embedded within Match entities, so we query through MatchMetadataReader
type TeamQueryService struct {
	common.BaseQueryService[replay_entity.Team]
}

// NewTeamQueryService creates a new TeamQueryService
// It wraps the match reader to extract team data
func NewTeamQueryService(matchReader replay_out.MatchMetadataReader) replay_in.TeamReader {
	queryableFields := map[string]bool{
		"ID":                 true,
		"NetworkID":          true,
		"NetworkTeamID":      true,
		"TeamHashID":         true,
		"Name":               true,
		"ShortName":          true,
		"CurrentDisplayName": true,
		"ResourceOwner":      true,
		"CreatedAt":          true,
		"UpdatedAt":          true,
	}

	readableFields := map[string]bool{
		"ID":                 true,
		"NetworkID":          true,
		"NetworkTeamID":      true,
		"TeamHashID":         true,
		"Name":               true,
		"ShortName":          true,
		"CurrentDisplayName": true,
		"NameHistory":        true,
		"Players":            true,
		"ResourceOwner":      common.DENY,
		"CreatedAt":          true,
		"UpdatedAt":          true,
	}

	// Create a team searchable adapter that extracts teams from matches
	teamAdapter := &TeamSearchableAdapter{
		matchReader: matchReader,
	}

	return &common.BaseQueryService[replay_entity.Team]{
		Reader:          teamAdapter,
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}

// TeamSearchableAdapter adapts MatchMetadataReader to provide Team search capabilities
type TeamSearchableAdapter struct {
	matchReader replay_out.MatchMetadataReader
}

// GetByID implements common.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) GetByID(ctx context.Context, id uuid.UUID) (*replay_entity.Team, error) {
	// Teams are embedded in matches, so we would need to search matches to find the team
	// For now, return nil as teams don't have their own collection
	return nil, nil
}

// Search implements common.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) Search(ctx context.Context, s common.Search) ([]replay_entity.Team, error) {
	// For now, return empty results since teams are embedded in matches
	// A full implementation would query matches and extract unique teams
	return []replay_entity.Team{}, nil
}

// Compile implements common.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	// Return a basic search object
	return &common.Search{
		SearchParams:  searchParams,
		ResultOptions: resultOptions,
		VisibilityOptions: common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(ctx),
			IntendedAudience: common.UserAudienceIDKey,
		},
	}, nil
}
