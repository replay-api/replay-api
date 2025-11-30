package matchmaking_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_vo "github.com/replay-api/replay-api/pkg/domain/matchmaking/value-objects"
)

// PlayerSlot represents a slot in a matchmaking lobby
type PlayerSlot struct {
	SlotNumber int       `json:"slot_number" bson:"slot_number"`
	PlayerID   *uuid.UUID `json:"player_id,omitempty" bson:"player_id,omitempty"`
	IsReady    bool      `json:"is_ready" bson:"is_ready"`
	JoinedAt   time.Time `json:"joined_at,omitempty" bson:"joined_at,omitempty"`
	MMR        *int      `json:"mmr,omitempty" bson:"mmr,omitempty"`
}

// LobbyStatus represents the lifecycle of a lobby
type LobbyStatus string

const (
	LobbyStatusOpen       LobbyStatus = "open"        // Accepting players
	LobbyStatusReadyCheck LobbyStatus = "ready_check" // Countdown active
	LobbyStatusStarting   LobbyStatus = "starting"    // Creating match
	LobbyStatusStarted    LobbyStatus = "started"     // Match in progress
	LobbyStatusCancelled  LobbyStatus = "cancelled"   // Lobby closed
)

// MatchmakingLobby is the aggregate root for custom matchmaking lobbies
// It manages player slots, ready checks, and prize pool setup
type MatchmakingLobby struct {
	common.BaseEntity `bson:",inline"`

	// Lobby Configuration
	CreatorID        uuid.UUID                        `json:"creator_id" bson:"creator_id"`
	GameID           string                           `json:"game_id" bson:"game_id"`
	Region           string                           `json:"region" bson:"region"`
	Tier             string                           `json:"tier" bson:"tier"`
	DistributionRule matchmaking_vo.DistributionRule  `json:"distribution_rule" bson:"distribution_rule"`
	
	// Player Management
	MaxPlayers       int           `json:"max_players" bson:"max_players"`
	PlayerSlots      []PlayerSlot  `json:"player_slots" bson:"player_slots"`
	
	// State Management
	Status           LobbyStatus   `json:"status" bson:"status"`
	ReadyCheckStart  *time.Time    `json:"ready_check_start,omitempty" bson:"ready_check_start,omitempty"`
	ReadyCheckEnd    *time.Time    `json:"ready_check_end,omitempty" bson:"ready_check_end,omitempty"`
	MatchID          *uuid.UUID    `json:"match_id,omitempty" bson:"match_id,omitempty"`
	CancelReason     string        `json:"cancel_reason,omitempty" bson:"cancel_reason,omitempty"`
	
	// Settings
	AutoFill         bool          `json:"auto_fill" bson:"auto_fill"`         // Fill empty slots with matchmaking
	ReadyTimeout     time.Duration `json:"ready_timeout" bson:"ready_timeout"` // Default 60 seconds
	InviteOnly       bool          `json:"invite_only" bson:"invite_only"`     // Only invited players can join
}

// NewMatchmakingLobby creates a new lobby with the creator as the first player
func NewMatchmakingLobby(
	resourceOwner common.ResourceOwner,
	creatorID uuid.UUID,
	gameID string,
	region string,
	tier string,
	distributionRule matchmaking_vo.DistributionRule,
	maxPlayers int,
	autoFill bool,
	inviteOnly bool,
) (*MatchmakingLobby, error) {
	if maxPlayers < 2 {
		return nil, fmt.Errorf("lobby must allow at least 2 players")
	}
	if maxPlayers > 10 {
		return nil, fmt.Errorf("lobby cannot exceed 10 players")
	}
	if !distributionRule.IsValid() {
		return nil, fmt.Errorf("invalid distribution rule: %s", distributionRule)
	}

	// Initialize player slots
	slots := make([]PlayerSlot, maxPlayers)
	for i := 0; i < maxPlayers; i++ {
		slots[i] = PlayerSlot{
			SlotNumber: i + 1,
			PlayerID:   nil,
			IsReady:    false,
		}
	}

	// Creator takes first slot
	now := time.Now().UTC()
	slots[0].PlayerID = &creatorID
	slots[0].JoinedAt = now

	lobby := &MatchmakingLobby{
		BaseEntity:       common.NewEntity(resourceOwner),
		CreatorID:        creatorID,
		GameID:           gameID,
		Region:           region,
		Tier:             tier,
		DistributionRule: distributionRule,
		MaxPlayers:       maxPlayers,
		PlayerSlots:      slots,
		Status:           LobbyStatusOpen,
		AutoFill:         autoFill,
		InviteOnly:       inviteOnly,
		ReadyTimeout:     60 * time.Second,
	}

	return lobby, nil
}

// GetID implements common.Entity interface
func (l *MatchmakingLobby) GetID() uuid.UUID {
	return l.ID
}

// AddPlayer adds a player to the first available slot
func (l *MatchmakingLobby) AddPlayer(playerID uuid.UUID, mmr int) error {
	if l.Status != LobbyStatusOpen {
		return fmt.Errorf("lobby is not open (status: %s)", l.Status)
	}

	// Check if player already in lobby
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == playerID {
			return fmt.Errorf("player already in lobby")
		}
	}

	// Find first empty slot
	for i := range l.PlayerSlots {
		if l.PlayerSlots[i].PlayerID == nil {
			now := time.Now().UTC()
			l.PlayerSlots[i].PlayerID = &playerID
			l.PlayerSlots[i].JoinedAt = now
			l.PlayerSlots[i].MMR = &mmr
			l.PlayerSlots[i].IsReady = false
			l.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("lobby is full")
}

// RemovePlayer removes a player from the lobby
func (l *MatchmakingLobby) RemovePlayer(playerID uuid.UUID) error {
	if l.Status == LobbyStatusStarted {
		return fmt.Errorf("cannot remove player from started match")
	}

	for i := range l.PlayerSlots {
		if l.PlayerSlots[i].PlayerID != nil && *l.PlayerSlots[i].PlayerID == playerID {
			l.PlayerSlots[i].PlayerID = nil
			l.PlayerSlots[i].IsReady = false
			l.PlayerSlots[i].MMR = nil
			l.UpdatedAt = time.Now().UTC()
			
			// If creator leaves, cancel lobby
			if playerID == l.CreatorID {
				return l.Cancel("creator left lobby")
			}
			
			return nil
		}
	}

	return fmt.Errorf("player not in lobby")
}

// SetPlayerReady marks a player as ready
func (l *MatchmakingLobby) SetPlayerReady(playerID uuid.UUID, isReady bool) error {
	if l.Status != LobbyStatusOpen && l.Status != LobbyStatusReadyCheck {
		return fmt.Errorf("lobby is not open or in ready check (status: %s)", l.Status)
	}

	for i := range l.PlayerSlots {
		if l.PlayerSlots[i].PlayerID != nil && *l.PlayerSlots[i].PlayerID == playerID {
			l.PlayerSlots[i].IsReady = isReady
			l.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("player not in lobby")
}

// StartReadyCheck begins the countdown for all players to ready up
func (l *MatchmakingLobby) StartReadyCheck() error {
	if l.Status != LobbyStatusOpen {
		return fmt.Errorf("lobby is not open (status: %s)", l.Status)
	}

	// Verify lobby has enough players
	playerCount := l.GetPlayerCount()
	if playerCount < 2 {
		return fmt.Errorf("need at least 2 players to start ready check")
	}

	now := time.Now().UTC()
	endTime := now.Add(l.ReadyTimeout)
	
	l.Status = LobbyStatusReadyCheck
	l.ReadyCheckStart = &now
	l.ReadyCheckEnd = &endTime
	l.UpdatedAt = time.Now().UTC()

	return nil
}

// CheckReadyStatus verifies if all players are ready
func (l *MatchmakingLobby) CheckReadyStatus() (bool, []uuid.UUID) {
	var notReadyPlayers []uuid.UUID

	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil {
			if !slot.IsReady {
				notReadyPlayers = append(notReadyPlayers, *slot.PlayerID)
			}
		}
	}

	allReady := len(notReadyPlayers) == 0
	return allReady, notReadyPlayers
}

// StartMatch transitions the lobby to starting/started status
func (l *MatchmakingLobby) StartMatch(matchID uuid.UUID) error {
	if l.Status != LobbyStatusReadyCheck {
		return fmt.Errorf("lobby is not in ready check (status: %s)", l.Status)
	}

	allReady, notReadyPlayers := l.CheckReadyStatus()
	if !allReady {
		return fmt.Errorf("not all players are ready (%d players not ready)", len(notReadyPlayers))
	}

	l.Status = LobbyStatusStarting
	l.MatchID = &matchID
	l.UpdatedAt = time.Now().UTC()

	return nil
}

// MarkMatchStarted confirms the match has started successfully
func (l *MatchmakingLobby) MarkMatchStarted() error {
	if l.Status != LobbyStatusStarting {
		return fmt.Errorf("lobby is not starting (status: %s)", l.Status)
	}
	if l.MatchID == nil {
		return fmt.Errorf("match ID not set")
	}

	l.Status = LobbyStatusStarted
	l.UpdatedAt = time.Now().UTC()

	return nil
}

// Cancel closes the lobby with a reason
func (l *MatchmakingLobby) Cancel(reason string) error {
	if l.Status == LobbyStatusStarted {
		return fmt.Errorf("cannot cancel started match")
	}

	l.Status = LobbyStatusCancelled
	l.CancelReason = reason
	l.UpdatedAt = time.Now().UTC()

	return nil
}

// GetPlayerCount returns the number of players currently in the lobby
func (l *MatchmakingLobby) GetPlayerCount() int {
	count := 0
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil {
			count++
		}
	}
	return count
}

// GetPlayerIDs returns all player IDs in the lobby
func (l *MatchmakingLobby) GetPlayerIDs() []uuid.UUID {
	var playerIDs []uuid.UUID
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil {
			playerIDs = append(playerIDs, *slot.PlayerID)
		}
	}
	return playerIDs
}

// IsFull returns true if all slots are occupied
func (l *MatchmakingLobby) IsFull() bool {
	return l.GetPlayerCount() == l.MaxPlayers
}

// CanStart returns true if the lobby can start a match
func (l *MatchmakingLobby) CanStart() bool {
	if l.Status != LobbyStatusOpen {
		return false
	}
	playerCount := l.GetPlayerCount()
	return playerCount >= 2 && (playerCount == l.MaxPlayers || !l.AutoFill)
}

// IsReadyCheckExpired returns true if the ready check timeout has passed
func (l *MatchmakingLobby) IsReadyCheckExpired() bool {
	if l.Status != LobbyStatusReadyCheck || l.ReadyCheckEnd == nil {
		return false
	}
	return time.Now().UTC().After(*l.ReadyCheckEnd)
}

// Validate ensures the lobby state is consistent
func (l *MatchmakingLobby) Validate() error {
	if l.MaxPlayers < 2 || l.MaxPlayers > 10 {
		return fmt.Errorf("max players must be between 2 and 10")
	}
	if len(l.PlayerSlots) != l.MaxPlayers {
		return fmt.Errorf("player slots count (%d) does not match max players (%d)", len(l.PlayerSlots), l.MaxPlayers)
	}
	if !l.DistributionRule.IsValid() {
		return fmt.Errorf("invalid distribution rule: %s", l.DistributionRule)
	}

	playerCount := l.GetPlayerCount()
	if l.Status == LobbyStatusReadyCheck && playerCount < 2 {
		return fmt.Errorf("ready check requires at least 2 players")
	}
	if l.Status == LobbyStatusStarted && l.MatchID == nil {
		return fmt.Errorf("started lobby must have match ID")
	}

	return nil
}
