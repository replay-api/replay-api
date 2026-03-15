package scores_out

import (
	"context"

	"github.com/google/uuid"
)

// ScoreAuthorization defines the outbound port for score-related authorization checks.
// This port enables the scores domain to verify permissions without directly depending
// on the tournament or matchmaking domains, maintaining Clean Architecture boundaries.
//
// Permission Model (Financial-Grade):
//   - Submit: Tournament organizer OR platform admin (tournament_admin source);
//     Replay pipeline (system-level, replay_file source); External API (system-level)
//   - Verify: Tournament organizer for their tournament scores, OR platform admin
//   - Dispute: Match participant (player in team_results.players or player_results)
//   - Conciliate: Tournament organizer for their tournament scores, OR platform admin
//   - Finalize: Tournament organizer for their tournament scores, OR platform admin;
//     Auto-finalization by system after 72h dispute window
//   - Cancel: Tournament organizer for their tournament scores, OR platform admin
type ScoreAuthorization interface {
	// IsTournamentOrganizer checks if the given user is the organizer of the specified tournament
	IsTournamentOrganizer(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error)

	// IsMatchParticipant checks if the given user is a participant in the match
	// by looking at player_results or team_results.players in the match result
	IsMatchParticipant(ctx context.Context, userID uuid.UUID, matchResultID uuid.UUID) (bool, error)

	// IsPlatformAdmin checks if the current context has platform admin privileges
	IsPlatformAdmin(ctx context.Context) bool
}
