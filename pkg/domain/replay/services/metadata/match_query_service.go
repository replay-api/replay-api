package metadata

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"
)

// MatchQueryService handles read operations for Match entities.
//
// HEXAGONAL ARCHITECTURE:
// - Embeds BaseQueryService[Match] for common query functionality
// - Implements replay_in.MatchReader port
// - Registers itself with QueryServiceRegistry for schema discovery
//
// ENDPOINT: GET /games/{game_id}/matches
//
// QUERYABLE FIELDS:
//   - MapName: Map name for text search (e.g., "dust2", "inferno")
//   - GameID, Status, NetworkID: Exact match filters
//
// SECURITY:
//   - Error is DENY (internal error details - security risk)
//   - ResourceOwner is DENY in ReadableFields (not exposed in responses)
type MatchQueryService struct {
	shared.BaseQueryService[replay_entity.Match]
}

// NewMatchQueryService creates and registers a new MatchQueryService.
//
// CRITICAL: This function registers the service with the global QueryServiceRegistry.
//
// Parameters:
//   - matchReader: Repository adapter implementing MatchMetadataReader
//
// Returns:
//   - replay_in.MatchReader: The service implementing the port interface
func NewMatchQueryService(matchReader replay_out.MatchMetadataReader) replay_in.MatchReader {
	// QueryableFields: Fields available for search/filter queries
	queryableFields := map[string]bool{
		"ID":            true,               // Primary key
		"GameID":        true,               // Filter by game (CS2, Valorant)
		"NetworkID":     true,               // Network/region identifier
		"Status":        true,               // Match status
		"Error":         shared.DENY,        // SECURITY: Internal error details
		"MapName":       true,               // Map name for search
		"ResourceOwner": true,               // Allow querying by owner
		"CreatedAt":     true,               // Match creation timestamp
		"UpdatedAt":     true,               // Last update timestamp
	}

	// ReadableFields: Fields included in API responses
	readableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Status":        true,
		"Error":         shared.DENY,        // SECURITY: Never expose error details
		"MapName":       true,
		"ResourceOwner": shared.DENY,        // SECURITY: Hide resource ownership
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	service := &shared.BaseQueryService[replay_entity.Match]{
		Reader:          matchReader.(shared.Searchable[replay_entity.Match]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,

		// DefaultSearchFields: MapName is primary search target
		DefaultSearchFields: []string{"MapName"},

		// SortableFields: Common sorting options
		SortableFields: []string{"CreatedAt", "UpdatedAt"},

		// FilterableFields: For exact-match filtering
		FilterableFields: []string{"GameID", "Status", "NetworkID"},

		MaxPageSize: 100,
		Audience:    shared.UserAudienceIDKey,

		// EntityType: URL path segment
		EntityType: "matches",
	}

	// CRITICAL: Register with global registry for schema discovery
	shared.GetQueryServiceRegistry().Register("matches", service)

	return service
}
