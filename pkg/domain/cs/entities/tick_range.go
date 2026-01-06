package entities

import (
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

type CSTickRange struct {
	StartTick replay_common.TickIDType
	EndTick   replay_common.TickIDType
}
