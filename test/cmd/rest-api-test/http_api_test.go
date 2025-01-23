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

	"github.com/golobby/container/v3"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/routing"
	ioc "github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"

	google_out "github.com/psavelis/team-pro/replay-api/pkg/domain/google/ports/out"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"
)

type Tester struct {
	Container      container.Container
	RequestHandler http.Handler
}

func NewTester() *Tester {
	os.Setenv("DEV_ENV", "test")
	os.Setenv("MONGO_URI", "mongodb://127.0.0.1:37019/replay")
	os.Setenv("MONGO_DB_NAME", "replay")
	os.Setenv("STEAM_VHASH_SOURCE", "82DA0F0D0135FEA0F5DDF6F96528B48A")

	b := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts()
	c := b.Build()
	return &Tester{
		Container:      c,
		RequestHandler: routing.NewRouter(context.Background(), c),
	}
}

func (t *Tester) Exec(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()

	// if len(req.Header.Get(string(common.ResourceOwnerIDParamKey))) > 0 {

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

// func Test_GetEvents(t *testing.T) {
// 	tester := NewTester()

// 	req, _ := http.NewRequest("GET", routing.GameEvents, nil)

// 	response := tester.Exec(req)

// 	t.Logf("response: %v", response)

// 	expectStatus(t, http.StatusOK, response)
// }

func Test_Search_Player_SuccessEmpty(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "players/123", 1), nil)
	// req.Header.Add("")

	res := tester.Exec((req))

	t.Logf("response: %v", res)

	expectStatus(t, http.StatusNoContent, res)
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

	expectStatus(t, http.StatusNoContent, response)
}

func Test_SearchSteamUserRealName_Success(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "profiles", -1), nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusNoContent, response)
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
	expectUUIDHeader(t, "X-Resource-Owner-ID", response)

	// should query the profile
	req = httptest.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "profiles?Type=steam&Details.ID=12345", -1), nil)
	req.Header.Add("X-Resource-Owner-ID", response.Header().Get("X-Resource-Owner-ID"))

	response = tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusOK, response)

	// should query the steam user

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
	expectUUIDHeader(t, "X-Resource-Owner-ID", response)
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

// curl --form 'file=@"/Users/psavelis-adm/Desktop/go/src/github.com/psavelis/team-pro/replay-api/test/sample_replays/cs2/sound.dem"' "http://127.0.0.1:4991/games/cs2/replays"
