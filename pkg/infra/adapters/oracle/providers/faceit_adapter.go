package oracle_providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// FACEITAdapter fetches match scores from the FACEIT Data API
type FACEITAdapter struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Compile-time interface check
var _ oracle_out.ExternalScorePort = (*FACEITAdapter)(nil)

// NewFACEITAdapter creates a new FACEIT Data API adapter
func NewFACEITAdapter(apiKey string) *FACEITAdapter {
	return &FACEITAdapter{
		apiKey:  apiKey,
		baseURL: "https://open.faceit.com/data/v4",
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// --- FACEIT API response structs ---

type faceitMatchStats struct {
	Rounds []faceitRound `json:"rounds"`
}

type faceitRound struct {
	BestOf     string         `json:"best_of"`
	GameMode   string         `json:"game_mode"`
	MatchID    string         `json:"match_id"`
	MatchRound string         `json:"match_round"`
	Played     string         `json:"played"`
	RoundStats faceitRoundStats `json:"round_stats"`
	Teams      []faceitTeam     `json:"teams"`
}

type faceitRoundStats struct {
	Map    string `json:"Map"`
	Rounds string `json:"Rounds"`
	Score  string `json:"Score"`
	Winner string `json:"Winner"`
}

type faceitTeam struct {
	TeamID    string              `json:"team_id"`
	Premade   bool                `json:"premade"`
	TeamStats faceitTeamStats     `json:"team_stats"`
	Players   []faceitPlayerStats `json:"players"`
}

type faceitTeamStats struct {
	TeamWin     string `json:"Team Win"`
	FinalScore  string `json:"Final Score"`
	TeamHeadshots string `json:"Team Headshots"`
}

type faceitPlayerStats struct {
	PlayerID string            `json:"player_id"`
	Nickname string            `json:"nickname"`
	PlayerStats map[string]string `json:"player_stats"`
}

func (a *FACEITAdapter) FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error) {
	url := fmt.Sprintf("%s/matches/%s/stats", a.baseURL, externalMatchID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FACEIT API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FACEIT API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var stats faceitMatchStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse FACEIT response: %w", err)
	}

	if len(stats.Rounds) == 0 {
		return nil, fmt.Errorf("no rounds found for FACEIT match: %s", externalMatchID)
	}

	return a.mapToSubmission(externalMatchID, stats, body)
}

func (a *FACEITAdapter) SupportsGame(gameID replay_common.GameIDKey) bool {
	return faceitGameSlug(gameID) != ""
}

func (a *FACEITAdapter) ProviderID() oracle_vo.OracleSourceType {
	return oracle_vo.OracleSourceFACEIT
}

func (a *FACEITAdapter) ConfidenceWeight() float64 {
	return oracle_vo.SourceConfidenceWeights[oracle_vo.OracleSourceFACEIT]
}

// ListRecentMatches fetches recently completed matches from FACEIT.
// TODO: Implement using FACEIT Data API v4 /championships endpoint.
func (a *FACEITAdapter) ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]oracle_out.ExternalMatch, error) {
	slog.Warn("FACEIT ListRecentMatches not yet implemented")
	return nil, nil
}

func (a *FACEITAdapter) mapToSubmission(matchID string, stats faceitMatchStats, rawBody []byte) (*oracle_entities.ScoreSubmission, error) {
	if len(stats.Rounds) == 0 || len(stats.Rounds[0].Teams) < 2 {
		return nil, fmt.Errorf("insufficient team data in FACEIT response")
	}

	// Use the first round to get team IDs
	teamA := stats.Rounds[0].Teams[0]
	teamB := stats.Rounds[0].Teams[1]

	teamAID := deterministicUUID("faceit", teamA.TeamID)
	teamBID := deterministicUUID("faceit", teamB.TeamID)

	// Count map wins for series score
	teamAMapWins := 0
	teamBMapWins := 0
	totalRounds := 0

	gameDetails := make([]oracle_entities.SubmissionGameDetail, 0, len(stats.Rounds))
	for i, round := range stats.Rounds {
		teamAWon := false
		var teamAScore, teamBScore int
		for _, team := range round.Teams {
			if team.TeamID == teamA.TeamID {
				fmt.Sscanf(team.TeamStats.FinalScore, "%d", &teamAScore)
				if team.TeamStats.TeamWin == "1" {
					teamAWon = true
					teamAMapWins++
				}
			} else {
				fmt.Sscanf(team.TeamStats.FinalScore, "%d", &teamBScore)
				if team.TeamStats.TeamWin == "1" {
					teamBMapWins++
				}
			}
		}

		totalRounds += teamAScore + teamBScore

		gameDetails = append(gameDetails, oracle_entities.SubmissionGameDetail{
			Position:   i + 1,
			MapName:    round.RoundStats.Map,
			TeamAScore: teamAScore,
			TeamBScore: teamBScore,
			TeamAWon:   teamAWon,
		})
	}

	// Determine winner
	var winnerTeamID *uuid.UUID
	isDraw := false
	if teamAMapWins > teamBMapWins {
		winnerTeamID = &teamAID
	} else if teamBMapWins > teamAMapWins {
		winnerTeamID = &teamBID
	} else {
		isDraw = true
	}

	// Map player scores from first round (aggregate later for full matches)
	playerScores := make([]oracle_entities.SubmissionPlayerScore, 0)
	for _, team := range stats.Rounds[0].Teams {
		var tid uuid.UUID
		if team.TeamID == teamA.TeamID {
			tid = teamAID
		} else {
			tid = teamBID
		}
		for _, p := range team.Players {
			pid := deterministicUUID("faceit", p.PlayerID)
			var kills, deaths, assists int
			fmt.Sscanf(p.PlayerStats["Kills"], "%d", &kills)
			fmt.Sscanf(p.PlayerStats["Deaths"], "%d", &deaths)
			fmt.Sscanf(p.PlayerStats["Assists"], "%d", &assists)

			playerScores = append(playerScores, oracle_entities.SubmissionPlayerScore{
				PlayerID: pid,
				TeamID:   tid,
				Kills:    kills,
				Deaths:   deaths,
				Assists:  assists,
			})
		}
	}

	slog.Info("FACEIT match mapped",
		slog.String("match_id", matchID),
		slog.Int("team_a_wins", teamAMapWins),
		slog.Int("team_b_wins", teamBMapWins),
		slog.Int("maps_played", len(stats.Rounds)),
	)

	return &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourceFACEIT,
		ProviderMatchID: matchID,
		WinnerTeamID:    winnerTeamID,
		IsDraw:          isDraw,
		TeamAID:         teamAID,
		TeamBID:         teamBID,
		TeamAScore:      teamAMapWins,
		TeamBScore:      teamBMapWins,
		RoundsPlayed:    totalRounds,
		GameDetails:     gameDetails,
		PlayerScores:    playerScores,
		RawResponse:     rawBody,
		SourceHash:      fmt.Sprintf("%x", hashBytes(rawBody)),
	}, nil
}

// faceitGameSlug maps internal game IDs to FACEIT game slugs
func faceitGameSlug(gameID replay_common.GameIDKey) string {
	switch gameID {
	case replay_common.CS2_GAME_ID, replay_common.CSGO_GAME_ID:
		return "cs2"
	case replay_common.VLRNT_GAME_ID:
		return "valorant"
	default:
		return ""
	}
}
