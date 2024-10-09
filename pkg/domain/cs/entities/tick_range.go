package entities

import common "github.com/psavelis/team-pro/replay-api/pkg/domain"

type CSTickRange struct {
	StartTick common.TickIDType
	EndTick   common.TickIDType
}
