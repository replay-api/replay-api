package metadata

import (
	"context"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// TeamQueryService provides query capabilities for Team entities
// Teams are embedded within Match entities, so we query through MatchMetadataReader
type TeamQueryService struct {
	shared.BaseQueryService[replay_entity.Team]
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
		"ResourceOwner":      shared.DENY,
		"CreatedAt":          true,
		"UpdatedAt":          true,
	}

	// Create a team searchable adapter that extracts teams from matches
	teamAdapter := &TeamSearchableAdapter{
		matchReader: matchReader,
	}

	return &shared.BaseQueryService[replay_entity.Team]{
		Reader:          teamAdapter,
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        shared.UserAudienceIDKey,
	}
}

// TeamSearchableAdapter adapts MatchMetadataReader to provide Team search capabilities
type TeamSearchableAdapter struct {
	matchReader replay_out.MatchMetadataReader
}

// GetByID implements shared.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) GetByID(ctx context.Context, id uuid.UUID) (*replay_entity.Team, error) {
	// Teams are embedded in matches, so we would need to search matches to find the team
	// For now, return nil as teams don't have their own collection
	return nil, nil
}

// Search implements shared.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) Search(ctx context.Context, s shared.Search) ([]replay_entity.Team, error) {
	// For now, return empty results since teams are embedded in matches
	// A full implementation would query matches and extract unique teams
	return []replay_entity.Team{}, nil
}

// Compile implements shared.Searchable[replay_entity.Team]
func (a *TeamSearchableAdapter) Compile(ctx context.Context, searchParams []shared.SearchAggregation, resultOptions shared.SearchResultOptions) (*shared.Search, error) {
	// Return a basic search object
	return &shared.Search{
		SearchParams:  searchParams,
		ResultOptions: resultOptions,
		VisibilityOptions: shared.SearchVisibilityOptions{
			RequestSource:    shared.GetResourceOwner(ctx),
			IntendedAudience: shared.UserAudienceIDKey,
		},
	}, nil
}
