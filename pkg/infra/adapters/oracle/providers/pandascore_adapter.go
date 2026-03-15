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

// PandaScoreAdapter fetches match scores from the PandaScore API
type PandaScoreAdapter struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Compile-time interface check
var _ oracle_out.ExternalScorePort = (*PandaScoreAdapter)(nil)

// NewPandaScoreAdapter creates a new PandaScore adapter
func NewPandaScoreAdapter(apiKey string) *PandaScoreAdapter {
	return &PandaScoreAdapter{
		apiKey:  apiKey,
		baseURL: "https://api.pandascore.co",
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// --- PandaScore API response structs ---

type pandaScoreMatch struct {
	ID            int                      `json:"id"`
	Name          string                   `json:"name"`
	Status        string                   `json:"status"`
	WinnerID      *int                     `json:"winner_id"`
	Draw          bool                     `json:"draw"`
	NumberOfGames int                      `json:"number_of_games"`
	MatchType     string                   `json:"match_type,omitempty"` // "best_of"
	Opponents     []pandaScoreOpponent     `json:"opponents"`
	Games         []pandaScoreGame         `json:"games"`
	Results       []pandaScoreResult       `json:"results"`
	Tournament    *pandaScoreTournament    `json:"tournament,omitempty"`
	StreamsList   []pandaScoreStream       `json:"streams_list,omitempty"`
	OfficialStreamURL string               `json:"official_stream_url,omitempty"`
	BeginAt       *time.Time               `json:"begin_at,omitempty"`
	EndAt         *time.Time               `json:"end_at,omitempty"`
	Videogame     *pandaScoreVideogame     `json:"videogame,omitempty"`
}

type pandaScoreTournament struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type pandaScoreStream struct {
	RawURL   string `json:"raw_url"`
	Language string `json:"language"`
	Main     bool   `json:"main"`
}

type pandaScoreVideogame struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type pandaScoreOpponent struct {
	Opponent struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"opponent"`
	Type string `json:"type"`
}

type pandaScoreGame struct {
	ID       int    `json:"id"`
	Position int    `json:"position"`
	Status   string `json:"status"`
	WinnerType string `json:"winner_type"`
	Winner struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	} `json:"winner"`
	DetailedStats bool `json:"detailed_stats"`
}

type pandaScoreResult struct {
	TeamID int `json:"team_id"`
	Score  int `json:"score"`
}

func (a *PandaScoreAdapter) FetchMatchScore(ctx context.Context, externalMatchID string, gameID replay_common.GameIDKey) (*oracle_entities.ScoreSubmission, error) {
	gameSlug := pandaScoreGameSlug(gameID)
	if gameSlug == "" {
		return nil, fmt.Errorf("unsupported game for PandaScore: %s", gameID)
	}

	url := fmt.Sprintf("%s/%s/matches/%s", a.baseURL, gameSlug, externalMatchID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PandaScore API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PandaScore API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var match pandaScoreMatch
	if err := json.Unmarshal(body, &match); err != nil {
		return nil, fmt.Errorf("failed to parse PandaScore response: %w", err)
	}

	return a.mapToSubmission(match, body)
}

func (a *PandaScoreAdapter) SupportsGame(gameID replay_common.GameIDKey) bool {
	return pandaScoreGameSlug(gameID) != ""
}

func (a *PandaScoreAdapter) ProviderID() oracle_vo.OracleSourceType {
	return oracle_vo.OracleSourcePandaScore
}

func (a *PandaScoreAdapter) ConfidenceWeight() float64 {
	return oracle_vo.SourceConfidenceWeights[oracle_vo.OracleSourcePandaScore]
}

// ListRecentMatches fetches recently completed matches from PandaScore.
// Uses the /csgo/matches/past endpoint (or equivalent for other games).
func (a *PandaScoreAdapter) ListRecentMatches(ctx context.Context, gameID replay_common.GameIDKey, since time.Time, limit int) ([]oracle_out.ExternalMatch, error) {
	gameSlug := pandaScoreGameSlug(gameID)
	if gameSlug == "" {
		return nil, fmt.Errorf("unsupported game for PandaScore: %s", gameID)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	sinceStr := since.Format(time.RFC3339)
	url := fmt.Sprintf("%s/%s/matches/past?filter[begin_at]=%s&sort=-begin_at&per_page=%d",
		a.baseURL, gameSlug, sinceStr, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PandaScore API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PandaScore API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matches []pandaScoreMatch
	if err := json.Unmarshal(body, &matches); err != nil {
		return nil, fmt.Errorf("failed to parse PandaScore response: %w", err)
	}

	result := make([]oracle_out.ExternalMatch, 0, len(matches))
	for _, m := range matches {
		ext := a.mapToExternalMatch(m, gameID)
		if ext != nil {
			result = append(result, *ext)
		}
	}

	slog.Info("PandaScore recent matches fetched",
		slog.String("game_id", string(gameID)),
		slog.Int("count", len(result)),
		slog.String("since", sinceStr),
	)

	return result, nil
}

// mapToExternalMatch converts a PandaScore match to an ExternalMatch
func (a *PandaScoreAdapter) mapToExternalMatch(m pandaScoreMatch, gameID replay_common.GameIDKey) *oracle_out.ExternalMatch {
	if len(m.Opponents) < 2 {
		return nil
	}

	teamAID := deterministicUUID("pandascore", fmt.Sprintf("%d", m.Opponents[0].Opponent.ID))
	teamBID := deterministicUUID("pandascore", fmt.Sprintf("%d", m.Opponents[1].Opponent.ID))

	var teamAScore, teamBScore int
	for _, r := range m.Results {
		if r.TeamID == m.Opponents[0].Opponent.ID {
			teamAScore = r.Score
		} else if r.TeamID == m.Opponents[1].Opponent.ID {
			teamBScore = r.Score
		}
	}

	var winnerTeamID *uuid.UUID
	if m.WinnerID != nil {
		if *m.WinnerID == m.Opponents[0].Opponent.ID {
			winnerTeamID = &teamAID
		} else {
			winnerTeamID = &teamBID
		}
	}

	var playedAt time.Time
	if m.BeginAt != nil {
		playedAt = *m.BeginAt
	} else {
		playedAt = time.Now().UTC()
	}

	// Extract stream URL (prefer main stream)
	var streamURL string
	for _, s := range m.StreamsList {
		if s.Main && s.RawURL != "" {
			streamURL = s.RawURL
			break
		}
	}
	if streamURL == "" && m.OfficialStreamURL != "" {
		streamURL = m.OfficialStreamURL
	}
	if streamURL == "" && len(m.StreamsList) > 0 {
		streamURL = m.StreamsList[0].RawURL
	}

	var tournamentName, tournamentID string
	if m.Tournament != nil {
		tournamentName = m.Tournament.Name
		tournamentID = fmt.Sprintf("%d", m.Tournament.ID)
	}

	seriesType := fmt.Sprintf("bo%d", m.NumberOfGames)

	return &oracle_out.ExternalMatch{
		ExternalMatchID: fmt.Sprintf("%d", m.ID),
		GameID:          gameID,
		Provider:        oracle_vo.OracleSourcePandaScore,
		TeamAName:       m.Opponents[0].Opponent.Name,
		TeamBName:       m.Opponents[1].Opponent.Name,
		TeamAID:         teamAID,
		TeamBID:         teamBID,
		TeamAScore:      teamAScore,
		TeamBScore:      teamBScore,
		WinnerTeamID:    winnerTeamID,
		IsDraw:          m.Draw,
		Status:          m.Status,
		TournamentName:  tournamentName,
		TournamentID:    tournamentID,
		SeriesType:      seriesType,
		StreamURL:       streamURL,
		PlayedAt:        playedAt,
		NumberOfGames:   m.NumberOfGames,
	}
}

func (a *PandaScoreAdapter) mapToSubmission(match pandaScoreMatch, rawBody []byte) (*oracle_entities.ScoreSubmission, error) {
	if len(match.Opponents) < 2 {
		return nil, fmt.Errorf("match has fewer than 2 opponents")
	}

	// Map team IDs (deterministic UUID from PandaScore integer IDs)
	teamAID := deterministicUUID("pandascore", fmt.Sprintf("%d", match.Opponents[0].Opponent.ID))
	teamBID := deterministicUUID("pandascore", fmt.Sprintf("%d", match.Opponents[1].Opponent.ID))

	var teamAScore, teamBScore int
	for _, r := range match.Results {
		if r.TeamID == match.Opponents[0].Opponent.ID {
			teamAScore = r.Score
		} else if r.TeamID == match.Opponents[1].Opponent.ID {
			teamBScore = r.Score
		}
	}

	var winnerTeamID *uuid.UUID
	if match.WinnerID != nil {
		if *match.WinnerID == match.Opponents[0].Opponent.ID {
			winnerTeamID = &teamAID
		} else {
			winnerTeamID = &teamBID
		}
	}

	// Map game details
	gameDetails := make([]oracle_entities.SubmissionGameDetail, 0, len(match.Games))
	for _, g := range match.Games {
		gd := oracle_entities.SubmissionGameDetail{
			Position: g.Position,
		}
		// Determine winner per game
		if g.Winner.ID == match.Opponents[0].Opponent.ID {
			gd.TeamAWon = true
		}
		gameDetails = append(gameDetails, gd)
	}

	slog.Info("PandaScore match mapped",
		slog.String("match_id", fmt.Sprintf("%d", match.ID)),
		slog.Int("team_a_score", teamAScore),
		slog.Int("team_b_score", teamBScore),
		slog.Int("games", len(match.Games)),
	)

	return &oracle_entities.ScoreSubmission{
		SourceType:      oracle_vo.OracleSourcePandaScore,
		ProviderMatchID: fmt.Sprintf("%d", match.ID),
		WinnerTeamID:    winnerTeamID,
		IsDraw:          match.Draw,
		TeamAID:         teamAID,
		TeamBID:         teamBID,
		TeamAScore:      teamAScore,
		TeamBScore:      teamBScore,
		RoundsPlayed:    0, // PandaScore doesn't always provide round-level data
		GameDetails:     gameDetails,
		RawResponse:     rawBody,
		SourceHash:      fmt.Sprintf("%x", hashBytes(rawBody)),
	}, nil
}

// pandaScoreGameSlug maps internal game IDs to PandaScore slug prefixes
func pandaScoreGameSlug(gameID replay_common.GameIDKey) string {
	switch gameID {
	case replay_common.CS2_GAME_ID, replay_common.CSGO_GAME_ID:
		return "csgo"
	case replay_common.VLRNT_GAME_ID:
		return "valorant"
	default:
		return ""
	}
}
