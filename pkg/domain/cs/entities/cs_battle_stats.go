package entities

import common "github.com/psavelis/team-pro/replay-api/pkg/domain"

type CSHitBoxType byte

const (
	HeadHitBox CSHitBoxType = iota
	BodyHitBox
	LegHitBox
	ArmHitBox
	// REVIEW
)

type HitStageType byte

const (
	HitStageTypeEntry HitStageType = iota
	HitStageEntryProgress
	HitStageTypeTrade
	HitStageTradeProgress
	HitStageTypeTradeFragLastHit
	HitStageTypeEntryFragLastHit
)

type CSBattleStats struct {
	KDA  float64
	DMR  float64
	UDMR float64 // TODO: definir DMR para utilitarios ***
	KAST float64

	/// other breakdowns... wpn most used, most damge etc, breakdown utilities?

	AssistanceFrags float64
	EntryFrags      float64
	TradeFrags      float64
	TotalFrags      float64

	OpponentAssistanceFrags float64
	OpponentEntryFrags      float64
	OpponentTradeFrags      float64
	TotalOpponentFrags      float64

	FragStatsByVictim map[common.PlayerIDType]map[common.TickIDType]CSHitStats
}

type CSHitStats struct {
	// TODO: [importante] para saber se foi no scope, HitStats > PlayerAngle > ScopeState (no-scope, quickscope, scoped)

	// KillerAngle r3.Vector // já tem tudo isso nos stats de cada player
	// VictimAngle r3.Vector
	// WeaponClass      string
	// WeaponCategory   string
	// WeaponPrice      int
	// WeaponKillReward int
	// NoScope          bool
	// QuickScope       bool
	// 	Weapon   WeaponIDType
	// TODO: mover para CSAngleStats (Killer and Victim)

	Damage           int
	Location         CSHitBoxType
	HitStage         HitStageType
	PenetrationLevel int
	SourceItemID     string // todo: type
	TargetPlayerID   string
	SourcePlayerID   string // todo: type (neste momento ainda não tem obj id)
}
