package achievement_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// AchievementCategory represents the category of achievement
type AchievementCategory string

const (
	CategoryCombat     AchievementCategory = "combat"
	CategoryObjective  AchievementCategory = "objective"
	CategoryTeamwork   AchievementCategory = "teamwork"
	CategoryProgress   AchievementCategory = "progress"
	CategoryCompetitive AchievementCategory = "competitive"
	CategorySocial     AchievementCategory = "social"
	CategoryMilestone  AchievementCategory = "milestone"
)

// AchievementRarity represents the rarity of an achievement
type AchievementRarity string

const (
	RarityCommon    AchievementRarity = "common"
	RarityUncommon  AchievementRarity = "uncommon"
	RarityRare      AchievementRarity = "rare"
	RarityEpic      AchievementRarity = "epic"
	RarityLegendary AchievementRarity = "legendary"
)

// Achievement represents an achievement definition
type Achievement struct {
	ID          uuid.UUID           `json:"id" bson:"_id"`
	Code        string              `json:"code" bson:"code"`
	Name        string              `json:"name" bson:"name"`
	Description string              `json:"description" bson:"description"`
	Category    AchievementCategory `json:"category" bson:"category"`
	Rarity      AchievementRarity   `json:"rarity" bson:"rarity"`
	XPReward    int                 `json:"xp_reward" bson:"xp_reward"`
	IconURL     string              `json:"icon_url" bson:"icon_url"`
	Criteria    AchievementCriteria `json:"criteria" bson:"criteria"`
	IsActive    bool                `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" bson:"updated_at"`
}

// AchievementCriteria defines the unlock criteria
type AchievementCriteria struct {
	Type       string                 `json:"type" bson:"type"`               // e.g., "count", "streak", "unique"
	Target     int                    `json:"target" bson:"target"`           // e.g., 100 kills, 10 matches
	Conditions map[string]interface{} `json:"conditions" bson:"conditions"`   // Additional conditions
}

// PlayerAchievement represents an achievement unlocked by a player
type PlayerAchievement struct {
	ID            uuid.UUID           `json:"id" bson:"_id"`
	PlayerID      uuid.UUID           `json:"player_id" bson:"player_id"`
	AchievementID uuid.UUID           `json:"achievement_id" bson:"achievement_id"`
	Achievement   *Achievement        `json:"achievement,omitempty" bson:"-"` // Populated for API responses
	Progress      int                 `json:"progress" bson:"progress"`
	TargetValue   int                 `json:"target_value" bson:"target_value"`
	UnlockedAt    *time.Time          `json:"unlocked_at" bson:"unlocked_at"`
	CreatedAt     time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" bson:"updated_at"`
	ResourceOwner common.ResourceOwner `json:"-" bson:"resource_owner"`
}

// IsUnlocked returns true if the achievement has been unlocked
func (pa *PlayerAchievement) IsUnlocked() bool {
	return pa.UnlockedAt != nil
}

// GetProgressPercentage returns the progress as a percentage
func (pa *PlayerAchievement) GetProgressPercentage() float64 {
	if pa.TargetValue == 0 {
		return 100.0
	}
	return float64(pa.Progress) / float64(pa.TargetValue) * 100.0
}

// NewPlayerAchievement creates a new player achievement tracking record
func NewPlayerAchievement(playerID, achievementID uuid.UUID, targetValue int, resourceOwner common.ResourceOwner) *PlayerAchievement {
	now := time.Now()
	return &PlayerAchievement{
		ID:            uuid.New(),
		PlayerID:      playerID,
		AchievementID: achievementID,
		Progress:      0,
		TargetValue:   targetValue,
		UnlockedAt:    nil,
		CreatedAt:     now,
		UpdatedAt:     now,
		ResourceOwner: resourceOwner,
	}
}

// IncrementProgress increments the achievement progress
func (pa *PlayerAchievement) IncrementProgress(amount int) {
	pa.Progress += amount
	pa.UpdatedAt = time.Now()
	
	if pa.Progress >= pa.TargetValue && pa.UnlockedAt == nil {
		now := time.Now()
		pa.UnlockedAt = &now
	}
}

// PlayerAchievementSummary provides a summary of player achievements
type PlayerAchievementSummary struct {
	PlayerID          uuid.UUID `json:"player_id"`
	TotalAchievements int       `json:"total_achievements"`
	UnlockedCount     int       `json:"unlocked_count"`
	TotalXP           int       `json:"total_xp"`
	RecentUnlocks     []PlayerAchievement `json:"recent_unlocks,omitempty"`
	ByCategory        map[AchievementCategory]AchievementCategoryProgress `json:"by_category"`
}

// AchievementCategoryProgress shows progress for a category
type AchievementCategoryProgress struct {
	Total    int `json:"total"`
	Unlocked int `json:"unlocked"`
}

// Predefined achievements
var PredefinedAchievements = []Achievement{
	// Combat achievements
	{Code: "FIRST_BLOOD", Name: "First Blood", Description: "Get your first kill", Category: CategoryCombat, Rarity: RarityCommon, XPReward: 50},
	{Code: "HEADHUNTER_10", Name: "Headhunter", Description: "Get 10 headshot kills", Category: CategoryCombat, Rarity: RarityUncommon, XPReward: 100},
	{Code: "HEADHUNTER_100", Name: "Precision Master", Description: "Get 100 headshot kills", Category: CategoryCombat, Rarity: RarityRare, XPReward: 500},
	{Code: "ACE", Name: "Ace", Description: "Eliminate all 5 enemy players in a single round", Category: CategoryCombat, Rarity: RarityEpic, XPReward: 300},
	{Code: "CLUTCH_1V3", Name: "Clutch Expert", Description: "Win a 1v3 clutch situation", Category: CategoryCombat, Rarity: RarityRare, XPReward: 250},
	{Code: "CLUTCH_1V5", Name: "Legendary Clutch", Description: "Win a 1v5 clutch situation", Category: CategoryCombat, Rarity: RarityLegendary, XPReward: 1000},
	
	// Objective achievements
	{Code: "BOMB_PLANT_10", Name: "Demolition Expert", Description: "Plant the bomb 10 times", Category: CategoryObjective, Rarity: RarityUncommon, XPReward: 100},
	{Code: "BOMB_DEFUSE_10", Name: "Defusal Specialist", Description: "Defuse the bomb 10 times", Category: CategoryObjective, Rarity: RarityUncommon, XPReward: 100},
	
	// Progress achievements  
	{Code: "MATCHES_10", Name: "Getting Started", Description: "Complete 10 matches", Category: CategoryProgress, Rarity: RarityCommon, XPReward: 75},
	{Code: "MATCHES_50", Name: "Regular Player", Description: "Complete 50 matches", Category: CategoryProgress, Rarity: RarityUncommon, XPReward: 200},
	{Code: "MATCHES_100", Name: "Veteran", Description: "Complete 100 matches", Category: CategoryProgress, Rarity: RarityRare, XPReward: 500},
	{Code: "MATCHES_500", Name: "Dedicated", Description: "Complete 500 matches", Category: CategoryProgress, Rarity: RarityEpic, XPReward: 1500},
	
	// Competitive achievements
	{Code: "WIN_STREAK_3", Name: "On Fire", Description: "Win 3 matches in a row", Category: CategoryCompetitive, Rarity: RarityUncommon, XPReward: 150},
	{Code: "WIN_STREAK_5", Name: "Unstoppable", Description: "Win 5 matches in a row", Category: CategoryCompetitive, Rarity: RarityRare, XPReward: 350},
	{Code: "WIN_STREAK_10", Name: "Dominator", Description: "Win 10 matches in a row", Category: CategoryCompetitive, Rarity: RarityLegendary, XPReward: 1000},
	{Code: "TOURNAMENT_WIN", Name: "Champion", Description: "Win a tournament", Category: CategoryCompetitive, Rarity: RarityEpic, XPReward: 750},
	
	// Social achievements
	{Code: "FIRST_SQUAD", Name: "Squad Up", Description: "Join or create your first squad", Category: CategorySocial, Rarity: RarityCommon, XPReward: 50},
	{Code: "SQUAD_MATCHES_10", Name: "Team Player", Description: "Complete 10 matches with your squad", Category: CategorySocial, Rarity: RarityUncommon, XPReward: 150},
	
	// Milestone achievements
	{Code: "MVP_10", Name: "Rising Star", Description: "Earn MVP 10 times", Category: CategoryMilestone, Rarity: RarityRare, XPReward: 400},
	{Code: "PRIZE_FIRST", Name: "First Win", Description: "Win your first prize pool", Category: CategoryMilestone, Rarity: RarityUncommon, XPReward: 200},
}

