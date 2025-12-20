# Controller Patterns

This guide explains the standardized controller patterns used in the LeetGaming Replay API, based on best practices from the [match-making-api](https://github.com/match-making-api/match-making-api) project.

## Standard Controller Structure

All controllers should follow this pattern:

```go
type MyController struct {
    container container.Container
    // Dependencies resolved from container
    useCaseHandler SomeUseCaseHandler
    repository     SomeRepository  // Only for read-only queries
}

func NewMyController(container container.Container) *MyController {
    ctrl := &MyController{
        container: container,
    }
    
    // Resolve dependencies in constructor
    if err := container.Resolve(&ctrl.useCaseHandler); err != nil {
        slog.Error("Failed to resolve use case handler", "err", err)
    }
    
    // Resolve repositories only for read-only queries
    if err := container.Resolve(&ctrl.repository); err != nil {
        slog.Error("Failed to resolve repository", "err", err)
    }
    
    return ctrl
}
```

## Key Principles

### 1. Use Container Directly

Controllers receive `container.Container` from `golobby/container/v3`:

✅ **Good**:
```go
func NewController(container container.Container) *Controller
```

❌ **Bad**:
```go
func NewController(container *container.Container) *Controller  // Pointer not needed
func NewController(container ioc.Container) *Controller         // Interface not used
```

### 2. Resolve Dependencies in Constructor

All dependencies should be resolved in the constructor, not in handlers:

✅ **Good**:
```go
func NewController(container container.Container) *Controller {
    ctrl := &Controller{container: container}
    
    // Resolve required dependencies
    if err := container.Resolve(&ctrl.requiredHandler); err != nil {
        slog.Error("Failed to resolve required handler", "err", err)
        // Panic or handle error appropriately
    }
    
    // Resolve optional dependencies
    if err := container.Resolve(&ctrl.optionalHandler); err != nil {
        slog.Warn("Optional handler not available", "err", err)
    }
    
    return ctrl
}
```

❌ **Bad**:
```go
func (c *Controller) Handler(ctx context.Context) http.HandlerFunc {
    var handler SomeHandler
    c.container.Resolve(&handler)  // Resolving in handler
    // ...
}
```

### 3. Store Container for Advanced Cases

Keep the container in the struct for cases where you need dynamic resolution:

```go
type MyController struct {
    container container.Container
    handler   SomeHandler
}

func NewMyController(container container.Container) *MyController {
    ctrl := &MyController{container: container}
    container.Resolve(&ctrl.handler)
    return ctrl
}
```

### 4. Use Use Cases, Not Repositories

Controllers should use use case handlers (command/query handlers) for business logic, not repositories directly:

✅ **Good** (Command Handler):
```go
func (c *Controller) CreateHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req CreateRequest
        json.NewDecoder(r.Body).Decode(&req)
        
        cmd := SomeCommand{...}
        result, err := c.commandHandler.Exec(r.Context(), cmd)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }
}
```

✅ **Acceptable** (Read-only queries):
```go
// For read-only queries, repositories can be used directly
func (c *Controller) GetHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := mux.Vars(r)["id"]
        result, err := c.repository.FindByID(r.Context(), id)
        // ...
    }
}
```

❌ **Bad** (Write operations via repository):
```go
func (c *Controller) CreateHandler(ctx context.Context) http.HandlerFunc {
    // Don't use repository for write operations
    result, err := c.repository.Save(ctx, entity)  // Use command handler instead
    // ...
}
```

### 5. Add Swagger Annotations

All handlers should have Swagger annotations for API documentation:

```go
// JoinQueueHandler handles POST /match-making/queue
// @Summary Join matchmaking queue
// @Description Adds a player or squad to the matchmaking queue
// @Tags Matchmaking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body JoinQueueRequest true "Join queue request"
// @Success 200 {object} JoinQueueResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-making/queue [post]
func (c *Controller) JoinQueueHandler(ctx context.Context) http.HandlerFunc {
    // ...
}
```

### 6. Handle Errors Consistently

Handle errors appropriately and return proper HTTP status codes:

```go
func (c *Controller) Handler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        result, err := c.handler.Exec(r.Context(), cmd)
        if err != nil {
            // Handle specific error types
            if errors.Is(err, ErrNotFound) {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
            }
            if errors.Is(err, ErrUnauthorized) {
                http.Error(w, err.Error(), http.StatusUnauthorized)
                return
            }
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        json.NewEncoder(w).Encode(result)
    }
}
```

## Example: Complete Controller

```go
package cmd_controllers

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    
    "github.com/golobby/container/v3"
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    common "github.com/replay-api/replay-api/pkg/domain"
    matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
    matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
)

type MatchmakingController struct {
    container               container.Container
    joinQueueHandler        matchmaking_in.JoinMatchmakingQueueCommandHandler
    leaveQueueHandler       matchmaking_in.LeaveMatchmakingQueueCommandHandler
    sessionRepo             matchmaking_out.MatchmakingSessionRepository  // For read-only queries
    poolRepo                matchmaking_out.MatchmakingPoolRepository     // For read-only queries
}

func NewMatchmakingController(container container.Container) *MatchmakingController {
    ctrl := &MatchmakingController{container: container}
    
    // Resolve command handlers (primary - use case layer)
    if err := container.Resolve(&ctrl.joinQueueHandler); err != nil {
        slog.Error("Failed to resolve JoinMatchmakingQueueCommandHandler", "err", err)
    }
    
    if err := container.Resolve(&ctrl.leaveQueueHandler); err != nil {
        slog.Error("Failed to resolve LeaveMatchmakingQueueCommandHandler", "err", err)
    }
    
    // Resolve repositories (for read-only queries only)
    if err := container.Resolve(&ctrl.sessionRepo); err != nil {
        slog.Error("Failed to resolve MatchmakingSessionRepository", "err", err)
    }
    
    if err := container.Resolve(&ctrl.poolRepo); err != nil {
        slog.Error("Failed to resolve MatchmakingPoolRepository", "err", err)
    }
    
    return ctrl
}

// JoinQueueRequest represents the request
type JoinQueueRequest struct {
    PlayerID string `json:"player_id" example:"550e8400-e29b-41d4-a716-446655440000"`
    GameID   string `json:"game_id" example:"cs2"`
    Region   string `json:"region" example:"us-east"`
}

// JoinQueueResponse represents the response
type JoinQueueResponse struct {
    SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
    Status    string `json:"status" example:"queued"`
    Position  int    `json:"position" example:"5"`
}

// JoinQueueHandler handles POST /match-making/queue
// @Summary Join matchmaking queue
// @Description Adds a player or squad to the matchmaking queue. Returns session ID and queue position.
// @Tags Matchmaking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body JoinQueueRequest true "Join queue request"
// @Success 200 {object} JoinQueueResponse "Successfully joined queue"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 409 {object} ErrorResponse "Already in queue"
// @Router /match-making/queue [post]
func (ctrl *MatchmakingController) JoinQueueHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        var req JoinQueueRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request body", http.StatusBadRequest)
            return
        }
        
        playerID, err := uuid.Parse(req.PlayerID)
        if err != nil {
            http.Error(w, "invalid player_id", http.StatusBadRequest)
            return
        }
        
        // Build command
        cmd := matchmaking_in.JoinMatchmakingQueueCommand{
            PlayerID: playerID,
            GameID:   req.GameID,
        }
        
        // Execute use case (command handler)
        session, err := ctrl.joinQueueHandler.Exec(r.Context(), cmd)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Return response
        response := JoinQueueResponse{
            SessionID: session.ID.String(),
            Status:    string(session.Status),
        }
        
        json.NewEncoder(w).Encode(response)
    }
}

// GetSessionStatusHandler handles GET /match-making/session/{session_id}
// @Summary Get matchmaking session status
// @Description Returns the current status of a matchmaking session
// @Tags Matchmaking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} SessionStatusResponse
// @Failure 404 {object} ErrorResponse "Session not found"
// @Router /match-making/session/{session_id} [get]
func (ctrl *MatchmakingController) GetSessionStatusHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        // For read-only queries, repository can be used directly
        sessionID := mux.Vars(r)["session_id"]
        id, err := uuid.Parse(sessionID)
        if err != nil {
            http.Error(w, "invalid session_id", http.StatusBadRequest)
            return
        }
        
        session, err := ctrl.sessionRepo.FindByID(r.Context(), id)
        if err != nil {
            http.Error(w, "session not found", http.StatusNotFound)
            return
        }
        
        json.NewEncoder(w).Encode(session)
    }
}
```

## Swagger Annotations

All handlers must include Swagger annotations for API documentation generation. Follow this pattern:

```go
// HandlerName handles METHOD /path/to/endpoint
// @Summary Short description
// @Description Detailed description of what the endpoint does
// @Tags TagName
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param param_name path/query/body type required "description"
// @Success status_code {object} ResponseModel "description"
// @Failure status_code {object} ErrorResponse "description"
// @Router /path/to/endpoint [method]
func (c *Controller) HandlerName(ctx context.Context) http.HandlerFunc {
    // ...
}
```

### Common Annotation Patterns

**Path Parameters**:
```go
// @Param id path string true "Resource ID"
```

**Query Parameters**:
```go
// @Param limit query int false "Results limit" default(20)
// @Param offset query int false "Results offset" default(0)
```

**Request Body**:
```go
// @Param request body RequestModel true "Request body"
```

**Response Models**:
```go
// @Success 200 {object} ResponseModel
// @Success 201 {object} CreatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
```

Generate documentation with:
```bash
make docs-generate
```

## Migration Guide

### From Old Pattern to New Pattern

**Old Pattern** (Resolving in handler):
```go
type Controller struct {
    container *container.Container
}

func NewController(container *container.Container) *Controller {
    return &Controller{container: container}
}

func (c *Controller) Handler(ctx context.Context) http.HandlerFunc {
    var handler SomeHandler
    c.container.Resolve(&handler)  // ❌ Resolving in handler
    // ...
}
```

**New Pattern** (Resolving in constructor):
```go
type Controller struct {
    container container.Container
    handler   SomeHandler
}

func NewController(container container.Container) *Controller {
    ctrl := &Controller{container: container}
    
    // ✅ Resolve dependencies in constructor
    if err := container.Resolve(&ctrl.handler); err != nil {
        slog.Error("Failed to resolve handler", "err", err)
    }
    
    return ctrl
}

func (c *Controller) Handler(ctx context.Context) http.HandlerFunc {
    // ✅ Use resolved handler directly
    result, err := c.handler.Exec(r.Context(), cmd)
    // ...
}
```

## Best Practices

1. **Resolve in Constructor**: Always resolve dependencies in the constructor, not in handlers
2. **Use Command/Query Handlers**: Prefer use case handlers over direct repository access
3. **Repositories for Read-Only**: Only use repositories directly for simple read-only queries
4. **Add Swagger Annotations**: Document all endpoints with Swagger annotations
5. **Error Handling**: Handle errors appropriately with proper HTTP status codes
6. **Request/Response Models**: Define clear request and response models with JSON tags
7. **Context Usage**: Always use `r.Context()` for request context, not the handler's context parameter

## Benefits

1. **Testability**: Easy to inject mocks via container
2. **Consistency**: All controllers follow the same pattern
3. **Performance**: Dependencies resolved once at startup, not per request
4. **Maintainability**: Clear separation of concerns
5. **Documentation**: Swagger annotations generate API docs automatically
6. **Type Safety**: Compile-time dependency checking

## Summary

- Use `container.Container` from `golobby/container/v3`
- Resolve dependencies in constructors, not in handlers
- Use command/query handlers for business logic
- Repositories can be used directly only for read-only queries
- Add Swagger annotations to all handlers
- Follow the standard controller structure
- Handle errors with appropriate HTTP status codes

For more information:
- [Dependency Injection Guide](./DEPENDENCY_INJECTION.md)
- [Swagger Workflow Guide](./SWAGGER_WORKFLOW.md)
- [Architecture Overview](../architecture/OVERVIEW.md)
