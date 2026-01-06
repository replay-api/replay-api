package query_controllers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	query_controllers "github.com/replay-api/replay-api/cmd/rest-api/controllers/query"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type RoutingTestCase struct {
	Path             string
	Name             string
	ExpectedResource string
}

type ParamsTestCase struct {
	Path           string
	Name           string
	ExpectedSearch shared.Search
}

var basePath = "http://localhost:4991"

func TestVectorGetResourceStringFromPath(t *testing.T) {
	tcs := []RoutingTestCase{
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches", basePath), // http://localhost:4991/search/players
			Name:             "Valid_Match_Root",
			ExpectedResource: "matches",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1", basePath),
			Name:             "Valid_Match_Leaf",
			ExpectedResource: "matches",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds", basePath),
			Name:             "Valid_Round_Root",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds/1", basePath),
			Name:             "Valid_Round_Leaf",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds/1?tag=test", basePath),
			Name:             "Valid_Round_Leaf_With_QueryStrings",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds/1?tag=test&t=1", basePath),
			Name:             "Valid_Round_Leaf_With_More_QueryStrings",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds?tag=test&t=1", basePath),
			Name:             "Valid_Round_Root_With_QueryStrings",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/games/cs2/matches/1/rounds/?tag=test&t=1", basePath),
			Name:             "Valid_Round_Root_With_QueryStrings_And_Ending_Slash",
			ExpectedResource: "rounds",
		},
		{
			Path:             fmt.Sprintf("%s/search/users?Steam.RealName=test&t=1", basePath),
			Name:             "Valid_SteamUser_Root_With_QueryStrings",
			ExpectedResource: "users",
		},
		{
			Path:             fmt.Sprintf("%s/search/users/?Steam.RealName=test&t=1", basePath),
			Name:             "Valid_SteamUser_Root_With_QueryStrings_And_Ending_Slash",
			ExpectedResource: "users",
		},
		{
			Path:             fmt.Sprintf("%s/search/users/1", basePath),
			Name:             "Valid_SteamUser_Leaf",
			ExpectedResource: "users",
		},
		{
			Path:             fmt.Sprintf("%s/search/users/1?Steam.RealName=test&t=1", basePath),
			Name:             "Valid_SteamUser_Leaf_With_QueryStrings",
			ExpectedResource: "users",
		},
		{
			Path:             fmt.Sprintf("%s/search/users/1/?Steam.RealName=test&t=1", basePath),
			Name:             "Valid_SteamUser_Leaf_With_QueryStrings_And_Ending_Slash",
			ExpectedResource: "users",
		},
		{
			Path:             fmt.Sprintf("%s/search/profiles?RIDSource=steam&Details.realname=Test&filter=out", basePath),
			Name:             "Valid_Profile_Root_With_QueryStrings",
			ExpectedResource: "profiles",
		},
	}

	types := []shared.ResourceType{
		replay_common.ResourceTypeBadge,
		replay_common.ResourceTypeChannel,
		replay_common.ResourceTypeGame,
		replay_common.ResourceTypeGameEvent,
		shared.ResourceTypeGroup,
		replay_common.ResourceTypeLeague,
		replay_common.ResourceTypeMatch,
		replay_common.ResourceTypeRound,
		replay_common.ResourceTypePlayerMetadata,
		replay_common.ResourceTypePlayerProfile,
		replay_common.ResourceTypeReplayFile,
		replay_common.ResourceTypeTeam,
		replay_common.ResourceTypeTournament,
		shared.ResourceTypeUser,
		replay_common.ResourceTypeProfile,
	}

	for _, tc := range tcs {
		res := query_controllers.GetResourceStringFromPath(types, tc.Path)

		if res != tc.ExpectedResource {
			t.Errorf("Test Case %s failed: expected %s, but received %s. (Path: %s)", tc.Name, tc.ExpectedResource, res, tc.Path)
		}

		ctx := context.WithValue(context.TODO(), shared.TenantIDKey, replay_common.TeamPROTenantID)
		ctx = context.WithValue(ctx, shared.ClientIDKey, replay_common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, shared.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, shared.UserIDKey, uuid.New())

		req, _ := http.NewRequestWithContext(ctx, "GET", tc.Path, &io.PipeReader{})

		s, err := query_controllers.GetSearchParams(req)

		if err != nil {
			t.Errorf("Test Case %s failed: %v", tc.Name, err)
		}

		if s == nil || len(s.SearchParams) == 0 {
			t.Errorf("Test Case %s has no search.SearchParams: %v", tc.Name, err)
		}

		t.Logf("√ Passed with tc.search: %v", s)
	}
}
