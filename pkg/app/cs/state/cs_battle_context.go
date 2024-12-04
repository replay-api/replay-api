package state

import (
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	cs_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/cs/entities"
)

// out: series p/ plot do heatmap (2d), battlelog (sumarizado),

// todo: dominio comum / generico (refact)
type StatsReader[T any] interface {
	GetStatistics() (interface{}, error)
	// GetStatsEntity() (T, error)
}

type CS2BattleContext struct {
	StatsReader[cs_entity.CSBattleStats]
	Hits map[common.TickIDType]cs_entity.CSHitStats
}

// func (ctx *CS2BattleContext) GetStatsEntity()

func (ctx *CS2BattleContext) GetStatistics() (interface{}, error) {
	return cs_entity.CSBattleStats{}, nil
}
