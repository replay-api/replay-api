// Package tournament_in defines inbound command and query interfaces
package tournament_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	tournament_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/entities"
	wallet_vo "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/value-objects"
)

// CreateTournamentCommand represents the data needed to create a tournament
type CreateTournamentCommand struct {
	ResourceOwner     common.ResourceOwner
	Name              string
	Description       string
	GameID            common.GameIDKey
	GameMode          string
	Region            string
	Format            tournament_entities.TournamentFormat
	MaxParticipants   int
	MinParticipants   int
	EntryFee          wallet_vo.Amount
	Currency          wallet_vo.Currency
	StartTime         time.Time
	RegistrationOpen  time.Time
	RegistrationClose time.Time
	Rules             tournament_entities.TournamentRules
	OrganizerID       uuid.UUID
}

// UpdateTournamentCommand represents updates to a tournament
type UpdateTournamentCommand struct {
	TournamentID      uuid.UUID
	Name              *string
	Description       *string
	MaxParticipants   *int
	StartTime         *time.Time
	RegistrationClose *time.Time
	Rules             *tournament_entities.TournamentRules
}

// RegisterPlayerCommand represents a player registration
type RegisterPlayerCommand struct {
	TournamentID uuid.UUID
	PlayerID     uuid.UUID
	DisplayName  string
}

// UnregisterPlayerCommand represents a player unregistration
type UnregisterPlayerCommand struct {
	TournamentID uuid.UUID
	PlayerID     uuid.UUID
}

// CompleteTournamentCommand represents tournament completion
type CompleteTournamentCommand struct {
	TournamentID uuid.UUID
	Winners      []tournament_entities.TournamentWinner
}

// CancelTournamentCommand represents tournament cancellation
type CancelTournamentCommand struct {
	TournamentID uuid.UUID
	Reason       string
}

// TournamentCommand defines operations for managing tournaments
type TournamentCommand interface {
	// CreateTournament creates a new tournament
	CreateTournament(ctx context.Context, cmd CreateTournamentCommand) (*tournament_entities.Tournament, error)

	// UpdateTournament updates tournament details (only before start)
	UpdateTournament(ctx context.Context, cmd UpdateTournamentCommand) (*tournament_entities.Tournament, error)

	// DeleteTournament removes a tournament (only in draft/registration)
	DeleteTournament(ctx context.Context, tournamentID uuid.UUID) error

	// RegisterPlayer registers a player for the tournament
	RegisterPlayer(ctx context.Context, cmd RegisterPlayerCommand) error

	// UnregisterPlayer removes a player from the tournament
	UnregisterPlayer(ctx context.Context, cmd UnregisterPlayerCommand) error

	// OpenRegistration opens player registration
	OpenRegistration(ctx context.Context, tournamentID uuid.UUID) error

	// CloseRegistration closes player registration
	CloseRegistration(ctx context.Context, tournamentID uuid.UUID) error

	// StartTournament begins the tournament
	StartTournament(ctx context.Context, tournamentID uuid.UUID) error

	// CompleteTournament marks tournament as completed with winners
	CompleteTournament(ctx context.Context, cmd CompleteTournamentCommand) error

	// CancelTournament cancels the tournament
	CancelTournament(ctx context.Context, cmd CancelTournamentCommand) error
}

// TournamentReader defines query operations for tournaments
type TournamentReader interface {
	// GetTournament retrieves a tournament by ID
	GetTournament(ctx context.Context, id uuid.UUID) (*tournament_entities.Tournament, error)

	// ListTournaments retrieves tournaments with filters
	ListTournaments(ctx context.Context, gameID, region string, status []tournament_entities.TournamentStatus, limit int) ([]*tournament_entities.Tournament, error)

	// GetUpcomingTournaments retrieves tournaments accepting registrations or starting soon
	GetUpcomingTournaments(ctx context.Context, gameID string, limit int) ([]*tournament_entities.Tournament, error)

	// GetPlayerTournaments retrieves tournaments a player is registered in
	GetPlayerTournaments(ctx context.Context, playerID uuid.UUID) ([]*tournament_entities.Tournament, error)

	// GetOrganizerTournaments retrieves tournaments created by a specific organizer
	GetOrganizerTournaments(ctx context.Context, organizerID uuid.UUID) ([]*tournament_entities.Tournament, error)
}
