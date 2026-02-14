package scores_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	scores_vo "github.com/replay-api/replay-api/pkg/domain/scores/value-objects"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResourceOwner() shared.ResourceOwner {
	return shared.ResourceOwner{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
	}
}

func testTeamResults() []TeamResult {
	return []TeamResult{
		{TeamID: uuid.New(), TeamName: "Team Alpha", Score: 16, Players: []uuid.UUID{uuid.New(), uuid.New()}},
		{TeamID: uuid.New(), TeamName: "Team Beta", Score: 12, Players: []uuid.UUID{uuid.New(), uuid.New()}},
	}
}

func testPlayerResults(teams []TeamResult) []PlayerResult {
	var results []PlayerResult
	for _, team := range teams {
		for _, pid := range team.Players {
			results = append(results, PlayerResult{
				PlayerID: pid,
				TeamID:   team.TeamID,
				Score:    100,
				Kills:    20,
				Deaths:   15,
				Assists:  5,
				Rating:   1.15,
			})
		}
	}
	if len(results) > 0 {
		results[0].IsMVP = true
	}
	return results
}

func TestNewMatchResult(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scores_vo.ResultStatusSubmitted, result.Status)
	assert.NotNil(t, result.WinnerTeamID, "should determine winner")
	assert.False(t, result.IsDraw)
	assert.Equal(t, 1, result.TeamResults[0].Position, "higher score team should be position 1")
	assert.Equal(t, 2, result.TeamResults[1].Position)
}

func TestNewMatchResult_Draw(t *testing.T) {
	ro := testResourceOwner()
	teams := []TeamResult{
		{TeamID: uuid.New(), TeamName: "Team A", Score: 15, Players: []uuid.UUID{uuid.New()}},
		{TeamID: uuid.New(), TeamName: "Team B", Score: 15, Players: []uuid.UUID{uuid.New()}},
	}
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceReplayFile, ro.UserID,
		teams, players, time.Now(), 40*time.Minute,
	)

	require.NoError(t, err)
	assert.True(t, result.IsDraw)
	assert.Nil(t, result.WinnerTeamID)
	for _, tr := range result.TeamResults {
		assert.Equal(t, 1, tr.Position, "draw should give all teams position 1")
	}
}

func TestNewMatchResultFromReplay(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)
	replayID := uuid.New()

	result, err := NewMatchResultFromReplay(
		ro, uuid.New(), replayID, "cs2", "de_dust2", "competitive",
		teams, players, time.Now(), 50*time.Minute,
	)

	require.NoError(t, err)
	assert.Equal(t, scores_vo.ScoreSourceReplayFile, result.Source)
	assert.Equal(t, &replayID, result.SourceReplayID)
}

func TestStateTransitions(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusSubmitted, result.Status)

	// Submit → Under Review
	err = result.Review()
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusUnderReview, result.Status)

	// Under Review → Verified
	verifier := uuid.New()
	err = result.Verify(scores_vo.VerificationMethodManual, &verifier)
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)
	assert.NotNil(t, result.VerifiedAt)
	assert.Equal(t, &verifier, result.VerifiedBy)

	// Verified → Disputed
	disputer := uuid.New()
	err = result.Dispute("scores are wrong", disputer)
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusDisputed, result.Status)
	assert.Equal(t, 1, result.DisputeCount)

	// Disputed → Conciliated
	conciliator := uuid.New()
	err = result.Conciliate(conciliator, "reviewed and confirmed", nil)
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusConciliated, result.Status)
	assert.False(t, result.WasAdjusted())

	// Conciliated → Finalized
	err = result.Finalize()
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusFinalized, result.Status)
	assert.NotNil(t, result.FinalizedAt)
}

func TestAutoVerify(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceReplayFile, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	err = result.AutoVerify()
	require.NoError(t, err)
	assert.Equal(t, scores_vo.ResultStatusVerified, result.Status)
	method := scores_vo.VerificationMethodAutomatic
	assert.Equal(t, &method, result.VerificationMethod)
}

func TestConciliateWithAdjustment(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	// Fast-track to disputed
	_ = result.Verify(scores_vo.VerificationMethodManual, nil)
	_ = result.Dispute("wrong scores", uuid.New())

	// Conciliate with adjusted scores
	adjustedTeams := []TeamResult{
		{TeamID: teams[0].TeamID, TeamName: teams[0].TeamName, Score: 13, Players: teams[0].Players},
		{TeamID: teams[1].TeamID, TeamName: teams[1].TeamName, Score: 16, Players: teams[1].Players},
	}
	err = result.Conciliate(uuid.New(), "corrected scores", adjustedTeams)
	require.NoError(t, err)

	assert.True(t, result.WasAdjusted())
	assert.Len(t, result.OriginalTeamResults, 2)
	// Winner should have changed since Team Beta now has higher score
	assert.Equal(t, teams[1].TeamID, *result.WinnerTeamID)
}

func TestInvalidTransitions(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	// Cannot finalize from submitted
	err = result.Finalize()
	assert.Error(t, err)

	// Cannot dispute from submitted
	err = result.Dispute("reason", uuid.New())
	assert.Error(t, err)

	// Verify, then finalize — then try to dispute finalized
	_ = result.AutoVerify()
	_ = result.Finalize()

	err = result.Dispute("too late", uuid.New())
	assert.Error(t, err)

	// Cannot cancel finalized
	err = result.Cancel("reason")
	assert.Error(t, err)
}

func TestGetRankedResults(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	ranked := result.GetRankedResults()
	assert.NotEmpty(t, ranked)

	// All players from winning team should have position 1
	for _, r := range ranked {
		assert.True(t, r.Position >= 1 && r.Position <= 2)
		assert.NotEqual(t, uuid.Nil, r.UserID)
	}
}

func TestGetMVPPlayerID(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	mvp := result.GetMVPPlayerID()
	assert.NotNil(t, mvp)
	assert.Equal(t, players[0].PlayerID, *mvp)
}

func TestValidation(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	tests := []struct {
		name      string
		modify    func(*MatchResult)
		expectErr bool
	}{
		{
			name:      "valid result",
			modify:    func(m *MatchResult) {},
			expectErr: false,
		},
		{
			name:      "missing match_id",
			modify:    func(m *MatchResult) { m.MatchID = uuid.Nil },
			expectErr: true,
		},
		{
			name:      "missing game_id",
			modify:    func(m *MatchResult) { m.GameID = "" },
			expectErr: true,
		},
		{
			name:      "missing submitted_by",
			modify:    func(m *MatchResult) { m.SubmittedBy = uuid.Nil },
			expectErr: true,
		},
		{
			name:      "only 1 team",
			modify:    func(m *MatchResult) { m.TeamResults = m.TeamResults[:1] },
			expectErr: true,
		},
		{
			name:      "zero played_at",
			modify:    func(m *MatchResult) { m.PlayedAt = time.Time{} },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := NewMatchResult(
				ro, uuid.New(), "cs2", "de_dust2", "competitive",
				scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
				teams, players, time.Now(), 45*time.Minute,
			)
			if result == nil {
				if !tt.expectErr {
					t.Fatal("NewMatchResult returned nil unexpectedly")
				}
				return
			}
			tt.modify(result)
			err := result.Validate()
			if tt.expectErr {
				assert.Error(t, err, tt.name)
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}

func TestGetRankedPlayerIDs(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	ranked := result.GetRankedPlayerIDs()
	assert.Len(t, ranked, 4) // 2 players per team, 2 teams

	// First players should be from winning team (position 1)
	winningTeam := result.TeamResults[0]
	for i := 0; i < len(winningTeam.Players); i++ {
		found := false
		for _, pid := range winningTeam.Players {
			if ranked[i] == pid {
				found = true
				break
			}
		}
		assert.True(t, found, "player at index %d should be from winning team", i)
	}
}

func TestSetPrizeDistribution(t *testing.T) {
	ro := testResourceOwner()
	teams := testTeamResults()
	players := testPlayerResults(teams)

	result, err := NewMatchResult(
		ro, uuid.New(), "cs2", "de_dust2", "competitive",
		scores_vo.ScoreSourceTournamentAdmin, ro.UserID,
		teams, players, time.Now(), 45*time.Minute,
	)
	require.NoError(t, err)

	distID := uuid.New()
	result.SetPrizeDistribution(distID)

	assert.NotNil(t, result.PrizeDistributionID)
	assert.Equal(t, distID, *result.PrizeDistributionID)
}
