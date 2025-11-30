package tournament_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// Tournament is an aggregate root representing a competitive gaming tournament
type Tournament struct {
	common.BaseEntity
	Name              string                `json:"name" bson:"name"`
	Description       string                `json:"description" bson:"description"`
	GameID            common.GameIDKey      `json:"game_id" bson:"game_id"`
	GameMode          string                `json:"game_mode" bson:"game_mode"`
	Region            string                `json:"region" bson:"region"`
	Format            TournamentFormat      `json:"format" bson:"format"`                 // SingleElimination, DoubleElimination, RoundRobin, Swiss
	MaxParticipants   int                   `json:"max_participants" bson:"max_participants"`
	MinParticipants   int                   `json:"min_participants" bson:"min_participants"`
	EntryFee          wallet_vo.Amount      `json:"entry_fee" bson:"entry_fee"`
	Currency          wallet_vo.Currency    `json:"currency" bson:"currency"`
	PrizePool         wallet_vo.Amount      `json:"prize_pool" bson:"prize_pool"`
	Status            TournamentStatus      `json:"status" bson:"status"`
	StartTime         time.Time             `json:"start_time" bson:"start_time"`
	EndTime           *time.Time            `json:"end_time,omitempty" bson:"end_time,omitempty"`
	RegistrationOpen  time.Time             `json:"registration_open" bson:"registration_open"`
	RegistrationClose time.Time             `json:"registration_close" bson:"registration_close"`
	Participants      []TournamentPlayer    `json:"participants" bson:"participants"`
	Matches           []TournamentMatch     `json:"matches,omitempty" bson:"matches,omitempty"`
	Winners           []TournamentWinner    `json:"winners,omitempty" bson:"winners,omitempty"`
	Rules             TournamentRules       `json:"rules" bson:"rules"`
	OrganizerID       uuid.UUID             `json:"organizer_id" bson:"organizer_id"`
}

// TournamentStatus represents the lifecycle state of a tournament
type TournamentStatus string

const (
	TournamentStatusDraft        TournamentStatus = "draft"         // Being configured by organizer
	TournamentStatusRegistration TournamentStatus = "registration"  // Open for player registration
	TournamentStatusReady        TournamentStatus = "ready"         // Registration closed, waiting for start time
	TournamentStatusInProgress   TournamentStatus = "in_progress"   // Matches being played
	TournamentStatusCompleted    TournamentStatus = "completed"     // All matches finished
	TournamentStatusCancelled    TournamentStatus = "cancelled"     // Tournament cancelled
)

// TournamentFormat represents the tournament structure
type TournamentFormat string

const (
	TournamentFormatSingleElimination TournamentFormat = "single_elimination"
	TournamentFormatDoubleElimination TournamentFormat = "double_elimination"
	TournamentFormatRoundRobin        TournamentFormat = "round_robin"
	TournamentFormatSwiss             TournamentFormat = "swiss"
)

// TournamentPlayer represents a registered participant
type TournamentPlayer struct {
	PlayerID     uuid.UUID `json:"player_id" bson:"player_id"`
	DisplayName  string    `json:"display_name" bson:"display_name"`
	RegisteredAt time.Time `json:"registered_at" bson:"registered_at"`
	Seed         int       `json:"seed,omitempty" bson:"seed,omitempty"` // For seeded tournaments
	Status       string    `json:"status" bson:"status"`                 // registered, checked_in, disqualified
}

// TournamentMatch represents a match within the tournament
type TournamentMatch struct {
	MatchID     uuid.UUID   `json:"match_id" bson:"match_id"`
	Round       int         `json:"round" bson:"round"`
	BracketPos  string      `json:"bracket_pos,omitempty" bson:"bracket_pos,omitempty"` // e.g., "winners_r1_m1", "losers_r2_m3"
	Player1ID   uuid.UUID   `json:"player1_id" bson:"player1_id"`
	Player2ID   uuid.UUID   `json:"player2_id" bson:"player2_id"`
	WinnerID    *uuid.UUID  `json:"winner_id,omitempty" bson:"winner_id,omitempty"`
	ScheduledAt time.Time   `json:"scheduled_at" bson:"scheduled_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	Status      MatchStatus `json:"status" bson:"status"`
}

// MatchStatus represents the state of a tournament match
type MatchStatus string

const (
	MatchStatusScheduled  MatchStatus = "scheduled"
	MatchStatusInProgress MatchStatus = "in_progress"
	MatchStatusCompleted  MatchStatus = "completed"
	MatchStatusCancelled  MatchStatus = "cancelled"
)

// TournamentWinner represents a prize recipient
type TournamentWinner struct {
	PlayerID  uuid.UUID        `json:"player_id" bson:"player_id"`
	Placement int              `json:"placement" bson:"placement"` // 1 = 1st place, 2 = 2nd, etc.
	Prize     wallet_vo.Amount `json:"prize" bson:"prize"`
	PaidAt    *time.Time       `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
}

// TournamentRules represents tournament-specific rules and settings
type TournamentRules struct {
	BestOf              int      `json:"best_of" bson:"best_of"`                               // Best of N games
	MapPool             []string `json:"map_pool,omitempty" bson:"map_pool,omitempty"`         // Available maps
	BanPickEnabled      bool     `json:"ban_pick_enabled" bson:"ban_pick_enabled"`             // Map ban/pick phase
	CheckInRequired     bool     `json:"check_in_required" bson:"check_in_required"`           // Players must check in
	CheckInWindowMins   int      `json:"check_in_window_mins,omitempty" bson:"check_in_window_mins,omitempty"`
	MatchTimeoutMins    int      `json:"match_timeout_mins" bson:"match_timeout_mins"`         // Time to complete match
	DisconnectGraceMins int      `json:"disconnect_grace_mins" bson:"disconnect_grace_mins"`   // Disconnect grace period
}

// NewTournament creates a new tournament
func NewTournament(
	resourceOwner common.ResourceOwner,
	name, description string,
	gameID common.GameIDKey,
	gameMode, region string,
	format TournamentFormat,
	maxParticipants, minParticipants int,
	entryFee wallet_vo.Amount,
	currency wallet_vo.Currency,
	startTime, registrationOpen, registrationClose time.Time,
	rules TournamentRules,
	organizerID uuid.UUID,
) (*Tournament, error) {
	if maxParticipants < minParticipants {
		return nil, fmt.Errorf("max_participants must be >= min_participants")
	}

	if registrationClose.After(startTime) {
		return nil, fmt.Errorf("registration must close before tournament start")
	}

	if registrationOpen.After(registrationClose) {
		return nil, fmt.Errorf("registration open must be before registration close")
	}

	baseEntity := common.NewUnrestrictedEntity(resourceOwner) // Tournaments are public

	// Calculate initial prize pool (platform may add seed money)
	platformSeed := wallet_vo.NewAmount(0) // Could be configurable
	prizePool := platformSeed

	return &Tournament{
		BaseEntity:        baseEntity,
		Name:              name,
		Description:       description,
		GameID:            gameID,
		GameMode:          gameMode,
		Region:            region,
		Format:            format,
		MaxParticipants:   maxParticipants,
		MinParticipants:   minParticipants,
		EntryFee:          entryFee,
		Currency:          currency,
		PrizePool:         prizePool,
		Status:            TournamentStatusDraft,
		StartTime:         startTime,
		RegistrationOpen:  registrationOpen,
		RegistrationClose: registrationClose,
		Participants:      make([]TournamentPlayer, 0),
		Matches:           make([]TournamentMatch, 0),
		Winners:           make([]TournamentWinner, 0),
		Rules:             rules,
		OrganizerID:       organizerID,
	}, nil
}

// RegisterPlayer adds a player to the tournament
func (t *Tournament) RegisterPlayer(playerID uuid.UUID, displayName string) error {
	if t.Status != TournamentStatusDraft && t.Status != TournamentStatusRegistration {
		return fmt.Errorf("registration is not open, status: %s", t.Status)
	}

	now := time.Now()
	if now.Before(t.RegistrationOpen) {
		return fmt.Errorf("registration has not opened yet")
	}

	if now.After(t.RegistrationClose) {
		return fmt.Errorf("registration has closed")
	}

	if len(t.Participants) >= t.MaxParticipants {
		return fmt.Errorf("tournament is full (%d/%d)", len(t.Participants), t.MaxParticipants)
	}

	// Check if player already registered
	for _, p := range t.Participants {
		if p.PlayerID == playerID {
			return fmt.Errorf("player already registered")
		}
	}

	participant := TournamentPlayer{
		PlayerID:     playerID,
		DisplayName:  displayName,
		RegisteredAt: now,
		Status:       "registered",
	}

	t.Participants = append(t.Participants, participant)

	// Add entry fee to prize pool
	if !t.EntryFee.IsZero() {
		t.PrizePool = t.PrizePool.Add(t.EntryFee)
	}

	t.UpdatedAt = now

	return nil
}

// UnregisterPlayer removes a player from the tournament
func (t *Tournament) UnregisterPlayer(playerID uuid.UUID) error {
	if t.Status != TournamentStatusDraft && t.Status != TournamentStatusRegistration {
		return fmt.Errorf("cannot unregister after registration closes, status: %s", t.Status)
	}

	for i, p := range t.Participants {
		if p.PlayerID == playerID {
			// Remove player
			t.Participants = append(t.Participants[:i], t.Participants[i+1:]...)

			// Refund entry fee from prize pool
			if !t.EntryFee.IsZero() {
				t.PrizePool = t.PrizePool.Subtract(t.EntryFee)
			}

			t.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("player not found in participants")
}

// OpenRegistration opens the tournament for player registration
func (t *Tournament) OpenRegistration() error {
	if t.Status != TournamentStatusDraft {
		return fmt.Errorf("can only open registration from draft status, current: %s", t.Status)
	}

	now := time.Now()
	if now.Before(t.RegistrationOpen) {
		return fmt.Errorf("registration open time has not been reached yet")
	}

	t.Status = TournamentStatusRegistration
	t.UpdatedAt = now

	return nil
}

// CloseRegistration closes registration and prepares tournament for start
func (t *Tournament) CloseRegistration() error {
	if t.Status != TournamentStatusRegistration {
		return fmt.Errorf("can only close registration from registration status, current: %s", t.Status)
	}

	if len(t.Participants) < t.MinParticipants {
		return fmt.Errorf("not enough participants: %d (min: %d)", len(t.Participants), t.MinParticipants)
	}

	t.Status = TournamentStatusReady
	t.UpdatedAt = time.Now()

	return nil
}

// Start begins the tournament and generates initial matches
func (t *Tournament) Start() error {
	if t.Status != TournamentStatusReady {
		return fmt.Errorf("can only start from ready status, current: %s", t.Status)
	}

	now := time.Now()
	if now.Before(t.StartTime) {
		return fmt.Errorf("tournament start time has not been reached")
	}

	// Generate bracket matches based on format
	// (Implementation would depend on format - single elimination, etc.)
	// For now, just change status

	t.Status = TournamentStatusInProgress
	t.UpdatedAt = now

	return nil
}

// Complete marks the tournament as finished
func (t *Tournament) Complete(winners []TournamentWinner) error {
	if t.Status != TournamentStatusInProgress {
		return fmt.Errorf("can only complete from in_progress status, current: %s", t.Status)
	}

	now := time.Now()
	t.Status = TournamentStatusCompleted
	t.Winners = winners
	t.EndTime = &now
	t.UpdatedAt = now

	return nil
}

// Cancel cancels the tournament (refunds should be issued)
func (t *Tournament) Cancel(reason string) error {
	if t.Status == TournamentStatusCompleted {
		return fmt.Errorf("cannot cancel completed tournament")
	}

	now := time.Now()
	t.Status = TournamentStatusCancelled
	t.EndTime = &now
	t.UpdatedAt = now

	return nil
}

// GetParticipantCount returns the current number of registered players
func (t *Tournament) GetParticipantCount() int {
	return len(t.Participants)
}

// IsFull returns true if tournament has reached max participants
func (t *Tournament) IsFull() bool {
	return len(t.Participants) >= t.MaxParticipants
}

// IsPlayerRegistered checks if a player is registered
func (t *Tournament) IsPlayerRegistered(playerID uuid.UUID) bool {
	for _, p := range t.Participants {
		if p.PlayerID == playerID {
			return true
		}
	}
	return false
}

// Validate ensures tournament invariants
func (t *Tournament) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("tournament name cannot be empty")
	}

	if t.MaxParticipants < t.MinParticipants {
		return fmt.Errorf("max_participants must be >= min_participants")
	}

	if t.MinParticipants < 2 {
		return fmt.Errorf("min_participants must be at least 2")
	}

	if !t.Currency.IsValid() {
		return fmt.Errorf("invalid currency: %s", t.Currency)
	}

	if t.EntryFee.IsNegative() {
		return fmt.Errorf("entry fee cannot be negative")
	}

	return nil
}
