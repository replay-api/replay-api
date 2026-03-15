package oracle_out

import (
	"context"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// TeamRef represents a resolved team identity from fuzzy name matching
type TeamRef struct {
	TeamID      uuid.UUID `json:"team_id"`
	MatchedName string    `json:"matched_name"` // The canonical name that was matched
	Confidence  float64   `json:"confidence"`   // 0.0-1.0 match quality
}

// TeamResolverPort defines the contract for resolving team names to team IDs
type TeamResolverPort interface {
	// ResolveTeam attempts to match a raw (possibly OCR-extracted) team name to a known team
	ResolveTeam(ctx context.Context, name string, gameID replay_common.GameIDKey) (*TeamRef, error)
}
