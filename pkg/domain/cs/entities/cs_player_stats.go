package entities

import (
	"github.com/golang/geo/r3"
	"github.com/google/uuid"
)

type CSPlayerStats struct {
	Efficiency      map[CSTeamSideIDType]StatEfficiencyUnit
	NetworkPlayerID string     `json:"network_player_id" bson:"network_player_id"` // (SteamID64) ID is the unique identifier of the player scoreboard
	PlayerID        *uuid.UUID `json:"player" bson:"player"`                       // PlayerID is the Domain ID of the player
	Health          int        `json:"health" bson:"health"`
	Armor           int        `json:"armor" bson:"armor"`
	// Inventory          []dto.Weapon `json:"inventory" bson:"inventory"` // Inventory is the list of weapons a player has
	Money             int       `json:"money" bson:"money"`
	TimesFragged      int       `json:"frags" bson:"frags"`                             // frag is the number of times a player frag an enemy
	TimesEliminated   int       `json:"times_eliminated" bson:"times_eliminated"`       // TimesEliminated is the number of times a player is killed
	Assists           int       `json:"assists" bson:"assists"`                         // Assists is the number of frag a player gets when they damage an enemy and a teammate finishes them off
	Headshots         int       `json:"headshots" bson:"headshots"`                     // Headshots is the number of frag a player gets by shooting an enemy in the head
	TotalRoundsPlayed int       `json:"total_rounds_played" bson:"total_rounds_played"` // TotalRoundsPlayed is the total number of rounds a player has played
	TotalDamage       int       `json:"total_damage" bson:"total_damage"`               // TotalDamage is the total damage a player has done to enemies
	DMR               float64   `json:"dmr" bson:"dmr"`                                 // DMR is the average damage per round
	LastAlivePosition r3.Vector `json:"last_alive_position" bson:"last_alive_position"` // LastAlivePosition is the position where the player was last alive
	KAST              int       `json:"kast" bson:"kast"`                               // KAST is a percentage of rounds in which a player either had a kill, assist, survived or was traded
	EntryTimesFragged int       `json:"entry_TimesFragged" bson:"entry_TimesFragged"`   // EntryTimesFragged is the number of frag a player gets when they are the first to kill an enemy
	TradeTimesFragged int       `json:"trade_TimesFragged" bson:"trade_TimesFragged"`   // TradeTimesFragged is the number of frag a player gets when they are the second to kill an enemy
	FirstTimesFragged int       `json:"first_frag" bson:"first_frag"`                   // Firstfrag is the number of times a player gets the first kill in a round
	Clutches          int       `json:"clutches" bson:"clutches"`                       // Clutches is the number of times a player wins a round when they are the last player alive
	Flashes           int       `json:"flashes" bson:"flashes"`                         // Flashes is the number of times a player flashes an enemy
	KDA               float64   `json:"kda" bson:"kda"`                                 // KDA is the kill-death-assist ratio
	// HSPercentage  or Percentages breakdown? PercentagesBreakdown!!!
	// HIghtlights breakdown
	// EconomyStats breakdown
	// Utility usage/Damage breakdown
	// TeamplayStats/Trade breakdown
	// GameSense breakdown
	// AimStats breakdown
	// Comm Stats breakdown
	// Movement/Positioning/Trajectories Stats breakdown
	// Weapon/Damage Stats breakdown
	// Round Stats breakdown
	// Spike Stats breakdown
	// Nades Stats breakdown

}

func CalculateADR(totalDamage int, roundsPlayed int) float64 {
	if roundsPlayed <= 0 {
		return 0.0
	}

	return float64(totalDamage) / float64(roundsPlayed)
}

func CalculateKDR(d, k int) float64 {
	if k == 0 {
		return 0.0
	}

	return float64(d) / float64(k)
}

func CalculateKAST(frag, assists, survived, traded int, roundsPlayed int) int {
	if roundsPlayed <= 0 {
		return 0
	}

	return (frag + assists + survived + traded) / roundsPlayed
}

// type Player struct {
// 	demoInfoProvider demoInfoProvider // provider for demo info such as tick-rate or current tick

// 	SteamID64             uint64             // 64-bit representation of the user's Steam ID. See https://developer.valvesoftware.com/wiki/SteamID
// 	LastAlivePosition     r3.Vector          // The location where the player was last alive. Should be equal to Position if the player is still alive.
// 	UserID                int                // Mostly used in game-events to address this player
// 	Name                  string             // Steam / in-game user name
// 	Inventory             map[int]*Equipment // All weapons / equipment the player is currently carrying. See also Weapons().
// 	AmmoLeft              [32]int            // Ammo left for special weapons (e.g. grenades), index corresponds Equipment.AmmoType
// 	EntityID              int                // Usually the same as Entity.ID() but may be different between player death and re-spawn.
// 	Entity                st.Entity          // May be nil between player-death and re-spawn
// 	FlashDuration         float32            // Blindness duration from the flashbang currently affecting the player (seconds)
// 	FlashTick             int                // In-game tick at which the player was last flashed
// 	TeamState             *TeamState         // When keeping the reference make sure you notice when the player changes teams
// 	Team                  Team               // Team identifier for the player (e.g. TeamTerrorists or TeamCounterTerrorists).
// 	IsBot                 bool               // True if this is a bot-entity. See also IsControllingBot and ControlledBot().
// 	IsConnected           bool
// 	IsDefusing            bool
// 	IsPlanting            bool
// 	IsReloading           bool
// 	IsUnknown             bool      // Used to identify unknown/broken players. see https://github.com/markus-wa/demoinfocs-golang/issues/162
// 	PreviousFramePosition r3.Vector // CS2 only, used to compute velocity as it's not networked in CS2 demos
// }
