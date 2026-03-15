package oracle_in

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	oracle_entities "github.com/replay-api/replay-api/pkg/domain/oracle/entities"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// GameImportCommandHandler coordinates the end-to-end flow:
// discovery → oracle ingestion → consensus → match creation → score bridging.
type GameImportCommandHandler interface {
	// ImportDiscoveredMatch imports a match discovered from an external provider.
	// Creates the Match, OracleResult, and triggers full ingestion + consensus.
	ImportDiscoveredMatch(ctx context.Context, cmd ImportDiscoveredMatchCommand) error

	// ImportFromYouTubeVOD creates an OCR stream config and triggers VOD processing.
	ImportFromYouTubeVOD(ctx context.Context, cmd ImportFromYouTubeVODCommand) error

	// BridgeOracleToMatchResult converts a finalized OracleResult into a MatchResult.
	BridgeOracleToMatchResult(ctx context.Context, cmd BridgeOracleToMatchResultCommand) error
}

// ImportDiscoveredMatchCommand is issued when the discovery worker finds a new external match.
type ImportDiscoveredMatchCommand struct {
	ExternalMatch   oracle_out.ExternalMatch `json:"external_match"`
	TriggerOCR      bool                     `json:"trigger_ocr"`      // Whether to also process VODs via OCR
	TriggerAPIIngest bool                    `json:"trigger_api_ingest"` // Whether to ingest via API providers
}

func (c ImportDiscoveredMatchCommand) Validate() error {
	if c.ExternalMatch.ExternalMatchID == "" {
		return fmt.Errorf("external_match_id is required")
	}
	if c.ExternalMatch.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	return nil
}

// ImportFromYouTubeVODCommand triggers OCR processing of a YouTube VOD for score extraction.
type ImportFromYouTubeVODCommand struct {
	VideoURL        string                     `json:"video_url"`
	GameID          replay_common.GameIDKey     `json:"game_id"`
	ExternalMatchID string                     `json:"external_match_id"`
	TeamAHint       string                     `json:"team_a_hint,omitempty"`
	TeamBHint       string                     `json:"team_b_hint,omitempty"`
	ScoreboardRegion *oracle_entities.ScoreboardRegion `json:"scoreboard_region,omitempty"`
}

func (c ImportFromYouTubeVODCommand) Validate() error {
	if c.VideoURL == "" {
		return fmt.Errorf("video_url is required")
	}
	if c.GameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if c.ExternalMatchID == "" {
		return fmt.Errorf("external_match_id is required")
	}
	return nil
}

// BridgeOracleToMatchResultCommand bridges a finalized oracle result to a platform MatchResult.
type BridgeOracleToMatchResultCommand struct {
	OracleResultID uuid.UUID  `json:"oracle_result_id"`
	MatchID        *uuid.UUID `json:"match_id,omitempty"` // If nil, a new Match will be created
}

func (c BridgeOracleToMatchResultCommand) Validate() error {
	if c.OracleResultID == uuid.Nil {
		return fmt.Errorf("oracle_result_id is required")
	}
	return nil
}

// GameImportEvent represents an event emitted during the import lifecycle.
type GameImportEvent struct {
	EventID         uuid.UUID              `json:"event_id"`
	ExternalMatchID string                 `json:"external_match_id"`
	GameID          replay_common.GameIDKey `json:"game_id"`
	EventType       string                 `json:"event_type"` // "discovered", "imported", "ocr_started", "bridged"
	MatchID         *uuid.UUID             `json:"match_id,omitempty"`
	OracleResultID  *uuid.UUID             `json:"oracle_result_id,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]string      `json:"metadata,omitempty"`
}
