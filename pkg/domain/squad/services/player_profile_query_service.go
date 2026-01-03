// Package squad_services provides domain services for squad-related entities.
//
// HEXAGONAL ARCHITECTURE:
// This package is part of the DOMAIN LAYER. Services here implement business
// logic and define what fields are externally queryable/readable.
package squad_services

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

// PlayerProfileQueryService handles read operations for PlayerProfile entities.
//
// HEXAGONAL ARCHITECTURE:
// - Embeds BaseQueryService[PlayerProfile] for common query functionality
// - Implements squad_in.PlayerProfileReader port
// - Registers itself with QueryServiceRegistry for schema discovery
//
// ENDPOINT: GET /players
//
// QUERYABLE FIELDS:
//   - Nickname, Description: Text search (fuzzy match)
//   - GameID, VisibilityLevel, VisibilityType: Exact match filters
//   - All fields support projection (select specific fields)
//
// SECURITY:
//   - ResourceOwner is DENY in ReadableFields (not exposed in API responses)
//   - All queryable fields are safe for external exposure
type PlayerProfileQueryService struct {
	common.BaseQueryService[squad_entities.PlayerProfile]
}

// NewPlayerProfileQueryService creates and registers a new PlayerProfileQueryService.
//
// CRITICAL: This function registers the service with the global QueryServiceRegistry.
// This enables the SearchSchemaController to expose the schema to frontends.
//
// Parameters:
//   - eventReader: Repository adapter implementing PlayerProfileReader
//
// Returns:
//   - squad_in.PlayerProfileReader: The service implementing the port interface
//
// HEXAGONAL ARCHITECTURE:
// The eventReader is an ADAPTER (infrastructure layer) that is injected here.
// The service doesn't know about MongoDB or any specific database - it only
// knows about the Searchable[PlayerProfile] interface.
//
// Example usage in dependency injection:
//
//	// In wire.go or main.go:
//	playerRepo := mongodb.NewPlayerProfileRepository(db)
//	playerQueryService := squad_services.NewPlayerProfileQueryService(playerRepo)
func NewPlayerProfileQueryService(eventReader squad_out.PlayerProfileReader) squad_in.PlayerProfileReader {
	// QueryableFields defines which fields can appear in search queries.
	// These are the fields that external clients can use to filter/search.
	// SECURITY: Only fields with `true` are exposed in the API schema.
	queryableFields := map[string]bool{
		"ID":              true,               // Primary key - always queryable
		"GameID":          true,               // Filter by game (e.g., CS2, Valorant)
		"Nickname":        true,               // Player display name - text search
		"SlugURI":         true,               // URL-friendly identifier
		"Avatar":          true,               // Avatar URL
		"Roles":           true,               // Player roles (e.g., IGL, AWPer)
		"Description":     true,               // Player bio - text search
		"VisibilityLevel": true,               // Public/Private/Friends
		"VisibilityType":  true,               // Visibility type enum
		"ResourceOwner":   true,               // Allow querying by owner
		"CreatedAt":       true,               // Creation timestamp
		"UpdatedAt":       true,               // Last update timestamp
	}

	// ReadableFields defines which fields are included in API responses.
	// SECURITY: ResourceOwner is DENY to hide internal ownership data.
	readableFields := map[string]bool{
		"ID":              true,
		"GameID":          true,
		"Nickname":        true,
		"SlugURI":         true,
		"Avatar":          true,
		"Roles":           true,
		"Description":     true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   common.DENY, // SECURITY: Hide resource ownership from responses
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}

	service := &common.BaseQueryService[squad_entities.PlayerProfile]{
		Reader:          eventReader.(common.Searchable[squad_entities.PlayerProfile]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,

		// DefaultSearchFields: Used when user searches without specifying fields.
		// Example: GET /players?q=ninja → searches Nickname and Description
		DefaultSearchFields: []string{"Nickname", "Description"},

		// SortableFields: Fields that support ORDER BY.
		// Example: GET /players?sort=Nickname:asc
		SortableFields: []string{"Nickname", "CreatedAt", "UpdatedAt"},

		// FilterableFields: Fields that support exact-match filtering.
		// Example: GET /players?GameID=cs2&VisibilityLevel=public
		FilterableFields: []string{"GameID", "VisibilityLevel", "VisibilityType"},

		MaxPageSize: 100,
		Audience:    common.UserAudienceIDKey,

		// EntityType: URL path segment - MUST match route (e.g., /players)
		EntityType: "players",
	}

	// CRITICAL: Register with global registry for schema discovery.
	// This enables GET /api/search/schema to return this service's schema.
	// Without this, the frontend SDK won't know what fields are available.
	common.GetQueryServiceRegistry().Register("players", service)

	return service
}
