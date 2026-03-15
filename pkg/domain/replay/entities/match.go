package entities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

type MatchVisibility string

const (
	MatchVisibilityPublic  MatchVisibility = "public"
	MatchVisibilitySquad   MatchVisibility = "squad"
	MatchVisibilityPrivate MatchVisibility = "private"
	MatchVisibilityCustom  MatchVisibility = "custom"
)

// MatchStatus represents the processing status of a match
type MatchStatus string

const (
	MatchStatusPending    MatchStatus = "pending"
	MatchStatusProcessing MatchStatus = "processing"
	MatchStatusCompleted  MatchStatus = "completed"
	MatchStatusFailed     MatchStatus = "failed"
)

// MatchSource indicates how the match was created/sourced
type MatchSource string

const (
	// MatchSourceReplay - Match created from replay file processing
	MatchSourceReplay MatchSource = "replay"
	// MatchSourceMatchmaking - Match created via matchmaking system
	MatchSourceMatchmaking MatchSource = "matchmaking"
	// MatchSourceExternalAPI - Match data imported from external API (e.g., Valve, FACEIT)
	MatchSourceExternalAPI MatchSource = "external_api"
	// MatchSourceManual - Match manually created by user/admin
	MatchSourceManual MatchSource = "manual"
	// MatchSourceOCRStream - Match imported from live stream OCR analysis
	MatchSourceOCRStream MatchSource = "ocr_stream"
	// MatchSourceOCRScreenshot - Match imported from user-uploaded screenshot OCR
	MatchSourceOCRScreenshot MatchSource = "ocr_screenshot"
	// MatchSourceYouTubeVOD - Match imported from YouTube VOD OCR analysis
	MatchSourceYouTubeVOD MatchSource = "youtube_vod"
	// MatchSourceDemo - Match imported from demo file analysis
	MatchSourceDemo MatchSource = "demo"
)

// ReconciliationOutcome indicates the result of a reconciliation attempt
type ReconciliationOutcome string

const (
	ReconciliationCreated             ReconciliationOutcome = "created"
	ReconciliationReconciled          ReconciliationOutcome = "reconciled"
	ReconciliationReconciledConflict  ReconciliationOutcome = "reconciled_with_conflict"
	ReconciliationReconciledDateShift ReconciliationOutcome = "reconciled_date_variant"
	ReconciliationReconciledExtID     ReconciliationOutcome = "reconciled_via_external_id"
)

// SourceConfirmation records a single source's contribution to a match record.
// Each time the same match is discovered from a different source, a new confirmation
// is appended. This provides full provenance tracking and enables conflict detection.
type SourceConfirmation struct {
	Source          MatchSource `json:"source" bson:"source"`
	ExternalMatchID string      `json:"external_match_id,omitempty" bson:"external_match_id,omitempty"`
	Provider        string      `json:"provider,omitempty" bson:"provider,omitempty"`
	TeamAName       string      `json:"team_a_name,omitempty" bson:"team_a_name,omitempty"`
	TeamBName       string      `json:"team_b_name,omitempty" bson:"team_b_name,omitempty"`
	TeamAScore      int         `json:"team_a_score" bson:"team_a_score"`
	TeamBScore      int         `json:"team_b_score" bson:"team_b_score"`
	MapName         string      `json:"map_name,omitempty" bson:"map_name,omitempty"`
	ConfirmedAt     time.Time   `json:"confirmed_at" bson:"confirmed_at"`
	Confidence      float64     `json:"confidence" bson:"confidence"`
}

// AggregateRoot
type Match struct {
	shared.BaseEntity `json:",inline" bson:",inline"`
	RegionID          replay_common.RegionIDKey `json:"region_id" bson:"region_id"`
	ReplayFileID      uuid.UUID                 `json:"replay_file_id" bson:"replay_file_id"`
	GameID            replay_common.GameIDKey   `json:"game_id" bson:"game_id"`
	MapName           string                    `json:"map_name,omitempty" bson:"map_name,omitempty"`
	Duration          float64                   `json:"duration,omitempty" bson:"duration,omitempty"`
	PlayedAt          time.Time                 `json:"played_at,omitempty" bson:"played_at,omitempty"`
	Mode              string                    `json:"mode,omitempty" bson:"mode,omitempty"`
	Status            MatchStatus               `json:"status,omitempty" bson:"status,omitempty"`
	ServerName        string                    `json:"server_name,omitempty" bson:"server_name,omitempty"`
	Scoreboard        Scoreboard                `json:"scoreboard" bson:"scoreboard"`
	Teams             []Team                    `json:"teams" bson:"teams"`
	EventCount        int                       `json:"event_count" bson:"event_count"`
	Visibility        MatchVisibility           `json:"visibility" bson:"visibility"`
	ShareTokens       []ShareToken              `json:"share_tokens" bson:"share_tokens"`
	// Source tracking
	Source          MatchSource `json:"source" bson:"source"`
	LinkedReplayID  *uuid.UUID  `json:"linked_replay_id,omitempty" bson:"linked_replay_id,omitempty"`
	ExternalMatchID string      `json:"external_match_id,omitempty" bson:"external_match_id,omitempty"`
	Slug            string      `json:"slug,omitempty" bson:"slug,omitempty"`
	LinkedMatchIDs  []uuid.UUID `json:"linked_match_ids,omitempty" bson:"linked_match_ids,omitempty"`
	// Multi-source provenance & conflict detection
	SourceConfirmations []SourceConfirmation `json:"source_confirmations,omitempty" bson:"source_confirmations,omitempty"`
	NeedsReview         bool                 `json:"needs_review,omitempty" bson:"needs_review,omitempty"`
	ConflictDetails     string               `json:"conflict_details,omitempty" bson:"conflict_details,omitempty"`
}

func (m Match) GetID() uuid.UUID {
	return m.ID
}

// AddSourceConfirmation appends a source confirmation and detects conflicts.
// Returns true if a conflict was detected (scores disagree with existing confirmations).
func (m *Match) AddSourceConfirmation(sc SourceConfirmation) bool {
	// Check for duplicate source
	for _, existing := range m.SourceConfirmations {
		if existing.Source == sc.Source && existing.ExternalMatchID == sc.ExternalMatchID {
			return false // Already confirmed from this exact source
		}
	}

	m.SourceConfirmations = append(m.SourceConfirmations, sc)

	// Detect score conflicts across confirmations
	if len(m.SourceConfirmations) > 1 {
		return m.detectConflicts()
	}

	return false
}

// detectConflicts checks whether any source confirmations have conflicting scores.
// Only compares confirmations that both have non-zero scores.
func (m *Match) detectConflicts() bool {
	type scoreKey struct{ a, b int }
	scores := make(map[scoreKey][]string)

	for _, sc := range m.SourceConfirmations {
		if sc.TeamAScore == 0 && sc.TeamBScore == 0 {
			continue
		}
		key := scoreKey{sc.TeamAScore, sc.TeamBScore}
		scores[key] = append(scores[key], fmt.Sprintf("%s(%s)", sc.Source, sc.Provider))
	}

	if len(scores) > 1 {
		details := "Score conflict detected: "
		first := true
		for score, sources := range scores {
			if !first {
				details += " vs "
			}
			srcList := ""
			for i, s := range sources {
				if i > 0 {
					srcList += ", "
				}
				srcList += s
			}
			details += fmt.Sprintf("%d-%d from [%s]", score.a, score.b, srcList)
			first = false
		}
		m.NeedsReview = true
		m.ConflictDetails = details
		return true
	}

	return false
}

type Scoreboard struct {
	TeamScoreboards []TeamScoreboard `json:"team_scoreboards" bson:"team_scoreboards"`
	MatchMVP        *PlayerMetadata  `json:"match_mvp" bson:"match_mvp"`
}

// PlayerStatsEntry contains comprehensive esports stats for a single player
type PlayerStatsEntry struct {
	PlayerID      string  `json:"player_id" bson:"player_id"`
	Kills         int     `json:"kills" bson:"kills"`
	Deaths        int     `json:"deaths" bson:"deaths"`
	Assists       int     `json:"assists" bson:"assists"`
	KDRatio       float64 `json:"kd_ratio" bson:"kd_ratio"`
	Headshots     int     `json:"headshots" bson:"headshots"`
	HeadshotPct   float64 `json:"headshot_pct" bson:"headshot_pct"`       // Headshot percentage
	TotalDamage   int     `json:"total_damage" bson:"total_damage"`
	UtilityDamage int     `json:"utility_damage" bson:"utility_damage"`
	ADR           float64 `json:"adr" bson:"adr"`
	MVPCount      int     `json:"mvp_count" bson:"mvp_count"`
	Score         int     `json:"score" bson:"score"`
	// Advanced esports stats
	KAST            float64 `json:"kast" bson:"kast"`                           // Kill/Assist/Survived/Traded %
	ImpactRating    float64 `json:"impact_rating" bson:"impact_rating"`         // Impact rating (HLTV-style)
	OpeningKills    int     `json:"opening_kills" bson:"opening_kills"`         // First kills of a round
	OpeningDeaths   int     `json:"opening_deaths" bson:"opening_deaths"`       // First deaths of a round
	TradeKills      int     `json:"trade_kills" bson:"trade_kills"`             // Kills avenging teammates
	ClutchWins      int     `json:"clutch_wins" bson:"clutch_wins"`             // 1vX clutch wins
	ClutchAttempts  int     `json:"clutch_attempts" bson:"clutch_attempts"`     // 1vX clutch attempts
	FlashAssists    int     `json:"flash_assists" bson:"flash_assists"`         // Kills assisted by flashes
	EnemiesFlashed  int     `json:"enemies_flashed" bson:"enemies_flashed"`     // Total enemies flashed
	EntryAttempts   int     `json:"entry_attempts" bson:"entry_attempts"`       // Entry duel attempts
	EntrySuccesses  int     `json:"entry_successes" bson:"entry_successes"`     // Entry duel wins
	MultiKills      int     `json:"multi_kills" bson:"multi_kills"`             // 2+ kills in a round
	Rating2         float64 `json:"rating_2" bson:"rating_2"`                   // HLTV 2.0 rating
}

type TeamScoreboard struct {
	Team        Team               `json:"team" bson:"team"`
	Side        string             `json:"side" bson:"side"`
	TeamScore   int                `json:"team_score" bson:"team_score"`
	TeamMVP     *PlayerMetadata    `json:"team_mvp" bson:"team_mvp"`
	Players     []PlayerMetadata   `json:"players" bson:"players"`
	PlayerStats []PlayerStatsEntry `json:"player_stats" bson:"player_stats"`
	Rounds      []RoundInfo        `json:"rounds" bson:"rounds"`
	RoundStats  map[int]interface{} `json:"round_stats" bson:"round_stats"`
}

type RoundInfo struct {
	RoundNumber      int         `json:"round_number" bson:"round_number"`
	WinnerTeamID     *uuid.UUID  `json:"winner" bson:"winner"`
	RoundMVPPlayerID *uuid.UUID  `json:"round_mvp_player_id" bson:"round_mvp_player_id"`
	Events           []GameEvent `json:"events" bson:"events"`
}

func NewCS2Match(userContext context.Context, replayFileID uuid.UUID) *Match {
	resourceOwner := shared.GetResourceOwner(userContext)
	// Use NewUnrestrictedEntity for public visibility by default
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:   entity,
		ReplayFileID: replayFileID,
		GameID:       replay_common.CS2.ID,
		Source:       MatchSourceReplay, // Matches from replay processing
	}
}

func NewCS2MatchWithOwner(resourceOwner shared.ResourceOwner, replayFileID uuid.UUID) *Match {
	// Use NewUnrestrictedEntity for public visibility by default
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:   entity,
		ReplayFileID: replayFileID,
		GameID:       replay_common.CS2.ID,
		Source:       MatchSourceReplay, // Matches from replay processing
	}
}

// NewMatchFromMatchmaking creates a match from the matchmaking system
func NewMatchFromMatchmaking(resourceOwner shared.ResourceOwner, gameID replay_common.GameIDKey, externalMatchID string) *Match {
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:      entity,
		GameID:          gameID,
		Source:          MatchSourceMatchmaking,
		ExternalMatchID: externalMatchID,
		Status:          MatchStatusPending,
	}
}

// NewMatchFromExternalAPI creates a match from external API data (FACEIT, Valve, etc.)
func NewMatchFromExternalAPI(resourceOwner shared.ResourceOwner, gameID replay_common.GameIDKey, externalMatchID string) *Match {
	entity := shared.NewUnrestrictedEntity(resourceOwner)
	return &Match{
		BaseEntity:      entity,
		GameID:          gameID,
		Source:          MatchSourceExternalAPI,
		ExternalMatchID: externalMatchID,
		Status:          MatchStatusCompleted,
	}
}

// NewMatchFromOCRImport creates an enriched match from OCR/import pipeline data.
// Unlike NewMatchFromExternalAPI, this populates team names, map, scores, and played_at from the source data.
func NewMatchFromOCRImport(
	resourceOwner shared.ResourceOwner,
	gameID replay_common.GameIDKey,
	source MatchSource,
	externalMatchID string,
	slug string,
	teamAName string,
	teamBName string,
	teamAScore int,
	teamBScore int,
	mapName string,
	playedAt time.Time,
) *Match {
	entity := shared.NewUnrestrictedEntity(resourceOwner)

	teams := make([]Team, 0, 2)
	if teamAName != "" {
		teams = append(teams, Team{
			BaseEntity:         shared.NewUnrestrictedEntity(resourceOwner),
			Name:               teamAName,
			CurrentDisplayName: teamAName,
		})
	}
	if teamBName != "" {
		teams = append(teams, Team{
			BaseEntity:         shared.NewUnrestrictedEntity(resourceOwner),
			Name:               teamBName,
			CurrentDisplayName: teamBName,
		})
	}

	var scoreboard Scoreboard
	if len(teams) == 2 {
		scoreboard = Scoreboard{
			TeamScoreboards: []TeamScoreboard{
				{Team: teams[0], TeamScore: teamAScore},
				{Team: teams[1], TeamScore: teamBScore},
			},
		}
	}

	// Canonicalize map name for storage
	canonMap := CanonicalizeMapName(mapName)
	if canonMap == "" {
		canonMap = mapName
	}

	m := &Match{
		BaseEntity:      entity,
		GameID:          gameID,
		Source:          source,
		ExternalMatchID: externalMatchID,
		Slug:            slug,
		MapName:         canonMap,
		PlayedAt:        playedAt,
		Teams:           teams,
		Scoreboard:      scoreboard,
		Status:          MatchStatusCompleted,
		// Initial source confirmation
		SourceConfirmations: []SourceConfirmation{
			{
				Source:          source,
				ExternalMatchID: externalMatchID,
				TeamAName:       teamAName,
				TeamBName:       teamBName,
				TeamAScore:      teamAScore,
				TeamBScore:      teamBScore,
				MapName:         canonMap,
				ConfirmedAt:     time.Now().UTC(),
				Confidence:      1.0,
			},
		},
	}

	return m
}

// LinkReplay links a replay file to a matchmaking or external API match
func (m *Match) LinkReplay(replayID uuid.UUID) {
	m.LinkedReplayID = &replayID
	m.ReplayFileID = replayID // Also set the main replay file ID for backwards compatibility
}

// HasReplay returns true if the match has an associated replay file
func (m *Match) HasReplay() bool {
	return m.ReplayFileID != uuid.Nil || (m.LinkedReplayID != nil && *m.LinkedReplayID != uuid.Nil)
}

// LinkMatch links another match to this one for reconciliation purposes
func (m *Match) LinkMatch(matchID uuid.UUID) {
	for _, id := range m.LinkedMatchIDs {
		if id == matchID {
			return // Already linked
		}
	}
	m.LinkedMatchIDs = append(m.LinkedMatchIDs, matchID)
}

// MatchSourceFromOracleSource maps an OracleSourceType to the corresponding MatchSource.
// This is used during import to determine the correct MatchSource based on the data provider.
func MatchSourceFromOracleSource(oracleSource string) MatchSource {
	switch oracleSource {
	case "ocr_stream":
		return MatchSourceOCRStream
	case "ocr_screenshot":
		return MatchSourceOCRScreenshot
	case "pandascore", "steam_web_api", "faceit_data_api", "sportsdataio", "grid", "abios":
		return MatchSourceExternalAPI
	default:
		return MatchSourceExternalAPI
	}
}

// FinalScoreboardPayload contains the final match scoreboard data
// This is the canonical type used for scoreboard extraction during replay processing
type FinalScoreboardPayload struct {
	CTScore      int                          `json:"ct_score"`
	TScore       int                          `json:"t_score"`
	CTTeamName   string                       `json:"ct_team_name"`
	TTeamName    string                       `json:"t_team_name"`
	TotalRounds  int                          `json:"total_rounds"`
	WinnerSide   string                       `json:"winner_side"`
	Players      []PlayerScoreboardDataEntry  `json:"players"`
	Duration     float64                      `json:"duration"`
	MapName      string                       `json:"map_name"`
}

// PlayerScoreboardDataEntry contains individual player statistics for scoreboard
type PlayerScoreboardDataEntry struct {
	NetworkUserID  string  `json:"network_user_id"`
	Name           string  `json:"name"`
	Team           string  `json:"team"`
	Side           string  `json:"side"` // "CT" or "T"
	Kills          int     `json:"kills"`
	Deaths         int     `json:"deaths"`
	Assists        int     `json:"assists"`
	KDRatio        float64 `json:"kd_ratio"`
	Headshots      int     `json:"headshots"`
	TotalDamage    int     `json:"total_damage"`
	UtilityDamage  int     `json:"utility_damage"`
	ADR            float64 `json:"adr"` // Average Damage per Round
	MVPCount       int     `json:"mvp_count"`
	Score          int     `json:"score"`
	FirstKills     int     `json:"first_kills"`
	FirstDeaths    int     `json:"first_deaths"`
	TradeKills     int     `json:"trade_kills"`
	TradeDeaths    int     `json:"trade_deaths"`
	FlashAssists   int     `json:"flash_assists"`
	EnemiesFlashed int     `json:"enemies_flashed"`
}
