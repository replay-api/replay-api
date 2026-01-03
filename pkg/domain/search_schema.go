package common

// ExternalQueryableFields defines fields that can be queried from outside the backend
// This is a security layer on top of repository-level QueryableFields
// Repository fields are for internal use, service fields are exposed to frontend SDK
type ExternalQueryableFields interface {
	// GetExternalQueryableFields returns field names that are safe to expose to frontend
	// The keys are PascalCase field names, values indicate if searchable
	GetExternalQueryableFields() map[string]bool

	// GetDefaultSearchFields returns the default fields to search when no specific fields are provided
	// This is used for fuzzy/global search functionality
	GetDefaultSearchFields() []string
}

// SearchSchemaProvider combines Searchable with ExternalQueryableFields
// Services implementing this can expose their schema to the frontend
type SearchSchemaProvider interface {
	ExternalQueryableFields
}

// EntitySearchSchema represents the search configuration for a single entity type
// This is returned to the frontend SDK
type EntitySearchSchema struct {
	EntityType          string   `json:"entity_type"`           // e.g., "players", "teams", "replays"
	QueryableFields     []string `json:"queryable_fields"`      // Fields that can be searched
	DefaultSearchFields []string `json:"default_search_fields"` // Fields used in fuzzy/global search
	SortableFields      []string `json:"sortable_fields"`       // Fields that can be sorted (subset of queryable)
	FilterableFields    []string `json:"filterable_fields"`     // Fields that support exact match filtering
}

// SearchSchema represents the complete search configuration for all entities
// Fetched once by frontend SDK and cached
type SearchSchema struct {
	Version  string                        `json:"version"`
	Entities map[string]EntitySearchSchema `json:"entities"`
}
