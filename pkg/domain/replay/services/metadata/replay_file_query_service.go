// Package metadata provides domain services for replay file and match metadata.
//
// HEXAGONAL ARCHITECTURE:
// This package is part of the DOMAIN LAYER. Services here define what fields
// are externally queryable/readable for replay-related entities.
package metadata

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
)

// ReplayFileQueryService handles read operations for ReplayFile entities.
//
// HEXAGONAL ARCHITECTURE:
// - Embeds BaseQueryService[ReplayFile] for common query functionality
// - Implements replay_in.ReplayFileReader port
// - Registers itself with QueryServiceRegistry for schema discovery
//
// ENDPOINT: GET /games/{game_id}/replays
//
// QUERYABLE FIELDS:
//   - Header: Contains map name and other game-specific info (text search)
//   - GameID, Status, NetworkID: Exact match filters
//
// SECURITY:
//   - InternalURI is DENY (internal storage path - security risk)
//   - Error is DENY (internal error details - security risk)
//   - ResourceOwner is DENY in ReadableFields (not exposed in responses)
type ReplayFileQueryService struct {
	shared.BaseQueryService[replay_entity.ReplayFile]
}

// NewReplayFileQueryService creates and registers a new ReplayFileQueryService.
//
// CRITICAL: This function registers the service with the global QueryServiceRegistry.
//
// Parameters:
//   - fileMetadataReader: Repository adapter implementing Searchable[ReplayFile]
//
// Returns:
//   - replay_in.ReplayFileReader: The service implementing the port interface
func NewReplayFileQueryService(fileMetadataReader shared.Searchable[replay_entity.ReplayFile]) replay_in.ReplayFileReader {
	// QueryableFields: Fields available for search/filter queries
	// SECURITY: InternalURI and Error are DENY to prevent information leakage
	queryableFields := map[string]bool{
		"ID":            true,               // Primary key
		"GameID":        true,               // Filter by game (CS2, Valorant)
		"NetworkID":     true,               // Network/region identifier
		"Size":          true,               // File size in bytes
		"InternalURI":   shared.DENY,        // SECURITY: Internal storage path
		"Status":        true,               // Processing status
		"Error":         shared.DENY,        // SECURITY: Internal error details
		"Header":        true,               // Game-specific header (map name, etc.)
		"ResourceOwner": true,               // Allow querying by owner
		"CreatedAt":     true,               // Upload timestamp
		"UpdatedAt":     true,               // Last update timestamp
	}

	// ReadableFields: Fields included in API responses
	readableFields := map[string]bool{
		"ID":            true,
		"GameID":        true,
		"NetworkID":     true,
		"Size":          true,
		"InternalURI":   shared.DENY,        // SECURITY: Never expose storage paths
		"Status":        true,
		"Error":         shared.DENY,        // SECURITY: Never expose error details
		"Header":        true,
		"ResourceOwner": shared.DENY,        // SECURITY: Hide resource ownership
		"CreatedAt":     true,
		"UpdatedAt":     true,
	}

	service := &shared.BaseQueryService[replay_entity.ReplayFile]{
		Reader:          fileMetadataReader,
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,

		// DefaultSearchFields: Header contains map name for fuzzy search
		DefaultSearchFields: []string{"Header"},

		// SortableFields: Common sorting options
		SortableFields: []string{"CreatedAt", "UpdatedAt", "Size"},

		// FilterableFields: For exact-match filtering
		FilterableFields: []string{"GameID", "Status", "NetworkID"},

		MaxPageSize: 100,
		Audience:    shared.UserAudienceIDKey,

		// EntityType: URL path segment
		EntityType: "replays",
	}

	// CRITICAL: Register with global registry for schema discovery
	shared.GetQueryServiceRegistry().Register("replays", service)

	return service
}
