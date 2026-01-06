package db_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestMatchMetadataRepository_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := getClient()
	assert.NoError(t, err, "Failed to connect to MongoDB")

	dbName := "replay"

	collectionName := "match_metadata"
	repo := db.NewMatchMetadataRepository(client, dbName, replay_entity.Match{}, collectionName)

	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	setContextWithValues := func(tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		newCtx := context.TODO()
		newCtx = context.WithValue(newCtx, shared.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, shared.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, shared.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, shared.UserIDKey, userID)
		return newCtx
	}

	defaultContext := setContextWithValues(tenantID, clientID, uuid.Nil, userID)
	squadAContext := setContextWithValues(tenantID, clientID, groupID, userID)

	sampleMatches := []replay_entity.Match{
		{ID: uuid.New(), GameID: replay_common.CS2_GAME_ID, Visibility: replay_entity.MatchVisibilityPublic, ResourceOwner: shared.GetResourceOwner(defaultContext)},
		{ID: uuid.New(), GameID: replay_common.CS2_GAME_ID, Visibility: replay_entity.MatchVisibilitySquad, ResourceOwner: shared.GetResourceOwner(squadAContext)},
		{ID: uuid.New(), GameID: replay_common.VLRNT_GAME_ID, Visibility: replay_entity.MatchVisibilityPublic, ResourceOwner: shared.GetResourceOwner(defaultContext)},
	}

	tests := []struct {
		name          string
		search        shared.Search
		expectedIDs   []uuid.UUID
		expectedError error
		context       context.Context
	}{
		{
			name:          "Filter by GameID",
			search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "GameID", Values: []interface{}{replay_common.CS2_GAME_ID}}}, shared.NewSearchResultOptions(0, 100), shared.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[0].ID, sampleMatches[1].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		{
			name:          "Filter by Visibility (As User)",
			search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, shared.NewSearchResultOptions(0, 100), shared.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[1].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		{
			name:          "Filter by Visibility (As Group)",
			search:        shared.NewSearchByValues(squadAContext, []shared.SearchableValue{{Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, shared.NewSearchResultOptions(0, 100), shared.GroupAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[1].ID},
			expectedError: nil,
			context:       squadAContext,
		},
		{
			name:          "Filter by GameID and Visibility",
			search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "GameID", Values: []interface{}{replay_common.CS2_GAME_ID}}, {Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilityPublic}}}, shared.NewSearchResultOptions(0, 100), shared.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[0].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		// },
		// {
		// 	name:          "Filter by GameID and Visibility with no matches",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "GameID", Values: []interface{}{shared.VLRNT_GAME_ID}}, {Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value"}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with invalid value",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "NonExistentField", Values: []interface{}{1}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with multiple values",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value1", "value2"}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with multiple values with invalid value",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value1", 2}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter nested field",
		// 	search:        shared.NewSearchByValues(defaultContext, []shared.SearchableValue{{Field: "Scoreboard.MVP", Values: []interface{}{uuid.New()}}}, shared.NewSearchResultOptions(0, 1), shared.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
	}

	collection := client.Database(dbName).Collection(collectionName)
	defer func() { _ = collection.Drop(defaultContext) }() // Clean up after tests

	interfaceMap := make([]interface{}, len(sampleMatches))
	for i, m := range sampleMatches {
		interfaceMap[i] = m
	}

	// Insert sample data
	_, err = collection.InsertMany(defaultContext, interfaceMap)
	assert.NoError(t, err, "Failed to insert sample matches")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.context == nil {
				tt.context = defaultContext
			}
			// Perform the search
			matches, err := repo.Search(tt.context, tt.search)

			// Assertions
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)

				// Check if the returned matches have the expected IDs
				assert.Len(t, matches, len(tt.expectedIDs))
				for _, expectedID := range tt.expectedIDs {
					found := false
					for _, match := range matches {
						if match.ID == expectedID {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected match ID %s not found in results", expectedID)
				}
			}
		})
	}
}
