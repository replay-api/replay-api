//go:build integration

package query_controllers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	query_controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers/query"
	"github.com/replay-api/replay-api/cmd/rest-api/routing"
	"github.com/replay-api/replay-api/pkg/infra/ioc"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDefaultTestContext(reqContext context.Context, tenantID, clientID, groupID, userID uuid.UUID) context.Context {
	reqContext = context.WithValue(reqContext, shared.TenantIDKey, tenantID)
	reqContext = context.WithValue(reqContext, shared.ClientIDKey, clientID)
	reqContext = context.WithValue(reqContext, shared.GroupIDKey, groupID)
	reqContext = context.WithValue(reqContext, shared.UserIDKey, userID)
	return reqContext
}

func TestReplaySearchHandler(t *testing.T) {
	// Skip test if MongoDB is not available
	if os.Getenv("MONGO_URI") == "" || os.Getenv("MONGO_URI") == "mongodb://host.docker.internal:37019" {
		// Try to connect to MongoDB to see if it's available
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://host.docker.internal:37019").SetServerSelectionTimeout(2*time.Second))
		if err != nil || client.Ping(context.TODO(), nil) != nil {
			t.Skip("MongoDB not available, skipping integration test")
		}
		if client != nil {
			client.Disconnect(context.TODO())
		}
	}

	controller := setup()

	req, err := http.NewRequest(http.MethodGet, routing.Match, nil)
	if err != nil {
		t.Fatal(err)
	}

	setContextWithValues := func(tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		newCtx := context.TODO()
		newCtx = context.WithValue(newCtx, shared.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, shared.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, shared.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, shared.UserIDKey, userID)
		return newCtx
	}

	tenantID := uuid.New()
	clientID := uuid.New()
	groupID := uuid.New()
	userID := uuid.New()

	s := shared.NewSearchByValues(
		setContextWithValues(tenantID, clientID, groupID, userID),
		[]shared.SearchableValue{{Field: "GameID", Values: []interface{}{replay_common.CS2_GAME_ID}}},
		shared.SearchResultOptions{Limit: 10},
		shared.UserAudienceIDKey,
	)

	var searchJSON []byte
	searchJSON, err = json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	slog.Info("searchJSON", "searchJSON", searchJSON)

	base64Encoded := base64.StdEncoding.EncodeToString(searchJSON)
	req = req.WithContext(GetDefaultTestContext(req.Context(), tenantID, clientID, groupID, userID))

	req.Header.Set("x-search", base64Encoded)

	recorder := httptest.NewRecorder()

	controller.DefaultSearchHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d. Response: %v", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	expectedContentType := "application/json"
	actualContentType := recorder.Header().Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Errorf("expected content type %s, got %s", expectedContentType, actualContentType)
	}
}
func TestPlayerSearchHandler(t *testing.T) {
	controller := setup()

	req, err := http.NewRequest(http.MethodGet, "/search/"+routing.PlayerProfile, nil)
	if err != nil {
		t.Fatal(err)
	}

	setContextWithValues := func(tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		newCtx := context.TODO()
		newCtx = context.WithValue(newCtx, shared.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, shared.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, shared.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, shared.UserIDKey, userID)
		return newCtx
	}

	tenantID := uuid.New()
	clientID := uuid.New()
	groupID := uuid.New()
	userID := uuid.New()

	s := shared.NewSearchByValues(
		setContextWithValues(tenantID, clientID, groupID, userID),
		[]shared.SearchableValue{},
		shared.SearchResultOptions{Limit: 10},
		shared.UserAudienceIDKey,
	)

	var searchJSON []byte
	searchJSON, err = json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	slog.Info("searchJSON", "searchJSON", searchJSON)

	base64Encoded := base64.StdEncoding.EncodeToString(searchJSON)
	req = req.WithContext(GetDefaultTestContext(req.Context(), tenantID, clientID, groupID, userID))

	req.Header.Set("x-search", base64Encoded)

	recorder := httptest.NewRecorder()

	controller.DefaultSearchHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d. Response: %v", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	expectedContentType := "application/json"
	actualContentType := recorder.Header().Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Errorf("expected content type %s, got %s", expectedContentType, actualContentType)
	}
}
func setup() *query_controllers.ReplayMetadataQueryController {
	os.Setenv("DEV_ENV", "test")
	os.Setenv("MONGO_URI", "mongodb://host.docker.internal:37019")
	os.Setenv("MONGO_DB_NAME", "replay")
	os.Setenv("STEAM_VHASH_SOURCE", "82DA0F0D0135FEA0F5DDF6F96528B48A")

	c := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().WithSquadAPI().Build()

	controller := query_controllers.NewReplayMetadataQueryController(c)
	return controller
}
