# Generate OpenAPI from Routes

This guide explains how to use the script to automatically generate a new OpenAPI file with all missing routes.

## Script

The `scripts/generate-openapi-from-routes.go` script does the following:

1. **Extracts routes** from `router.go`
2. **Reads existing routes** from `openapi.yaml`
3. **Identifies missing routes**
4. **Generates new OpenAPI** with missing routes added

## How to Use

```bash
make docs-generate-routes
```

## What the Script Does

1. **Analyzes router.go** and extracts all routes
2. **Compares with existing openapi.yaml**
3. **Generates missing routes** with basic structure:
   - Tags based on path
   - Automatic summary
   - Unique operation ID
   - Default responses (200, 400, 401, 500)

4. **Creates new file**: `cmd/rest-api/docs/openapi.yaml`

## Generated Structure

Each generated route will have:

```yaml
  /path/to/endpoint:
    get:
      tags: [TagName]
      summary: Auto-generated summary
      operationId: get_path_to_endpoint
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalServerError'
```

## Next Steps After Generating

1. **Review the generated file**:
   ```bash
   # View the file
   cat cmd/rest-api/docs/openapi.yaml.new
   ```

2. **Add details**:
   - Request bodies (if needed)
   - Specific response schemas
   - Query/path parameters
   - Examples
   - More detailed descriptions

3. **Replace the original file**:
   ```bash
   # Windows PowerShell
   Move-Item -Force cmd/rest-api/docs/openapi.yaml cmd/rest-api/docs/openapi.yaml
   
   # Linux/Mac
   mv cmd/rest-api/docs/openapi.yaml cmd/rest-api/docs/openapi.yaml
   ```

4. **Validate**:
   ```bash
   make docs-validate
   ```

## Automatic Tags

The script infers tags based on the path:

- `/auth/*` → `Authentication`
- `/players/*` → `Players`
- `/match*` → `Matches`
- `/match-making/*` → `Matchmaking`
- `/tournament*` → `Tournaments`
- `/prize-pool*` → `Prize Pools`
- `/squad*` → `Squads`
- `/replay*` → `Replays`
- `/billing*`, `/withdrawal*`, `/payment*` → `Billing`
- `/leaderboard*`, `/rating*` → `Leaderboard`
- `/wallet*` → `Wallet`
- `/lobby*` → `Lobbies`
- `/health*` → `Health`
- `/group*` → `IAM`
- Others → `General`

## Complete Example

```bash
# 1. Generate OpenAPI with missing routes
make docs-generate-routes

# 2. Review the generated file
code cmd/rest-api/docs/openapi.yaml

# 3. Add schemas and details manually

# 4. Replace original file
Move-Item -Force cmd/rest-api/docs/openapi.yaml cmd/rest-api/docs/openapi.yaml

# 5. Validate
make docs-validate
```

## Limitations

The script generates a **basic** structure. You will need to:

- ✅ Add request bodies
- ✅ Define specific response schemas
- ✅ Add parameters (query, path, header)
- ✅ Improve descriptions
- ✅ Add examples
- ✅ Define required authentication

## Tips

1. **Use the script regularly** to keep it synchronized
2. **Always review** before replacing the original file
3. **Add schemas** in `components` when needed
4. **Validate** after each change

## Troubleshooting

### Error: "No routes found"
- Check if `router.go` is in the correct path
- Check if routes are in the expected format

### Routes don't appear
- Check if routes are not commented out
- Check if they are not OPTIONS routes

### Incorrect paths
- The script tries to resolve constants, but may need adjustments
- Manually review paths that seem incorrect
