package kafka

import (
	"context"

	"github.com/google/uuid"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
)

// WalletEventPublisherAdapter adapts the Kafka EventPublisher to the domain-layer WalletEventPublisher port
// This bridges the infrastructure layer (Kafka) with the domain layer (WalletService)
type WalletEventPublisherAdapter struct {
	publisher *EventPublisher
}

// NewWalletEventPublisherAdapter creates a new adapter
func NewWalletEventPublisherAdapter(publisher *EventPublisher) wallet_services.WalletEventPublisher {
	if publisher == nil {
		return nil
	}
	return &WalletEventPublisherAdapter{publisher: publisher}
}

func (a *WalletEventPublisherAdapter) PublishWalletCreated(ctx context.Context, walletID, userID uuid.UUID) error {
	return a.publisher.PublishWalletCreated(ctx, &WalletEvent{
		WalletID: walletID,
		UserID:   userID,
	})
}

func (a *WalletEventPublisherAdapter) PublishWalletDeposit(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, ledgerTxID uuid.UUID) error {
	return a.publisher.PublishWalletDeposit(ctx, &WalletEvent{
		WalletID:      walletID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		LedgerEntryID: &ledgerTxID,
	})
}

func (a *WalletEventPublisherAdapter) PublishWalletWithdrawal(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, toAddress string, ledgerTxID uuid.UUID) error {
	return a.publisher.PublishWalletWithdrawal(ctx, &WalletEvent{
		WalletID:      walletID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		ToAddress:     toAddress,
		LedgerEntryID: &ledgerTxID,
	})
}

func (a *WalletEventPublisherAdapter) PublishWalletEntryFee(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, matchID, tournamentID *uuid.UUID, ledgerTxID uuid.UUID) error {
	return a.publisher.PublishWalletEntryFee(ctx, &WalletEvent{
		WalletID:      walletID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		MatchID:       matchID,
		TournamentID:  tournamentID,
		LedgerEntryID: &ledgerTxID,
	})
}

func (a *WalletEventPublisherAdapter) PublishWalletPrize(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, matchID, tournamentID *uuid.UUID, ledgerTxID uuid.UUID) error {
	return a.publisher.PublishWalletPrize(ctx, &WalletEvent{
		WalletID:      walletID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		MatchID:       matchID,
		TournamentID:  tournamentID,
		LedgerEntryID: &ledgerTxID,
	})
}

func (a *WalletEventPublisherAdapter) PublishWalletRefund(ctx context.Context, walletID, userID uuid.UUID, amount float64, currency string, reason string, ledgerTxID uuid.UUID) error {
	return a.publisher.PublishWalletRefund(ctx, &WalletEvent{
		WalletID:      walletID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		Description:   reason,
		LedgerEntryID: &ledgerTxID,
	})
}

// Compile-time interface compliance check
var _ wallet_services.WalletEventPublisher = (*WalletEventPublisherAdapter)(nil)
