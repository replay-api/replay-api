package db_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	db "github.com/psavelis/team-pro/replay-api/pkg/infra/db/mongodb"
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
		newCtx = context.WithValue(newCtx, common.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, common.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, common.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, common.UserIDKey, userID)
		return newCtx
	}

	defaultContext := setContextWithValues(tenantID, clientID, uuid.Nil, userID)
	squadAContext := setContextWithValues(tenantID, clientID, groupID, userID)

	sampleMatches := []replay_entity.Match{
		{ID: uuid.New(), GameID: common.CS2_GAME_ID, Visibility: replay_entity.MatchVisibilityPublic, ResourceOwner: common.GetResourceOwner(defaultContext)},
		{ID: uuid.New(), GameID: common.CS2_GAME_ID, Visibility: replay_entity.MatchVisibilitySquad, ResourceOwner: common.GetResourceOwner(squadAContext)},
		{ID: uuid.New(), GameID: common.VLRNT_GAME_ID, Visibility: replay_entity.MatchVisibilityPublic, ResourceOwner: common.GetResourceOwner(defaultContext)},
	}

	tests := []struct {
		name          string
		search        common.Search
		expectedIDs   []uuid.UUID
		expectedError error
		context       context.Context
	}{
		{
			name:          "Filter by GameID",
			search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}}}, common.NewSearchResultOptions(0, 100), common.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[0].ID, sampleMatches[1].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		{
			name:          "Filter by Visibility (As User)",
			search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, common.NewSearchResultOptions(0, 100), common.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[1].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		{
			name:          "Filter by Visibility (As Group)",
			search:        common.NewSearchByValues(squadAContext, []common.SearchableValue{{Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, common.NewSearchResultOptions(0, 100), common.GroupAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[1].ID},
			expectedError: nil,
			context:       squadAContext,
		},
		{
			name:          "Filter by GameID and Visibility",
			search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}}, {Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilityPublic}}}, common.NewSearchResultOptions(0, 100), common.UserAudienceIDKey),
			expectedIDs:   []uuid.UUID{sampleMatches[0].ID},
			expectedError: nil,
			context:       defaultContext,
		},
		// },
		// {
		// 	name:          "Filter by GameID and Visibility with no matches",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "GameID", Values: []interface{}{common.VLRNT_GAME_ID}}, {Field: "Visibility", Values: []interface{}{replay_entity.MatchVisibilitySquad}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value"}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with invalid value",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "NonExistentField", Values: []interface{}{1}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with multiple values",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value1", "value2"}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter by non-existent field with multiple values with invalid value",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "NonExistentField", Values: []interface{}{"value1", 2}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
		// {
		// 	name:          "Filter nested field",
		// 	search:        common.NewSearchByValues(defaultContext, []common.SearchableValue{{Field: "Scoreboard.MVP", Values: []interface{}{uuid.New()}}}, common.NewSearchResultOptions(0, 1), common.UserAudienceIDKey),
		// 	expectedIDs:   []uuid.UUID{},
		// 	expectedError: nil,
		// },
	}

	collection := client.Database(dbName).Collection(collectionName)
	defer collection.Drop(defaultContext) // Clean up after tests

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
