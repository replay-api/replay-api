package rest_api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/psavelis/team-pro/replay-api/cmd/rest-api/routing"
	ioc "github.com/psavelis/team-pro/replay-api/pkg/infra/ioc"

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

// func Test_GetEvents(t *testing.T) {
// 	tester := NewTester()

// 	req, _ := http.NewRequest("GET", routing.GameEvents, nil)

// 	response := tester.Exec(req)

// 	t.Logf("response: %v", response)

// 	expectStatus(t, http.StatusOK, response)
// }

func Test_Search_Player_Success(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "players/123", 1), nil)
	// req.Header.Add("")

	res := tester.Exec((req))

	t.Logf("response: %v", res)

	expectStatus(t, http.StatusOK, res)
}

func Test_SteamOnboarding_BadRequest(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("POST", routing.OnboardSteam, nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusBadRequest, response)
}

func Test_GameEventSearch_Success(t *testing.T) {
	tester := NewTester()

	req, _ := http.NewRequest("GET", strings.Replace(routing.Search, "{query:.*}", "GameEvents", 1), nil)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusOK, response)
}

func Test_SteamOnboarding_Success(t *testing.T) {
	tester := NewTester()

	var vHashWriter steam_out.VHashWriter
	err := tester.Container.Resolve(&vHashWriter)

	if err != nil {
		t.Errorf("fail to resolve steam_out.VHashWriter %v", err)
	}

	vhash := vHashWriter.CreateVHash(context.Background(), "1")

	steamUser := fmt.Sprintf(`{"steam": {"id": "1"}, "v_hash": "%s"}`, vhash)
	reader := strings.NewReader(steamUser)

	req, _ := http.NewRequest("POST", routing.OnboardSteam, reader)

	response := tester.Exec(req)

	t.Logf("response: %v", response)

	expectStatus(t, http.StatusCreated, response)
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
