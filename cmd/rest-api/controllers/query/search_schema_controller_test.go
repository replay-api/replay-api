package query_controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type mockSchemaProvider struct {
	schema common.QueryServiceSchema
}

func (m *mockSchemaProvider) GetQuerySchema() common.QueryServiceSchema {
	return m.schema
}

func setupTestRegistry() *common.QueryServiceRegistry {
	registry := common.GetQueryServiceRegistry()

	registry.Register("players", &mockSchemaProvider{
		schema: common.QueryServiceSchema{
			EntityType:          "players",
			EntityName:          "PlayerProfile",
			QueryableFields:     []string{"Nickname", "Description", "GameID"},
			DefaultSearchFields: []string{"Nickname", "Description"},
			SortableFields:      []string{"Nickname", "CreatedAt", "UpdatedAt"},
			FilterableFields:    []string{"GameID", "VisibilityLevel"},
		},
	})
	registry.Register("teams", &mockSchemaProvider{
		schema: common.QueryServiceSchema{
			EntityType:          "teams",
			EntityName:          "Squad",
			QueryableFields:     []string{"Name", "Symbol", "Description"},
			DefaultSearchFields: []string{"Name", "Symbol"},
			SortableFields:      []string{"Name", "CreatedAt"},
			FilterableFields:    []string{"GameID"},
		},
	})

	return registry
}

func TestSearchSchemaController_GetSearchSchemaHandler(t *testing.T) {
	setupTestRegistry()
	controller := NewSearchSchemaController()

	req := httptest.NewRequest(http.MethodGet, "/api/search/schema", nil)
	w := httptest.NewRecorder()

	controller.GetSearchSchemaHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("expected Cache-Control 'public, max-age=3600', got '%s'", cacheControl)
	}

	var response common.SearchSchema
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", response.Version)
	}

	if len(response.Entities) < 2 {
		t.Errorf("expected at least 2 entities, got %d", len(response.Entities))
	}

	players, exists := response.Entities["players"]
	if !exists {
		t.Fatal("expected 'players' entity to exist")
	}
	if players.EntityType != "players" {
		t.Errorf("expected players.EntityType 'players', got '%s'", players.EntityType)
	}
}

func TestSearchSchemaController_GetEntitySchemaHandler(t *testing.T) {
	setupTestRegistry()
	controller := NewSearchSchemaController()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		entityType     string
	}{
		{
			name:           "valid entity - players",
			path:           "/api/search/schema/players",
			expectedStatus: http.StatusOK,
			entityType:     "players",
		},
		{
			name:           "valid entity - teams",
			path:           "/api/search/schema/teams",
			expectedStatus: http.StatusOK,
			entityType:     "teams",
		},
		{
			name:           "invalid entity",
			path:           "/api/search/schema/nonexistent",
			expectedStatus: http.StatusNotFound,
			entityType:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			controller.GetEntitySchemaHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var schema common.EntitySearchSchema
				if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if schema.EntityType != tt.entityType {
					t.Errorf("expected EntityType '%s', got '%s'", tt.entityType, schema.EntityType)
				}
			}

			if tt.expectedStatus == http.StatusNotFound {
				var errorResp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if _, exists := errorResp["valid_types"]; !exists {
					t.Error("expected 'valid_types' in error response")
				}
			}
		})
	}
}

func TestSearchSchemaController_GetSchema(t *testing.T) {
	setupTestRegistry()
	controller := NewSearchSchemaController()

	schema, exists := controller.GetSchema("players")
	if !exists {
		t.Fatal("expected 'players' schema to exist")
	}
	if schema.EntityType != "players" {
		t.Errorf("expected EntityType 'players', got '%s'", schema.EntityType)
	}

	_, exists = controller.GetSchema("nonexistent")
	if exists {
		t.Error("expected 'nonexistent' schema to not exist")
	}
}

func TestSearchSchemaController_QueryableFieldsSorted(t *testing.T) {
	setupTestRegistry()
	controller := NewSearchSchemaController()

	schemas := controller.getSchemas()
	players := schemas["players"]

	for i := 1; i < len(players.QueryableFields); i++ {
		if players.QueryableFields[i-1] > players.QueryableFields[i] {
			t.Errorf("queryable fields not sorted: %s > %s",
				players.QueryableFields[i-1], players.QueryableFields[i])
		}
	}
}
