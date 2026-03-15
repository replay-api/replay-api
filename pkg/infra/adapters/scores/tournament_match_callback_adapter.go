package scores

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
)

// tournamentMatchCallbackAdapter implements TournamentMatchCallback
// by delegating to the tournament command service to record match results
// and advance brackets.
type tournamentMatchCallbackAdapter struct {
	tournamentCommand tournament_in.TournamentCommand
}

// NewTournamentMatchCallbackAdapter creates a new callback adapter
func NewTournamentMatchCallbackAdapter(
	tournamentCommand tournament_in.TournamentCommand,
) scores_out.TournamentMatchCallback {
	return &tournamentMatchCallbackAdapter{
		tournamentCommand: tournamentCommand,
	}
}

// OnMatchResultFinalized records the winner in the tournament match and advances the bracket
func (a *tournamentMatchCallbackAdapter) OnMatchResultFinalized(ctx context.Context, tournamentID uuid.UUID, matchID uuid.UUID, winnerID uuid.UUID) error {
	slog.InfoContext(ctx, "tournament match callback: recording finalized result",
		"tournament_id", tournamentID, "match_id", matchID, "winner_id", winnerID)

	// Record the match result in the tournament
	if err := a.tournamentCommand.RecordMatchResult(ctx, tournamentID, matchID, winnerID); err != nil {
		slog.ErrorContext(ctx, "tournament match callback: failed to record match result",
			"tournament_id", tournamentID, "match_id", matchID, "error", err)
		return fmt.Errorf("failed to record tournament match result: %w", err)
	}

	// Attempt to advance the bracket (may fail if not all matches in round are complete - that's OK)
	if err := a.tournamentCommand.AdvanceBracket(ctx, tournamentID); err != nil {
		slog.InfoContext(ctx, "tournament match callback: bracket advancement not ready (expected)",
			"tournament_id", tournamentID, "info", err.Error())
		// Not a failure - bracket may not be ready to advance yet
	}

	slog.InfoContext(ctx, "tournament match callback: completed successfully",
		"tournament_id", tournamentID, "match_id", matchID)
	return nil
}
