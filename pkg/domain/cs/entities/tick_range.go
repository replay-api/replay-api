package entities

import common "github.com/replay-api/replay-api/pkg/domain"

type CSTickRange struct {
	StartTick common.TickIDType
	EndTick   common.TickIDType
}
