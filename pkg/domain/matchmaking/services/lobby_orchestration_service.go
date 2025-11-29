package matchmaking_services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	matchmaking_entities "github.com/replay-api/replay-api/pkg/domain/matchmaking/entities"
	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
	ws "github.com/replay-api/replay-api/pkg/infra/websocket"
)

// LobbyOrchestrationService coordinates lobby → prize pool → wallet operations with Saga pattern
type LobbyOrchestrationService struct {
	lobbyRepo     matchmaking_out.LobbyRepository
	prizePoolRepo matchmaking_out.PrizePoolRepository
	walletCommand wallet_in.WalletCommand
	wsHub         *ws.WebSocketHub
}

func NewLobbyOrchestrationService(
lobbyRepo matchmaking_out.LobbyRepository,
prizePoolRepo matchmaking_out.PrizePoolRepository,
walletCommand wallet_in.WalletCommand,
wsHub *ws.WebSocketHub,
) matchmaking_in.LobbyCommand {
	return &LobbyOrchestrationService{
		lobbyRepo:     lobbyRepo,
		prizePoolRepo: prizePoolRepo,
		walletCommand: walletCommand,
		wsHub:         wsHub,
	}
}

func (s *LobbyOrchestrationService) CreateLobby(ctx context.Context, cmd matchmaking_in.CreateLobbyCommand) (*matchmaking_entities.MatchmakingLobby, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	lobby, err := matchmaking_entities.NewMatchmakingLobby(
resourceOwner,
cmd.CreatorID,
cmd.GameID,
cmd.Region,
cmd.Tier,
cmd.DistributionRule,
cmd.MaxPlayers,
cmd.AutoFill,
cmd.InviteOnly,
)
	if err != nil {
		return nil, fmt.Errorf("failed to create lobby: %w", err)
	}

	if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
		return nil, fmt.Errorf("failed to save lobby: %w", err)
	}

	// Create associated prize pool
	prizePoolCmd := matchmaking_in.CreatePrizePoolCommand{
		MatchID:              lobby.ID,
		GameID:               cmd.GameID,
		Region:               cmd.Region,
		Currency:             "USD",
		PlatformContribution: 0.50,
		DistributionRule:     cmd.DistributionRule,
	}

	_, err = s.createPrizePool(ctx, prizePoolCmd)
	if err != nil {
		// Rollback: delete lobby
		_ = s.lobbyRepo.Delete(ctx, lobby.ID)
		return nil, fmt.Errorf("failed to create prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Lobby created", "lobby_id", lobby.ID, "creator", cmd.CreatorID)
	s.wsHub.BroadcastLobbyUpdate(lobby.ID, lobby)

	return lobby, nil
}

// JoinLobby implements Saga pattern: Deduct entry fee → Add to lobby → Add to prize pool (rollback on failure)
func (s *LobbyOrchestrationService) JoinLobby(ctx context.Context, cmd matchmaking_in.JoinLobbyCommand) error {
	// Step 1: Load lobby
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	// Step 2: Load prize pool
	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Step 3: Calculate entry fee by tier
	entryFee := getEntryFeeByTier(lobby.Tier)

	// Step 4: Deduct entry fee from wallet (with rollback support)
	walletCmd := wallet_in.DeductEntryFeeCommand{
		UserID:   cmd.PlayerID,
		Currency: string(prizePool.Currency),
		Amount:   entryFee,
	}

	if err := s.walletCommand.DeductEntryFee(ctx, walletCmd); err != nil {
		return fmt.Errorf("insufficient balance: %w", err)
	}

	// Step 5: Add player to lobby
	if err := lobby.AddPlayer(cmd.PlayerID, cmd.MMR); err != nil {
		// Rollback: refund entry fee
		refundCmd := wallet_in.RefundCommand{
			UserID:   cmd.PlayerID,
			Currency: string(prizePool.Currency),
			Amount:   entryFee,
			Reason:   "failed to join lobby",
		}
		_ = s.walletCommand.Refund(ctx, refundCmd)
		return fmt.Errorf("failed to add player: %w", err)
	}

	// Step 6: Add contribution to prize pool
	prizePoolAmount := wallet_vo.NewAmount(entryFee)

	if err := prizePool.AddPlayerContribution(cmd.PlayerID, prizePoolAmount); err != nil {
		// Rollback: remove from lobby + refund
		_ = lobby.RemovePlayer(cmd.PlayerID)
		_ = s.lobbyRepo.Update(ctx, lobby)
		refundCmd := wallet_in.RefundCommand{
			UserID:   cmd.PlayerID,
			Currency: string(prizePool.Currency),
			Amount:   entryFee,
			Reason:   "failed to add prize contribution",
		}
		_ = s.walletCommand.Refund(ctx, refundCmd)
		return fmt.Errorf("failed to add prize contribution: %w", err)
	}

	// Step 7: Persist changes
	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}
	if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Player joined lobby", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)

	// Step 8: Broadcast updates via WebSocket
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)
	s.wsHub.BroadcastPrizePoolUpdate(cmd.LobbyID, prizePool)

	return nil
}

func (s *LobbyOrchestrationService) LeaveLobby(ctx context.Context, cmd matchmaking_in.LeaveLobbyCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Get player contribution before removal
	var playerContribution float64
	if contribution, exists := prizePool.PlayerContributions[cmd.PlayerID]; exists {
		playerContribution = contribution.ToFloat()
	}

	if err := lobby.RemovePlayer(cmd.PlayerID); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	// Refund entry fee
	if playerContribution > 0 {
		refundCmd := wallet_in.RefundCommand{
			UserID:   cmd.PlayerID,
			Currency: string(prizePool.Currency),
			Amount:   playerContribution,
			Reason:   "left lobby",
		}
		if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
			slog.ErrorContext(ctx, "Failed to refund player", "error", err)
		}
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Player left lobby", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) SetPlayerReady(ctx context.Context, cmd matchmaking_in.SetPlayerReadyCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	if err := lobby.SetPlayerReady(cmd.PlayerID, cmd.IsReady); err != nil {
		return fmt.Errorf("failed to set ready: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Player ready status changed", "lobby_id", cmd.LobbyID, "player_id", cmd.PlayerID, "is_ready", cmd.IsReady)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) StartReadyCheck(ctx context.Context, cmd matchmaking_in.StartReadyCheckCommand) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, cmd.LobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	if err := lobby.StartReadyCheck(); err != nil {
		return fmt.Errorf("failed to start ready check: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Ready check started", "lobby_id", cmd.LobbyID)
	s.wsHub.BroadcastLobbyUpdate(cmd.LobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) StartMatch(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error) {
	lobby, err := s.lobbyRepo.FindByID(ctx, lobbyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, lobbyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("prize pool not found: %w", err)
	}

	matchID := uuid.New()

	if err := lobby.StartMatch(matchID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to start match: %w", err)
	}

	if err := prizePool.Lock(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to lock prize pool: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update lobby: %w", err)
	}

	if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update prize pool: %w", err)
	}

	if err := lobby.MarkMatchStarted(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to mark match started: %w", err)
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update lobby: %w", err)
	}

	slog.InfoContext(ctx, "Match started", "lobby_id", lobbyID, "match_id", matchID)
	s.wsHub.BroadcastLobbyUpdate(lobbyID, lobby)

	return matchID, nil
}

func (s *LobbyOrchestrationService) CancelLobby(ctx context.Context, lobbyID uuid.UUID, reason string) error {
	lobby, err := s.lobbyRepo.FindByID(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	prizePool, err := s.prizePoolRepo.FindByMatchID(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if err := lobby.Cancel(reason); err != nil {
		return fmt.Errorf("failed to cancel lobby: %w", err)
	}

	if err := prizePool.Cancel(reason); err != nil {
		return fmt.Errorf("failed to cancel prize pool: %w", err)
	}

	// Refund all players
	for playerID, contribution := range prizePool.PlayerContributions {
		refundCmd := wallet_in.RefundCommand{
			UserID:   playerID,
			Currency: string(prizePool.Currency),
			Amount:   contribution.ToFloat(),
			Reason:   fmt.Sprintf("lobby cancelled: %s", reason),
		}
		if err := s.walletCommand.Refund(ctx, refundCmd); err != nil {
			slog.ErrorContext(ctx, "Failed to refund player", "player_id", playerID, "error", err)
		}
	}

	if err := s.lobbyRepo.Update(ctx, lobby); err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	if err := s.prizePoolRepo.Update(ctx, prizePool); err != nil {
		return fmt.Errorf("failed to update prize pool: %w", err)
	}

	slog.InfoContext(ctx, "Lobby cancelled", "lobby_id", lobbyID, "reason", reason)
	s.wsHub.BroadcastLobbyUpdate(lobbyID, lobby)

	return nil
}

func (s *LobbyOrchestrationService) createPrizePool(ctx context.Context, cmd matchmaking_in.CreatePrizePoolCommand) (*matchmaking_entities.PrizePool, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	currency, err := wallet_vo.ParseCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	platformAmount := wallet_vo.NewAmount(cmd.PlatformContribution)

	pool := matchmaking_entities.NewPrizePool(
		resourceOwner,
		cmd.MatchID,
		common.GameIDKey(cmd.GameID),
		cmd.Region,
		currency,
		cmd.DistributionRule,
		platformAmount,
	)

	if err := s.prizePoolRepo.Save(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to save prize pool: %w", err)
	}

	return pool, nil
}

func getEntryFeeByTier(tier string) float64 {
	fees := map[string]float64{
		"free":    0.00,
		"premium": 1.00,
		"pro":     2.00,
		"elite":   5.00,
	}
	if fee, exists := fees[tier]; exists {
		return fee
	}
	return 0.00
}
