package state

import (
	cs_entity "github.com/replay-api/replay-api/pkg/domain/cs/entities"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// out: series p/ plot do heatmap (2d), battlelog (sumarizado),

// todo: dominio comum / generico (refact)
type StatsReader[T any] interface {
	GetStatistics() (interface{}, error)
	// GetStatsEntity() (T, error)
}

type CS2BattleContext struct {
	StatsReader[cs_entity.CSBattleStats]
	Hits map[replay_common.TickIDType]cs_entity.CSHitStats
}

// func (ctx *CS2BattleContext) GetStatsEntity()

func (ctx *CS2BattleContext) GetStatistics() (interface{}, error) {
	return cs_entity.CSBattleStats{}, nil
}
