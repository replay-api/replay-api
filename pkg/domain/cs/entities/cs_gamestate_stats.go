package entities

import (
	"github.com/golang/geo/r2"
	"github.com/golang/geo/r3"
	"github.com/google/uuid"
)

type CSGameState struct {
	MatchID            *uuid.UUID  `json:"match_id" bson:"match_id"`
	TickID             int64       `json:"tick_id" bson:"tick_id"`
	Rules              CSGameRules `json:"rules" bson:"rules"`
	Nades              []Nade      `json:"nades" bson:"nades"`
	Mollies            []Molly     `json:"mollies" bson:"mollies"`
	Equipments         []Equipment `json:"equipments" bson:"equipments"`
	PackagePosition    r3.Vector   `json:"package_position" bson:"package_position"`
	TotalRoundsPlayed  int         `json:"total_rounds_played" bson:"total_rounds_played"`
	Phase              string      `json:"phase" bson:"phase"`
	IsWarmupPeriod     bool        `json:"is_warmup_period" bson:"is_warmup_period"`
	IsFreezetimePeriod bool        `json:"is_freezetime_period" bson:"is_freezetime_period"`
	IsMatchStarted     bool        `json:"is_match_started" bson:"is_match_started"`
	OvertimeCount      int         `json:"overtime_count" bson:"overtime_count"`
}

type Nade struct {
	ThrowerNetworkUserID *uuid.UUID  `json:"thrower_id" bson:"thrower_id"`
	Position             r3.Vector   `json:"position" bson:"position"`
	Equipment            Equipment   `json:"equipment" bson:"equipment"`
	OwnerNetworkUserID   uuid.UUID   `json:"owner_id" bson:"owner_id"`
	Trajectory           []r3.Vector `json:"trajectory_a" bson:"trajectory_a"`
	Trajectory2          []r3.Vector `json:"trajectory_b" bson:"trajectory_b"`
}

type Equipment struct {
	OwnerNetworkUserID *uuid.UUID `json:"owner_id" bson:"owner_id"`
	Position           *r3.Vector `json:"position" bson:"position"`
	Name               string     `json:"name" bson:"name"`
	Type               string     `json:"type" bson:"type"`
	CurrentSupply      int        `json:"current_supply_level" bson:"current_supply_level"`
	BackupSupply       int        `json:"backup_supply_level" bson:"backup_supply_level"`
	SupplyType         int        `json:"supply_type" bson:"supply_type"`
	ZoomLevel          int        `json:"zoom_level" bson:"zoom_level"`
	RecoilIndex        float64    `json:"recoil_index" bson:"recoil_index"`
}

type Molly struct {
	ThrowerNetworkUserID *uuid.UUID  `json:"thrower_id" bson:"thrower_id"`
	Equipment            Equipment   `json:"equipment" bson:"equipment"`
	OwnerNetworkUserID   *uuid.UUID  `json:"owner_id" bson:"owner_id"`
	Trajectory           []r3.Vector `json:"trajectory_a" bson:"trajectory_a"`
	Trajectory2          []r3.Vector `json:"trajectory_b" bson:"trajectory_b"`
	ConvexHull2D         []r2.Point  `json:"hull_2d" bson:"hull_2d"`
	ConvexHull3D         []r3.Vector `json:"hull_3d" bson:"hull_3d"`
}

type Hull3D struct {
	Vertices []r3.Vector `json:"points" bson:"points"`
	Index    []int       `json:"index" bson:"index"`
}
