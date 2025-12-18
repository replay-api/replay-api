package tournament_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	tournament_entities "github.com/replay-api/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
)

// GenerateBracketsHandler defines the interface for bracket generation
type GenerateBracketsHandler interface {
	Exec(ctx context.Context, tournamentID uuid.UUID) error
}

// TournamentService implements tournament management business logic
type TournamentService struct {
	tournamentRepo           tournament_out.TournamentRepository
	walletCommand            wallet_in.WalletCommand
	playerProfileReader      squad_in.PlayerProfileReader
	generateBracketsHandler  GenerateBracketsHandler
}

// NewTournamentService creates a new tournament service
func NewTournamentService(
	tournamentRepo tournament_out.TournamentRepository,
	walletCommand wallet_in.WalletCommand,
	playerProfileReader squad_in.PlayerProfileReader,
	generateBracketsHandler GenerateBracketsHandler,
) tournament_in.TournamentCommand {
	return &TournamentService{
		tournamentRepo:          tournamentRepo,
		walletCommand:           walletCommand,
		playerProfileReader:     playerProfileReader,
		generateBracketsHandler: generateBracketsHandler,
	}
}

// CreateTournament creates a new tournament
func (s *TournamentService) CreateTournament(ctx context.Context, cmd tournament_in.CreateTournamentCommand) (*tournament_entities.Tournament, error) {
	slog.InfoContext(ctx, "creating tournament", "name", cmd.Name, "game_id", cmd.GameID, "organizer", cmd.OrganizerID)

	// Create tournament entity
	tournament, err := tournament_entities.NewTournament(
		cmd.ResourceOwner,
		cmd.Name,
		cmd.Description,
		cmd.GameID,
		cmd.GameMode,
		cmd.Region,
		cmd.Format,
		cmd.MaxParticipants,
		cmd.MinParticipants,
		cmd.EntryFee,
		cmd.Currency,
		cmd.StartTime,
		cmd.RegistrationOpen,
		cmd.RegistrationClose,
		cmd.Rules,
		cmd.OrganizerID,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create tournament entity", "error", err)
		return nil, fmt.Errorf("invalid tournament data: %w", err)
	}

	// Validate entity
	if err := tournament.Validate(); err != nil {
		slog.ErrorContext(ctx, "tournament validation failed", "error", err)
		return nil, fmt.Errorf("tournament validation failed: %w", err)
	}

	// Persist tournament
	if err := s.tournamentRepo.Save(ctx, tournament); err != nil {
		slog.ErrorContext(ctx, "failed to save tournament", "tournament_id", tournament.ID, "error", err)
		return nil, fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament created successfully", "tournament_id", tournament.ID, "name", tournament.Name)
	return tournament, nil
}

// UpdateTournament updates tournament details (only before start)
func (s *TournamentService) UpdateTournament(ctx context.Context, cmd tournament_in.UpdateTournamentCommand) (*tournament_entities.Tournament, error) {
	slog.InfoContext(ctx, "updating tournament", "tournament_id", cmd.TournamentID)

	// Fetch existing tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	// Only allow updates before tournament starts
	if tournament.Status != tournament_entities.TournamentStatusDraft &&
		tournament.Status != tournament_entities.TournamentStatusRegistration {
		return nil, fmt.Errorf("cannot update tournament in status: %s", tournament.Status)
	}

	// Apply updates
	if cmd.Name != nil {
		tournament.Name = *cmd.Name
	}
	if cmd.Description != nil {
		tournament.Description = *cmd.Description
	}
	if cmd.MaxParticipants != nil {
		if *cmd.MaxParticipants < len(tournament.Participants) {
			return nil, fmt.Errorf("cannot reduce max_participants below current participant count")
		}
		tournament.MaxParticipants = *cmd.MaxParticipants
	}
	if cmd.StartTime != nil {
		tournament.StartTime = *cmd.StartTime
	}
	if cmd.RegistrationClose != nil {
		tournament.RegistrationClose = *cmd.RegistrationClose
	}
	if cmd.Rules != nil {
		tournament.Rules = *cmd.Rules
	}

	// Validate updated entity
	if err := tournament.Validate(); err != nil {
		return nil, fmt.Errorf("updated tournament validation failed: %w", err)
	}

	// Persist updates
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return nil, fmt.Errorf("failed to update tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament updated successfully", "tournament_id", tournament.ID)
	return tournament, nil
}

// DeleteTournament removes a tournament (only in draft/registration)
func (s *TournamentService) DeleteTournament(ctx context.Context, tournamentID uuid.UUID) error {
	slog.InfoContext(ctx, "deleting tournament", "tournament_id", tournamentID)

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Only allow deletion in early stages
	if tournament.Status != tournament_entities.TournamentStatusDraft &&
		tournament.Status != tournament_entities.TournamentStatusRegistration {
		return fmt.Errorf("cannot delete tournament in status: %s", tournament.Status)
	}

	// TODO: Issue refunds to registered players if entry fee was charged
	if len(tournament.Participants) > 0 && !tournament.EntryFee.IsZero() {
		slog.WarnContext(ctx, "deleting tournament with registered players - refunds should be issued",
			"tournament_id", tournamentID,
			"participant_count", len(tournament.Participants))
	}

	// Delete tournament
	if err := s.tournamentRepo.Delete(ctx, tournamentID); err != nil {
		return fmt.Errorf("failed to delete tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament deleted successfully", "tournament_id", tournamentID)
	return nil
}

// RegisterPlayer registers a player for the tournament
func (s *TournamentService) RegisterPlayer(ctx context.Context, cmd tournament_in.RegisterPlayerCommand) error {
	slog.InfoContext(ctx, "registering player for tournament",
		"tournament_id", cmd.TournamentID,
		"player_id", cmd.PlayerID)

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Register player (entity method handles validation)
	if err := tournament.RegisterPlayer(cmd.PlayerID, cmd.DisplayName); err != nil {
		slog.ErrorContext(ctx, "failed to register player", "error", err)
		return fmt.Errorf("registration failed: %w", err)
	}

	// Charge entry fee if required
	if !tournament.EntryFee.IsZero() {
		// Create wallet transaction command
		debitCmd := wallet_in.DebitWalletCommand{
			UserID:      cmd.PlayerID,
			Amount:      tournament.EntryFee,
			Currency:    string(tournament.Currency),
			Description: fmt.Sprintf("Tournament entry fee: %s", tournament.Name),
			Metadata: map[string]interface{}{
				"tournament_id": tournament.ID.String(),
				"type":          "tournament_entry",
			},
		}

		if _, err := s.walletCommand.DebitWallet(ctx, debitCmd); err != nil {
			// Rollback registration
			_ = tournament.UnregisterPlayer(cmd.PlayerID)
			slog.ErrorContext(ctx, "failed to charge entry fee", "player_id", cmd.PlayerID, "error", err)
			return fmt.Errorf("failed to charge entry fee: %w", err)
		}

		slog.InfoContext(ctx, "entry fee charged", "player_id", cmd.PlayerID, "amount", tournament.EntryFee.String())
	}

	// Persist updated tournament
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		// TODO: Consider refunding entry fee on persistence failure
		slog.ErrorContext(ctx, "failed to persist tournament registration", "error", err)
		return fmt.Errorf("failed to save registration: %w", err)
	}

	slog.InfoContext(ctx, "player registered successfully",
		"tournament_id", cmd.TournamentID,
		"player_id", cmd.PlayerID,
		"total_participants", len(tournament.Participants))

	return nil
}

// UnregisterPlayer removes a player from the tournament
func (s *TournamentService) UnregisterPlayer(ctx context.Context, cmd tournament_in.UnregisterPlayerCommand) error {
	slog.InfoContext(ctx, "unregistering player from tournament",
		"tournament_id", cmd.TournamentID,
		"player_id", cmd.PlayerID)

	// CRITICAL: Ownership validation - prevent impersonation
	// Verify the PlayerID belongs to the authenticated user
	playerSearch := squad_entities.NewSearchByID(ctx, cmd.PlayerID)
	players, err := s.playerProfileReader.Search(ctx, playerSearch)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find player profile for unregister", "error", err, "player_id", cmd.PlayerID)
		return fmt.Errorf("player not found")
	}
	if len(players) == 0 {
		return common.NewErrNotFound(common.ResourceTypePlayerProfile, "ID", cmd.PlayerID.String())
	}

	// Verify ownership - player must belong to authenticated user
	currentUserID := common.GetResourceOwner(ctx).UserID
	if players[0].ResourceOwner.UserID != currentUserID {
		slog.WarnContext(ctx, "Tournament unregistration impersonation attempt blocked",
			"attempted_player_id", cmd.PlayerID,
			"player_owner", players[0].ResourceOwner.UserID,
			"attacker_user_id", currentUserID,
			"tournament_id", cmd.TournamentID,
		)
		return common.NewErrUnauthorized()
	}

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Unregister player
	if err := tournament.UnregisterPlayer(cmd.PlayerID); err != nil {
		return fmt.Errorf("unregistration failed: %w", err)
	}

	// Refund entry fee if charged
	if !tournament.EntryFee.IsZero() {
		creditCmd := wallet_in.CreditWalletCommand{
			UserID:      cmd.PlayerID,
			Amount:      tournament.EntryFee,
			Currency:    string(tournament.Currency),
			Description: fmt.Sprintf("Tournament entry refund: %s", tournament.Name),
			Metadata: map[string]interface{}{
				"tournament_id": tournament.ID.String(),
				"type":          "tournament_refund",
			},
		}

		if _, err := s.walletCommand.CreditWallet(ctx, creditCmd); err != nil {
			slog.ErrorContext(ctx, "failed to refund entry fee", "player_id", cmd.PlayerID, "error", err)
			// Continue with unregistration even if refund fails (manual intervention needed)
		} else {
			slog.InfoContext(ctx, "entry fee refunded", "player_id", cmd.PlayerID, "amount", tournament.EntryFee.String())
		}
	}

	// Persist updated tournament
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save unregistration: %w", err)
	}

	slog.InfoContext(ctx, "player unregistered successfully",
		"tournament_id", cmd.TournamentID,
		"player_id", cmd.PlayerID)

	return nil
}

// OpenRegistration opens the tournament for player registration
func (s *TournamentService) OpenRegistration(ctx context.Context, tournamentID uuid.UUID) error {
	slog.InfoContext(ctx, "opening tournament registration", "tournament_id", tournamentID)

	tournament, err := s.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.OpenRegistration(); err != nil {
		return fmt.Errorf("failed to open registration: %w", err)
	}

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "registration opened", "tournament_id", tournamentID)
	return nil
}

// CloseRegistration closes player registration
func (s *TournamentService) CloseRegistration(ctx context.Context, tournamentID uuid.UUID) error {
	slog.InfoContext(ctx, "closing tournament registration", "tournament_id", tournamentID)

	tournament, err := s.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.CloseRegistration(); err != nil {
		return fmt.Errorf("failed to close registration: %w", err)
	}

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "registration closed", "tournament_id", tournamentID, "participants", len(tournament.Participants))
	return nil
}

// StartTournament begins the tournament
func (s *TournamentService) StartTournament(ctx context.Context, tournamentID uuid.UUID) error {
	slog.InfoContext(ctx, "starting tournament", "tournament_id", tournamentID)

	tournament, err := s.tournamentRepo.FindByID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.Start(); err != nil {
		return fmt.Errorf("failed to start tournament: %w", err)
	}

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament started", "tournament_id", tournamentID)
	return nil
}

// GenerateBrackets generates tournament brackets based on format
func (s *TournamentService) GenerateBrackets(ctx context.Context, tournamentID uuid.UUID) error {
	slog.InfoContext(ctx, "generating brackets", "tournament_id", tournamentID)

	if s.generateBracketsHandler == nil {
		return fmt.Errorf("bracket generation handler not configured")
	}

	if err := s.generateBracketsHandler.Exec(ctx, tournamentID); err != nil {
		slog.ErrorContext(ctx, "failed to generate brackets", "tournament_id", tournamentID, "error", err)
		return fmt.Errorf("failed to generate brackets: %w", err)
	}

	slog.InfoContext(ctx, "brackets generated successfully", "tournament_id", tournamentID)
	return nil
}

// CompleteTournament marks tournament as completed with winners
func (s *TournamentService) CompleteTournament(ctx context.Context, cmd tournament_in.CompleteTournamentCommand) error {
	slog.InfoContext(ctx, "completing tournament", "tournament_id", cmd.TournamentID)

	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.Complete(cmd.Winners); err != nil {
		return fmt.Errorf("failed to complete tournament: %w", err)
	}

	// Distribute prizes to winners
	for _, winner := range cmd.Winners {
		if !winner.Prize.IsZero() {
			creditCmd := wallet_in.CreditWalletCommand{
				UserID:      winner.PlayerID,
				Amount:      winner.Prize,
				Currency:    string(tournament.Currency),
				Description: fmt.Sprintf("Tournament prize: %s (Placement: %d)", tournament.Name, winner.Placement),
				Metadata: map[string]interface{}{
					"tournament_id": tournament.ID.String(),
					"placement":     winner.Placement,
					"type":          "tournament_prize",
				},
			}

			if _, err := s.walletCommand.CreditWallet(ctx, creditCmd); err != nil {
				slog.ErrorContext(ctx, "failed to distribute prize", "player_id", winner.PlayerID, "error", err)
				// Continue distributing other prizes even if one fails
			} else {
				slog.InfoContext(ctx, "prize distributed", "player_id", winner.PlayerID, "amount", winner.Prize.String())
			}
		}
	}

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament completed", "tournament_id", cmd.TournamentID, "winners", len(cmd.Winners))
	return nil
}

// CancelTournament cancels the tournament
func (s *TournamentService) CancelTournament(ctx context.Context, cmd tournament_in.CancelTournamentCommand) error {
	slog.InfoContext(ctx, "cancelling tournament", "tournament_id", cmd.TournamentID, "reason", cmd.Reason)

	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	if err := tournament.Cancel(cmd.Reason); err != nil {
		return fmt.Errorf("failed to cancel tournament: %w", err)
	}

	// Issue refunds to all participants
	if !tournament.EntryFee.IsZero() {
		for _, participant := range tournament.Participants {
			creditCmd := wallet_in.CreditWalletCommand{
				UserID:      participant.PlayerID,
				Amount:      tournament.EntryFee,
				Currency:    string(tournament.Currency),
				Description: fmt.Sprintf("Tournament cancellation refund: %s", tournament.Name),
				Metadata: map[string]interface{}{
					"tournament_id": tournament.ID.String(),
					"type":          "tournament_cancellation_refund",
					"reason":        cmd.Reason,
				},
			}

			if _, err := s.walletCommand.CreditWallet(ctx, creditCmd); err != nil {
				slog.ErrorContext(ctx, "failed to refund participant", "player_id", participant.PlayerID, "error", err)
			} else {
				slog.InfoContext(ctx, "participant refunded", "player_id", participant.PlayerID)
			}
		}
	}

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament cancelled", "tournament_id", cmd.TournamentID)
	return nil
}

// ScheduleMatches automatically schedules all tournament matches
func (s *TournamentService) ScheduleMatches(ctx context.Context, cmd tournament_in.ScheduleMatchesCommand) error {
	slog.InfoContext(ctx, "scheduling tournament matches", "tournament_id", cmd.TournamentID)

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Validate tournament state
	if tournament.Status != tournament_entities.TournamentStatusReady && 
	   tournament.Status != tournament_entities.TournamentStatusInProgress {
		return fmt.Errorf("tournament must be in ready or in_progress status, current: %s", tournament.Status)
	}

	if len(tournament.Matches) == 0 {
		return fmt.Errorf("no matches to schedule - generate brackets first")
	}

	// Determine start time
	startTime := tournament.StartTime
	if cmd.StartTime != nil {
		startTime = *cmd.StartTime
	}

	// Default scheduling parameters
	matchDuration := time.Duration(cmd.MatchDurationMins) * time.Minute
	if matchDuration == 0 {
		matchDuration = 45 * time.Minute // Default 45 min per match
	}

	breakDuration := time.Duration(cmd.BreakBetweenMins) * time.Minute
	if breakDuration == 0 {
		breakDuration = 15 * time.Minute // Default 15 min break
	}

	concurrentMatches := cmd.ConcurrentMatches
	if concurrentMatches == 0 {
		concurrentMatches = 1 // Default sequential
	}

	// Schedule matches by round
	currentTime := startTime
	matchesPerSlot := 0

	for i := range tournament.Matches {
		tournament.Matches[i].ScheduledAt = currentTime
		matchesPerSlot++

		// Move to next time slot if we've scheduled enough concurrent matches
		if matchesPerSlot >= concurrentMatches {
			currentTime = currentTime.Add(matchDuration + breakDuration)
			matchesPerSlot = 0
		}
	}

	// Save updated tournament
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament schedule: %w", err)
	}

	slog.InfoContext(ctx, "tournament matches scheduled",
		"tournament_id", cmd.TournamentID,
		"matches_scheduled", len(tournament.Matches),
		"start_time", startTime,
	)

	return nil
}

// RescheduleMatch reschedules a specific match
func (s *TournamentService) RescheduleMatch(ctx context.Context, cmd tournament_in.RescheduleMatchCommand) error {
	slog.InfoContext(ctx, "rescheduling match",
		"tournament_id", cmd.TournamentID,
		"match_id", cmd.MatchID,
		"new_time", cmd.NewTime,
	)

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Find and update the match
	matchFound := false
	for i := range tournament.Matches {
		if tournament.Matches[i].MatchID == cmd.MatchID {
			// Can only reschedule scheduled or in_progress matches
			if tournament.Matches[i].Status == tournament_entities.MatchStatusCompleted {
				return fmt.Errorf("cannot reschedule completed match")
			}
			if tournament.Matches[i].Status == tournament_entities.MatchStatusCancelled {
				return fmt.Errorf("cannot reschedule cancelled match")
			}

			tournament.Matches[i].ScheduledAt = cmd.NewTime
			matchFound = true
			break
		}
	}

	if !matchFound {
		return fmt.Errorf("match not found: %s", cmd.MatchID)
	}

	// Save updated tournament
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save rescheduled match: %w", err)
	}

	slog.InfoContext(ctx, "match rescheduled",
		"tournament_id", cmd.TournamentID,
		"match_id", cmd.MatchID,
		"new_time", cmd.NewTime,
		"reason", cmd.Reason,
	)

	return nil
}

// ReportMatchResult reports the result of a completed match
func (s *TournamentService) ReportMatchResult(ctx context.Context, cmd tournament_in.ReportMatchResultCommand) error {
	slog.InfoContext(ctx, "reporting match result",
		"tournament_id", cmd.TournamentID,
		"match_id", cmd.MatchID,
		"winner_id", cmd.WinnerID,
		"score", cmd.Score,
	)

	// Fetch tournament
	tournament, err := s.tournamentRepo.FindByID(ctx, cmd.TournamentID)
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	// Validate tournament is in progress
	if tournament.Status != tournament_entities.TournamentStatusInProgress {
		return fmt.Errorf("tournament must be in_progress to report results, current: %s", tournament.Status)
	}

	// Find the match
	matchFound := false
	var matchIdx int
	for i := range tournament.Matches {
		if tournament.Matches[i].MatchID == cmd.MatchID {
			matchIdx = i
			matchFound = true
			break
		}
	}

	if !matchFound {
		return fmt.Errorf("match not found: %s", cmd.MatchID)
	}

	match := &tournament.Matches[matchIdx]

	// Validate match state
	if match.Status == tournament_entities.MatchStatusCompleted {
		return fmt.Errorf("match already completed")
	}
	if match.Status == tournament_entities.MatchStatusCancelled {
		return fmt.Errorf("cannot report result for cancelled match")
	}

	// Validate winner is one of the match participants
	if cmd.WinnerID != match.Player1ID && cmd.WinnerID != match.Player2ID {
		return fmt.Errorf("winner must be one of the match participants")
	}

	// Update match result
	now := time.Now()
	match.WinnerID = &cmd.WinnerID
	match.CompletedAt = &now
	match.Status = tournament_entities.MatchStatusCompleted

	// Check if all matches in the tournament are complete
	allComplete := true
	for _, m := range tournament.Matches {
		if m.Status != tournament_entities.MatchStatusCompleted {
			allComplete = false
			break
		}
	}

	// Save updated tournament
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save match result: %w", err)
	}

	slog.InfoContext(ctx, "match result reported",
		"tournament_id", cmd.TournamentID,
		"match_id", cmd.MatchID,
		"winner_id", cmd.WinnerID,
		"score", cmd.Score,
		"all_matches_complete", allComplete,
	)

	return nil
}

// Ensure TournamentService implements TournamentCommand interface
var _ tournament_in.TournamentCommand = (*TournamentService)(nil)
