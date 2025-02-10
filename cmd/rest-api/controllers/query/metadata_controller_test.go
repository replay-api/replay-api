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

	"github.com/google/uuid"
	query_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/query"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/routing"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	"github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"
)

func GetDefaultTestContext(reqContext context.Context, tenantID, clientID, groupID, userID uuid.UUID) context.Context {
	reqContext = context.WithValue(reqContext, common.TenantIDKey, tenantID)
	reqContext = context.WithValue(reqContext, common.ClientIDKey, clientID)
	reqContext = context.WithValue(reqContext, common.GroupIDKey, groupID)
	reqContext = context.WithValue(reqContext, common.UserIDKey, userID)
	return reqContext
}

func TestReplaySearchHandler(t *testing.T) {
	controller := setup()

	req, err := http.NewRequest(http.MethodGet, routing.Match, nil)
	if err != nil {
		t.Fatal(err)
	}

	setContextWithValues := func(tenantID, clientID, groupID, userID uuid.UUID) context.Context {
		newCtx := context.TODO()
		newCtx = context.WithValue(newCtx, common.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, common.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, common.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, common.UserIDKey, userID)
		return newCtx
	}

	tenantID := uuid.New()
	clientID := uuid.New()
	groupID := uuid.New()
	userID := uuid.New()

	s := common.NewSearchByValues(
		setContextWithValues(tenantID, clientID, groupID, userID),
		[]common.SearchableValue{{Field: "GameID", Values: []interface{}{common.CS2_GAME_ID}}},
		common.SearchResultOptions{Limit: 10},
		common.UserAudienceIDKey,
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
		newCtx = context.WithValue(newCtx, common.TenantIDKey, tenantID)
		newCtx = context.WithValue(newCtx, common.ClientIDKey, clientID)
		newCtx = context.WithValue(newCtx, common.GroupIDKey, groupID)
		newCtx = context.WithValue(newCtx, common.UserIDKey, userID)
		return newCtx
	}

	tenantID := uuid.New()
	clientID := uuid.New()
	groupID := uuid.New()
	userID := uuid.New()

	s := common.NewSearchByValues(
		setContextWithValues(tenantID, clientID, groupID, userID),
		[]common.SearchableValue{},
		common.SearchResultOptions{Limit: 10},
		common.UserAudienceIDKey,
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

	c := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().Build()

	controller := query_controllers.NewReplayMetadataQueryController(c)
	return controller
}
