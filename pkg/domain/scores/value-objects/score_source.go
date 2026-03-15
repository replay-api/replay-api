package scores_vo

import "fmt"

// ScoreSource represents the origin of submitted scores
type ScoreSource string

const (
	// ScoreSourceReplayFile indicates scores were extracted from a replay file
	ScoreSourceReplayFile ScoreSource = "replay_file"

	// ScoreSourceTournamentAdmin indicates scores were submitted by a tournament administrator
	ScoreSourceTournamentAdmin ScoreSource = "tournament_admin"

	// ScoreSourceExternalAPI indicates scores came from an external API (e.g., Valve, Faceit)
	ScoreSourceExternalAPI ScoreSource = "external_api"

	// ScoreSourceConsensus indicates scores were agreed upon by participating teams
	ScoreSourceConsensus ScoreSource = "consensus"

	// ScoreSourceMatchmaking indicates scores originated from the matchmaking system
	ScoreSourceMatchmaking ScoreSource = "matchmaking"

	// ScoreSourceOracle indicates scores were produced by the oracle consensus system (OCR + external APIs)
	ScoreSourceOracle ScoreSource = "oracle"
)

// IsValid returns true if the source is a known valid value
func (s ScoreSource) IsValid() bool {
	switch s {
	case ScoreSourceReplayFile, ScoreSourceTournamentAdmin, ScoreSourceExternalAPI, ScoreSourceConsensus, ScoreSourceMatchmaking, ScoreSourceOracle:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (s ScoreSource) String() string {
	return string(s)
}

// RequiresManualVerification returns true if the score source typically requires human review
func (s ScoreSource) RequiresManualVerification() bool {
	switch s {
	case ScoreSourceTournamentAdmin, ScoreSourceConsensus:
		return false // Admin submissions are pre-verified; consensus is agreed upon
	case ScoreSourceReplayFile:
		return false // Replay parsing is deterministic
	case ScoreSourceExternalAPI:
		return false // External APIs are trusted sources
	default:
		return true
	}
}

// IsAutomated returns true if the source is programmatic rather than human-submitted
func (s ScoreSource) IsAutomated() bool {
	return s == ScoreSourceReplayFile || s == ScoreSourceExternalAPI || s == ScoreSourceMatchmaking || s == ScoreSourceOracle
}

// Validate returns an error if the score source is invalid
func (s ScoreSource) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid score source: %s", s)
	}
	return nil
}
