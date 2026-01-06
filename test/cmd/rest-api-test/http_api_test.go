//go:build integration

package rest_api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"encoding/json"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/replay-api/replay-api/cmd/rest-api/controllers"
	"github.com/replay-api/replay-api/cmd/rest-api/routing"
	ioc "github.com/replay-api/replay-api/pkg/infra/ioc"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	google_out "github.com/replay-api/replay-api/pkg/domain/google/ports/out"
	iam_dtos "github.com/replay-api/replay-api/pkg/domain/iam/dtos"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	steam_out "github.com/replay-api/replay-api/pkg/domain/steam/ports/out"
)

type Tester struct {
	Container      container.Container
	RequestHandler http.Handler
}

func NewTester() *Tester {
	// Skip if MongoDB is not available
	if os.Getenv("MONGO_URI") == "" || os.Getenv("MONGO_URI") == "mongodb://127.0.0.1:37019/replay" {
		// Try to connect to MongoDB to see if it's available
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://127.0.0.1:37019").SetServerSelectionTimeout(2*time.Second))
		if err != nil || client.Ping(context.TODO(), nil) != nil {
			panic("MongoDB not available, skipping integration test")
		}
		if client != nil {
			client.Disconnect(context.TODO())
		}
	}

	os.Setenv("DEV_ENV", "test")
	os.Setenv("MONGO_URI", "mongodb://127.0.0.1:37019/replay")
	os.Setenv("MONGO_DB_NAME", "replay")
	os.Setenv("STEAM_VHASH_SOURCE", "82DA0F0D0135FEA0F5DDF6F96528B48A")

	b := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts().WithSquadAPI()
	c := b.Build()
	return &Tester{
		Container:      c,
		RequestHandler: routing.NewRouter(context.Background(), c),
	}
}

func (t *Tester) Exec(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()

	// if len(req.Header.Get(string(shared.ResourceOwnerIDParamKey))) > 0 {

	// }

	t.RequestHandler.ServeHTTP(rec, req)

	return rec

	// authReq := http.NewRequest()

}

func expectStatus(t *testing.T, expected int, r *httptest.ResponseRecorder) {
	if expected != r.Code {
		t.Errorf("Expected response code %d. Got %d\n Body=%v", expected, r.Code, r.Body)
	}
}

func expectUUIDHeader(t *testing.T, key string, r *httptest.ResponseRecorder) {
	if len(r.Header().Get(key)) == 0 {
		t.Errorf("Expected response header %s to be a valid UUID. Got %s", key, r.Header().Get(key))
	}

	uuidRegEx := `^[\w\d]{8}-[\w\d]{4}-[\w\d]{4}-[\w\d]{4}-[\w\d]{12}$`

	if !regexp.MustCompile(uuidRegEx).MatchString(r.Header().Get(key)) {
		t.Errorf("Expected response header %s to be a UUID. Got %s", key, r.Header().Get(key))
	}

}

func expectHeader(t *testing.T, key string, r *httptest.ResponseRecorder, aud shared.IntendedAudienceKey) {
	if len(r.Header().Get(key)) == 0 {
		t.Errorf("Expected response header %s to be a valid UUID. Got %s", key, r.Header().Get(key))
	}

	if r.Header().Get(key) != string(aud) {
		t.Errorf("Expected response header %s to be a UUID. Got %s", key, r.Header().Get(key))
	}
}

// func Test_GetEvents(t *testing.T) {
// 	tester := NewTester()

// 	req, _ := http.NewRequest("GET", routing.GameEvents, nil)

// 	response := tester.Exec(req)

// 	t.Logf("response: %v", response)

// 	expectStatus(t, http.StatusOK, response)
// }

func expectStatuses(t *testing.T, expected []int, r *httptest.ResponseRecorder) {
	for _, status := range expected {
		if status == r.Code {
			return
		}
	}
	t.Errorf("Expected one of response codes %v. Got %d\n Body=%v", expected, r.Code, r.Body)
}

func Test_Search_Player_SuccessEmpty(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "players/123", 1), nil)
	// req.Header.Add("")

	res := tester.Exec((req))

	t.Logf("response: %v", res)

	expectStatuses(t, []int{
		http.StatusNoContent, http.StatusOK,
	}, res)
}

func Test_SteamOnboarding_BadRequest(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("POST", routing.OnboardSteam, nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusBadRequest, response)
}

func Test_GameEventSearch_SuccessEmpty(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "GameEvents", 1), nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatuses(t, []int{
		http.StatusNoContent, http.StatusOK,
	}, response)
}

func Test_SearchSteamUserRealName_Success(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "profiles", -1), nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatuses(t, []int{
		http.StatusNoContent, http.StatusOK,
	}, response)
}

func Test_SteamOnboarding_Success(t *testing.T) {
	tester := NewTester()

	var vHashWriter steam_out.VHashWriter
	err := tester.Container.Resolve(&vHashWriter)

	if err != nil {
		t.Errorf("fail to resolve steam_out.VHashWriter %v", err)
	}

	mockSteamID := "12345"

	vhash := vHashWriter.CreateVHash(context.Background(), mockSteamID)

	steamUser := fmt.Sprintf(`{"steam": {"id": "%s"}, "v_hash": "%s"}`, mockSteamID, vhash)
	reader := strings.NewReader(steamUser)

	req, _ := http.NewRequest("POST", routing.OnboardSteam, reader)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusCreated, response)
	expectUUIDHeader(t, controllers.ResourceOwnerIDHeaderKey, response)
	expectHeader(t, controllers.ResourceOwnerAudTypeHeaderKey, response, shared.UserAudienceIDKey)

	// should query the profile

	rid := response.Header().Get(controllers.ResourceOwnerIDHeaderKey)
	req = httptest.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "profiles?Type=steam&Details.ID=12345", -1), nil)
	req.Header.Add(controllers.ResourceOwnerIDHeaderKey, rid)

	response = tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusOK, response)

	// should query groups by listing memberships from search controller

	req = httptest.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "memberships", -1), nil)
	req.Header.Add(controllers.ResourceOwnerIDHeaderKey, rid)
	response = tester.Exec(req)
	t.Logf("response: %v", response)

	expectStatus(t, http.StatusOK, response)

	// should list groups details & memberships directly to /groups
	req = httptest.NewRequest("GET", routing.Group, nil)
	req.Header.Add(controllers.ResourceOwnerIDHeaderKey, rid)

	response = tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusOK, response)

	var groupMembershipDTOs map[uuid.UUID]iam_dtos.GroupMembershipDTO

	if response.Body == nil {
		t.Errorf("response body of GET /groups is nil")
	}

	err = json.NewDecoder(response.Body).Decode(&groupMembershipDTOs)

	if err != nil {
		t.Errorf("failed to decode response body %v", err)
	}

	if len(groupMembershipDTOs) == 0 {
		t.Errorf("no groups found at GET /groups")
	}

	// validate private group content

	var defaultPrivateGroup *iam_entities.Group

	for _, groupMembershipDTO := range groupMembershipDTOs {
		if groupMembershipDTO.Group.Name == iam_entities.DefaultUserGroupName {
			defaultPrivateGroup = &groupMembershipDTO.Group
		}
	}

	if defaultPrivateGroup == nil {
		t.Errorf("no private group found")
	} else {
		if defaultPrivateGroup.Type != iam_entities.GroupTypeSystem {
			t.Errorf("private group type is not user")
		}
	}

}

func Test_GoogleOnboarding_Success(t *testing.T) {
	tester := NewTester()

	var vHashWriter google_out.VHashWriter
	err := tester.Container.Resolve(&vHashWriter)

	if err != nil {
		t.Errorf("fail to resolve google_out.VHashWriter %v", err)
	}

	mockGoogleEmail := "testiwewkksdj@gmail.com"

	vhash := vHashWriter.CreateVHash(context.Background(), mockGoogleEmail)

	googleUser := fmt.Sprintf(`{"email": "%s", "v_hash": "%s"}`, mockGoogleEmail, vhash)
	reader := strings.NewReader(googleUser)

	req, _ := http.NewRequest("POST", routing.OnboardGoogle, reader)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusCreated, response)
	expectUUIDHeader(t, controllers.ResourceOwnerIDHeaderKey, response)
}

func LoadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var fileContent []byte
	_, err = file.Read(fileContent)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

// func Test_UploadReplayFile(t *testing.T) {
// 	tester := NewTester()

// 	req, _ := http.NewRequest("POST", routing.Replay, nil)

// 	filePath := "../../../test/sample_replays/cs2/sound.dem"

// 	fileBytes, err := LoadFile(filePath)
// 	if err != nil {
// 		panic(err)
// 		t.Fatalf("Failed to open demo file: %v", err)
// 	}

// 	expectStatus(t, http.StatusOK, response.Code)
// }

// curl --form 'file=@"/Users/psavelis-adm/Desktop/go/src/github.com/replay-api/replay-api/test/sample_replays/cs2/sound.dem"' "http://127.0.0.1:4991/games/cs2/replays"
// curl --form 'file=@"C:\sources\replay-api\replay-api\test\sample_replays\cs2\sound.demo"' "http://127.0.0.1:4991/games/cs2/replays"
// func Test_UploadReplayFile(t *testing.T) {
// 	tester := NewTester()

// 	filePath := "../../../test/sample_replays/cs2/sound.dem"
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		t.Fatalf("Failed to open demo file: %v", err)
// 	}
// 	defer file.Close()

// 	body := &strings.Builder{}
// 	writer := multipart.NewWriter(body)
// 	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
// 	if err != nil {
// 		t.Fatalf("Failed to create form file: %v", err)
// 	}
// 	_, err = io.Copy(part, file)
// 	if err != nil {
// 		t.Fatalf("Failed to copy file content: %v", err)
// 	}
// 	writer.Close()

// 	req, _ := http.NewRequest("POST", "http://127.0.0.1:4991/games/cs2/replays", strings.NewReader(body.String()))
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	response := tester.Exec(req)

// 	t.Logf("response: %v", response)

// 	expectStatus(t, http.StatusOK, response)
// }
