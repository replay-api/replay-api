package prediction_usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"

	prediction_entities "github.com/replay-api/replay-api/pkg/domain/prediction/entities"
	prediction_in "github.com/replay-api/replay-api/pkg/domain/prediction/ports/in"
	prediction_out "github.com/replay-api/replay-api/pkg/domain/prediction/ports/out"
)

// BetCommandUseCase implements BetCommand
type BetCommandUseCase struct {
	betRepo        prediction_out.BetRepository
	marketRepo     prediction_out.MarketRepository
	eventPublisher prediction_out.PredictionEventPublisher
}

func NewBetCommandUseCase(
	betRepo prediction_out.BetRepository,
	marketRepo prediction_out.MarketRepository,
	eventPublisher prediction_out.PredictionEventPublisher,
) *BetCommandUseCase {
	return &BetCommandUseCase{
		betRepo:        betRepo,
		marketRepo:     marketRepo,
		eventPublisher: eventPublisher,
	}
}

func (uc *BetCommandUseCase) PlaceBet(ctx context.Context, cmd prediction_in.PlaceBetCommand) (*prediction_entities.Bet, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	userID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized: user_id not found in context")
	}

	// Load market and verify open
	market, err := uc.marketRepo.FindByID(ctx, cmd.MarketID)
	if err != nil {
		return nil, fmt.Errorf("market not found: %w", err)
	}

	if market.Status != prediction_entities.PredictionStatusOpen {
		return nil, fmt.Errorf("market is not open for betting (status: %s)", market.Status)
	}

	// Find the option and get current odds
	var oddsAtPlace float64
	found := false
	for _, opt := range market.Options {
		if opt.Key == cmd.OptionKey {
			oddsAtPlace = opt.Odds
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("option %q not found in market", cmd.OptionKey)
	}

	ro := shared.ResourceOwner{UserID: userID}

	bet, err := prediction_entities.NewBet(
		ro,
		market.ID,
		market.MatchID,
		userID,
		cmd.OptionKey,
		cmd.Amount,
		oddsAtPlace,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bet: %w", err)
	}

	// Update market pool
	if err := market.AddStake(cmd.OptionKey, cmd.Amount); err != nil {
		return nil, fmt.Errorf("failed to update market pool: %w", err)
	}

	// Save bet
	if err := uc.betRepo.Save(ctx, bet); err != nil {
		return nil, fmt.Errorf("failed to save bet: %w", err)
	}

	// Update market
	if err := uc.marketRepo.Update(ctx, market); err != nil {
		return nil, fmt.Errorf("failed to update market: %w", err)
	}

	// Publish event
	if uc.eventPublisher != nil {
		_ = uc.eventPublisher.PublishBetPlaced(ctx, bet)
	}

	return bet, nil
}

// BetQueryUseCase implements BetQuery
type BetQueryUseCase struct {
	betRepo    prediction_out.BetRepository
	marketRepo prediction_out.MarketRepository
}

func NewBetQueryUseCase(
	betRepo prediction_out.BetRepository,
	marketRepo prediction_out.MarketRepository,
) *BetQueryUseCase {
	return &BetQueryUseCase{
		betRepo:    betRepo,
		marketRepo: marketRepo,
	}
}

func (uc *BetQueryUseCase) GetUserBets(ctx context.Context, query prediction_in.GetUserBetsQuery) (*prediction_in.BetListResult, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	status := ""
	if query.Status != "" {
		status = string(query.Status)
	}

	bets, total, err := uc.betRepo.FindByUserID(ctx, query.UserID, status, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &prediction_in.BetListResult{
		Bets:       bets,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}

func (uc *BetQueryUseCase) GetMarketBets(ctx context.Context, marketID uuid.UUID, limit, offset int) (*prediction_in.BetListResult, error) {
	bets, total, err := uc.betRepo.FindByMarketID(ctx, marketID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &prediction_in.BetListResult{
		Bets:       bets,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (uc *BetQueryUseCase) GetUserBetSummary(ctx context.Context, marketID uuid.UUID, userID uuid.UUID) (*prediction_entities.UserBetSummary, error) {
	bets, err := uc.betRepo.FindByMarketAndUser(ctx, marketID, userID)
	if err != nil {
		return nil, err
	}

	summary := &prediction_entities.UserBetSummary{
		MarketID: marketID,
		UserID:   userID,
		Bets:     bets,
	}

	for _, bet := range bets {
		summary.TotalStaked += bet.Amount
		summary.TotalPayout += bet.Payout
		summary.BetCount++
	}

	return summary, nil
}

func (uc *BetQueryUseCase) GetLeaderboard(ctx context.Context, limit int) ([]*prediction_entities.BetLeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return uc.betRepo.GetLeaderboard(ctx, limit)
}
