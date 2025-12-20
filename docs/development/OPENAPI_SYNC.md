# OpenAPI Synchronization Guide

This guide explains how to keep the OpenAPI file synchronized with the actual routes in the code.

## Problem

The OpenAPI file (`cmd/rest-api/docs/openapi.yaml`) can become outdated when:
- New routes are added to the router
- Routes are removed or modified
- Paths are changed

## Solution

### 1. Check Synchronization

Run the synchronization script to identify differences:

```bash
make docs-sync
```

Or directly:

```bash
go run scripts/sync-openapi.go
```

The script will:
- Extract all routes from `router.go`
- Extract all paths from `openapi.yaml`
- Compare and list differences

### 2. Update OpenAPI

After identifying missing routes:

1. **Add missing routes** to `openapi.yaml`
2. **Remove or update routes** that no longer exist
3. **Add schemas** for request/response if necessary

### 3. Validate OpenAPI

After updating, validate the file:

```bash
make docs-validate
```

## OpenAPI Structure

The OpenAPI file is located at `cmd/rest-api/docs/openapi.yaml` and is served via `embed.FS` in `cmd/rest-api/docs/swagger_handler.go`.

### Main Sections

```yaml
openapi: 3.1.0
info:
  title: LeetGaming Pro API
  version: 2.0.0
servers:
  - url: http://localhost:8080
paths:
  /path/to/endpoint:
    get:
      tags: [TagName]
      summary: Description
      responses:
        '200':
          description: Success
components:
  schemas:
    # Define request/response models
```

## Adding a New Route

### 1. Add to Router

```go
r.HandleFunc("/new/endpoint", controller.Handler(ctx)).Methods("POST")
```

### 2. Add to OpenAPI

```yaml
paths:
  /new/endpoint:
    post:
      tags: [TagName]
      summary: Description of endpoint
      operationId: newEndpoint
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NewRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/NewResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
```

### 3. Add Schemas (if necessary)

```yaml
components:
  schemas:
    NewRequest:
      type: object
      properties:
        field:
          type: string
    NewResponse:
      type: object
      properties:
        result:
          type: string
```

## Best Practices

1. **Document everything**: Add descriptions, examples, and schemas
2. **Use tags**: Organize endpoints by functionality
3. **Validate**: Always validate after changes
4. **Versioning**: Update the version when there are breaking changes

## Useful Commands

```bash
# Validate OpenAPI
make docs-validate

# Generate documentation (if using swag)
make docs-generate
```

## Troubleshooting

### Routes don't appear in Swagger UI

1. Check if the route is in `openapi.yaml`
2. Check if the file was copied to `cmd/rest-api/docs/`
3. Restart the server
4. Clear browser cache

### Validation error

1. Run `make docs-validate` to see specific errors
2. Check YAML syntax
3. Check schema references (`$ref`)

### Different paths

Some paths may have format differences:
- Router: `/games/{game_id}/match`
- OpenAPI: `/games/{game_id}/match` (same format)

The synchronization script normalizes paths for comparison.
