# Query Service Architecture

## Overview

This document describes the **Query Service Architecture** used in the replay-api. It follows **Hexagonal Architecture** (Ports & Adapters) principles to ensure clean separation of concerns, testability, and consistency across all external interfaces.

## Table of Contents

1. [Architecture Diagram](#architecture-diagram)
2. [Core Components](#core-components)
3. [Data Flow](#data-flow)
4. [Creating a New Query Service](#creating-a-new-query-service)
5. [Security Considerations](#security-considerations)
6. [API Reference](#api-reference)
7. [Examples](#examples)

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL ADAPTERS                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
│  │ REST API     │  │ gRPC API     │  │ GraphQL API  │  (Future)         │
│  │ Controller   │  │ Handler      │  │ Resolver     │                   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                   │
│         │                 │                 │                           │
│         └─────────────────┼─────────────────┘                           │
│                           ▼                                             │
│              ┌────────────────────────┐                                 │
│              │  QueryServiceRegistry  │◄─── SINGLE POINT OF ACCESS     │
│              │  (query_service_       │                                 │
│              │   registry.go)         │                                 │
│              └────────────┬───────────┘                                 │
│                           │                                             │
│         ┌─────────────────┼─────────────────┐                           │
│         ▼                 ▼                 ▼                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
│  │ PlayerQuery  │  │ SquadQuery   │  │ ReplayQuery  │  ... more         │
│  │ Service      │  │ Service      │  │ Service      │                   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                   │
│         │                 │                 │                           │
│         └─────────────────┼─────────────────┘                           │
│                           ▼                                             │
│              ┌────────────────────────┐                                 │
│              │  BaseQueryService[T]   │◄─── IMPLEMENTS SchemaProvider   │
│              │  (base_query_service.  │                                 │
│              │   go)                  │                                 │
│              └────────────────────────┘                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. QueryServiceRegistry (`pkg/domain/query_service_registry.go`)

The **QueryServiceRegistry** is a singleton that holds references to all query services. It provides:

- **Thread-safe registration** of query services
- **Schema discovery** for external adapters
- **Consistent facade** across all API types (REST, gRPC, GraphQL)

```go
// Access the global registry
registry := common.GetQueryServiceRegistry()

// Register a service
registry.Register("players", playerQueryService)

// Get all schemas (used by controllers)
schemas := registry.GetAllSchemas()
```

### 2. SchemaProvider Interface

Every query service must implement the `SchemaProvider` interface:

```go
type SchemaProvider interface {
    GetQuerySchema() QueryServiceSchema
}
```

The `BaseQueryService[T]` already implements this, so embedding it is sufficient.

### 3. BaseQueryService[T] (`pkg/domain/base_query_service.go`)

The generic base implementation for all query services. It provides:

- **Field visibility control** (QueryableFields, ReadableFields)
- **Search configuration** (DefaultSearchFields, SortableFields, FilterableFields)
- **Schema generation** via `GetQuerySchema()`
- **Common query methods** (GetByID, Search, etc.)

```go
type BaseQueryService[T any] struct {
    Reader              Searchable[T]        // Data source (repository)
    QueryableFields     map[string]bool      // Fields for queries
    ReadableFields      map[string]bool      // Fields in responses
    DefaultSearchFields []string             // Fuzzy search fields
    SortableFields      []string             // ORDER BY fields
    FilterableFields    []string             // Exact match fields
    MaxPageSize         uint                 // Pagination limit
    Audience            IntendedAudienceKey  // Access control
    EntityType          string               // URL path segment
}
```

### 4. QueryServiceSchema

The schema struct that gets exposed to external clients:

```go
type QueryServiceSchema struct {
    EntityType          string   `json:"entity_type"`
    EntityName          string   `json:"entity_name"`
    QueryableFields     []string `json:"queryable_fields"`
    DefaultSearchFields []string `json:"default_search_fields"`
    SortableFields      []string `json:"sortable_fields"`
    FilterableFields    []string `json:"filterable_fields"`
    ReadableFields      []string `json:"readable_fields"`
}
```

---

## Data Flow

### Schema Registration Flow

```
1. Application starts
2. Dependency injection creates query services
3. Each New*QueryService() function:
   a. Defines QueryableFields, ReadableFields, etc.
   b. Creates BaseQueryService instance
   c. Calls GetQueryServiceRegistry().Register(entityType, service)
4. SearchSchemaController.GetSearchSchemaHandler() is called
5. Controller calls registry.GetAllSchemas()
6. Each service's GetQuerySchema() is invoked
7. Aggregated schemas are returned as JSON
```

### Query Execution Flow

```
1. HTTP request: GET /players?q=ninja&GameID=cs2
2. Controller extracts query parameters
3. Query service's Search() method is called
4. Service validates fields against QueryableFields
5. Repository executes MongoDB query
6. Results are filtered against ReadableFields
7. JSON response is returned
```

---

## Creating a New Query Service

### Step-by-Step Guide

#### 1. Define the Service Struct

```go
package myentity_services

import (
    common "github.com/replay-api/replay-api/pkg/domain"
    myentity "github.com/replay-api/replay-api/pkg/domain/myentity/entities"
    myentity_in "github.com/replay-api/replay-api/pkg/domain/myentity/ports/in"
    myentity_out "github.com/replay-api/replay-api/pkg/domain/myentity/ports/out"
)

// MyEntityQueryService handles read operations for MyEntity.
//
// HEXAGONAL ARCHITECTURE:
// - Embeds BaseQueryService[MyEntity] for common query functionality
// - Implements myentity_in.MyEntityReader port
// - Registers itself with QueryServiceRegistry for schema discovery
//
// ENDPOINT: GET /myentities
type MyEntityQueryService struct {
    common.BaseQueryService[myentity.MyEntity]
}
```

#### 2. Create the Constructor

```go
func NewMyEntityQueryService(reader myentity_out.MyEntityReader) myentity_in.MyEntityReader {
    // Define queryable fields
    // SECURITY: Only fields with `true` are exposed in API schema
    queryableFields := map[string]bool{
        "ID":          true,               // Primary key
        "Name":        true,               // Text search field
        "Description": true,               // Text search field
        "Status":      true,               // Filter field
        "GameID":      true,               // Filter field
        "InternalRef": common.DENY,        // SECURITY: Internal reference
        "CreatedAt":   true,               // Sort field
        "UpdatedAt":   true,               // Sort field
    }

    // Define readable fields
    // SECURITY: ResourceOwner should typically be DENY
    readableFields := map[string]bool{
        "ID":            true,
        "Name":          true,
        "Description":   true,
        "Status":        true,
        "GameID":        true,
        "InternalRef":   common.DENY,      // SECURITY: Hide internal data
        "ResourceOwner": common.DENY,      // SECURITY: Hide ownership
        "CreatedAt":     true,
        "UpdatedAt":     true,
    }

    service := &common.BaseQueryService[myentity.MyEntity]{
        Reader:          reader.(common.Searchable[myentity.MyEntity]),
        QueryableFields: queryableFields,
        ReadableFields:  readableFields,

        // DefaultSearchFields: Used when user provides text without fields
        // Example: GET /myentities?q=test → searches Name and Description
        DefaultSearchFields: []string{"Name", "Description"},

        // SortableFields: Fields that support ORDER BY
        // Example: GET /myentities?sort=CreatedAt:desc
        SortableFields: []string{"Name", "CreatedAt", "UpdatedAt"},

        // FilterableFields: Fields for exact-match filtering
        // Example: GET /myentities?GameID=cs2&Status=active
        FilterableFields: []string{"GameID", "Status"},

        MaxPageSize: 100,
        Audience:    common.UserAudienceIDKey,

        // EntityType: URL path segment - MUST match route
        EntityType: "myentities",
    }

    // CRITICAL: Register with global registry for schema discovery
    common.GetQueryServiceRegistry().Register("myentities", service)

    return service
}
```

#### 3. Wire Up in Dependency Injection

```go
// In cmd/rest-api/main.go or wire.go
myEntityRepo := mongodb.NewMyEntityRepository(db)
myEntityQueryService := myentity_services.NewMyEntityQueryService(myEntityRepo)
```

#### 4. The Schema is Automatically Exposed

Once registered, the schema appears in:

- `GET /api/search/schema` (all schemas)
- `GET /api/search/schema/myentities` (specific schema)

---

## Security Considerations

### Field Visibility

| Field Type  | QueryableFields | ReadableFields | Notes                  |
| ----------- | --------------- | -------------- | ---------------------- |
| Public Data | `true`          | `true`         | Normal fields          |
| Query-Only  | `true`          | `common.DENY`  | Can filter but not see |
| Read-Only   | `common.DENY`   | `true`         | Can see but not filter |
| Internal    | `common.DENY`   | `common.DENY`  | Completely hidden      |

### Sensitive Fields to ALWAYS DENY

- `InternalURI` - Storage paths
- `Error` - Internal error details
- `ResourceOwner` - Ownership information (in ReadableFields)
- Database internal fields
- Credentials or tokens

### Example Security Pattern

```go
queryableFields := map[string]bool{
    "ID":            true,
    "PublicName":    true,
    "InternalPath":  common.DENY,  // SECURITY: Never expose
    "ErrorDetails":  common.DENY,  // SECURITY: Never expose
    "ResourceOwner": true,         // Can query by owner
}

readableFields := map[string]bool{
    "ID":            true,
    "PublicName":    true,
    "InternalPath":  common.DENY,  // SECURITY: Never expose
    "ErrorDetails":  common.DENY,  // SECURITY: Never expose
    "ResourceOwner": common.DENY,  // SECURITY: Hide in responses
}
```

---

## API Reference

### GET /api/search/schema

Returns all registered entity schemas.

**Response:**

```json
{
  "version": "1.0.0",
  "entities": {
    "players": {
      "entity_type": "players",
      "entity_name": "PlayerProfile",
      "queryable_fields": ["CreatedAt", "Description", "GameID", "ID", "Nickname", ...],
      "default_search_fields": ["Nickname", "Description"],
      "sortable_fields": ["Nickname", "CreatedAt", "UpdatedAt"],
      "filterable_fields": ["GameID", "VisibilityLevel", "VisibilityType"]
    },
    "teams": { ... },
    "replays": { ... },
    "matches": { ... }
  }
}
```

**Headers:**

- `Content-Type: application/json`
- `Cache-Control: public, max-age=3600`

### GET /api/search/schema/{entity_type}

Returns schema for a specific entity.

**Parameters:**

- `entity_type`: One of `players`, `teams`, `replays`, `matches`, etc.

**Response (200 OK):**

```json
{
  "entity_type": "players",
  "entity_name": "PlayerProfile",
  "queryable_fields": ["CreatedAt", "Description", "GameID", "ID", "Nickname", ...],
  "default_search_fields": ["Nickname", "Description"],
  "sortable_fields": ["Nickname", "CreatedAt", "UpdatedAt"],
  "filterable_fields": ["GameID", "VisibilityLevel", "VisibilityType"]
}
```

**Response (404 Not Found):**

```json
{
  "error": "entity type not found",
  "valid_types": ["matches", "players", "replays", "teams"]
}
```

---

## Examples

### Frontend SDK Usage

```typescript
// Fetch schema from backend
const response = await fetch("/api/search/schema");
const schema = await response.json();

// Use schema to build dynamic search UI
const playerFields = schema.entities.players;
console.log("Can search:", playerFields.queryable_fields);
console.log("Default search in:", playerFields.default_search_fields);
console.log("Can sort by:", playerFields.sortable_fields);
console.log("Can filter by:", playerFields.filterable_fields);
```

### Search Query Examples

```bash
# Search players by nickname
curl "http://localhost:8080/players?q=ninja&search_fields=Nickname"

# Search with default fields (Nickname, Description)
curl "http://localhost:8080/players?q=pro gamer"

# Filter by game
curl "http://localhost:8080/players?GameID=cs2"

# Combined search and filter
curl "http://localhost:8080/players?q=ninja&GameID=cs2&sort=CreatedAt:desc"

# Search teams
curl "http://localhost:8080/teams?q=faze&search_fields=Name,Symbol"
```

---

## Key Principles for AI Agents

### MUST DO:

1. **Embed BaseQueryService[T]** in all query services
2. **Define QueryableFields** with explicit `true` or `common.DENY`
3. **Set EntityType** to match the URL path segment
4. **Call `GetQueryServiceRegistry().Register()`** in constructor
5. **Use `common.DENY`** for sensitive fields (InternalURI, Error, ResourceOwner in ReadableFields)

### MUST NOT DO:

1. **Never hardcode schemas** in controllers
2. **Never expose internal fields** (storage paths, error details)
3. **Never skip registration** with the global registry
4. **Never define schemas** in multiple places

### Verification Checklist

- [ ] Service embeds `BaseQueryService[T]`
- [ ] All fields in `QueryableFields` are intentional
- [ ] Sensitive fields are `common.DENY`
- [ ] `EntityType` matches URL path
- [ ] Service is registered with `GetQueryServiceRegistry().Register()`
- [ ] Schema appears in `GET /api/search/schema`

---

## File Reference

| File                                                               | Purpose                                      |
| ------------------------------------------------------------------ | -------------------------------------------- |
| `pkg/domain/query_service_registry.go`                             | Global registry and SchemaProvider interface |
| `pkg/domain/base_query_service.go`                                 | Generic base implementation                  |
| `pkg/domain/squad/services/player_profile_query_service.go`        | Player query service                         |
| `pkg/domain/squad/services/squad_query_service.go`                 | Team/Squad query service                     |
| `pkg/domain/replay/services/metadata/replay_file_query_service.go` | Replay file query service                    |
| `pkg/domain/replay/services/metadata/match_query_service.go`       | Match query service                          |
| `cmd/rest-api/controllers/query/search_schema_controller.go`       | REST API schema endpoint                     |

---

## Changelog

- **2026-01-03**: Initial documentation created
- **2026-01-03**: Refactored from hardcoded controller schemas to service-based registry
