package squad_entities

import (
	"time"

	"github.com/google/uuid"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// SkillCategory represents one of the five skill radar axes
type SkillCategory string

const (
	SkillCategoryMechanical  SkillCategory = "mechanical"
	SkillCategoryTactical    SkillCategory = "tactical"
	SkillCategoryLeadership  SkillCategory = "leadership"
	SkillCategoryUtility     SkillCategory = "utility"
	SkillCategoryConsistency SkillCategory = "consistency"
)

// SkillDataSource indicates how the skill level was determined
type SkillDataSource string

const (
	SkillDataSourceAuto   SkillDataSource = "auto"
	SkillDataSourceManual SkillDataSource = "manual"
	SkillDataSourceHybrid SkillDataSource = "hybrid"
)

// PlayerSkill represents an individual skill with endorsement support
type PlayerSkill struct {
	ID               uuid.UUID              `json:"id" bson:"_id"`
	PlayerID         uuid.UUID              `json:"player_id" bson:"player_id"`
	GameID           replay_common.GameIDKey `json:"game_id" bson:"game_id"`
	SkillName        string                 `json:"skill_name" bson:"skill_name"`
	SkillKey         string                 `json:"skill_key" bson:"skill_key"`
	Category         SkillCategory          `json:"category" bson:"category"`
	Level            int                    `json:"level" bson:"level"` // 0–100
	DataSource       SkillDataSource        `json:"data_source" bson:"data_source"`
	EndorsementCount int                    `json:"endorsement_count" bson:"endorsement_count"`
	EndorsedByUsers  []uuid.UUID            `json:"-" bson:"endorsed_by_users"`
	LastEvaluated    time.Time              `json:"last_evaluated" bson:"last_evaluated"`
	CreatedAt        time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" bson:"updated_at"`
}

func (s PlayerSkill) GetID() uuid.UUID {
	return s.ID
}

// IsEndorsedBy checks whether a specific user has endorsed this skill
func (s PlayerSkill) IsEndorsedBy(userID uuid.UUID) bool {
	for _, uid := range s.EndorsedByUsers {
		if uid == userID {
			return true
		}
	}
	return false
}

// ToggleEndorsement adds or removes an endorsement from a user
func (s *PlayerSkill) ToggleEndorsement(userID uuid.UUID) bool {
	for i, uid := range s.EndorsedByUsers {
		if uid == userID {
			// Remove endorsement
			s.EndorsedByUsers = append(s.EndorsedByUsers[:i], s.EndorsedByUsers[i+1:]...)
			s.EndorsementCount--
			return false
		}
	}
	// Add endorsement
	s.EndorsedByUsers = append(s.EndorsedByUsers, userID)
	s.EndorsementCount++
	return true
}

// NewPlayerSkill creates a new skill for a player
func NewPlayerSkill(playerID uuid.UUID, gameID replay_common.GameIDKey, skillName, skillKey string, category SkillCategory, level int, source SkillDataSource) *PlayerSkill {
	now := time.Now()
	return &PlayerSkill{
		ID:               uuid.New(),
		PlayerID:         playerID,
		GameID:           gameID,
		SkillName:        skillName,
		SkillKey:         skillKey,
		Category:         category,
		Level:            clampLevel(level),
		DataSource:       source,
		EndorsementCount: 0,
		EndorsedByUsers:  make([]uuid.UUID, 0),
		LastEvaluated:    now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// SkillProfile is the aggregated skill profile for radar chart display
type SkillProfile struct {
	PlayerID          uuid.UUID                `json:"player_id" bson:"player_id"`
	Categories        map[SkillCategory]int    `json:"categories" bson:"categories"` // 0–100 per category
	TopSkills         []PlayerSkill            `json:"top_skills" bson:"top_skills"`
	TotalEndorsements int                      `json:"total_endorsements" bson:"total_endorsements"`
	LastUpdated       time.Time                `json:"last_updated" bson:"last_updated"`
}

func clampLevel(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
