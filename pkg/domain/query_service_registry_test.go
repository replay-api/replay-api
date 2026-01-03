package common

import (
	"testing"
)

// MockSchemaProvider implements SchemaProvider for testing
type MockSchemaProvider struct {
	schema QueryServiceSchema
}

func (m *MockSchemaProvider) GetQuerySchema() QueryServiceSchema {
	return m.schema
}

func TestQueryServiceRegistry_Register(t *testing.T) {
	registry := &QueryServiceRegistry{
		services: make(map[string]SchemaProvider),
	}

	mockService := &MockSchemaProvider{
		schema: QueryServiceSchema{
			EntityType:          "players",
			EntityName:          "PlayerProfile",
			QueryableFields:     []string{"Nickname", "Description", "GameID"},
			DefaultSearchFields: []string{"Nickname", "Description"},
			SortableFields:      []string{"Nickname", "CreatedAt"},
			FilterableFields:    []string{"GameID"},
		},
	}

	registry.Register("players", mockService)

	service, exists := registry.Get("players")
	if !exists {
		t.Fatal("expected service to be registered")
	}

	if service != mockService {
		t.Error("expected registered service to match")
	}
}

func TestQueryServiceRegistry_Get(t *testing.T) {
	registry := &QueryServiceRegistry{
		services: make(map[string]SchemaProvider),
	}

	_, exists := registry.Get("nonexistent")
	if exists {
		t.Error("expected service to not exist")
	}

	mockService := &MockSchemaProvider{
		schema: QueryServiceSchema{EntityType: "teams"},
	}
	registry.Register("teams", mockService)

	service, exists := registry.Get("teams")
	if !exists {
		t.Fatal("expected service to exist after registration")
	}

	if service.GetQuerySchema().EntityType != "teams" {
		t.Error("expected EntityType to be 'teams'")
	}
}

func TestQueryServiceRegistry_GetAll(t *testing.T) {
	registry := &QueryServiceRegistry{
		services: make(map[string]SchemaProvider),
	}

	registry.Register("players", &MockSchemaProvider{
		schema: QueryServiceSchema{EntityType: "players"},
	})
	registry.Register("teams", &MockSchemaProvider{
		schema: QueryServiceSchema{EntityType: "teams"},
	})
	registry.Register("replays", &MockSchemaProvider{
		schema: QueryServiceSchema{EntityType: "replays"},
	})

	all := registry.GetAll()

	if len(all) != 3 {
		t.Errorf("expected 3 services, got %d", len(all))
	}

	for _, entityType := range []string{"players", "teams", "replays"} {
		if _, exists := all[entityType]; !exists {
			t.Errorf("expected %s to be in GetAll result", entityType)
		}
	}
}

func TestQueryServiceRegistry_GetAllSchemas(t *testing.T) {
	registry := &QueryServiceRegistry{
		services: make(map[string]SchemaProvider),
	}

	registry.Register("players", &MockSchemaProvider{
		schema: QueryServiceSchema{
			EntityType:          "players",
			EntityName:          "PlayerProfile",
			QueryableFields:     []string{"Nickname", "GameID"},
			DefaultSearchFields: []string{"Nickname"},
			SortableFields:      []string{"CreatedAt"},
			FilterableFields:    []string{"GameID"},
		},
	})
	registry.Register("teams", &MockSchemaProvider{
		schema: QueryServiceSchema{
			EntityType:          "teams",
			EntityName:          "Squad",
			QueryableFields:     []string{"Name", "Symbol"},
			DefaultSearchFields: []string{"Name", "Symbol"},
			SortableFields:      []string{"Name"},
			FilterableFields:    []string{"GameID"},
		},
	})

	schemas := registry.GetAllSchemas()

	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}

	playerSchema, exists := schemas["players"]
	if !exists {
		t.Fatal("expected players schema to exist")
	}
	if playerSchema.EntityType != "players" {
		t.Errorf("expected EntityType 'players', got '%s'", playerSchema.EntityType)
	}
	if playerSchema.EntityName != "PlayerProfile" {
		t.Errorf("expected EntityName 'PlayerProfile', got '%s'", playerSchema.EntityName)
	}
	if len(playerSchema.QueryableFields) != 2 {
		t.Errorf("expected 2 queryable fields, got %d", len(playerSchema.QueryableFields))
	}

	teamSchema, exists := schemas["teams"]
	if !exists {
		t.Fatal("expected teams schema to exist")
	}
	if teamSchema.EntityType != "teams" {
		t.Errorf("expected EntityType 'teams', got '%s'", teamSchema.EntityType)
	}
}

func TestQueryServiceRegistry_ThreadSafety(t *testing.T) {
	registry := &QueryServiceRegistry{
		services: make(map[string]SchemaProvider),
	}

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			registry.Register("players", &MockSchemaProvider{
				schema: QueryServiceSchema{EntityType: "players"},
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			registry.Get("players")
			registry.GetAll()
			registry.GetAllSchemas()
		}
		done <- true
	}()

	<-done
	<-done
}

func TestGetQueryServiceRegistry_Singleton(t *testing.T) {
	registry1 := GetQueryServiceRegistry()
	registry2 := GetQueryServiceRegistry()

	if registry1 != registry2 {
		t.Error("expected GetQueryServiceRegistry to return same instance")
	}
}
