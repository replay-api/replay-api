package squad_services

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
)

// SquadQueryService handles read operations for Squad (Team) entities.
//
// HEXAGONAL ARCHITECTURE:
// - Embeds BaseQueryService[Squad] for common query functionality
// - Implements squad_in.SquadReader port
// - Registers itself with QueryServiceRegistry for schema discovery
//
// ENDPOINT: GET /teams
//
// QUERYABLE FIELDS:
//   - Name, Symbol, Description: Text search (fuzzy match)
//   - GameID, VisibilityLevel: Exact match filters
//   - All fields support projection (select specific fields)
//
// SECURITY:
//   - ResourceOwner is DENY in ReadableFields (not exposed in API responses)
type SquadQueryService struct {
	shared.BaseQueryService[squad_entities.Squad]
}

// NewSquadQueryService creates and registers a new SquadQueryService.
//
// CRITICAL: This function registers the service with the global QueryServiceRegistry.
// This enables the SearchSchemaController to expose the schema to frontends.
//
// Parameters:
//   - eventReader: Repository adapter implementing SquadReader
//
// Returns:
//   - squad_in.SquadReader: The service implementing the port interface
func NewSquadQueryService(eventReader squad_out.SquadReader) squad_in.SquadReader {
	// QueryableFields: Fields available for search/filter queries
	queryableFields := map[string]bool{
		"ID":              true,               // Primary key
		"GroupID":         true,               // Parent group reference
		"GameID":          true,               // Filter by game (CS2, Valorant)
		"Name":            true,               // Team name - text search
		"Symbol":          true,               // Team symbol/tag (e.g., "NaVi")
		"SlugURI":         true,               // URL-friendly identifier
		"Description":     true,               // Team bio - text search
		"Membership":      true,               // Membership structure
		"LogoURI":         true,               // Team logo URL
		"BannerURI":       true,               // Team banner URL
		"VisibilityLevel": true,               // Public/Private/Friends
		"VisibilityType":  true,               // Visibility type enum
		"ResourceOwner":   true,               // Allow querying by owner
		"CreatedAt":       true,               // Creation timestamp
		"UpdatedAt":       true,               // Last update timestamp
	}

	// ReadableFields: Fields included in API responses
	readableFields := map[string]bool{
		"ID":              true,
		"GroupID":         true,
		"GameID":          true,
		"Name":            true,
		"Symbol":          true,
		"SlugURI":         true,
		"Description":     true,
		"Membership":      true,
		"LogoURI":         true,
		"BannerURI":       true,
		"VisibilityLevel": true,
		"VisibilityType":  true,
		"ResourceOwner":   shared.DENY, // SECURITY: Hide resource ownership
		"CreatedAt":       true,
		"UpdatedAt":       true,
	}

	service := &shared.BaseQueryService[squad_entities.Squad]{
		Reader:          eventReader.(shared.Searchable[squad_entities.Squad]),
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,

		// DefaultSearchFields: Searched when user provides text without fields
		// Example: GET /teams?q=faze → searches Name, Symbol, Description
		DefaultSearchFields: []string{"Name", "Symbol", "Description"},

		// SortableFields: Fields supporting ORDER BY
		SortableFields: []string{"Name", "CreatedAt", "UpdatedAt"},

		// FilterableFields: Fields for exact-match filtering
		FilterableFields: []string{"GameID", "VisibilityLevel"},

		MaxPageSize: 100,
		Audience:    shared.UserAudienceIDKey,

		// EntityType: URL path segment - MUST match route (/teams)
		EntityType: "teams",
	}

	// CRITICAL: Register with global registry for schema discovery
	shared.GetQueryServiceRegistry().Register("teams", service)

	return service
}
