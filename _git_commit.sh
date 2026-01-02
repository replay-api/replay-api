#!/bin/bash
cd /Users/psavelis/github.com/leetgaming-pro/replay-api

# Stage specific files for the fix
git add cmd/rest-api/routing/router.go
git add cmd/rest-api/middlewares/cors_middleware.go
git add cmd/rest-api/controllers/default_search_controller.go
git add pkg/domain/squad/services/squad_query_service.go
git add pkg/infra/db/mongodb/squad_mongodb.go

# Show what's staged
echo "=== Staged files ==="
git diff --cached --name-only

# Commit
git commit -m "fix: CORS middleware and search API query param support

- Apply CORS middleware handler in router.go (was defined but not used)
- Add X-Search and x-search headers to CORS allowed headers
- Add buildSearchFromQueryParams() to DefaultSearchController for SDK compatibility
- Fix squad field mappings: FullName→Name, ShortName→Symbol
- Add missing field mappings: LogoURI, BannerURI
- Fix typo: create_at→created_at in squad repository"

echo "=== Current branch ==="
git branch --show-current
