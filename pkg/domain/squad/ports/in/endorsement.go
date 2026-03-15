package squad_in

import (
	"context"

	"github.com/google/uuid"
)

// SkillEndorsementCommand represents a request to endorse/un-endorse a skill
type SkillEndorsementCommand struct {
	PlayerID uuid.UUID `json:"player_id"`
	SkillID  uuid.UUID `json:"skill_id"`
	UserID   uuid.UUID `json:"user_id"` // The endorser
}

// TraitEndorsementCommand represents a request to endorse/un-endorse a trait
type TraitEndorsementCommand struct {
	PlayerID uuid.UUID `json:"player_id"`
	TraitID  uuid.UUID `json:"trait_id"`
	UserID   uuid.UUID `json:"user_id"` // The endorser
}

// EndorsementWriter defines the interface for writing endorsements
type EndorsementWriter interface {
	// EndorseSkill toggles an endorsement on a player skill (returns true if now endorsed)
	EndorseSkill(ctx context.Context, cmd SkillEndorsementCommand) (bool, error)
	// EndorseTrait toggles an endorsement on a player trait (returns true if now endorsed)
	EndorseTrait(ctx context.Context, cmd TraitEndorsementCommand) (bool, error)
}
