// Package query_controllers provides HTTP handlers for query-related endpoints.
//
// HEXAGONAL ARCHITECTURE:
// This package is part of the ADAPTER LAYER (infrastructure). Controllers here
// translate HTTP requests to domain service calls and format responses.
//
// IMPORTANT FOR AI AGENTS:
// Controllers MUST NOT define business logic or schemas. They MUST read from
// the domain layer (services) via the QueryServiceRegistry.
package query_controllers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	common "github.com/replay-api/replay-api/pkg/domain"
)

// SearchSchemaController exposes the search schema API to frontend clients.
//
// HEXAGONAL ARCHITECTURE:
// This is an ADAPTER that exposes domain capabilities via HTTP REST.
// It reads from the QueryServiceRegistry to get schemas defined by services.
//
// CRITICAL FOR AI AGENTS:
// - This controller MUST NOT define its own schemas
// - All schema information comes from GetQueryServiceRegistry().GetAllSchemas()
// - This ensures consistency across all adapters (REST, gRPC, GraphQL)
//
// ENDPOINTS:
//   - GET /api/search/schema          → All entity schemas
//   - GET /api/search/schema/{entity} → Specific entity schema
//
// RESPONSE FORMAT:
//
//	{
//	  "version": "1.0.0",
//	  "entities": {
//	    "players": {
//	      "entity_type": "players",
//	      "queryable_fields": ["Nickname", "Description", ...],
//	      "default_search_fields": ["Nickname", "Description"],
//	      "sortable_fields": ["Nickname", "CreatedAt", "UpdatedAt"],
//	      "filterable_fields": ["GameID", "VisibilityLevel"]
//	    },
//	    "teams": { ... },
//	    "replays": { ... },
//	    "matches": { ... }
//	  }
//	}
//
// CACHING:
// Responses are cached for 1 hour (Cache-Control: public, max-age=3600)
// since schemas rarely change during runtime.
type SearchSchemaController struct {
	// registry is the global QueryServiceRegistry singleton.
	// All schema information is read from here.
	registry *common.QueryServiceRegistry
}

// NewSearchSchemaController creates a controller that reads from service registry.
//
// HEXAGONAL ARCHITECTURE:
// The registry is injected here, allowing for testing with mock registries.
// In production, this uses the global singleton from GetQueryServiceRegistry().
//
// Example usage:
//
//	controller := query_controllers.NewSearchSchemaController()
//	router.HandleFunc("/api/search/schema", controller.GetSearchSchemaHandler)
func NewSearchSchemaController() *SearchSchemaController {
	return &SearchSchemaController{
		registry: common.GetQueryServiceRegistry(),
	}
}

// getSchemas returns EntitySearchSchema map from the service registry.
//
// SINGLE SOURCE OF TRUTH:
// This method reads from QueryServiceRegistry which aggregates schemas
// from all registered query services. Services define their own schemas
// in their New*QueryService constructors.
//
// DATA FLOW:
//
//	Service defines schema → Registers with registry → Controller reads here
//
// Thread Safety: Safe for concurrent calls.
func (c *SearchSchemaController) getSchemas() map[string]common.EntitySearchSchema {
	schemas := make(map[string]common.EntitySearchSchema)

	// Read all schemas from the registry (populated by query services)
	for entityType, serviceSchema := range c.registry.GetAllSchemas() {
		// Copy and sort queryable fields for consistent output
		queryableFields := make([]string, len(serviceSchema.QueryableFields))
		copy(queryableFields, serviceSchema.QueryableFields)
		sort.Strings(queryableFields)

		schemas[entityType] = common.EntitySearchSchema{
			EntityType:          serviceSchema.EntityType,
			QueryableFields:     queryableFields,
			DefaultSearchFields: serviceSchema.DefaultSearchFields,
			SortableFields:      serviceSchema.SortableFields,
			FilterableFields:    serviceSchema.FilterableFields,
		}
	}

	return schemas
}

// GetSearchSchemaHandler returns the queryable fields for all registered entities.
//
// HTTP: GET /api/search/schema
//
// Response: 200 OK with SearchSchema JSON
//
// Headers:
//   - Content-Type: application/json
//   - Cache-Control: public, max-age=3600
//
// Used by frontend SDK to discover available search fields dynamically.
func (c *SearchSchemaController) GetSearchSchemaHandler(w http.ResponseWriter, r *http.Request) {
	schemas := c.getSchemas()

	response := common.SearchSchema{
		Version:  "1.0.0",
		Entities: schemas,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetEntitySchemaHandler returns the queryable fields for a specific entity type.
//
// HTTP: GET /api/search/schema/{entity_type}
//
// Path Parameters:
//   - entity_type: One of "players", "teams", "replays", "matches", etc.
//
// Response:
//   - 200 OK: EntitySearchSchema JSON
//   - 404 Not Found: Invalid entity type with list of valid types
//
// Headers:
//   - Content-Type: application/json
//   - Cache-Control: public, max-age=3600
func (c *SearchSchemaController) GetEntitySchemaHandler(w http.ResponseWriter, r *http.Request) {
	// Extract entity type from path (e.g., "players", "teams", "replays")
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "entity type required", http.StatusBadRequest)
		return
	}
	entityType := parts[len(parts)-1]

	schemas := c.getSchemas()
	schema, exists := schemas[entityType]
	if !exists {
		// Return list of valid entity types to help debugging
		validTypes := make([]string, 0, len(schemas))
		for k := range schemas {
			validTypes = append(validTypes, k)
		}
		sort.Strings(validTypes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "entity type not found",
			"valid_types": validTypes,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(schema)
}

// GetSchema returns the schema for a given entity type (for internal use).
//
// This method is for internal Go code that needs schema information.
// For HTTP access, use GetSearchSchemaHandler or GetEntitySchemaHandler.
//
// Parameters:
//   - entityType: The entity type to look up (e.g., "players", "teams")
//
// Returns:
//   - schema: Pointer to EntitySearchSchema if found
//   - exists: true if the entity type exists, false otherwise
func (c *SearchSchemaController) GetSchema(entityType string) (*common.EntitySearchSchema, bool) {
	schemas := c.getSchemas()
	schema, exists := schemas[entityType]
	if !exists {
		return nil, false
	}
	return &schema, true
}
