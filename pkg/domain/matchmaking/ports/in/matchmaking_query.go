// Package matchmaking_in defines inbound query interfaces for matchmaking
package matchmaking_in

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
)

// --- Validation Error ---

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// --- Query Definitions ---

// GetLobbyQuery retrieves a single lobby by ID
type GetLobbyQuery struct {
	LobbyID uuid.UUID
	UserID  uuid.UUID // For authorization
}

func (q *GetLobbyQuery) Validate() error {
	if q.LobbyID == uuid.Nil {
		return &ValidationError{Field: "lobby_id", Message: "lobby_id is required"}
	}
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// GetUserLobbiesQuery retrieves lobbies created by a user
type GetUserLobbiesQuery struct {
	UserID  uuid.UUID
	Filters LobbyQueryFilters
}

func (q *GetUserLobbiesQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// SearchLobbiesQuery searches for open lobbies
type SearchLobbiesQuery struct {
	UserID  uuid.UUID // For rate limiting/access
	Filters LobbySearchFilters
}

func (q *SearchLobbiesQuery) Validate() error {
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// GetSessionQuery retrieves a matchmaking session
type GetSessionQuery struct {
	SessionID uuid.UUID
	UserID    uuid.UUID // For authorization - must be session owner
}

func (q *GetSessionQuery) Validate() error {
	if q.SessionID == uuid.Nil {
		return &ValidationError{Field: "session_id", Message: "session_id is required"}
	}
	if q.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Message: "user_id is required"}
	}
	return nil
}

// GetPoolStatsQuery retrieves matchmaking pool statistics
type GetPoolStatsQuery struct {
	GameID   string
	GameMode string
	Region   string
}

func (q *GetPoolStatsQuery) Validate() error {
	if q.GameID == "" {
		return &ValidationError{Field: "game_id", Message: "game_id is required"}
	}
	return nil
}

// --- Query Filters ---

// LobbyQueryFilters for filtering lobby queries
type LobbyQueryFilters struct {
	Status    *matchmaking_entities.LobbyStatus
	GameID    *string
	Region    *string
	Limit     int
	Offset    int
	SortBy    string // created_at, player_count
	SortOrder string // asc, desc
}

// LobbySearchFilters for searching open lobbies
type LobbySearchFilters struct {
	GameID     string
	Region     string
	Tier       string
	MinPlayers *int
	MaxPlayers *int
	Limit      int
}

// --- DTOs ---

// LobbyDTO represents a lobby for external use
type LobbyDTO struct {
	ID               string            `json:"id"`
	CreatorID        string            `json:"creator_id"`
	GameID           string            `json:"game_id"`
	Region           string            `json:"region"`
	Tier             string            `json:"tier"`
	DistributionRule string            `json:"distribution_rule"`
	MaxPlayers       int               `json:"max_players"`
	PlayerCount      int               `json:"player_count"`
	Players          []PlayerSlotDTO   `json:"players"`
	Status           string            `json:"status"`
	ReadyCheckStart  *time.Time        `json:"ready_check_start,omitempty"`
	ReadyCheckEnd    *time.Time        `json:"ready_check_end,omitempty"`
	MatchID          *string           `json:"match_id,omitempty"`
	AutoFill         bool              `json:"auto_fill"`
	InviteOnly       bool              `json:"invite_only"`
	CanStart         bool              `json:"can_start"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// PlayerSlotDTO represents a player slot in a lobby
type PlayerSlotDTO struct {
	SlotNumber int        `json:"slot_number"`
	PlayerID   *string    `json:"player_id,omitempty"`
	IsReady    bool       `json:"is_ready"`
	JoinedAt   *time.Time `json:"joined_at,omitempty"`
	MMR        *int       `json:"mmr,omitempty"`
}

// SessionDTO represents a matchmaking session for external use
type SessionDTO struct {
	ID            string             `json:"id"`
	PlayerID      string             `json:"player_id"`
	SquadID       *string            `json:"squad_id,omitempty"`
	Status        string             `json:"status"`
	GameID        string             `json:"game_id"`
	GameMode      string             `json:"game_mode"`
	Region        string             `json:"region"`
	Tier          string             `json:"tier"`
	PlayerMMR     int                `json:"player_mmr"`
	EstimatedWait int                `json:"estimated_wait_seconds"`
	QueuePosition int                `json:"queue_position,omitempty"`
	MatchID       *string            `json:"match_id,omitempty"`
	QueuedAt      time.Time          `json:"queued_at"`
	MatchedAt     *time.Time         `json:"matched_at,omitempty"`
	ExpiresAt     time.Time          `json:"expires_at"`
}

// PoolStatsDTO represents matchmaking pool statistics
type PoolStatsDTO struct {
	PoolID                string            `json:"pool_id"`
	GameID                string            `json:"game_id"`
	GameMode              string            `json:"game_mode"`
	Region                string            `json:"region"`
	TotalPlayers          int               `json:"total_players"`
	AverageWaitSeconds    int               `json:"average_wait_time_seconds"`
	PlayersByTier         map[string]int    `json:"players_by_tier"`
	EstimatedMatchSeconds int               `json:"estimated_match_time_seconds"`
	QueueHealth           string            `json:"queue_health"` // healthy, moderate, slow, degraded
	Timestamp             time.Time         `json:"timestamp"`
}

// LobbiesResult is a paginated result of lobbies
type LobbiesResult struct {
	Lobbies    []LobbyDTO `json:"lobbies"`
	TotalCount int        `json:"total_count"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// --- Query Interface ---

// LobbyQuery defines read operations for lobbies
type LobbyQuery interface {
	GetLobby(ctx context.Context, query GetLobbyQuery) (*LobbyDTO, error)
	GetUserLobbies(ctx context.Context, query GetUserLobbiesQuery) (*LobbiesResult, error)
	SearchLobbies(ctx context.Context, query SearchLobbiesQuery) (*LobbiesResult, error)
}

// MatchmakingQuery defines read operations for matchmaking sessions
type MatchmakingQuery interface {
	GetSession(ctx context.Context, query GetSessionQuery) (*SessionDTO, error)
	GetPoolStats(ctx context.Context, query GetPoolStatsQuery) (*PoolStatsDTO, error)
}

// --- Helper Functions ---

// LobbyToDTO converts a Lobby entity to DTO
func LobbyToDTO(lobby *matchmaking_entities.MatchmakingLobby) LobbyDTO {
	players := make([]PlayerSlotDTO, len(lobby.PlayerSlots))
	for i, slot := range lobby.PlayerSlots {
		var playerID *string
		if slot.PlayerID != nil {
			id := slot.PlayerID.String()
			playerID = &id
		}
		var joinedAt *time.Time
		if !slot.JoinedAt.IsZero() {
			joinedAt = &slot.JoinedAt
		}
		players[i] = PlayerSlotDTO{
			SlotNumber: slot.SlotNumber,
			PlayerID:   playerID,
			IsReady:    slot.IsReady,
			JoinedAt:   joinedAt,
			MMR:        slot.MMR,
		}
	}

	var matchID *string
	if lobby.MatchID != nil {
		id := lobby.MatchID.String()
		matchID = &id
	}

	return LobbyDTO{
		ID:               lobby.ID.String(),
		CreatorID:        lobby.CreatorID.String(),
		GameID:           lobby.GameID,
		Region:           lobby.Region,
		Tier:             lobby.Tier,
		DistributionRule: string(lobby.DistributionRule),
		MaxPlayers:       lobby.MaxPlayers,
		PlayerCount:      lobby.GetPlayerCount(),
		Players:          players,
		Status:           string(lobby.Status),
		ReadyCheckStart:  lobby.ReadyCheckStart,
		ReadyCheckEnd:    lobby.ReadyCheckEnd,
		MatchID:          matchID,
		AutoFill:         lobby.AutoFill,
		InviteOnly:       lobby.InviteOnly,
		CanStart:         lobby.CanStart(),
		CreatedAt:        lobby.CreatedAt,
		UpdatedAt:        lobby.UpdatedAt,
	}
}

// SessionToDTO converts a MatchmakingSession entity to DTO
func SessionToDTO(session *matchmaking_entities.MatchmakingSession) SessionDTO {
	var squadID *string
	if session.SquadID != nil {
		id := session.SquadID.String()
		squadID = &id
	}

	var matchID *string
	if session.MatchID != nil {
		id := session.MatchID.String()
		matchID = &id
	}

	return SessionDTO{
		ID:            session.ID.String(),
		PlayerID:      session.PlayerID.String(),
		SquadID:       squadID,
		Status:        string(session.Status),
		GameID:        session.Preferences.GameID,
		GameMode:      session.Preferences.GameMode,
		Region:        session.Preferences.Region,
		Tier:          string(session.Preferences.Tier),
		PlayerMMR:     session.PlayerMMR,
		EstimatedWait: session.EstimatedWait,
		MatchID:       matchID,
		QueuedAt:      session.QueuedAt,
		MatchedAt:     session.MatchedAt,
		ExpiresAt:     session.ExpiresAt,
	}
}

// ValidateLobbyFilters ensures filter values are valid
func ValidateLobbyFilters(f *LobbyQueryFilters) error {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
	validSortBy := map[string]bool{"created_at": true, "player_count": true, "updated_at": true}
	if !validSortBy[f.SortBy] {
		return errors.New("invalid sort_by field")
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		return errors.New("sort_order must be 'asc' or 'desc'")
	}
	return nil
}

// ValidateLobbySearchFilters ensures search filter values are valid
func ValidateLobbySearchFilters(f *LobbySearchFilters) error {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 50 {
		f.Limit = 50
	}
	return nil
}
