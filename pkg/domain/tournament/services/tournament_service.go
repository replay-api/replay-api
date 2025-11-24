package tournament_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	tournament_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/entities"
	tournament_in "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/psavelis/team-pro/replay-api/pkg/domain/tournament/ports/out"
	wallet_in "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/ports/in"
)

// TournamentService implements tournament management business logic
type TournamentService struct {
	tournamentRepo tournament_out.TournamentRepository
	walletCommand  wallet_in.WalletCommand
}

// NewTournamentService creates a new tournament service
func NewTournamentService(
	tournamentRepo tournament_out.TournamentRepository,
	walletCommand wallet_in.WalletCommand,
) tournament_in.TournamentCommand {
	return &TournamentService{
		tournamentRepo: tournamentRepo,
		walletCommand:  walletCommand,
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
		deductCmd := wallet_in.DeductEntryFeeCommand{
			UserID:   cmd.PlayerID,
			Currency: string(tournament.Currency),
			Amount:   tournament.EntryFee.Dollars(),
		}

		if err := s.walletCommand.DeductEntryFee(ctx, deductCmd); err != nil {
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
		refundCmd := wallet_in.RefundCommand{
			UserID:   cmd.PlayerID,
			Currency: string(tournament.Currency),
			Amount:   tournament.EntryFee.Dollars(),
			Reason:   fmt.Sprintf("Tournament entry refund: %s", tournament.Name),
		}

		if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
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

	// TODO: Generate bracket matches based on tournament format

	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		return fmt.Errorf("failed to save tournament: %w", err)
	}

	slog.InfoContext(ctx, "tournament started", "tournament_id", tournamentID)
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
			addPrizeCmd := wallet_in.AddPrizeCommand{
				UserID:   winner.PlayerID,
				Currency: string(tournament.Currency),
				Amount:   winner.Prize.Dollars(),
			}

			if err := s.walletCommand.AddPrize(ctx, addPrizeCmd); err != nil {
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
			refundCmd := wallet_in.RefundCommand{
				UserID:   participant.PlayerID,
				Currency: string(tournament.Currency),
				Amount:   tournament.EntryFee.Dollars(),
				Reason:   fmt.Sprintf("Tournament cancellation refund: %s - %s", tournament.Name, cmd.Reason),
			}

			if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
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

// Ensure TournamentService implements TournamentCommand interface
var _ tournament_in.TournamentCommand = (*TournamentService)(nil)
