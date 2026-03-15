package squad_entities

import (
	"time"

	"github.com/google/uuid"
)

// PlayerTeamHistoryEntry represents a player's tenure on a specific team
type PlayerTeamHistoryEntry struct {
	ID            uuid.UUID `json:"id" bson:"_id"`
	PlayerID      uuid.UUID `json:"player_id" bson:"player_id"`
	SquadID       uuid.UUID `json:"squad_id" bson:"squad_id"`
	SquadName     string    `json:"squad_name" bson:"squad_name"`
	SquadTag      string    `json:"squad_tag" bson:"squad_tag"`
	SquadLogoURI  string    `json:"squad_logo_uri,omitempty" bson:"squad_logo_uri,omitempty"`
	Role          string    `json:"role" bson:"role"`
	JoinedAt      time.Time `json:"joined_at" bson:"joined_at"`
	LeftAt        *time.Time `json:"left_at,omitempty" bson:"left_at,omitempty"` // nil = current team
	MatchesPlayed int       `json:"matches_played" bson:"matches_played"`
	WinRate       float64   `json:"win_rate" bson:"win_rate"`
	Achievements  []string  `json:"achievements" bson:"achievements"` // Titles/events won with this team
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

func (e PlayerTeamHistoryEntry) GetID() uuid.UUID {
	return e.ID
}

// IsCurrent returns whether this is the player's current team
func (e PlayerTeamHistoryEntry) IsCurrent() bool {
	return e.LeftAt == nil
}

// TenureMonths returns the tenure duration in months
func (e PlayerTeamHistoryEntry) TenureMonths() int {
	end := time.Now()
	if e.LeftAt != nil {
		end = *e.LeftAt
	}
	months := int(end.Sub(e.JoinedAt).Hours() / (24 * 30))
	if months < 1 {
		return 1
	}
	return months
}

// NewPlayerTeamHistoryEntry creates a new team history entry
func NewPlayerTeamHistoryEntry(playerID, squadID uuid.UUID, squadName, squadTag, squadLogoURI, role string) *PlayerTeamHistoryEntry {
	now := time.Now()
	return &PlayerTeamHistoryEntry{
		ID:            uuid.New(),
		PlayerID:      playerID,
		SquadID:       squadID,
		SquadName:     squadName,
		SquadTag:      squadTag,
		SquadLogoURI:  squadLogoURI,
		Role:          role,
		JoinedAt:      now,
		MatchesPlayed: 0,
		WinRate:       0,
		Achievements:  make([]string, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// TeamRosterHistoryEntry represents a roster entry for team profile display
type TeamRosterHistoryEntry struct {
	ID                 uuid.UUID  `json:"id" bson:"_id"`
	SquadID            uuid.UUID  `json:"squad_id" bson:"squad_id"`
	PlayerID           uuid.UUID  `json:"player_id" bson:"player_id"`
	PlayerNickname     string     `json:"player_nickname" bson:"player_nickname"`
	PlayerAvatar       string     `json:"player_avatar,omitempty" bson:"player_avatar,omitempty"`
	Role               string     `json:"role" bson:"role"`
	JoinedAt           time.Time  `json:"joined_at" bson:"joined_at"`
	LeftAt             *time.Time `json:"left_at,omitempty" bson:"left_at,omitempty"`
	MatchesPlayed      int        `json:"matches_played" bson:"matches_played"`
	ContributionRating float64    `json:"contribution_rating" bson:"contribution_rating"` // 0.00–2.00 (HLTV-style)
}

func (e TeamRosterHistoryEntry) GetID() uuid.UUID {
	return e.ID
}

// IsActive returns whether this member is currently active
func (e TeamRosterHistoryEntry) IsActive() bool {
	return e.LeftAt == nil
}

// NewTeamRosterHistoryEntry creates a new roster history entry
func NewTeamRosterHistoryEntry(squadID, playerID uuid.UUID, nickname, avatar, role string) *TeamRosterHistoryEntry {
	return &TeamRosterHistoryEntry{
		ID:                 uuid.New(),
		SquadID:            squadID,
		PlayerID:           playerID,
		PlayerNickname:     nickname,
		PlayerAvatar:       avatar,
		Role:               role,
		JoinedAt:           time.Now(),
		MatchesPlayed:      0,
		ContributionRating: 1.00,
	}
}
