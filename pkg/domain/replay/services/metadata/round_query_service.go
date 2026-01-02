package metadata

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// RoundQueryService provides query capabilities for Round entities
// Rounds are embedded within Match entities, so we query through MatchMetadataReader
type RoundQueryService struct {
	common.BaseQueryService[replay_entity.Round]
}

// NewRoundQueryService creates a new RoundQueryService
// It wraps the match reader to extract round data
func NewRoundQueryService(matchReader replay_out.MatchMetadataReader) replay_in.RoundReader {
	queryableFields := map[string]bool{
		"ID":          true,
		"GameID":      true,
		"MatchID":     true,
		"Title":       true,
		"Description": true,
		"CreatedAt":   true,
		"UpdatedAt":   true,
	}

	readableFields := map[string]bool{
		"ID":          true,
		"GameID":      true,
		"MatchID":     true,
		"Title":       true,
		"Description": true,
		"ImageURL":    true,
		"Events":      common.DENY,
		"CreatedAt":   true,
		"UpdatedAt":   true,
	}

	// Create a round searchable adapter that extracts rounds from matches
	roundAdapter := &RoundSearchableAdapter{
		matchReader: matchReader,
	}

	return &common.BaseQueryService[replay_entity.Round]{
		Reader:          roundAdapter,
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}

// RoundSearchableAdapter adapts MatchMetadataReader to provide Round search capabilities
type RoundSearchableAdapter struct {
	matchReader replay_out.MatchMetadataReader
}

// GetByID implements common.Searchable[replay_entity.Round]
func (a *RoundSearchableAdapter) GetByID(ctx context.Context, id uuid.UUID) (*replay_entity.Round, error) {
	// Rounds are embedded in matches, so we would need to search matches to find the round
	// For now, return nil as rounds don't have their own collection
	return nil, nil
}

// Search implements common.Searchable[replay_entity.Round]
func (a *RoundSearchableAdapter) Search(ctx context.Context, s common.Search) ([]replay_entity.Round, error) {
	// For now, return empty results since rounds are embedded in matches
	// A full implementation would query matches and extract rounds
	return []replay_entity.Round{}, nil
}

// Compile implements common.Searchable[replay_entity.Round]
func (a *RoundSearchableAdapter) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
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
