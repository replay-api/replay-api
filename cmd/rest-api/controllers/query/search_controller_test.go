package query_controllers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	query_controllers "github.com/psavelis/team-pro/replay-api/cmd/rest-api/controllers/query"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type RoutingTestCase struct {
	Path             string
	Name             string
	ExpectedResource string
}

type ParamsTestCase struct {
	Path           string
	Name           string
	ExpectedSearch common.Search
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
	}

	types := []common.ResourceType{
		common.ResourceTypeBadge,
		common.ResourceTypeChannel,
		common.ResourceTypeGame,
		common.ResourceTypeGameEvent,
		common.ResourceTypeGroup,
		common.ResourceTypeLeague,
		common.ResourceTypeMatch,
		common.ResourceTypeRound,
		common.ResourceTypePlayer,
		common.ResourceTypeReplayFile,
		common.ResourceTypeTeam,
		common.ResourceTypeTournament,
		common.ResourceTypeUser,
	}

	for _, tc := range tcs {
		res := query_controllers.GetResourceStringFromPath(types, tc.Path)

		if res != tc.ExpectedResource {
			t.Errorf("Test Case %s failed: expected %s, but received %s. (Path: %s)", tc.Name, tc.ExpectedResource, res, tc.Path)
		}

		ctx := context.WithValue(context.TODO(), common.TenantIDKey, common.TeamPROTenantID)
		ctx = context.WithValue(ctx, common.ClientIDKey, common.TeamPROAppClientID)
		ctx = context.WithValue(ctx, common.GroupIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

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
