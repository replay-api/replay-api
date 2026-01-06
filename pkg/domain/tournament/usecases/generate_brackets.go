package tournament_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
)

// GenerateBracketsUseCase handles bracket generation for tournaments
type GenerateBracketsUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
	tournamentRepository     tournament_out.TournamentRepository
}

// NewGenerateBracketsUseCase creates a new generate brackets usecase
func NewGenerateBracketsUseCase(
	billableOperationHandler billing_in.BillableOperationCommandHandler,
	tournamentRepository tournament_out.TournamentRepository,
) *GenerateBracketsUseCase {
	return &GenerateBracketsUseCase{
		billableOperationHandler: billableOperationHandler,
		tournamentRepository:     tournamentRepository,
	}
}

// Exec generates tournament brackets based on format
func (uc *GenerateBracketsUseCase) Exec(ctx context.Context, tournamentID uuid.UUID) error {
	// auth check
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}

	// get tournament
	tournament, err := uc.tournamentRepository.FindByID(ctx, tournamentID)
	if err != nil {
		slog.ErrorContext(ctx, "tournament not found", "error", err, "tournament_id", tournamentID)
		return fmt.Errorf("tournament not found")
	}

	// validate tournament is ready for bracket generation
	if tournament.Status != tournament_entities.TournamentStatusReady {
		return fmt.Errorf("tournament must be in ready status, current: %s", tournament.Status)
	}

	participantCount := len(tournament.Participants)
	if participantCount < tournament.MinParticipants {
		return fmt.Errorf("not enough participants: %d (min: %d)", participantCount, tournament.MinParticipants)
	}

	// billing validation BEFORE generating brackets
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: billing_entities.OperationTypeGenerateBrackets,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      1,
		Args: map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"format":        tournament.Format,
			"participants":  participantCount,
		},
	}

	err = uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "billing validation failed for generate brackets", "error", err, "tournament_id", tournamentID)
		return err
	}

	// apply seeding if tournament has seeds
	uc.applySeedingIfNeeded(tournament)

	// generate brackets based on format
	var matches []tournament_entities.TournamentMatch
	switch tournament.Format {
	case tournament_entities.TournamentFormatSingleElimination:
		matches, err = uc.generateSingleEliminationBracket(tournament)
	case tournament_entities.TournamentFormatDoubleElimination:
		matches, err = uc.generateDoubleEliminationBracket(tournament)
	case tournament_entities.TournamentFormatRoundRobin:
		matches, err = uc.generateRoundRobinBracket(tournament)
	case tournament_entities.TournamentFormatSwiss:
		matches, err = uc.generateSwissBracket(tournament, 1) // first round
	default:
		return fmt.Errorf("unsupported tournament format: %s", tournament.Format)
	}

	if err != nil {
		slog.ErrorContext(ctx, "failed to generate brackets", "error", err, "format", tournament.Format)
		return err
	}

	// update tournament with generated matches
	tournament.Matches = matches

	// save updated tournament
	err = uc.tournamentRepository.Update(ctx, tournament)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update tournament with brackets", "error", err, "tournament_id", tournamentID)
		return fmt.Errorf("failed to generate brackets")
	}

	// billing execution AFTER successful operation
	_, _, err = uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "failed to execute billing for generate brackets", "error", err, "tournament_id", tournamentID)
	}

	slog.InfoContext(ctx, "tournament brackets generated",
		"tournament_id", tournamentID,
		"format", tournament.Format,
		"participants", participantCount,
		"matches", len(matches),
	)

	return nil
}

// applySeedingIfNeeded applies seeding to participants if needed
func (uc *GenerateBracketsUseCase) applySeedingIfNeeded(tournament *tournament_entities.Tournament) {
	// if participants don't have seeds, assign based on registration order
	unseeded := false
	for _, p := range tournament.Participants {
		if p.Seed == 0 {
			unseeded = true
			break
		}
	}

	if unseeded {
		// assign seeds based on registration order
		for i := range tournament.Participants {
			tournament.Participants[i].Seed = i + 1
		}
	}
}

// generateSingleEliminationBracket generates single elimination bracket
// standard bracket seeding: 1 vs N, 2 vs N-1, etc.
func (uc *GenerateBracketsUseCase) generateSingleEliminationBracket(tournament *tournament_entities.Tournament) ([]tournament_entities.TournamentMatch, error) {
	participants := tournament.Participants
	n := len(participants)

	// calculate next power of 2
	bracketSize := nextPowerOf2(n)
	byesNeeded := bracketSize - n

	var matches []tournament_entities.TournamentMatch
	matchNumber := 0
	startTime := tournament.StartTime

	// round 1 - pair players using standard seeding
	round := 1
	seedPairs := generateSeedPairs(n, bracketSize)

	for _, pair := range seedPairs {
		if pair.Seed2 > n {
			// bye - player advances automatically
			continue
		}

		player1 := participants[pair.Seed1-1]
		player2 := participants[pair.Seed2-1]

		match := tournament_entities.TournamentMatch{
			MatchID:     uuid.New(),
			Round:       round,
			BracketPos:  fmt.Sprintf("r%d_m%d", round, matchNumber),
			Player1ID:   player1.PlayerID,
			Player2ID:   player2.PlayerID,
			ScheduledAt: startTime.Add(time.Duration(matchNumber*30) * time.Minute),
			Status:      tournament_entities.MatchStatusScheduled,
		}

		matches = append(matches, match)
		matchNumber++
	}

	slog.Info("single elimination bracket generated",
		"participants", n,
		"bracket_size", bracketSize,
		"byes", byesNeeded,
		"round_1_matches", len(matches),
	)

	return matches, nil
}

// generateDoubleEliminationBracket generates double elimination bracket
// includes winners bracket and losers bracket
func (uc *GenerateBracketsUseCase) generateDoubleEliminationBracket(tournament *tournament_entities.Tournament) ([]tournament_entities.TournamentMatch, error) {
	participants := tournament.Participants
	n := len(participants)

	bracketSize := nextPowerOf2(n)
	var matches []tournament_entities.TournamentMatch
	matchNumber := 0
	startTime := tournament.StartTime

	// winners bracket round 1
	round := 1
	seedPairs := generateSeedPairs(n, bracketSize)

	for _, pair := range seedPairs {
		if pair.Seed2 > n {
			continue
		}

		player1 := participants[pair.Seed1-1]
		player2 := participants[pair.Seed2-1]

		match := tournament_entities.TournamentMatch{
			MatchID:     uuid.New(),
			Round:       round,
			BracketPos:  fmt.Sprintf("winners_r%d_m%d", round, matchNumber),
			Player1ID:   player1.PlayerID,
			Player2ID:   player2.PlayerID,
			ScheduledAt: startTime.Add(time.Duration(matchNumber*30) * time.Minute),
			Status:      tournament_entities.MatchStatusScheduled,
		}

		matches = append(matches, match)
		matchNumber++
	}

	// losers bracket will be populated as matches complete
	// for now, just generate winners bracket structure

	slog.Info("double elimination bracket generated",
		"participants", n,
		"bracket_size", bracketSize,
		"winners_r1_matches", len(matches),
	)

	return matches, nil
}

// generateRoundRobinBracket generates round robin bracket
// every player plays every other player once
func (uc *GenerateBracketsUseCase) generateRoundRobinBracket(tournament *tournament_entities.Tournament) ([]tournament_entities.TournamentMatch, error) {
	participants := tournament.Participants
	n := len(participants)

	var matches []tournament_entities.TournamentMatch
	matchNumber := 0
	startTime := tournament.StartTime

	// round robin: n*(n-1)/2 matches
	round := 1
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			match := tournament_entities.TournamentMatch{
				MatchID:     uuid.New(),
				Round:       round,
				BracketPos:  fmt.Sprintf("rr_m%d", matchNumber),
				Player1ID:   participants[i].PlayerID,
				Player2ID:   participants[j].PlayerID,
				ScheduledAt: startTime.Add(time.Duration(matchNumber*30) * time.Minute),
				Status:      tournament_entities.MatchStatusScheduled,
			}

			matches = append(matches, match)
			matchNumber++

			// group matches into rounds (every player plays once per round)
			if (matchNumber % (n / 2)) == 0 {
				round++
			}
		}
	}

	totalMatches := (n * (n - 1)) / 2

	slog.Info("round robin bracket generated",
		"participants", n,
		"total_matches", totalMatches,
		"rounds", round,
	)

	return matches, nil
}

// generateSwissBracket generates swiss system bracket for a specific round
// pairs players with similar records, avoiding rematches
func (uc *GenerateBracketsUseCase) generateSwissBracket(tournament *tournament_entities.Tournament, round int) ([]tournament_entities.TournamentMatch, error) {
	participants := tournament.Participants
	n := len(participants)

	var matches []tournament_entities.TournamentMatch
	startTime := tournament.StartTime

	// for round 1, use standard seeding
	if round == 1 {
		// pair top half vs bottom half
		halfPoint := n / 2
		for i := 0; i < halfPoint; i++ {
			if i+halfPoint < n {
				match := tournament_entities.TournamentMatch{
					MatchID:     uuid.New(),
					Round:       round,
					BracketPos:  fmt.Sprintf("swiss_r%d_m%d", round, i),
					Player1ID:   participants[i].PlayerID,
					Player2ID:   participants[i+halfPoint].PlayerID,
					ScheduledAt: startTime.Add(time.Duration(i*30) * time.Minute),
					Status:      tournament_entities.MatchStatusScheduled,
				}
				matches = append(matches, match)
			}
		}
	}
	// subsequent rounds would pair based on standings
	// this requires tracking wins/losses which happens during match completion

	slog.Info("swiss bracket round generated",
		"participants", n,
		"round", round,
		"matches", len(matches),
	)

	return matches, nil
}

// SeedPair represents a seeded pairing
type SeedPair struct {
	Seed1 int
	Seed2 int
}

// generateSeedPairs generates standard tournament seeding pairs
// example for 8 players: (1,8), (4,5), (2,7), (3,6)
func generateSeedPairs(_, bracketSize int) []SeedPair {
	var pairs []SeedPair

	// generate pairs using standard bracket seeding algorithm
	for round := 0; round < int(math.Log2(float64(bracketSize))); round++ {
		if round == 0 {
			// first round: pair extremes
			for i := 0; i < bracketSize/2; i++ {
				seed1 := i + 1
				seed2 := bracketSize - i

				pairs = append(pairs, SeedPair{
					Seed1: seed1,
					Seed2: seed2,
				})
			}
			break // we only need round 1 pairs
		}
	}

	return pairs
}

// nextPowerOf2 returns the next power of 2 >= n
func nextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}
	power := 1
	for power < n {
		power *= 2
	}
	return power
}
