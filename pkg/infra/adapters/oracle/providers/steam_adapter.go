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

// SteamWebAPIAdapter fetches match scores from the Steam Web API.
// Authoritative for Valve games (CS2, CSGO) with highest confidence weight (0.95).
type SteamWebAPIAdapter struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Compile-time interface check
var _ oracle_out.ExternalScorePort = (*SteamWebAPIAdapter)(nil)

// NewSteamWebAPIAdapter creates a new Steam Web API adapter
func NewSteamWebAPIAdapter(apiKey string) *SteamWebAPIAdapter {
	return &SteamWebAPIAdapter{
		apiKey:  apiKey,
		baseURL: "https://api.steampowered.com",
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// --- Steam API response structs ---

type steamMatchResponse struct {
	Result struct {
		Matches []steamMatch `json:"matches"`
	} `json:"result"`
}

type steamMatch struct {
	MatchID        int64           `json:"match_id"`
	Duration       int             `json:"duration"`
	StartTime      int64           `json:"start_time"`
	RadiantWin     bool            `json:"radiant_win"`           // For Dota 2
	CTWin          *bool           `json:"ct_win,omitempty"`      // For CS2
	WinnerSide     string          `json:"winner_side,omitempty"` // "CT" or "T"
	ScoreCT        int             `json:"score_ct"`
	ScoreT         int             `json:"score_t"`
	MapName        string          `json:"map_name"`
	RoundsPlayed   int             `json:"rounds_played"`
	Players        []steamPlayer   `json:"players"`
}

type steamPlayer struct {
	AccountID  int64   `json:"account_id"`
	TeamNumber int     `json:"team_number"` // 0 or 1
	Kills      int     `json:"kills"`
	Deaths     int     `json:"deaths"`
	Assists    int     `json:"assists"`
	MVPs       int     `json:"mvps"`
	Score      int     `json:"score"`
}

func (a *SteamWebAPIAdapter) FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error) {
	appID := steamGameID(gameID)
	if appID == "" {
		return nil, fmt.Errorf("unsupported game for Steam Web API: %s", gameID)
	}

	url := fmt.Sprintf("%s/IDOTA2Match_570/GetMatchDetails/v1/?key=%s&match_id=%s",
		a.baseURL, a.apiKey, externalMatchID)

	// For CS2, use different endpoint
	if gameID == replay_common.CS2_GAME_ID || gameID == replay_common.CSGO_GAME_ID {
		url = fmt.Sprintf("%s/ICSGOPlayers_730/GetMatchHistory/v1/?key=%s&match_id=%s",
			a.baseURL, a.apiKey, externalMatchID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Steam Web API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Steam Web API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matchResp steamMatchResponse
	if err := json.Unmarshal(body, &matchResp); err != nil {
		return nil, fmt.Errorf("failed to parse Steam response: %w", err)
	}

	if len(matchResp.Result.Matches) == 0 {
		return nil, fmt.Errorf("no match found for ID: %s", externalMatchID)
	}

	return a.mapToSubmission(matchResp.Result.Matches[0], body)
}

func (a *SteamWebAPIAdapter) SupportsGame(gameID replay_common.GameIDKey) bool {
	return steamGameID(gameID) != ""
}

func (a *SteamWebAPIAdapter) ProviderID() oracle_vo.OracleSourceType {
	return oracle_vo.OracleSourceSteamWebAPI
}

func (a *SteamWebAPIAdapter) ConfidenceWeight() float64 {
	return oracle_vo.SourceConfidenceWeights[oracle_vo.OracleSourceSteamWebAPI]
}

// ListRecentMatches is not supported by the Steam Web API (no tournament match listing).
// Returns empty to signal no matches from this source.
func (a *SteamWebAPIAdapter) ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]oracle_out.ExternalMatch, error) {
	return nil, nil
}

func (a *SteamWebAPIAdapter) mapToSubmission(match steamMatch, rawBody []byte) (*oracle_entities.ScoreSubmission, error) {
	teamAID := deterministicUUID("steam", fmt.Sprintf("team_ct_%d", match.MatchID))
	teamBID := deterministicUUID("steam", fmt.Sprintf("team_t_%d", match.MatchID))

	teamAScore := match.ScoreCT
	teamBScore := match.ScoreT
	roundsPlayed := match.RoundsPlayed
	if roundsPlayed == 0 {
		roundsPlayed = teamAScore + teamBScore
	}

	var winnerTeamID *uuid.UUID
	isDraw := false
	if teamAScore > teamBScore {
		winnerTeamID = &teamAID
	} else if teamBScore > teamAScore {
		winnerTeamID = &teamBID
	} else {
		isDraw = true
	}

	// Map player scores
	playerScores := make([]oracle_entities.SubmissionPlayerScore, 0, len(match.Players))
	var mvpPlayerID *uuid.UUID
	maxMVPs := 0
	for _, p := range match.Players {
		pid := deterministicUUID("steam", fmt.Sprintf("player_%d", p.AccountID))
		var teamID uuid.UUID
		if p.TeamNumber == 0 {
			teamID = teamAID
		} else {
			teamID = teamBID
		}

		playerScores = append(playerScores, oracle_entities.SubmissionPlayerScore{
			PlayerID: pid,
			TeamID:   teamID,
			Kills:    p.Kills,
			Deaths:   p.Deaths,
			Assists:  p.Assists,
		})

		if p.MVPs > maxMVPs {
			maxMVPs = p.MVPs
			id := pid
			mvpPlayerID = &id
		}
	}

	slog.Info("Steam match mapped",
		slog.Int64("match_id", match.MatchID),
		slog.Int("score_ct", teamAScore),
		slog.Int("score_t", teamBScore),
		slog.Int("rounds", roundsPlayed),
	)

	return &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourceSteamWebAPI,
		ProviderMatchID: fmt.Sprintf("%d", match.MatchID),
		WinnerTeamID:    winnerTeamID,
		IsDraw:          isDraw,
		TeamAID:         teamAID,
		TeamBID:         teamBID,
		TeamAScore:      teamAScore,
		TeamBScore:      teamBScore,
		RoundsPlayed:    roundsPlayed,
		MVPPlayerID:     mvpPlayerID,
		PlayerScores:    playerScores,
		GameDetails: []oracle_entities.SubmissionGameDetail{
			{
				Position:   1,
				MapName:    match.MapName,
				TeamAScore: teamAScore,
				TeamBScore: teamBScore,
				TeamAWon:   teamAScore > teamBScore,
			},
		},
		RawResponse: rawBody,
		SourceHash:  fmt.Sprintf("%x", hashBytes(rawBody)),
	}, nil
}

// steamGameID maps internal game IDs to Steam App IDs
func steamGameID(gameID replay_common.GameIDKey) string {
	switch gameID {
	case replay_common.CS2_GAME_ID, replay_common.CSGO_GAME_ID:
		return "730"
	default:
		return ""
	}
}
