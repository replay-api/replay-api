# Search Architecture Documentation

## Overview

The search system uses a **generic, layered architecture** where:

1. **SDK (Frontend)** specifies which fields to search
2. **Controller (REST API)** is generic - passes fields through without transformation
3. **Query Service (Domain)** defines which fields are allowed via `queryableFields`
4. **Repository (Infrastructure)** builds MongoDB aggregation pipelines

## Key Principle

> **The SDK is responsible for specifying field names in PascalCase** that match the Go struct fields.
> The backend validates these against `queryableFields` but does NOT transform field names.

---

## Query Parameters

### Text Search (OR logic)

- `q`: The search term value
- `search_fields`: Comma-separated list of field names to search (uses `$or` + `$regex`)

Example:

```
GET /players?q=Nova&search_fields=Nickname
GET /squads?q=Phoenix&search_fields=Name,Symbol
```

### Exact Filters (AND logic)

Any query parameter not in the reserved list is treated as an exact filter:

```
GET /players?GameID=cs2
GET /squads?GameID=cs2&Symbol=ABC
```

### Reserved Parameters

These are NOT treated as filters:

- `q` - text search value
- `search_fields` - fields for text search
- `limit`, `offset`, `page` - pagination
- `sort`, `order` - sorting

---

## Entity Search Configuration

### BaseEntity Fields (inherited by all entities)

All entities inherit from `BaseEntity` and have these queryable fields:
| Field | Queryable | Readable | Notes |
|-------|-----------|----------|-------|
| ID | ✅ | ✅ | UUID, stored as `_id` in MongoDB |
| VisibilityLevel | ❌ | ✅ | Internal use |
| VisibilityType | ❌ | ✅ | Internal use |
| ResourceOwner | ✅ | ❌ | Used for tenancy, never returned |
| CreatedAt | ✅ | ✅ | |
| UpdatedAt | ✅ | ✅ | |

---

### PlayerProfile

**Collection**: `player_profiles`
**Route**: `GET /players`

| Field         | Queryable | Readable | SDK search_fields      |
| ------------- | --------- | -------- | ---------------------- |
| ID            | ✅        | ✅       | -                      |
| GameID        | ✅        | ✅       | -                      |
| **Nickname**  | ✅        | ✅       | ✅ Primary text search |
| SlugURI       | ❌        | ✅       | -                      |
| Avatar        | ✅        | ✅       | -                      |
| Roles         | ✅        | ✅       | -                      |
| Description   | ✅        | ✅       | -                      |
| ResourceOwner | ✅        | ❌       | -                      |
| CreatedAt     | ✅        | ✅       | -                      |
| UpdatedAt     | ✅        | ✅       | -                      |

**SDK Usage**:

```typescript
// Search by nickname
sdk.playerProfiles.searchPlayerProfiles({ nickname: "Nova" });
// Sends: GET /players?q=Nova&search_fields=Nickname
```

---

### Squad

**Collection**: `squads`
**Route**: `GET /squads`

| Field         | Queryable | Readable | SDK search_fields        |
| ------------- | --------- | -------- | ------------------------ |
| ID            | ✅        | ✅       | -                        |
| GroupID       | ✅        | ✅       | -                        |
| GameID        | ✅        | ✅       | -                        |
| **Name**      | ✅        | ✅       | ✅ Primary text search   |
| **Symbol**    | ✅        | ✅       | ✅ Secondary text search |
| SlugURI       | ✅        | ✅       | -                        |
| Description   | ✅        | ✅       | -                        |
| Membership    | ✅        | ✅       | -                        |
| LogoURI       | ✅        | ✅       | -                        |
| BannerURI     | ✅        | ✅       | -                        |
| ResourceOwner | ✅        | ❌       | -                        |
| CreatedAt     | ✅        | ✅       | -                        |
| UpdatedAt     | ✅        | ✅       | -                        |

**SDK Usage**:

```typescript
// Search by name or symbol
sdk.squads.searchSquads({ name: "Phoenix" });
// Sends: GET /squads?q=Phoenix&search_fields=Name,Symbol
```

---

### ReplayFile

**Collection**: `replay_files`
**Route**: `GET /games/{game_id}/replays`

| Field         | Queryable | Readable | SDK search_fields          |
| ------------- | --------- | -------- | -------------------------- |
| ID            | ✅        | ✅       | -                          |
| GameID        | ✅        | ✅       | -                          |
| NetworkID     | ✅        | ✅       | -                          |
| Size          | ✅        | ✅       | -                          |
| InternalURI   | ❌        | ❌       | Internal storage path      |
| Status        | ✅        | ✅       | -                          |
| Error         | ❌        | ❌       | Internal error details     |
| **Header.\*** | ✅        | ✅       | Wildcard for header fields |
| ResourceOwner | ✅        | ✅       | -                          |
| CreatedAt     | ✅        | ✅       | -                          |
| UpdatedAt     | ✅        | ✅       | -                          |

**Note**: ReplayFile doesn't have a text-searchable "name" field by default. Search by NetworkID or Header fields.

---

### Match

**Collection**: `match_metadata`
**Route**: `GET /games/{game_id}/matches`

| Field          | Queryable | Readable | SDK search_fields        |
| -------------- | --------- | -------- | ------------------------ |
| ID             | ✅        | ✅       | -                        |
| GameID         | ✅        | ✅       | -                        |
| ReplayFileID   | ✅        | ✅       | -                        |
| Visibility     | ✅        | ✅       | -                        |
| Scoreboard     | ✅        | ✅       | -                        |
| Scoreboard.\*  | ✅        | ✅       | Nested scoreboard fields |
| Events         | ✅        | ✅       | -                        |
| ShareTokens.\* | ✅        | ✅       | -                        |
| ResourceOwner  | ✅        | ✅       | -                        |
| CreatedAt      | ✅        | ✅       | -                        |
| UpdatedAt      | ✅        | ✅       | -                        |

**Note**: Match doesn't have NetworkID or Status - those are on ReplayFile. Use ReplayFileID to link.

---

### Team

**Collection**: `teams` (embedded in matches)
**Route**: `GET /teams`

| Field              | Queryable | Readable | SDK search_fields        |
| ------------------ | --------- | -------- | ------------------------ |
| ID                 | ✅        | ✅       | -                        |
| NetworkID          | ✅        | ✅       | -                        |
| NetworkTeamID      | ✅        | ✅       | -                        |
| TeamHashID         | ✅        | ✅       | -                        |
| **Name**           | ✅        | ✅       | ✅ Primary text search   |
| **ShortName**      | ✅        | ✅       | ✅ Secondary text search |
| CurrentDisplayName | ✅        | ✅       | -                        |
| NameHistory        | ❌        | ✅       | -                        |
| Players            | ❌        | ✅       | -                        |
| ResourceOwner      | ✅        | ❌       | -                        |
| CreatedAt          | ✅        | ✅       | -                        |
| UpdatedAt          | ✅        | ✅       | -                        |

---

## Search Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                           FRONTEND SDK                               │
│                                                                       │
│  searchPlayerProfiles({ nickname: 'Nova' })                          │
│  → GET /players?q=Nova&search_fields=Nickname                        │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    DEFAULT SEARCH CONTROLLER                         │
│  (cmd/rest-api/controllers/default_search_controller.go)            │
│                                                                       │
│  1. Parse query params                                               │
│  2. Build Search object with:                                        │
│     - Text search: q + search_fields → OR aggregation                │
│     - Exact filters: other params → AND aggregation                  │
│  3. Call Compile() for validation                                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       BASE QUERY SERVICE                             │
│  (pkg/domain/base_query_service.go)                                 │
│                                                                       │
│  Compile(searchParams, resultOptions):                               │
│  1. ValidateSearchParameters(params, queryableFields)                │
│     → Checks each field against queryableFields map                  │
│  2. ValidateResultOptions(options, readableFields)                   │
│  3. Build Search with visibility/tenancy options                     │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      MONGODB REPOSITORY                              │
│  (pkg/infra/db/mongodb/mongodb_repository.go)                       │
│                                                                       │
│  setMatchValues():                                                   │
│  1. For each SearchParameter:                                        │
│     - Build inner clauses from ValueParams                           │
│     - Combine using parameter's AggregationClause (OR/AND)           │
│  2. Combine all parameters using outer AggregationClause             │
│                                                                       │
│  EnsureTenancy():                                                    │
│  - mergeTenancyCondition() combines search + tenancy with $and       │
│                                                                       │
│  Final MongoDB query:                                                │
│  { $match: { $and: [                                                 │
│      { $or: [{ nickname: /Nova/i }] },  // Search filter             │
│      { $or: [...tenancy conditions...] } // Visibility               │
│  ]}}                                                                 │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Adding a New Searchable Entity

### 1. Create Query Service

```go
func NewMyEntityQueryService(reader MyEntityReader) MyEntityReader {
    queryableFields := map[string]bool{
        "ID":            true,
        "Name":          true,  // Make searchable
        "Status":        true,
        "ResourceOwner": true,
        "CreatedAt":     true,
        "UpdatedAt":     true,
    }

    readableFields := map[string]bool{
        "ID":            true,
        "Name":          true,
        "Status":        true,
        "InternalField": common.DENY,  // Hide from responses
        "ResourceOwner": common.DENY,  // Hide tenancy info
        "CreatedAt":     true,
        "UpdatedAt":     true,
    }

    return &common.BaseQueryService[MyEntity]{
        Reader:          reader,
        QueryableFields: queryableFields,
        ReadableFields:  readableFields,
        MaxPageSize:     100,
        Audience:        common.UserAudienceIDKey,
    }
}
```

### 2. Register Route in Router

```go
myEntityController := query_controllers.NewMyEntityQueryController(container)
r.HandleFunc("/my-entities", myEntityController.DefaultSearchHandler).Methods("GET")
```

### 3. Update SDK

```typescript
async searchMyEntities(filters: {
  name?: string;
  status?: string;
  limit?: number;
}): Promise<MyEntity[]> {
  const params = new URLSearchParams();
  if (filters.name) {
    params.append('q', filters.name);
    params.append('search_fields', 'Name');  // PascalCase!
  }
  if (filters.status) params.append('Status', filters.status);  // PascalCase!
  if (filters.limit) params.append('limit', String(filters.limit));

  const response = await this.client.get<MyEntity[]>(`/my-entities?${params}`);
  return response.data || [];
}
```

---

## Common Mistakes to Avoid

### ❌ Hardcoding field names in controller

```go
// WRONG - don't hardcode field transformations
if gameID := query.Get("game_id"); gameID != "" {
    valueParams = append(valueParams, SearchableValue{Field: "GameID", ...})
}
```

### ✅ Let SDK specify fields directly

```typescript
// CORRECT - SDK specifies PascalCase field names
params.append("GameID", filters.game_id); // SDK handles the mapping
```

### ❌ Forgetting search_fields for text search

```typescript
// WRONG - q without search_fields does nothing
params.append("q", searchTerm);
```

### ✅ Always specify search_fields with q

```typescript
// CORRECT
params.append("q", searchTerm);
params.append("search_fields", "Nickname");
```

### ❌ Using snake_case for field names

```typescript
// WRONG
params.append("game_id", value);
```

### ✅ Using PascalCase matching Go struct

```typescript
// CORRECT
params.append("GameID", value);
```

---

## Field Name Reference

| SDK Parameter | Go Field  | MongoDB Field | Notes         |
| ------------- | --------- | ------------- | ------------- |
| ID            | ID        | \_id          | UUID          |
| GameID        | GameID    | game_id       |               |
| Nickname      | Nickname  | nickname      | PlayerProfile |
| Name          | Name      | name          | Squad, Team   |
| Symbol        | Symbol    | symbol        | Squad         |
| ShortName     | ShortName | short_name    | Team          |
| Status        | Status    | status        |               |
| CreatedAt     | CreatedAt | created_at    |               |
| UpdatedAt     | UpdatedAt | updated_at    |               |

---

## Wildcard Fields

Some entities support wildcard queries on nested objects:

```
Header.*  → Matches any field under Header (e.g., Header.MapName, Header.Duration)
```

This is useful for ReplayFile and Match entities that have dynamic header structures.
