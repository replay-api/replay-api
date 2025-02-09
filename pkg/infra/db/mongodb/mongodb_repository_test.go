package db_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	db "github.com/psavelis/team-pro/replay-api/pkg/infra/db/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName = "replay"
)

func Test_Mongo_QueryBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, err := getClient()
	if err != nil {
		failErr(t, err)
	}

	r := db.NewReplayFileMetadataRepository(client, dbName, replay_entity.ReplayFile{}, "replay_files")

	r.InitQueryableFields(map[string]bool{
		"ID":               true,
		"GameID":           true,
		"NetworkID":        true,
		"Size":             true,
		"InternalURI":      true,
		"Status":           true,
		"Error":            true,
		"Header":           true,
		"Header.Filestamp": true,
		"ResourceOwner":    true,
		"CreatedAt":        true,
		"UpdatedAt":        true,
	}, map[string]string{
		"ID":                     "_id",
		"GameID":                 "game_id",
		"NetworkID":              "network_id",
		"Size":                   "size",
		"InternalURI":            "uri",
		"Status":                 "status",
		"Error":                  "error",
		"Header":                 "header",
		"ResourceOwner":          "resource_owner",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
		"Header.Filestamp":       "header.filestamp",
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
	})

	fieldName, err := r.GetBSONFieldName("GameID")
	if err != nil {
		failErr(t, err)
	}

	if fieldName != "game_id" {
		t.Fatalf("expected bsonFieldName, got %s", fieldName)
	}

	ctx := context.WithValue(context.Background(), common.TenantIDKey, common.TeamPROTenantID)
	ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

	s := common.NewSearchByID(ctx, uuid.New(), common.ClientApplicationAudienceIDKey)

	if err != nil {
		failErr(t, err)
	}

	results, err := r.Query(ctx, s)

	if err != nil {
		failErr(t, err)
	}

	t.Logf("result: %v", results)
}

func failErr(t *testing.T, e error) {
	t.Fatalf("test failed %s %v", e.Error(), e)
}

var (
	clientInstance *mongo.Client
	clientOnce     sync.Once
)

func getClient() (*mongo.Client, error) {
	var err error
	if clientInstance == nil {
		clientOnce.Do(func() {
			opt := options.Client().ApplyURI("mongodb://127.0.0.1:37019/replay")
			// review: refactor (dry/config)
			clientInstance, err = mongo.Connect(context.Background(), opt)
		})
	}

	return clientInstance, err
}

func TestMongoDBRepository_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Establish a connection to a real MongoDB database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://127.0.0.1:37019/replay"))
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Use a dedicated collection for testing
	collectionName := "replay_files"
	repo := db.NewReplayFileMetadataRepository(client, dbName, replay_entity.ReplayFile{}, collectionName)

	setContextWithValues := func(ctx context.Context, tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, clientID)
		ctx = context.WithValue(ctx, common.UserIDKey, userID)
		ctx = context.WithValue(ctx, common.GroupIDKey, groupID)
		return ctx
	}

	repo.InitQueryableFields(map[string]bool{
		"ID":               true,
		"GameID":           true,
		"NetworkID":        true,
		"Size":             true,
		"InternalURI":      true,
		"Status":           true,
		"Error":            true,
		"Header":           true,
		"Header.Filestamp": true,
		"ResourceOwner":    true,
		"CreatedAt":        true,
		"UpdatedAt":        true,
	}, map[string]string{
		"ID":                     "_id",
		"GameID":                 "game_id",
		"NetworkID":              "network_id",
		"Size":                   "size",
		"InternalURI":            "uri",
		"Status":                 "status",
		"Error":                  "error",
		"Header":                 "header",
		"ResourceOwner":          "resource_owner",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
		"Header.Filestamp":       "header.filestamp",
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
	})

	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	sampleData := []replay_entity.ReplayFile{
		{
			ID:            uuid.MustParse("fcad61ef-67fe-4405-9a4e-1b51774bb46a"),
			GameID:        common.CS2_GAME_ID,
			NetworkID:     common.SteamNetworkIDKey,
			Size:          99999999999,
			Header:        cs_entity.CSReplayFileHeader{Filestamp: "HLTV-1.0.0"},
			InternalURI:   "https://leetgaming.pro/replays/fcad61ef-67fe-4405-9a4e-1b51774bb46a",
			ResourceOwner: common.ResourceOwner{TenantID: tenantID, ClientID: clientID, GroupID: groupID, UserID: userID},
			CreatedAt:     time.Now().Add(-96 * time.Hour),
			UpdatedAt:     time.Now().Add(-72 * time.Hour),
		},
		{
			ID:            uuid.MustParse("8097926d-5958-45fb-bf17-416659336058"),
			GameID:        common.CS2_GAME_ID,
			NetworkID:     common.FaceItNetworkIDKey,
			InternalURI:   "https://leetgaming.pro/replays/8097926d-5958-45fb-bf17-416659336058",
			Size:          1,
			Header:        cs_entity.CSReplayFileHeader{Filestamp: "HLTV-1.0.1"},
			ResourceOwner: common.ResourceOwner{TenantID: tenantID, GroupID: groupID, ClientID: clientID, UserID: userID},
			CreatedAt:     time.Now().Add(-48 * time.Hour),
			UpdatedAt:     time.Now().Add(-25 * time.Hour),
		},
		{
			ID:            uuid.MustParse("5c54807d-0339-451c-9f4b-47a2c05d9291"),
			GameID:        common.VLRNT_GAME_ID,
			NetworkID:     common.FaceItNetworkIDKey,
			InternalURI:   "https://leetgaming.pro/replays/5c54807d-0339-451c-9f4b-47a2c05d9291",
			Size:          1,
			Header:        cs_entity.CSReplayFileHeader{Filestamp: "HLTV-1.0.1"},
			ResourceOwner: common.ResourceOwner{TenantID: tenantID, ClientID: clientID, GroupID: groupID, UserID: userID},
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now(),
		},
	}

	tests := []struct {
		name              string
		search            common.Search
		expectedResults   []replay_entity.ReplayFile
		expectedError     error
		mockData          []replay_entity.ReplayFile
		contextValues     map[interface{}]uuid.UUID
		maxRecursiveDepth int
	}{
		{
			name: "Valid Query - GameID",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[:2],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},
		{
			name: "Valid Query - Header Wildcard (User Level)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "Header.Filestamp", Values: []interface{}{"HLTV-1.0.0"}}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[0:1],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},
		{
			name: "Valid Query - Header Wildcard (Client Level)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "Header.Filestamp", Values: []interface{}{"HLTV-1.0.0"}}},
				common.SearchResultOptions{Limit: 10, PickFields: []string{"Header", "GameID"}},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[0:1],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		{
			name: "Valid Query - Date Range",
			search: common.NewSearchByRange(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableDateRange{
					{Field: "CreatedAt", Min: &sampleData[2].CreatedAt, Max: &sampleData[2].UpdatedAt},
				},
				common.SearchResultOptions{
					Skip:       0,
					Limit:      10,
					PickFields: []string{},
					OmitFields: []string{},
				},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[2:3],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},
		// 1. Basic Valid Query - Filter by GameID
		{
			name: "Valid Query - GameID",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[:2], // Both CS2 games
			mockData:        sampleData,
			contextValues: map[interface{}]uuid.UUID{
				common.TenantIDKey: tenantID,
				common.ClientIDKey: clientID,
				common.UserIDKey:   userID,
			},
		},

		// 2. Nested Field Query - Header.Filestamp
		{
			name: "Valid Nested Query - Header.Filestamp",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "Header.Filestamp", Values: []interface{}{"HLTV-1.0.0"}}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[:1], // First game only
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},

		// 3. Multiple Values - Filtering by NetworkID
		{
			name: "Multiple Values - NetworkID (OR)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "NetworkID", Values: []interface{}{common.SteamNetworkIDKey, common.FaceItNetworkIDKey}, Operator: common.InOperator}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData, // All games match
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},

		// 5. String Field - Filtering by Error (Contains) - NOT TESTED due to no sampleData containing Error value
		{
			name: "String Filter - Error (Contains)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{
					{Field: "Error", Values: []interface{}{"connection"}, Operator: common.ContainsOperator},
				},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[0:0],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},

		// 6. Date Range Query - CreatedAt
		{
			name: "Date Range - CreatedAt (Between)",
			search: common.NewSearchByRange(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableDateRange{
					{Field: "CreatedAt", Min: &sampleData[2].CreatedAt, Max: &sampleData[2].UpdatedAt},
				},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[2:],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		// 7. Boolean Field Query - Filtering by Status
		{
			name: "Boolean Filter - Status (True)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "Status", Values: []interface{}{true}}},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[0:0],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},

		{
			name: "Empty Values Slice",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, groupID, userID),
				[]common.SearchableValue{{Field: "GameID", Values: []interface{}{}}},
				common.SearchResultOptions{Limit: 10},
				common.UserAudienceIDKey,
			),
			expectedResults: sampleData[0:0],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID, common.UserIDKey: userID},
		},
		{
			name: "Numeric Filter - Size (Greater Than)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, uuid.Nil, uuid.Nil),
				[]common.SearchableValue{
					{Field: "Size", Values: []interface{}{10}, Operator: common.GreaterThanOperator},
				},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[0:1],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		{
			name: "Numeric Filter - Size (Less Than)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, uuid.Nil, uuid.Nil),
				[]common.SearchableValue{
					{Field: "Size", Values: []interface{}{10000}, Operator: common.LessThanOperator},
				},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[1:],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		{
			name: "Numeric Filter - Size (Equals)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, uuid.Nil, uuid.Nil),
				[]common.SearchableValue{
					{Field: "Size", Values: []interface{}{1}, Operator: common.EqualsOperator},
				},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[1:],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},

		// 5. String Field - Filtering (All variations)
		{
			name: "String Filter - InternalURI (Contains)",
			search: common.NewSearchByValues(
				setContextWithValues(context.Background(), tenantID, clientID, uuid.Nil, uuid.Nil),
				[]common.SearchableValue{
					{Field: "InternalURI", Values: []interface{}{"/replays/fcad61ef-67fe-4405-9a4e-1b51774bb46a"}, Operator: common.ContainsOperator},
				},
				common.SearchResultOptions{Limit: 10},
				common.ClientApplicationAudienceIDKey,
			),
			expectedResults: sampleData[0:1],
			mockData:        sampleData,
			contextValues:   map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},

		// {
		// 	name: "Tenancy with OR Aggregation",
		// 	search: common.NewSearchByValues(
		// 		setContextWithValues(context.Background(), tenantID, clientID, uuid.Nil, uuid.Nil),
		// 		[]common.SearchableValue{
		// 			[]common.SearchableValue{
		// 				{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}},
		// 				{Field: "NetworkID", Values: []interface{}{common.FaceItNetworkIDKey}},
		// 			},
		// 		},
		// 		common.SearchResultOptions{Limit: 10},
		// 		common.ClientApplicationAudienceIDKey,
		// 	),
		// 	expectedResults: []replay_entity.ReplayFile{sampleData[0], sampleData[1]}, // Only User 1's games
		// 	mockData:        sampleData,
		// 	contextValues: map[interface{}]uuid.UUID{
		// 		common.TenantIDKey: tenantID,
		// 		common.ClientIDKey: clientID,
		// 		common.UserIDKey:   userID,
		// 	},
		// },
	}

	collection := client.Database(dbName).Collection(collectionName)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setContextWithValues(context.Background(), tt.contextValues[common.TenantIDKey], tt.contextValues[common.ClientIDKey], tt.contextValues[common.GroupIDKey], tt.contextValues[common.UserIDKey])

			data := make([]interface{}, len(tt.mockData))

			for i, d := range tt.mockData {
				data[i] = d
			}

			// clear
			collection.DeleteMany(ctx, bson.M{})

			rs, err := collection.InsertMany(ctx, data)
			if err != nil {
				t.Fatalf("Error inserting mock data: %v", err)
			}

			if len(rs.InsertedIDs) != len(tt.mockData) {
				t.Fatalf("Expected %d inserted documents, got %d", len(tt.mockData), len(rs.InsertedIDs))
			}

			// Query the database
			cursor, err := repo.Query(ctx, tt.search)
			if err != tt.expectedError {
				t.Fatalf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Check the results
			results := make([]replay_entity.ReplayFile, 0)
			for cursor.Next(ctx) {
				var result replay_entity.ReplayFile
				if err := cursor.Decode(&result); err != nil {
					t.Fatalf("Error decoding result: %v", err)
				}
				results = append(results, result)
			}

			insertedUUids := make([]interface{}, len(tt.mockData))
			for i, data := range tt.mockData {
				insertedUUids[i] = data.ID
			}

			deleteOnlyInserted := bson.M{"_id": bson.M{"$in": insertedUUids}}

			r, err := collection.DeleteMany(ctx, deleteOnlyInserted)
			if err != nil {
				t.Fatalf("Error deleting mock data: %v", err)
			}

			if r.DeletedCount != int64(len(tt.mockData)) {
				t.Fatalf("Expected %d deleted documents, got %d", len(tt.mockData), r.DeletedCount)
			}

			if len(results) != len(tt.expectedResults) {
				t.Fatalf("Expected %d results, got %d", len(tt.expectedResults), len(results))
			}

			for i, expected := range tt.expectedResults {
				if results[i].ID != expected.ID {
					t.Fatalf("Expected ID %v, got %v", expected.ID, results[i].ID)
				}
			}
		})
	}
}

func TestGetBSONFieldNameFromSearchableValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Establish a connection to a real MongoDB database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://127.0.0.1:37019/replay"))
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Use a dedicated collection for testing
	collectionName := "replay_files"
	repo := db.NewReplayFileMetadataRepository(client, dbName, replay_entity.ReplayFile{}, collectionName)

	repo.InitQueryableFields(map[string]bool{
		"ID":               true,
		"GameID":           true,
		"NetworkID":        true,
		"Size":             true,
		"InternalURI":      true,
		"Status":           true,
		"Error":            true,
		"Header":           true,
		"Header.Filestamp": true,
		"ResourceOwner":    true,
		"CreatedAt":        true,
		"UpdatedAt":        true,
	}, map[string]string{
		"ID":                     "_id",
		"GameID":                 "game_id",
		"NetworkID":              "network_id",
		"Size":                   "size",
		"InternalURI":            "uri",
		"Status":                 "status",
		"Error":                  "error",
		"Header":                 "header",
		"ResourceOwner":          "resource_owner",
		"CreatedAt":              "created_at",
		"UpdatedAt":              "updated_at",
		"Header.Filestamp":       "header.filestamp",
		"ResourceOwner.TenantID": "resource_owner.tenant_id",
		"ResourceOwner.UserID":   "resource_owner.user_id",
		"ResourceOwner.GroupID":  "resource_owner.group_id",
		"ResourceOwner.ClientID": "resource_owner.client_id",
	})

	testCases := []struct {
		name            string
		searchableValue common.SearchableValue
		expectedName    string
		expectedError   error
	}{
		{
			name:            "Valid Field",
			searchableValue: common.SearchableValue{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}},
			expectedName:    "game_id",
			expectedError:   nil,
		},
		{
			name:            "Valid Field with Header",
			searchableValue: common.SearchableValue{Field: "Header.Filestamp", Values: []interface{}{"HLTV-1.0.0"}},
			expectedName:    "header.filestamp",
			expectedError:   nil,
		},
		{
			name:            "Valid Field",
			searchableValue: common.SearchableValue{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}},
			expectedName:    "game_id",
			expectedError:   nil,
		},
		{
			name:            "Valid Nested Field",
			searchableValue: common.SearchableValue{Field: "Header.Filestamp", Values: []interface{}{"HLTV-1.0.0"}},
			expectedName:    "header.filestamp",
			expectedError:   nil,
		},
		{
			name:            "Multiple Values (Ignored)",
			searchableValue: common.SearchableValue{Field: "NetworkID", Values: []interface{}{"value1", "value2"}},
			expectedName:    "network_id",
			expectedError:   nil,
		},
		{
			name:            "Numeric Value",
			searchableValue: common.SearchableValue{Field: "Size", Values: []interface{}{12345}},
			expectedName:    "size",
			expectedError:   nil,
		},
		{
			name:            "Boolean Value",
			searchableValue: common.SearchableValue{Field: "Status", Values: []interface{}{true}},
			expectedName:    "status",
			expectedError:   nil,
		},
		{
			name:            "Nonexistent Field",
			searchableValue: common.SearchableValue{Field: "NonexistentField", Values: []interface{}{"value"}},
			expectedName:    "",
			expectedError:   fmt.Errorf("field NonexistentField not found or not queryable in Entity: ReplayFile (Collection: replay_files. Queryable Fields: map[CreatedAt:true Error:true GameID:true Header:true Header.Filestamp:true ID:true InternalURI:true NetworkID:true ResourceOwner:true Size:true Status:true UpdatedAt:true])"),
		},
		{
			name:            "Time/Date Value",
			searchableValue: common.SearchableValue{Field: "CreatedAt", Values: []interface{}{time.Now()}},
			expectedName:    "created_at",
			expectedError:   nil,
		},
		{
			name:            "Wildcard Query",
			searchableValue: common.SearchableValue{Field: "InternalURI.*", Values: []interface{}{"value"}},
			expectedName:    "uri",
			expectedError:   nil,
		},
		{
			name:            "Invalid Nested Field",
			searchableValue: common.SearchableValue{Field: "Invalid.Nested", Values: []interface{}{"value"}},
			expectedName:    "",
			expectedError:   fmt.Errorf("field Invalid.Nested not found or not queryable in Entity: ReplayFile (Collection: replay_files. Queryable Fields: map[CreatedAt:true Error:true GameID:true Header:true Header.Filestamp:true ID:true InternalURI:true NetworkID:true ResourceOwner:true Size:true Status:true UpdatedAt:true])"),
		},
		{
			name:            "Invalid Subfield",
			searchableValue: common.SearchableValue{Field: "GameID.Invalid", Values: []interface{}{"value"}},
			expectedName:    "",
			expectedError:   fmt.Errorf("field GameID.Invalid not found or not queryable in Entity: ReplayFile (Collection: replay_files. Queryable Fields: map[CreatedAt:true Error:true GameID:true Header:true Header.Filestamp:true ID:true InternalURI:true NetworkID:true ResourceOwner:true Size:true Status:true UpdatedAt:true])"),
		},
		{
			name:            "Empty Field Name",
			searchableValue: common.SearchableValue{Field: "", Values: []interface{}{"value"}},
			expectedName:    "",
			expectedError:   fmt.Errorf("empty field not allowed. cant query"),
		},
		{
			name:            "Empty Values Slice",
			searchableValue: common.SearchableValue{Field: "GameID", Values: []interface{}{}},
			expectedName:    "game_id",
			expectedError:   nil, // Should still return the BSON field name
		},
		{
			name:            "ResourceOwner Field",
			searchableValue: common.SearchableValue{Field: "ResourceOwner.TenantID", Values: []interface{}{uuid.New()}},
			expectedName:    "resource_owner.tenant_id", // Use the correct BSON name here
			expectedError:   nil,
		},
		{
			name:            "ResourceOwner Nested Field with Nil",
			searchableValue: common.SearchableValue{Field: "ResourceOwner.TenantID", Values: []interface{}{nil}},
			expectedName:    "resource_owner.tenant_id",
			expectedError:   fmt.Errorf("field ResourceOwner.TenantID not found or not queryable"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name, err := repo.GetBSONFieldNameFromSearchableValue(tc.searchableValue)
			if !errors.Is(err, tc.expectedError) && err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
			if name != tc.expectedName {
				t.Errorf("Expected name: %s, got: %s", tc.expectedName, name)
			}
		})
	}
}

func TestMongoDBRepository_EnsureTenancy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Establish a connection to a real MongoDB database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://127.0.0.1:37019/replay"))
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Use a dedicated collection for testing
	collectionName := "replay_files"
	repo := db.NewReplayFileMetadataRepository(client, dbName, replay_entity.ReplayFile{}, collectionName)

	setContextWithValues := func(ctx context.Context, tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, clientID)
		ctx = context.WithValue(ctx, common.UserIDKey, userID)
		ctx = context.WithValue(ctx, common.GroupIDKey, groupID)
		return ctx
	}

	tenantID := uuid.New()
	clientID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	testCases := []struct {
		name              string
		agg               bson.M
		search            common.Search
		expectedAgg       bson.M
		expectedError     error
		expectedErrorPart string
		contextValues     map[interface{}]uuid.UUID
		maxRecursiveDepth int
	}{
		{
			name:          "Success - ClientApplicationAudienceIDKey",
			agg:           bson.M{},
			search:        common.Search{VisibilityOptions: common.SearchVisibilityOptions{IntendedAudience: common.ClientApplicationAudienceIDKey, RequestSource: common.ResourceOwner{TenantID: tenantID, ClientID: clientID}}},
			expectedAgg:   bson.M{"resource_owner.tenant_id": tenantID, "resource_owner.client_id": clientID},
			expectedError: nil,
			contextValues: map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		{
			name:          "Success - GroupAudienceIDKey",
			agg:           bson.M{},
			search:        common.Search{VisibilityOptions: common.SearchVisibilityOptions{IntendedAudience: common.GroupAudienceIDKey, RequestSource: common.ResourceOwner{TenantID: tenantID, GroupID: groupID}}},
			expectedAgg:   bson.M{"resource_owner.tenant_id": tenantID, "resource_owner.group_id": groupID},
			expectedError: nil,
			contextValues: map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.GroupIDKey: groupID},
		},
		{
			name:          "Success - UserAudienceIDKey",
			agg:           bson.M{},
			search:        common.Search{VisibilityOptions: common.SearchVisibilityOptions{IntendedAudience: common.UserAudienceIDKey, RequestSource: common.ResourceOwner{TenantID: tenantID, UserID: userID}}},
			expectedAgg:   bson.M{"resource_owner.tenant_id": tenantID, "resource_owner.user_id": userID},
			expectedError: nil,
			contextValues: map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.UserIDKey: userID},
		},
		{
			name:              "Error - Empty TenantID in Search",
			agg:               bson.M{},
			search:            common.Search{VisibilityOptions: common.SearchVisibilityOptions{RequestSource: common.ResourceOwner{}}},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.RequestSource: valid tenant_id is required in queryCtx",
			contextValues:     map[interface{}]uuid.UUID{},
		},
		{
			name:              "Error - Empty ClientID in Search",
			agg:               bson.M{},
			search:            common.Search{VisibilityOptions: common.SearchVisibilityOptions{IntendedAudience: common.ClientApplicationAudienceIDKey, RequestSource: common.ResourceOwner{TenantID: tenantID}}},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.ApplicationLevel: valid client_id is required in queryCtx",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: tenantID},
		},
		{
			name:              "Error - No Audience Provided",
			agg:               bson.M{},
			search:            common.Search{VisibilityOptions: common.SearchVisibilityOptions{RequestSource: common.ResourceOwner{TenantID: tenantID, ClientID: clientID}}},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.Unknown: intended audience",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.ClientIDKey: clientID},
		},
		{
			name: "Success - EnsureUserAndGroupIDTenancy",
			agg:  bson.M{},
			search: common.Search{
				VisibilityOptions: common.SearchVisibilityOptions{
					IntendedAudience: common.UserAudienceIDKey,
					RequestSource:    common.ResourceOwner{TenantID: tenantID, UserID: userID, GroupID: groupID},
				},
			},
			expectedAgg:   bson.M{"resource_owner.tenant_id": tenantID, "$or": bson.A{bson.M{"resource_owner.group_id": groupID}, bson.M{"resource_owner.user_id": userID}}},
			expectedError: nil,
			contextValues: map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.UserIDKey: userID, common.GroupIDKey: groupID},
		},
		{
			name: "Error - EnsureUserAndGroupIDTenancy with Empty UserID",
			agg:  bson.M{},
			search: common.Search{
				VisibilityOptions: common.SearchVisibilityOptions{
					IntendedAudience: common.UserAudienceIDKey,
					RequestSource:    common.ResourceOwner{TenantID: tenantID, GroupID: groupID},
				},
			},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.UserLevel: user_id is required in search parameters for intended audience:",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.GroupIDKey: groupID},
		},
		{
			name: "Error - EnsureUserAndGroupIDTenancy with Empty GroupID",
			agg:  bson.M{},
			search: common.Search{
				VisibilityOptions: common.SearchVisibilityOptions{
					IntendedAudience: common.GroupAudienceIDKey,
					RequestSource:    common.ResourceOwner{TenantID: tenantID, UserID: userID},
				},
			},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.GroupLevel: group_id is required in search parameters for intended audience:",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: tenantID, common.UserIDKey: userID},
		},
		{
			name: "Error - EnsureUserAndGroupIDTenancy with Mismatched TenantID",
			agg:  bson.M{},
			search: common.Search{
				VisibilityOptions: common.SearchVisibilityOptions{
					IntendedAudience: common.UserAudienceIDKey,
					RequestSource:    common.ResourceOwner{TenantID: uuid.New(), UserID: userID},
				},
			},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.RequestSource: `tenant_id` in queryCtx does not match `tenant_id` in `common.Search`",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: uuid.New(), common.UserIDKey: userID},
		},
		{
			name: "Error - EnsureUserAndGroupIDTenancy with Nil TenantID",
			agg:  bson.M{},
			search: common.Search{
				VisibilityOptions: common.SearchVisibilityOptions{
					IntendedAudience: common.UserAudienceIDKey,
					RequestSource:    common.ResourceOwner{TenantID: uuid.Nil, UserID: userID},
				},
			},
			expectedAgg:       bson.M{},
			expectedErrorPart: "TENANCY.RequestSource: valid tenant_id is required in queryCtx",
			contextValues:     map[interface{}]uuid.UUID{common.TenantIDKey: uuid.Nil, common.UserIDKey: userID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set context
			ctx := setContextWithValues(context.Background(), tc.contextValues[common.TenantIDKey], tc.contextValues[common.ClientIDKey], tc.contextValues[common.GroupIDKey], tc.contextValues[common.UserIDKey])
			result, err := repo.EnsureTenancy(ctx, tc.agg, tc.search)

			if tc.expectedError != nil {
				assert.Error(t, err, tc.name)
				assert.EqualError(t, err, tc.expectedError.Error(), tc.name)
			} else if tc.expectedErrorPart != "" {
				if err == nil {
					assert.Fail(t, "expectedErrorPart is set but error is nil")
				}
				assert.Contains(t, err.Error(), tc.expectedErrorPart, tc.name)
			} else if tc.expectedAgg != nil {
				assert.NoError(t, err, tc.name)
				assert.Equal(t, tc.expectedAgg, result, tc.name)
			} else {
				assert.Fail(t, "expectedError, expectedErrorPart expectedAgg must be set")
			}

		})
	}
}
