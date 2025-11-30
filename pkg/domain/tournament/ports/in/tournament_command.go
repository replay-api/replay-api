// Package tournament_in defines inbound command and query interfaces
package tournament_in

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
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

// Validate validates the CreateTournamentCommand
func (c *CreateTournamentCommand) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("name is required")
	}
	if len(c.Name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}
	if c.GameID == "" {
		return errors.New("game_id is required")
	}
	if c.GameMode == "" {
		return errors.New("game_mode is required")
	}
	if c.Region == "" {
		return errors.New("region is required")
	}
	if c.MaxParticipants < 2 {
		return errors.New("max_participants must be at least 2")
	}
	if c.MinParticipants < 2 {
		return errors.New("min_participants must be at least 2")
	}
	if c.MinParticipants > c.MaxParticipants {
		return errors.New("min_participants cannot exceed max_participants")
	}
	if c.OrganizerID == uuid.Nil {
		return errors.New("organizer_id is required")
	}
	if c.StartTime.IsZero() {
		return errors.New("start_time is required")
	}
	if c.RegistrationClose.IsZero() {
		return errors.New("registration_close is required")
	}
	if c.RegistrationClose.After(c.StartTime) {
		return errors.New("registration_close must be before start_time")
	}
	return nil
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

// Validate validates the UpdateTournamentCommand
func (c *UpdateTournamentCommand) Validate() error {
	if c.TournamentID == uuid.Nil {
		return errors.New("tournament_id is required")
	}
	if c.Name != nil && strings.TrimSpace(*c.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if c.Name != nil && len(*c.Name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}
	if c.MaxParticipants != nil && *c.MaxParticipants < 2 {
		return errors.New("max_participants must be at least 2")
	}
	return nil
}

// RegisterPlayerCommand represents a player registration
type RegisterPlayerCommand struct {
	TournamentID uuid.UUID
	PlayerID     uuid.UUID
	DisplayName  string
}

// Validate validates the RegisterPlayerCommand
func (c *RegisterPlayerCommand) Validate() error {
	if c.TournamentID == uuid.Nil {
		return errors.New("tournament_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	if strings.TrimSpace(c.DisplayName) == "" {
		return errors.New("display_name is required")
	}
	if len(c.DisplayName) > 50 {
		return errors.New("display_name cannot exceed 50 characters")
	}
	return nil
}

// UnregisterPlayerCommand represents a player unregistration
type UnregisterPlayerCommand struct {
	TournamentID uuid.UUID
	PlayerID     uuid.UUID
}

// Validate validates the UnregisterPlayerCommand
func (c *UnregisterPlayerCommand) Validate() error {
	if c.TournamentID == uuid.Nil {
		return errors.New("tournament_id is required")
	}
	if c.PlayerID == uuid.Nil {
		return errors.New("player_id is required")
	}
	return nil
}

// CompleteTournamentCommand represents tournament completion
type CompleteTournamentCommand struct {
	TournamentID uuid.UUID
	Winners      []tournament_entities.TournamentWinner
}

// Validate validates the CompleteTournamentCommand
func (c *CompleteTournamentCommand) Validate() error {
	if c.TournamentID == uuid.Nil {
		return errors.New("tournament_id is required")
	}
	if len(c.Winners) == 0 {
		return errors.New("winners is required")
	}
	return nil
}

// CancelTournamentCommand represents tournament cancellation
type CancelTournamentCommand struct {
	TournamentID uuid.UUID
	Reason       string
}

// Validate validates the CancelTournamentCommand
func (c *CancelTournamentCommand) Validate() error {
	if c.TournamentID == uuid.Nil {
		return errors.New("tournament_id is required")
	}
	if strings.TrimSpace(c.Reason) == "" {
		return errors.New("reason is required")
	}
	if len(c.Reason) > 500 {
		return errors.New("reason cannot exceed 500 characters")
	}
	return nil
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
