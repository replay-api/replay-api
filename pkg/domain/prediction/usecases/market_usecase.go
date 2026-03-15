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

// MarketCommandUseCase implements MarketCommand
type MarketCommandUseCase struct {
	marketRepo     prediction_out.MarketRepository
	betRepo        prediction_out.BetRepository
	eventPublisher prediction_out.PredictionEventPublisher
}

func NewMarketCommandUseCase(
	marketRepo prediction_out.MarketRepository,
	betRepo prediction_out.BetRepository,
	eventPublisher prediction_out.PredictionEventPublisher,
) *MarketCommandUseCase {
	return &MarketCommandUseCase{
		marketRepo:     marketRepo,
		betRepo:        betRepo,
		eventPublisher: eventPublisher,
	}
}

func (uc *MarketCommandUseCase) CreateMarket(ctx context.Context, cmd prediction_in.CreateMarketCommand) (*prediction_entities.PredictionMarket, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	userID, ok := ctx.Value(shared.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized: user_id not found in context")
	}

	ro := shared.ResourceOwner{UserID: userID}

	market, err := prediction_entities.NewPredictionMarket(
		ro,
		cmd.MatchID,
		cmd.GameID,
		cmd.BetType,
		cmd.Title,
		cmd.Description,
		cmd.Options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create market: %w", err)
	}

	if err := uc.marketRepo.Save(ctx, market); err != nil {
		return nil, fmt.Errorf("failed to save market: %w", err)
	}

	return market, nil
}

func (uc *MarketCommandUseCase) LockMarket(ctx context.Context, marketID uuid.UUID) error {
	market, err := uc.marketRepo.FindByID(ctx, marketID)
	if err != nil {
		return fmt.Errorf("market not found: %w", err)
	}

	if err := market.Lock(); err != nil {
		return err
	}

	return uc.marketRepo.Update(ctx, market)
}

func (uc *MarketCommandUseCase) ResolveMarket(ctx context.Context, cmd prediction_in.ResolveMarketCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	market, err := uc.marketRepo.FindByID(ctx, cmd.MarketID)
	if err != nil {
		return fmt.Errorf("market not found: %w", err)
	}

	if err := market.Resolve(cmd.OutcomeKey); err != nil {
		return err
	}

	if err := uc.marketRepo.Update(ctx, market); err != nil {
		return fmt.Errorf("failed to update market: %w", err)
	}

	// Resolve all pending bets
	pendingBets, err := uc.betRepo.FindPendingByMarketID(ctx, cmd.MarketID)
	if err != nil {
		return fmt.Errorf("failed to find pending bets: %w", err)
	}

	for _, bet := range pendingBets {
		bet.Resolve(cmd.OutcomeKey)
		if err := uc.betRepo.Update(ctx, bet); err != nil {
			return fmt.Errorf("failed to update bet %s: %w", bet.ID, err)
		}
		// TODO: credit wallet for winning bets
	}

	// Publish event
	if uc.eventPublisher != nil {
		_ = uc.eventPublisher.PublishMarketResolved(ctx, market)
	}

	return nil
}

func (uc *MarketCommandUseCase) CancelMarket(ctx context.Context, marketID uuid.UUID) error {
	market, err := uc.marketRepo.FindByID(ctx, marketID)
	if err != nil {
		return fmt.Errorf("market not found: %w", err)
	}

	if err := market.Cancel(); err != nil {
		return err
	}

	if err := uc.marketRepo.Update(ctx, market); err != nil {
		return fmt.Errorf("failed to update market: %w", err)
	}

	// Refund all pending bets
	pendingBets, err := uc.betRepo.FindPendingByMarketID(ctx, marketID)
	if err != nil {
		return fmt.Errorf("failed to find pending bets: %w", err)
	}

	for _, bet := range pendingBets {
		bet.Refund()
		if err := uc.betRepo.Update(ctx, bet); err != nil {
			return fmt.Errorf("failed to refund bet %s: %w", bet.ID, err)
		}
		// TODO: credit wallet refund
	}

	return nil
}

// MarketQueryUseCase implements MarketQuery
type MarketQueryUseCase struct {
	marketRepo prediction_out.MarketRepository
}

func NewMarketQueryUseCase(marketRepo prediction_out.MarketRepository) *MarketQueryUseCase {
	return &MarketQueryUseCase{marketRepo: marketRepo}
}

func (uc *MarketQueryUseCase) GetMarket(ctx context.Context, marketID uuid.UUID) (*prediction_entities.PredictionMarket, error) {
	return uc.marketRepo.FindByID(ctx, marketID)
}

func (uc *MarketQueryUseCase) ListMatchMarkets(ctx context.Context, query prediction_in.ListMatchMarketsQuery) (*prediction_in.MarketListResult, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	status := ""
	if query.Status != "" {
		status = string(query.Status)
	}

	markets, total, err := uc.marketRepo.FindByMatchID(ctx, query.MatchID, status, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &prediction_in.MarketListResult{
		Markets:    markets,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}
