package squad_entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// TraitTier represents the difficulty tier of an auto-awarded trait
type TraitTier string

const (
	TraitTierBronze  TraitTier = "bronze"
	TraitTierSilver  TraitTier = "silver"
	TraitTierGold    TraitTier = "gold"
	TraitTierDiamond TraitTier = "diamond"
)

// PlayerTrait represents a professional trait, auto-awarded and endorseable
type PlayerTrait struct {
	ID               uuid.UUID              `json:"id" bson:"_id"`
	PlayerID         uuid.UUID              `json:"player_id" bson:"player_id"`
	GameID           replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	TraitKey         string                 `json:"trait_key" bson:"trait_key"`
	DisplayName      string                 `json:"display_name" bson:"display_name"`
	Description      string                 `json:"description" bson:"description"`
	Icon             string                 `json:"icon" bson:"icon"` // Iconify icon name
	Tier             TraitTier              `json:"tier" bson:"tier"`
	AwardedAt        time.Time              `json:"awarded_at" bson:"awarded_at"`
	AwardedCriteria  string                 `json:"awarded_criteria" bson:"awarded_criteria"`
	EndorsementCount int                    `json:"endorsement_count" bson:"endorsement_count"`
	EndorsedByUsers  []uuid.UUID            `json:"-" bson:"endorsed_by_users"`
	CreatedAt        time.Time              `json:"created_at" bson:"created_at"`
}

func (t PlayerTrait) GetID() uuid.UUID {
	return t.ID
}

// IsEndorsedBy checks whether a specific user has endorsed this trait
func (t PlayerTrait) IsEndorsedBy(userID uuid.UUID) bool {
	for _, uid := range t.EndorsedByUsers {
		if uid == userID {
			return true
		}
	}
	return false
}

// ToggleEndorsement adds or removes an endorsement from a user
func (t *PlayerTrait) ToggleEndorsement(userID uuid.UUID) bool {
	for i, uid := range t.EndorsedByUsers {
		if uid == userID {
			t.EndorsedByUsers = append(t.EndorsedByUsers[:i], t.EndorsedByUsers[i+1:]...)
			t.EndorsementCount--
			return false
		}
	}
	t.EndorsedByUsers = append(t.EndorsedByUsers, userID)
	t.EndorsementCount++
	return true
}

// NewPlayerTrait creates a new trait for a player
func NewPlayerTrait(playerID uuid.UUID, gameID replay_common.GameIDKey, traitKey, displayName, description, icon string, tier TraitTier, criteria string) *PlayerTrait {
	now := time.Now()
	return &PlayerTrait{
		ID:               uuid.New(),
		PlayerID:         playerID,
		GameID:           gameID,
		TraitKey:         traitKey,
		DisplayName:      displayName,
		Description:      description,
		Icon:             icon,
		Tier:             tier,
		AwardedAt:        now,
		AwardedCriteria:  criteria,
		EndorsementCount: 0,
		EndorsedByUsers:  make([]uuid.UUID, 0),
		CreatedAt:        now,
	}
}
