package scores_out

import (
	"context"

	"github.com/google/uuid"
)

// TournamentMatchCallback defines the outbound port for notifying the tournament domain
// when a score-related event occurs that affects a tournament match.
// This enables the scores domain to trigger bracket advancement without coupling
// directly to the tournament domain (Clean Architecture boundary).
//
// Financial-Grade Considerations:
//   - Callback only fires AFTER the match result is fully finalized (post dispute window)
//   - WinnerID is derived from finalized match result rankings (immutable at this point)
//   - If callback fails, match result finalization still succeeds (eventual consistency)
//   - Audit trail: the finalized match result ID links back to the tournament match
type TournamentMatchCallback interface {
	// OnMatchResultFinalized notifies the tournament domain that a match result
	// associated with a tournament has been finalized, providing the winner.
	// The tournament domain should update the TournamentMatch.WinnerID and
	// optionally advance the bracket.
	OnMatchResultFinalized(ctx context.Context, tournamentID uuid.UUID, matchID uuid.UUID, winnerID uuid.UUID) error
}
