package scores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// scoreAuthorizationAdapter implements ScoreAuthorization using tournament repository
// and resource ownership context to enforce financial-grade RBAC on score operations.
type scoreAuthorizationAdapter struct {
	tournamentRepo     tournament_out.TournamentRepository
	matchResultRepo    scores_out.MatchResultRepository
}

// NewScoreAuthorizationAdapter creates a new authorization adapter
func NewScoreAuthorizationAdapter(
	tournamentRepo tournament_out.TournamentRepository,
	matchResultRepo scores_out.MatchResultRepository,
) scores_out.ScoreAuthorization {
	return &scoreAuthorizationAdapter{
		tournamentRepo:  tournamentRepo,
		matchResultRepo: matchResultRepo,
	}
}

// IsTournamentOrganizer checks if the user is the organizer of the specified tournament
func (a *scoreAuthorizationAdapter) IsTournamentOrganizer(ctx context.Context, userID uuid.UUID, tournamentID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || tournamentID == uuid.Nil {
		return false, nil
	}

	tournament, err := a.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return false, fmt.Errorf("failed to find tournament %s: %w", tournamentID, err)
	}

	if tournament == nil {
		return false, nil
	}

	return tournament.OrganizerID == userID, nil
}

// IsMatchParticipant checks if the user is a participant in the match result
// by looking at player_results or team_results.players
func (a *scoreAuthorizationAdapter) IsMatchParticipant(ctx context.Context, userID uuid.UUID, matchResultID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || matchResultID == uuid.Nil {
		return false, nil
	}

	result, err := a.matchResultRepo.FindByID(ctx, matchResultID)
	if err != nil {
		return false, fmt.Errorf("failed to find match result %s: %w", matchResultID, err)
	}

	if result == nil {
		return false, nil
	}

	// Check player_results
	for _, pr := range result.PlayerResults {
		if pr.PlayerID == userID {
			return true, nil
		}
	}

	// Check team_results.players
	for _, tr := range result.TeamResults {
		for _, playerID := range tr.Players {
			if playerID == userID {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsPlatformAdmin checks if the current context has platform admin privileges
func (a *scoreAuthorizationAdapter) IsPlatformAdmin(ctx context.Context) bool {
	return shared.IsAdmin(ctx)
}
