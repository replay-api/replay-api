package entities

import (
	"github.com/golang/geo/r3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type CSPositioningStats struct {
	// TickID                       TickIDType
	PlayerTrajectory             map[common.PlayerIDType][]CSPositioningTrajectoryStats `json:"player_trajectory" bson:"player_trajectory"` // TODO: review se não está redundante esse array, talvez pode ser só um objeto
	PlayerZoneFrequencies        map[common.PlayerIDType]map[ZoneCodeType]int           `json:"player_zone_frequencies"`                    // Player -> Zone -> Count
	PlayerZoneDwellTimes         map[common.PlayerIDType]map[ZoneCodeType]float64       `json:"player_zone_dwell_times"`                    // Player -> Zone -> Total Time
	PlayerAverageSpeed           map[common.PlayerIDType]float64                        `json:"player_average_speed"`
	PlayerAverageSpeedCrouching  map[common.PlayerIDType]float64
	PlayerAverageSpeedWalking    map[common.PlayerIDType]float64
	PlayerAverageSpeedRunning    map[common.PlayerIDType]float64
	PlayerAverageSpeedAirborne   map[common.PlayerIDType]float64
	PlayerAverageSpeedJumping    map[common.PlayerIDType]float64
	PlayerAverageSpeedStrafing   map[common.PlayerIDType]float64
	PlayerAverageSpeedInBombZone map[common.PlayerIDType]float64

	PlayerAverageSpeedInBuyZone map[common.PlayerIDType]float64
}

type CSTick struct {
	TickID           float64   `json:"tick_id" bson:"tick_id"`
	RoundTime        float64   `json:"round_time" bson:"round_time"`
	RoundID          uuid.UUID `json:"round_id" bson:"round_id"`
	RoundNumber      int       `json:"round_number" bson:"round_number"`
	MatchTime        float64   `json:"match_time" bson:"match_time"`
	FileTime         float64   `json:"file_time" bson:"file_time"`
	WorldUTCTimeTick float64   `json:"utc_tick" bson:"utc_tick"`
	Rate             float64   `json:"rate" bson:"rate"`
}
type ZoneCodeType string
type CSZone struct {
	ZoneID   uuid.UUID    `json:"zone_id" bson:"zone_id"`
	ZoneCode ZoneCodeType `json:"zone_code" bson:"zone_code"`
	ZoneName string       `json:"zone_name" bson:"zone_name"`
}

type CSPositioningTrajectoryStats struct {
	TickID             float64       `json:"tick_id" bson:"tick_id"`
	Position           *r3.Vector    `json:"position" bson:"position"`
	GameSpeed          *float64      `json:"game_speed" bson:"game_speed"`
	Velocity           *r3.Vector    `json:"velocity" bson:"velocity"`
	IsRushing          *bool         `json:"is_rushing" bson:"is_rushing"`
	IsCrouching        *bool         `json:"is_crouching" bson:"is_crouching"`
	Angle              *r3.Vector    `json:"angle" bson:"angle"`
	Zone               *ZoneCodeType `json:"zone_code" bson:"zone_code"`
	IsAlive            *bool         `json:"is_alive" bson:"is_alive"`
	IsWalking          *bool         `json:"is_walking" bson:"is_walking"`
	IsAirborne         *bool         `json:"is_airborne" bson:"is_airborne"`
	IsJumping          *bool         `json:"is_jumping" bson:"is_jumping"`
	IsDefusing         *bool         `json:"is_defusing" bson:"is_defusing"`
	IsPlanting         *bool         `json:"is_planting" bson:"is_planting"`
	IsInBombZone       *bool         `json:"is_in_bomb_zone" bson:"is_in_bomb_zone"`
	IsInBuyZone        *bool         `json:"is_in_buy_zone" bson:"is_in_buy_zone"`
	IsInNoClip         *bool         `json:"is_in_no_clip" bson:"is_in_no_clip"`
	IsInHostageZone    *bool         `json:"is_in_hostage_zone" bson:"is_in_hostage_zone"`
	IsInCombatZone     bool          `json:"is_in_combat_zone" bson:"is_in_combat_zone"`
	IsInSpawnZone      bool          `json:"is_in_spawn_zone" bson:"is_in_spawn_zone"`
	IsInTransitionZone bool          `json:"is_in_transition_zone" bson:"is_in_transition_zone"`
	IsInBombSiteA      *bool         `json:"is_in_bombsite_a" bson:"is_in_bombsite_a"`
	IsInBombSiteB      *bool         `json:"is_in_bombsite_b" bson:"is_in_bombsite_b"`

	IsCounterStrafing bool  `json:"is_counter_strafing" bson:"is_counter_strafing"`
	IsInAir           *bool `json:"is_in_air" bson:"is_in_air"`
	IsInCorner        *bool `json:"is_in_corner" bson:"is_in_corner"`
	IsInWindow        *bool `json:"is_in_window" bson:"is_in_window"`
	IsInDoor          *bool `json:"is_in_door" bson:"is_in_door"`
	IsInWater         *bool `json:"is_in_water" bson:"is_in_water"`
	IsInLadder        *bool `json:"is_in_ladder" bson:"is_in_ladder"`
	IsTakingDamage    bool  `json:"is_taking_damage" bson:"is_taking_damage"`
	IsFleeing         bool  `json:"is_fleeing" bson:"is_fleeing"`
	IsDeflecting      bool  `json:"is_deflecting" bson:"is_deflecting"`
}

type CSAngleStats struct {
	TickID       float64 `json:"tick_id" bson:"tick_id"`
	Angle        *r3.Vector
	Velocity     *r3.Vector
	IsScoped     *bool `json:"is_scoped" bson:"is_scoped"`
	IsFlashed    bool  `json:"is_flashed" bson:"is_flashed"`
	IsInCover    bool  `json:"is_in_cover" bson:"is_in_cover"`
	IsInSmoke    bool  `json:"is_in_smoke" bson:"is_in_smoke"`
	IsInFire     bool  `json:"is_in_fire" bson:"is_in_fire"`
	IsFleeing    bool  `json:"is_fleeing" bson:"is_fleeing"`
	IsDeflecting bool  `json:"is_deflecting" bson:"is_deflecting"`

	IsOffTarget       bool       `json:"is_off_target" bson:"is_off_target"`
	IsOnTarget        bool       `json:"is_on_target" bson:"is_on_target"`
	IsFiring          bool       `json:"is_firing" bson:"is_firing"`
	IsReloading       bool       `json:"is_reloading" bson:"is_reloading"`
	DistToTarget      float64    // Distance to the nearest enemy (if visible)
	TargetPositioning *r3.Vector // 3D position of the nearest enemy (if visible)
	TargetVelocity    *r3.Vector // 3D velocity of the nearest enemy (if visible)
	TargetHealth      int        // Health of the nearest enemy (if visible)
	TargetArmor       int        // Armor of the nearest enemy (if visible)
	TargetMoney       int        // Money of the nearest enemy (if visible)
	TargetTeam        string     // Team of the nearest enemy (if visible)
	TargetName        string     // Name of the nearest enemy (if visible)
	AngleType         string     // Categorize the angle (e.g., "wide peek", "narrow angle")

}

// func analyzeAngles(stats []CSAngleStats) {
// 	for _, stat := range stats {
// 		if stat.IsOnTarget && stat.IsFiring && stat.Weapon == "AK-47" {
// 			// Analyze successful AK-47 shots on target
// 		}

// 		if stat.AngleType == "wide peek" && stat.DistToTarget > 20 {
// 			// Analyze wide peeks from long distances
// 		}
// 		// ... more analysis based on the rich context available
// 	}
// }
