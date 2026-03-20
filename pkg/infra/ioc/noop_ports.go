package ioc

import (
	"context"
	"fmt"

	replay_common "github.com/replay-api/replay-common/pkg/replay"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
)

type noopTeamResolver struct{}

func (noopTeamResolver) ResolveTeam(_ context.Context, name string, _ replay_common.GameIDKey) (*oracle_out.TeamRef, error) {
	return nil, fmt.Errorf("team resolver unavailable for %q", name)
}