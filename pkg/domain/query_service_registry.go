// Package common provides shared domain types and infrastructure for the replay-api.
//
// HEXAGONAL ARCHITECTURE - QUERY SERVICE REGISTRY
// ================================================
//
// This file implements the Query Service Registry pattern, which is a core component
// of the hexagonal (ports & adapters) architecture used in this application.
//
// ARCHITECTURE OVERVIEW:
//
//	┌─────────────────────────────────────────────────────────────────────────┐
//	│                         EXTERNAL ADAPTERS                               │
//	│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
//	│  │ REST API     │  │ gRPC API     │  │ GraphQL API  │  (Future)         │
//	│  │ Controller   │  │ Handler      │  │ Resolver     │                   │
//	│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                   │
//	│         │                 │                 │                           │
//	│         └─────────────────┼─────────────────┘                           │
//	│                           ▼                                             │
//	│              ┌────────────────────────┐                                 │
//	│              │  QueryServiceRegistry  │◄─── SINGLE POINT OF ACCESS     │
//	│              │  (This File)           │                                 │
//	│              └────────────┬───────────┘                                 │
//	│                           │                                             │
//	│         ┌─────────────────┼─────────────────┐                           │
//	│         ▼                 ▼                 ▼                           │
//	│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
//	│  │ PlayerQuery  │  │ SquadQuery   │  │ ReplayQuery  │  ... more         │
//	│  │ Service      │  │ Service      │  │ Service      │                   │
//	│  └──────────────┘  └──────────────┘  └──────────────┘                   │
//	│         │                 │                 │                           │
//	│         └─────────────────┼─────────────────┘                           │
//	│                           ▼                                             │
//	│              ┌────────────────────────┐                                 │
//	│              │  BaseQueryService[T]   │◄─── IMPLEMENTS SchemaProvider   │
//	│              │  (base_query_service)  │                                 │
//	│              └────────────────────────┘                                 │
//	└─────────────────────────────────────────────────────────────────────────┘
//
// KEY PRINCIPLES:
//
//  1. SINGLE SOURCE OF TRUTH: Services define their own queryable fields.
//     Controllers/adapters MUST read from the registry, never hardcode schemas.
//
//  2. SELF-REGISTRATION: Each query service registers itself during initialization.
//     This ensures the registry is always up-to-date with actual capabilities.
//
//  3. FACADE CONSISTENCY: Any external adapter (REST, gRPC, GraphQL) gets the same
//     schema information from the registry, ensuring API consistency.
//
//  4. SECURITY BY DESIGN: Only fields explicitly marked as queryable in the service
//     are exposed. The repository may have additional internal-only fields.
//
// USAGE FOR AI AGENTS:
//
// When creating a new query service:
//  1. Embed BaseQueryService[T] in your service struct
//  2. Define QueryableFields, ReadableFields, DefaultSearchFields, etc.
//  3. Set EntityType (e.g., "players", "teams")
//  4. Call common.GetQueryServiceRegistry().Register(entityType, service)
//
// Example:
//
//	service := &common.BaseQueryService[MyEntity]{
//	    Reader:              reader,
//	    QueryableFields:     map[string]bool{"Name": true, "ID": true},
//	    DefaultSearchFields: []string{"Name"},
//	    EntityType:          "myentities",
//	}
//	common.GetQueryServiceRegistry().Register("myentities", service)
package common

import (
	"sync"
)

// QueryServiceSchema defines the schema exposed by a query service.
// This struct represents the "contract" between a query service and external consumers.
//
// IMPORTANT FOR AI AGENTS:
// - This is what gets serialized to JSON and sent to frontends/clients
// - All fields here are safe for external exposure
// - Never include internal fields like database paths or credentials
//
// Fields:
//   - EntityType: URL-friendly identifier (e.g., "players", "teams", "replays")
//   - EntityName: Go struct name for debugging (e.g., "PlayerProfile", "Squad")
//   - QueryableFields: Fields that can appear in search queries
//   - DefaultSearchFields: Fields searched when no specific field is provided
//   - SortableFields: Fields that support ORDER BY operations
//   - FilterableFields: Fields that support exact-match filtering
//   - ReadableFields: Fields included in API responses
type QueryServiceSchema struct {
	EntityType          string   `json:"entity_type"`           // URL path segment (e.g., "players")
	EntityName          string   `json:"entity_name"`           // Go struct name (e.g., "PlayerProfile")
	QueryableFields     []string `json:"queryable_fields"`      // Fields for search queries
	DefaultSearchFields []string `json:"default_search_fields"` // Default fuzzy search fields
	SortableFields      []string `json:"sortable_fields"`       // Fields supporting sort
	FilterableFields    []string `json:"filterable_fields"`     // Fields for exact filters
	ReadableFields      []string `json:"readable_fields"`       // Fields in responses
}

// SchemaProvider is the interface that query services MUST implement
// to participate in the registry system.
//
// HEXAGONAL ARCHITECTURE:
// This is a "port" - it defines the contract that the domain layer
// exposes to adapters (REST controllers, gRPC handlers, etc.)
//
// AI AGENTS MUST:
// - Ensure any new query service implements this interface
// - The BaseQueryService[T] already implements this, so embedding it is sufficient
type SchemaProvider interface {
	// GetQuerySchema returns the complete schema for this service.
	// This method is called by adapters to discover service capabilities.
	//
	// Returns:
	//   QueryServiceSchema with all field definitions
	//
	// Thread Safety:
	//   This method MUST be safe for concurrent calls
	GetQuerySchema() QueryServiceSchema
}

// QueryServiceRegistry is the central registry for all query services.
// It provides thread-safe access to registered services and their schemas.
//
// SINGLETON PATTERN:
// Use GetQueryServiceRegistry() to access the global instance.
// Do NOT create new instances directly.
//
// THREAD SAFETY:
// All methods are safe for concurrent use via sync.RWMutex.
type QueryServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]SchemaProvider
}

// queryServiceRegistry is the global singleton instance.
// IMPORTANT: Always use GetQueryServiceRegistry() to access this.
var queryServiceRegistry = &QueryServiceRegistry{
	services: make(map[string]SchemaProvider),
}

// GetQueryServiceRegistry returns the global query service registry singleton.
//
// USAGE:
//
//	// In a query service constructor:
//	common.GetQueryServiceRegistry().Register("players", service)
//
//	// In a controller/adapter:
//	registry := common.GetQueryServiceRegistry()
//	schemas := registry.GetAllSchemas()
//
// Thread Safety: Safe for concurrent use.
func GetQueryServiceRegistry() *QueryServiceRegistry {
	return queryServiceRegistry
}

// Register adds a query service to the registry.
//
// MUST be called during service initialization (typically in New*QueryService).
// The entityType should match the URL path segment (e.g., "players" for /players).
//
// Parameters:
//   - entityType: URL-friendly identifier (lowercase, plural, e.g., "players")
//   - service: The query service implementing SchemaProvider
//
// Example:
//
//	func NewPlayerQueryService(reader Searchable[Player]) *PlayerQueryService {
//	    service := &PlayerQueryService{...}
//	    common.GetQueryServiceRegistry().Register("players", service)
//	    return service
//	}
//
// Thread Safety: Safe for concurrent use.
func (r *QueryServiceRegistry) Register(entityType string, service SchemaProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[entityType] = service
}

// Get returns a specific service by entity type.
//
// Parameters:
//   - entityType: The entity type to look up (e.g., "players", "teams")
//
// Returns:
//   - service: The SchemaProvider if found
//   - ok: true if the service exists, false otherwise
//
// Thread Safety: Safe for concurrent use.
func (r *QueryServiceRegistry) Get(entityType string) (SchemaProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	service, ok := r.services[entityType]
	return service, ok
}

// GetAll returns a copy of all registered services.
//
// Returns a new map to prevent external modification of the registry.
//
// Thread Safety: Safe for concurrent use.
func (r *QueryServiceRegistry) GetAll() map[string]SchemaProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to avoid race conditions
	copy := make(map[string]SchemaProvider, len(r.services))
	for k, v := range r.services {
		copy[k] = v
	}
	return copy
}

// GetAllSchemas returns QueryServiceSchema from all registered services.
//
// This is the PRIMARY method used by external adapters (REST, gRPC, GraphQL)
// to discover all available entity schemas.
//
// HEXAGONAL ARCHITECTURE:
// This method is the "facade" that adapters use to get schema information.
// It ensures all adapters see the same schema definitions.
//
// Returns:
//   - map[string]QueryServiceSchema: entityType -> schema mapping
//
// Example (in REST controller):
//
//	func (c *SchemaController) GetAllSchemas(w http.ResponseWriter, r *http.Request) {
//	    schemas := common.GetQueryServiceRegistry().GetAllSchemas()
//	    json.NewEncoder(w).Encode(schemas)
//	}
//
// Thread Safety: Safe for concurrent use.
func (r *QueryServiceRegistry) GetAllSchemas() map[string]QueryServiceSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	schemas := make(map[string]QueryServiceSchema, len(r.services))
	for entityType, service := range r.services {
		schemas[entityType] = service.GetQuerySchema()
	}
	return schemas
}
