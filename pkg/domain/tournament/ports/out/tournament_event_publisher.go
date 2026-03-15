// Package tournament_out defines outbound interfaces for tournament domain
package tournament_out

import (
	"context"

	"github.com/google/uuid"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
)

// TournamentEventPublisher defines the contract for publishing tournament lifecycle events.
// Events are published asynchronously for downstream consumers (notifications, analytics, etc.)
//
// Event Flow:
//   - Created → RegistrationOpened → RegistrationClosed → Started → MatchCompleted → Completed
//   - At any point before completion: Cancelled
//   - During registration: PlayerRegistered, PlayerUnregistered
//   - During in_progress: MatchResultRecorded, BracketAdvanced
type TournamentEventPublisher interface {
	// PublishTournamentCreated fires when a new tournament is created
	PublishTournamentCreated(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishRegistrationOpened fires when registration opens for a tournament
	PublishRegistrationOpened(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishRegistrationClosed fires when registration closes
	PublishRegistrationClosed(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishTournamentStarted fires when the tournament begins
	PublishTournamentStarted(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishTournamentCompleted fires when all matches are done and winners determined
	PublishTournamentCompleted(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishTournamentCancelled fires when a tournament is cancelled
	PublishTournamentCancelled(ctx context.Context, tournament *tournament_entities.Tournament) error

	// PublishPlayerRegistered fires when a player joins a tournament
	PublishPlayerRegistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error

	// PublishPlayerUnregistered fires when a player leaves a tournament
	PublishPlayerUnregistered(ctx context.Context, tournament *tournament_entities.Tournament, playerID uuid.UUID) error

	// PublishMatchResultRecorded fires when a match result is recorded within a tournament
	PublishMatchResultRecorded(ctx context.Context, tournament *tournament_entities.Tournament, matchID uuid.UUID, winnerID uuid.UUID) error

	// PublishBracketAdvanced fires when the bracket advances to the next round
	PublishBracketAdvanced(ctx context.Context, tournament *tournament_entities.Tournament, newMatches []tournament_entities.TournamentMatch) error
}
