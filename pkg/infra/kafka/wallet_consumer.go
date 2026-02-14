package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
	"github.com/segmentio/kafka-go"
)

// WalletConsumer processes wallet events from Kafka
type WalletConsumer struct {
	consumer       *Consumer
	walletRepo     wallet_out.WalletRepository
	ledgerRepo     wallet_out.LedgerRepository
	walletCommand  wallet_in.WalletCommand
	eventPublisher *EventPublisher
}

// WalletConsumerConfig holds configuration for wallet consumer
type WalletConsumerConfig struct {
	GroupID        string
	WalletRepo     wallet_out.WalletRepository
	LedgerRepo     wallet_out.LedgerRepository
	WalletCommand  wallet_in.WalletCommand
	EventPublisher *EventPublisher
}

// NewWalletConsumer creates a new wallet consumer
func NewWalletConsumer(client *Client, config *WalletConsumerConfig) *WalletConsumer {
	topics := GetAllWalletTopics()

	consumerConfig := DefaultConsumerConfig(config.GroupID, topics)
	consumer := NewConsumer(client, consumerConfig)

	wc := &WalletConsumer{
		consumer:       consumer,
		walletRepo:     config.WalletRepo,
		ledgerRepo:     config.LedgerRepo,
		walletCommand:  config.WalletCommand,
		eventPublisher: config.EventPublisher,
	}

	// Register handlers for each wallet topic
	consumer.RegisterHandler(TopicWalletCreated, wc.handleWalletCreated)
	consumer.RegisterHandler(TopicWalletDeposit, wc.handleDeposit)
	consumer.RegisterHandler(TopicWalletWithdrawal, wc.handleWithdrawal)
	consumer.RegisterHandler(TopicWalletWithdrawalPending, wc.handleWithdrawalPending)
	consumer.RegisterHandler(TopicWalletEntryFee, wc.handleEntryFee)
	consumer.RegisterHandler(TopicWalletPrize, wc.handlePrize)
	consumer.RegisterHandler(TopicWalletRefund, wc.handleRefund)
	consumer.RegisterHandler(TopicWalletLocked, wc.handleWalletLocked)
	consumer.RegisterHandler(TopicWalletUnlocked, wc.handleWalletUnlocked)

	return wc
}

func (wc *WalletConsumer) handleWalletCreated(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal wallet created event", "error", err)
		return fmt.Errorf("failed to unmarshal wallet created event: %w", err)
	}

	slog.InfoContext(ctx, "Processing wallet created event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
	)

	// Audit trail for wallet creation
	slog.InfoContext(ctx, "[AUDIT] Wallet created",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleDeposit(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal deposit event", "error", err)
		return fmt.Errorf("failed to unmarshal deposit event: %w", err)
	}

	slog.InfoContext(ctx, "Processing deposit event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
	)

	// Audit trail for deposit
	slog.InfoContext(ctx, "[AUDIT] Deposit processed",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"balance_before", event.BalanceBefore,
		"balance_after", event.BalanceAfter,
		"tx_hash", event.TxHash,
		"ledger_entry_id", event.LedgerEntryID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleWithdrawal(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal withdrawal event", "error", err)
		return fmt.Errorf("failed to unmarshal withdrawal event: %w", err)
	}

	slog.InfoContext(ctx, "Processing withdrawal event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"to_address", event.ToAddress,
	)

	// Audit trail for withdrawal
	slog.InfoContext(ctx, "[AUDIT] Withdrawal processed",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"to_address", event.ToAddress,
		"balance_before", event.BalanceBefore,
		"balance_after", event.BalanceAfter,
		"tx_hash", event.TxHash,
		"ledger_entry_id", event.LedgerEntryID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleWithdrawalPending(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal pending withdrawal event", "error", err)
		return fmt.Errorf("failed to unmarshal pending withdrawal event: %w", err)
	}

	slog.InfoContext(ctx, "Processing pending withdrawal event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"to_address", event.ToAddress,
	)

	// Track pending withdrawals for fraud detection
	slog.InfoContext(ctx, "[AUDIT] Withdrawal pending review",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"to_address", event.ToAddress,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleEntryFee(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal entry fee event", "error", err)
		return fmt.Errorf("failed to unmarshal entry fee event: %w", err)
	}

	slog.InfoContext(ctx, "Processing entry fee event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"match_id", event.MatchID,
		"tournament_id", event.TournamentID,
	)

	// Audit trail for entry fee
	slog.InfoContext(ctx, "[AUDIT] Entry fee deducted",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"match_id", event.MatchID,
		"tournament_id", event.TournamentID,
		"balance_before", event.BalanceBefore,
		"balance_after", event.BalanceAfter,
		"ledger_entry_id", event.LedgerEntryID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handlePrize(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal prize event", "error", err)
		return fmt.Errorf("failed to unmarshal prize event: %w", err)
	}

	slog.InfoContext(ctx, "Processing prize event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"match_id", event.MatchID,
		"tournament_id", event.TournamentID,
	)

	// Check for fraud patterns
	if event.Amount > 1000 {
		slog.WarnContext(ctx, "[FRAUD_CHECK] Large prize detected - manual review required",
			"event_id", event.EventID,
			"wallet_id", event.WalletID,
			"user_id", event.UserID,
			"amount", event.Amount,
		)
	}

	// Audit trail for prize
	slog.InfoContext(ctx, "[AUDIT] Prize credited",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"match_id", event.MatchID,
		"tournament_id", event.TournamentID,
		"balance_before", event.BalanceBefore,
		"balance_after", event.BalanceAfter,
		"ledger_entry_id", event.LedgerEntryID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleRefund(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal refund event", "error", err)
		return fmt.Errorf("failed to unmarshal refund event: %w", err)
	}

	slog.InfoContext(ctx, "Processing refund event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
	)

	// Audit trail for refund
	slog.InfoContext(ctx, "[AUDIT] Refund processed",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"amount", event.Amount,
		"currency", event.Currency,
		"description", event.Description,
		"balance_before", event.BalanceBefore,
		"balance_after", event.BalanceAfter,
		"ledger_entry_id", event.LedgerEntryID,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleWalletLocked(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal wallet locked event", "error", err)
		return fmt.Errorf("failed to unmarshal wallet locked event: %w", err)
	}

	slog.WarnContext(ctx, "Processing wallet locked event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"lock_reason", event.LockReason,
	)

	// Security audit for wallet lock
	slog.WarnContext(ctx, "[SECURITY] Wallet locked",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"lock_reason", event.LockReason,
		"timestamp", event.Timestamp,
	)

	return nil
}

func (wc *WalletConsumer) handleWalletUnlocked(ctx context.Context, msg *kafka.Message) error {
	var event WalletEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal wallet unlocked event", "error", err)
		return fmt.Errorf("failed to unmarshal wallet unlocked event: %w", err)
	}

	slog.InfoContext(ctx, "Processing wallet unlocked event",
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
	)

	// Security audit for wallet unlock
	slog.InfoContext(ctx, "[SECURITY] Wallet unlocked",
		"event_id", event.EventID,
		"wallet_id", event.WalletID,
		"user_id", event.UserID,
		"timestamp", event.Timestamp,
	)

	return nil
}

// Start begins consuming wallet events
func (wc *WalletConsumer) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting wallet consumer")
	return wc.consumer.Start(ctx)
}

// Close closes the wallet consumer
func (wc *WalletConsumer) Close() error {
	return wc.consumer.Close()
}

// setWalletResourceOwnerContext sets the resource owner context for wallet operations
func setWalletResourceOwnerContext(ctx context.Context, userID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, shared.UserIDKey, userID)
	ctx = context.WithValue(ctx, shared.AuthenticatedKey, true)
	return ctx
}
