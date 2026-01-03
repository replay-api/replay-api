// Package common provides shared domain types and infrastructure for the replay-api.
package common

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
)

// BaseQueryService is the generic base implementation for all query services.
//
// HEXAGONAL ARCHITECTURE:
// This is part of the DOMAIN LAYER. It implements the SchemaProvider interface
// and provides common query functionality that all entity-specific query services inherit.
//
// USAGE FOR AI AGENTS:
//
// When creating a new query service, embed this struct and configure the fields:
//
//	type MyEntityQueryService struct {
//	    common.BaseQueryService[MyEntity]
//	}
//
//	func NewMyEntityQueryService(reader common.Searchable[MyEntity]) *MyEntityQueryService {
//	    queryableFields := map[string]bool{
//	        "ID":        true,              // Allow querying
//	        "Name":      true,              // Allow querying
//	        "Secret":    common.DENY,       // DENY = false, blocks querying
//	    }
//
//	    service := &common.BaseQueryService[MyEntity]{
//	        Reader:              reader,
//	        QueryableFields:     queryableFields,
//	        ReadableFields:      readableFields,
//	        DefaultSearchFields: []string{"Name"},
//	        SortableFields:      []string{"Name", "CreatedAt"},
//	        FilterableFields:    []string{"Status", "GameID"},
//	        MaxPageSize:         100,
//	        Audience:            common.UserAudienceIDKey,
//	        EntityType:          "myentities",  // URL path segment
//	    }
//
//	    // CRITICAL: Register with global registry for schema discovery
//	    common.GetQueryServiceRegistry().Register("myentities", service)
//
//	    return service
//	}
//
// FIELD DEFINITIONS:
//
//   - QueryableFields: map[string]bool - Fields that can be used in search queries.
//     Use `true` to allow, `common.DENY` (false) to block.
//     SECURITY: Only fields with `true` are exposed to external API.
//
//   - ReadableFields: map[string]bool - Fields included in API responses.
//     Use `true` to include, `common.DENY` (false) to hide.
//     SECURITY: Fields like InternalURI, Error should be DENY.
//
//   - DefaultSearchFields: []string - Fields used when user provides search text
//     without specifying which fields to search. Typically text fields.
//
//   - SortableFields: []string - Fields that support ORDER BY in queries.
//     Usually timestamps and display names.
//
//   - FilterableFields: []string - Fields that support exact-match filtering.
//     Usually IDs, status fields, and enums.
//
//   - EntityType: string - URL path segment (lowercase, plural).
//     Example: "players", "teams", "replays", "matches"
//
// SECURITY CONSIDERATIONS:
//
//  1. Always use DENY for internal fields (InternalURI, Error, etc.)
//  2. Always use DENY for ResourceOwner in ReadableFields
//  3. Never expose database-internal fields
//  4. Review queryable fields for potential data leakage
type BaseQueryService[T any] struct {
	// Reader is the data source implementing the Searchable interface.
	// This is typically a repository adapter.
	Reader Searchable[T]

	// QueryableFields defines which entity fields can be searched.
	// Key: field name (must match entity struct field name)
	// Value: true = allowed, false (DENY) = blocked
	// SECURITY: Only fields with true are exposed in API schema
	QueryableFields map[string]bool

	// ReadableFields defines which entity fields are returned in responses.
	// Key: field name (must match entity struct field name)
	// Value: true = included, false (DENY) = hidden
	// SECURITY: Sensitive fields should be DENY
	ReadableFields map[string]bool

	// DefaultSearchFields are used for fuzzy search when no specific fields provided.
	// Example: ["Nickname", "Description"] for player search
	DefaultSearchFields []string

	// SortableFields are fields that support ORDER BY operations.
	// Example: ["CreatedAt", "UpdatedAt", "Name"]
	SortableFields []string

	// FilterableFields are fields that support exact-match filtering.
	// Example: ["GameID", "Status", "VisibilityLevel"]
	FilterableFields []string

	// MaxPageSize limits the number of results per query.
	// Default: 100
	MaxPageSize uint

	// Audience defines the intended audience for this service.
	// Controls visibility and access permissions.
	Audience IntendedAudienceKey

	// EntityType is the URL path segment for this entity (lowercase, plural).
	// Example: "players", "teams", "replays"
	// CRITICAL: Must be set for registry to work
	EntityType string

	// name is cached entity name from reflection (internal use)
	name string
}

// GetQuerySchema implements the SchemaProvider interface.
//
// This method is called by the QueryServiceRegistry to expose the service's
// schema to external adapters (REST, gRPC, etc.).
//
// HEXAGONAL ARCHITECTURE:
// This is how the domain layer communicates its capabilities to adapters.
// The schema returned here is the SINGLE SOURCE OF TRUTH for what this
// service can do.
//
// Returns:
//   - QueryServiceSchema with all field definitions
//
// Thread Safety: Safe for concurrent calls (read-only operations)
func (service *BaseQueryService[T]) GetQuerySchema() QueryServiceSchema {
	var queryable []string
	var readable []string

	// Convert map[string]bool to []string for external exposure
	for field, allowed := range service.QueryableFields {
		if allowed {
			queryable = append(queryable, field)
		}
	}

	for field, allowed := range service.ReadableFields {
		if allowed {
			readable = append(readable, field)
		}
	}

	// Use defaults if not explicitly set
	sortable := service.SortableFields
	if len(sortable) == 0 {
		sortable = []string{"CreatedAt", "UpdatedAt"}
	}

	filterable := service.FilterableFields
	if len(filterable) == 0 {
		// Default filterable fields are non-text queryable fields
		for field := range service.QueryableFields {
			if field != "Description" && field != "Name" && field != "Nickname" {
				filterable = append(filterable, field)
			}
		}
	}

	return QueryServiceSchema{
		EntityType:          service.EntityType,
		EntityName:          service.GetName(),
		QueryableFields:     queryable,
		DefaultSearchFields: service.DefaultSearchFields,
		SortableFields:      sortable,
		FilterableFields:    filterable,
		ReadableFields:      readable,
	}
}

// GetName returns the entity name using reflection.
// Used for debugging and schema generation.
func (service *BaseQueryService[T]) GetName() string {
	if service.name != "" {
		return service.name
	}

	var t T
	service.name = reflect.TypeOf(t).Name()

	return service.name
}

// GetByID returns a single entity by its ID.
// Uses ClientApplicationAudienceIDKey as the intended audience.
func (service *BaseQueryService[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	params := []SearchAggregation{
		{
			Params: []SearchParameter{
				{
					ValueParams: []SearchableValue{
						{
							Field: "ID",
							Values: []interface{}{
								id,
							},
						},
					},
				},
			},
		},
	}

	visibility := SearchVisibilityOptions{
		RequestSource:    GetResourceOwner(ctx),
		IntendedAudience: ClientApplicationAudienceIDKey,
	}

	result := SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	search := Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}

	entities, err := service.Reader.Search(ctx, search)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error searching. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	res := entities[0]

	return &res, nil
}

func (service *BaseQueryService[T]) Search(ctx context.Context, s Search) ([]T, error) {
	var omitFields []string
	var pickFields []string
	for fieldName, isReadable := range service.ReadableFields {
		if !isReadable {
			omitFields = append(omitFields, fieldName)
			continue
		}

		pickFields = append(pickFields, fieldName)
	}

	if len(omitFields) > 0 {
		slog.Info("Omitting fields", "fields", omitFields)
	}

	s.ResultOptions.OmitFields = omitFields
	s.ResultOptions.PickFields = pickFields

	entities, err := service.Reader.Search(ctx, s)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error filtering. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	return entities, nil
}

func (svc *BaseQueryService[T]) Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error) {
	err := ValidateSearchParameters(searchParams, svc.QueryableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating search parameters: %v", err)
	}

	err = ValidateResultOptions(resultOptions, svc.ReadableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating result options: %v", err)
	}

	intendedAud := GetIntendedAudience(ctx)

	if intendedAud == nil {
		intendedAud = &svc.Audience
	}

	s := NewSearchByAggregation(ctx, searchParams, resultOptions, *intendedAud)

	return &s, nil
}
