package common

import (
	"testing"
)

func TestBaseQueryService_GetQuerySchema(t *testing.T) {
	service := &BaseQueryService[struct{}]{
		QueryableFields: map[string]bool{
			"ID":          true,
			"Name":        true,
			"Description": true,
			"Secret":      DENY,
		},
		ReadableFields: map[string]bool{
			"ID":          true,
			"Name":        true,
			"Description": true,
			"Secret":      DENY,
		},
		DefaultSearchFields: []string{"Name", "Description"},
		SortableFields:      []string{"Name", "CreatedAt"},
		FilterableFields:    []string{"Status", "GameID"},
		EntityType:          "testentities",
	}

	schema := service.GetQuerySchema()

	if schema.EntityType != "testentities" {
		t.Errorf("expected EntityType 'testentities', got '%s'", schema.EntityType)
	}

	foundSecret := false
	for _, field := range schema.QueryableFields {
		if field == "Secret" {
			foundSecret = true
		}
	}
	if foundSecret {
		t.Error("DENY field 'Secret' should not appear in queryable fields")
	}

	expectedQueryable := map[string]bool{"ID": true, "Name": true, "Description": true}
	for _, field := range schema.QueryableFields {
		if !expectedQueryable[field] {
			t.Errorf("unexpected queryable field: %s", field)
		}
	}

	if len(schema.DefaultSearchFields) != 2 {
		t.Errorf("expected 2 default search fields, got %d", len(schema.DefaultSearchFields))
	}

	if len(schema.SortableFields) != 2 {
		t.Errorf("expected 2 sortable fields, got %d", len(schema.SortableFields))
	}

	if len(schema.FilterableFields) != 2 {
		t.Errorf("expected 2 filterable fields, got %d", len(schema.FilterableFields))
	}
}

func TestBaseQueryService_GetQuerySchema_Defaults(t *testing.T) {
	service := &BaseQueryService[struct{}]{
		QueryableFields: map[string]bool{
			"ID":     true,
			"Name":   true,
			"GameID": true,
		},
		EntityType: "defaults",
	}

	schema := service.GetQuerySchema()

	if len(schema.SortableFields) != 2 {
		t.Errorf("expected 2 default sortable fields, got %d", len(schema.SortableFields))
	}

	if len(schema.FilterableFields) == 0 {
		t.Error("expected auto-derived filterable fields")
	}
}

func TestBaseQueryService_GetQuerySchema_DenyFiltering(t *testing.T) {
	service := &BaseQueryService[struct{}]{
		QueryableFields: map[string]bool{
			"Public":      true,
			"Internal":    DENY,
			"InternalURI": DENY,
			"Error":       DENY,
		},
		ReadableFields: map[string]bool{
			"Public":        true,
			"ResourceOwner": DENY,
		},
		EntityType: "securitytest",
	}

	schema := service.GetQuerySchema()

	if len(schema.QueryableFields) != 1 {
		t.Errorf("expected 1 queryable field, got %d", len(schema.QueryableFields))
	}
	if len(schema.QueryableFields) > 0 && schema.QueryableFields[0] != "Public" {
		t.Errorf("expected 'Public' as only queryable field, got '%s'", schema.QueryableFields[0])
	}

	if len(schema.ReadableFields) != 1 {
		t.Errorf("expected 1 readable field, got %d", len(schema.ReadableFields))
	}
	if len(schema.ReadableFields) > 0 && schema.ReadableFields[0] != "Public" {
		t.Errorf("expected 'Public' as only readable field, got '%s'", schema.ReadableFields[0])
	}
}

func TestDENY_Constant(t *testing.T) {
	if DENY != false {
		t.Error("DENY constant should be false")
	}

	fields := map[string]bool{
		"allowed": true,
		"denied":  DENY,
	}

	if fields["allowed"] != true {
		t.Error("allowed field should be true")
	}
	if fields["denied"] != false {
		t.Error("denied field should be false (DENY)")
	}
}
