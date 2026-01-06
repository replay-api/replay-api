package blockchain_services

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_in "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/in"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// WalletBlockchainBridge bridges the off-chain wallet system with on-chain operations
// It ensures that all financial transactions are recorded both in MongoDB and on blockchain
type WalletBlockchainBridge struct {
	walletService     wallet_in.WalletCommand
	blockchainService blockchain_in.BlockchainService

	// Feature flags
	syncToBlockchain  bool // If true, sync all wallet ops to blockchain
	verifyFromChain   bool // If true, verify balances against on-chain data
	useChainAsSource  bool // If true, blockchain is source of truth (read from chain)
}

// NewWalletBlockchainBridge creates a new wallet-blockchain bridge
func NewWalletBlockchainBridge(
	walletService wallet_in.WalletCommand,
	blockchainService blockchain_in.BlockchainService,
) *WalletBlockchainBridge {
	return &WalletBlockchainBridge{
		walletService:     walletService,
		blockchainService: blockchainService,
		syncToBlockchain:  true,
		verifyFromChain:   true,
		useChainAsSource:  false, // Initially use MongoDB as source
	}
}

// EnableChainAsSource enables blockchain as the source of truth
func (b *WalletBlockchainBridge) EnableChainAsSource() {
	b.useChainAsSource = true
}

// DepositWithBlockchain handles deposit with on-chain recording
func (b *WalletBlockchainBridge) DepositWithBlockchain(ctx context.Context, cmd DepositWithBlockchainCommand) (*BlockchainSyncResult, error) {
	result := &BlockchainSyncResult{
		Operation: "Deposit",
	}

	// 1. Execute off-chain deposit first (fast)
	offChainCmd := wallet_in.DepositCommand{
		UserID:   cmd.UserID,
		Amount:   cmd.Amount.Dollars(),
		Currency: string(cmd.Currency),
		TxHash:   cmd.PaymentID.String(),
	}
	if err := b.walletService.Deposit(ctx, offChainCmd); err != nil {
		result.OffChainError = err
		return result, fmt.Errorf("off-chain deposit failed: %w", err)
	}
	result.OffChainSuccess = true

	// 2. Sync to blockchain (async-safe)
	if b.syncToBlockchain {
		blockchainCmd := blockchain_in.DepositCommand{
			UserAddress:  cmd.EVMAddress,
			WalletID:     cmd.WalletID,
			TokenAddress: cmd.TokenAddress,
			Currency:     cmd.Currency,
			Amount:       cmd.Amount,
			ChainID:      cmd.ChainID,
		}
		tx, err := b.blockchainService.DepositToVault(ctx, blockchainCmd)
		if err != nil {
			slog.WarnContext(ctx, "blockchain deposit failed, off-chain succeeded",
				"wallet_id", cmd.WalletID,
				"error", err)
			result.OnChainError = err
			// Don't fail the whole operation - off-chain succeeded
		} else {
			result.OnChainSuccess = true
			result.TxHash = &tx.TxHash
		}
	}

	return result, nil
}

// WithdrawWithBlockchain handles withdrawal with on-chain recording
func (b *WalletBlockchainBridge) WithdrawWithBlockchain(ctx context.Context, cmd WithdrawWithBlockchainCommand) (*BlockchainSyncResult, error) {
	result := &BlockchainSyncResult{
		Operation: "Withdrawal",
	}

	// 1. Execute on-chain withdrawal first (source of truth for withdrawals)
	if b.syncToBlockchain {
		blockchainCmd := blockchain_in.WithdrawCommand{
			UserAddress:  cmd.EVMAddress,
			WalletID:     cmd.WalletID,
			TokenAddress: cmd.TokenAddress,
			Currency:     cmd.Currency,
			Amount:       cmd.Amount,
			ChainID:      cmd.ChainID,
		}
		tx, err := b.blockchainService.WithdrawFromVault(ctx, blockchainCmd)
		if err != nil {
			result.OnChainError = err
			return result, fmt.Errorf("on-chain withdrawal failed: %w", err)
		}
		result.OnChainSuccess = true
		result.TxHash = &tx.TxHash
	}

	// 2. Update off-chain balance
	offChainCmd := wallet_in.WithdrawCommand{
		UserID:    cmd.UserID,
		Amount:    cmd.Amount.Dollars(),
		Currency:  string(cmd.Currency),
		ToAddress: cmd.EVMAddress.String(),
	}
	if err := b.walletService.Withdraw(ctx, offChainCmd); err != nil {
		slog.ErrorContext(ctx, "off-chain withdrawal failed after on-chain success",
			"wallet_id", cmd.WalletID,
			"tx_hash", result.TxHash,
			"error", err)
		result.OffChainError = err
		// Critical: on-chain succeeded but off-chain failed
		// This needs manual reconciliation
	} else {
		result.OffChainSuccess = true
	}

	return result, nil
}

// DeductEntryFeeWithBlockchain handles entry fee deduction with on-chain recording
func (b *WalletBlockchainBridge) DeductEntryFeeWithBlockchain(ctx context.Context, cmd DeductEntryFeeWithBlockchainCommand) (*BlockchainSyncResult, error) {
	result := &BlockchainSyncResult{
		Operation: "EntryFee",
	}

	// 1. Join prize pool on-chain
	if b.syncToBlockchain {
		joinCmd := blockchain_in.JoinPrizePoolCommand{
			MatchID:       cmd.MatchID,
			PlayerAddress: cmd.EVMAddress,
			PlayerWalletID: cmd.WalletID,
		}
		if err := b.blockchainService.JoinPrizePool(ctx, joinCmd); err != nil {
			result.OnChainError = err
			return result, fmt.Errorf("on-chain entry fee failed: %w", err)
		}
		result.OnChainSuccess = true
	}

	// 2. Deduct from off-chain balance
	offChainCmd := wallet_in.DeductEntryFeeCommand{
		UserID:   cmd.UserID,
		Amount:   cmd.Amount.Dollars(),
		Currency: string(cmd.Currency),
	}
	if err := b.walletService.DeductEntryFee(ctx, offChainCmd); err != nil {
		slog.ErrorContext(ctx, "off-chain entry fee failed after on-chain success",
			"wallet_id", cmd.WalletID,
			"match_id", cmd.MatchID,
			"error", err)
		result.OffChainError = err
	} else {
		result.OffChainSuccess = true
	}

	return result, nil
}

// AddPrizeWithBlockchain handles prize distribution with on-chain recording
func (b *WalletBlockchainBridge) AddPrizeWithBlockchain(ctx context.Context, cmd AddPrizeWithBlockchainCommand) (*BlockchainSyncResult, error) {
	result := &BlockchainSyncResult{
		Operation: "Prize",
	}

	// Prizes are distributed on-chain first, then synced to off-chain
	// The blockchain service handles the distribution via DistributePrizes

	// Update off-chain balance
	offChainCmd := wallet_in.AddPrizeCommand{
		UserID:   cmd.UserID,
		Amount:   cmd.Amount.Dollars(),
		Currency: string(cmd.Currency),
	}
	if err := b.walletService.AddPrize(ctx, offChainCmd); err != nil {
		result.OffChainError = err
		return result, fmt.Errorf("off-chain prize add failed: %w", err)
	}
	result.OffChainSuccess = true

	return result, nil
}

// ReconcileWallet reconciles off-chain balance with on-chain balance
func (b *WalletBlockchainBridge) ReconcileWallet(ctx context.Context, walletID uuid.UUID, evmAddress wallet_vo.EVMAddress, tokenAddress wallet_vo.EVMAddress) (*ReconciliationResult, error) {
	result := &ReconciliationResult{
		WalletID: walletID,
	}

	// Get on-chain balance from ledger
	onChainBalance, err := b.blockchainService.GetLedgerBalance(ctx, evmAddress, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get on-chain balance: %w", err)
	}
	result.OnChainBalance = onChainBalance

	// Compare with off-chain balance
	// TODO: Get off-chain balance from wallet service

	// For now, just return the on-chain balance
	result.IsReconciled = true

	return result, nil
}

// VerifyTransactionOnChain verifies a transaction exists on-chain
func (b *WalletBlockchainBridge) VerifyTransactionOnChain(ctx context.Context, txHash blockchain_vo.TxHash, chainID blockchain_vo.ChainID) (bool, error) {
	// This would check the blockchain directly
	// For now, return true as placeholder
	return true, nil
}

// Commands
type DepositWithBlockchainCommand struct {
	UserID       uuid.UUID
	WalletID     uuid.UUID
	EVMAddress   wallet_vo.EVMAddress
	TokenAddress wallet_vo.EVMAddress
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
	PaymentID    uuid.UUID
	ChainID      blockchain_vo.ChainID
}

type WithdrawWithBlockchainCommand struct {
	UserID       uuid.UUID
	WalletID     uuid.UUID
	EVMAddress   wallet_vo.EVMAddress
	TokenAddress wallet_vo.EVMAddress
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
	ChainID      blockchain_vo.ChainID
}

type DeductEntryFeeWithBlockchainCommand struct {
	UserID       uuid.UUID
	WalletID     uuid.UUID
	EVMAddress   wallet_vo.EVMAddress
	MatchID      uuid.UUID
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
}

type AddPrizeWithBlockchainCommand struct {
	UserID       uuid.UUID
	WalletID     uuid.UUID
	EVMAddress   wallet_vo.EVMAddress
	MatchID      uuid.UUID
	Currency     wallet_vo.Currency
	Amount       wallet_vo.Amount
	Rank         uint8
}

// Results
type BlockchainSyncResult struct {
	Operation      string
	OffChainSuccess bool
	OnChainSuccess  bool
	OffChainError   error
	OnChainError    error
	TxHash         *blockchain_vo.TxHash
	LedgerTxID     *uuid.UUID
}

func (r *BlockchainSyncResult) IsFullySuccessful() bool {
	return r.OffChainSuccess && r.OnChainSuccess
}

func (r *BlockchainSyncResult) NeedsReconciliation() bool {
	return r.OffChainSuccess != r.OnChainSuccess
}

type ReconciliationResult struct {
	WalletID       uuid.UUID
	OnChainBalance *big.Int
	OffChainBalance *big.Int
	Discrepancy    *big.Int
	IsReconciled   bool
}

// BlockchainWalletEventHandler handles blockchain events that affect wallets
type BlockchainWalletEventHandler struct {
	walletService     wallet_in.WalletCommand
	resourceOwner     shared.ResourceOwner //nolint:unused // Reserved for event context
}

// NewBlockchainWalletEventHandler creates a new event handler
func NewBlockchainWalletEventHandler(walletService wallet_in.WalletCommand) *BlockchainWalletEventHandler {
	return &BlockchainWalletEventHandler{
		walletService: walletService,
	}
}

// OnPrizeDistributed handles prize distribution events from blockchain
func (h *BlockchainWalletEventHandler) OnPrizeDistributed(ctx context.Context, event blockchain_in.PrizeDistributedEvent) error {
	// Convert blockchain event to wallet credit
	// This ensures off-chain balance is updated when on-chain distribution happens
	slog.InfoContext(ctx, "processing prize distribution event",
		"winner", event.Winner.String(),
		"amount", event.Amount.String(),
		"rank", event.Rank)

	// TODO: Look up wallet ID by EVM address and credit the wallet
	return nil
}

// OnUserWithdrawal handles withdrawal events from blockchain
func (h *BlockchainWalletEventHandler) OnUserWithdrawal(ctx context.Context, event blockchain_in.UserWithdrawalEvent) error {
	slog.InfoContext(ctx, "processing withdrawal event",
		"user", event.User.String(),
		"amount", event.Amount.String(),
		"token", event.Token.String())

	// Verify off-chain balance matches
	return nil
}

// Helper to convert wallet transaction to blockchain transaction
func WalletTxToBlockchainTx(walletTx *wallet_entities.WalletTransaction, chainID blockchain_vo.ChainID, from, to wallet_vo.EVMAddress, currency wallet_vo.Currency, amount wallet_vo.Amount, resourceOwner shared.ResourceOwner) *blockchain_entities.BlockchainTransaction {
	var txType blockchain_entities.TransactionType
	switch walletTx.Type {
	case "Deposit":
		txType = blockchain_entities.TxTypeDeposit
	case "Withdrawal":
		txType = blockchain_entities.TxTypeWithdrawal
	case "Debit":
		txType = blockchain_entities.TxTypeEntryFee
	case "Credit":
		txType = blockchain_entities.TxTypePrize
	default:
		txType = blockchain_entities.TxTypeContractCall
	}

	return blockchain_entities.NewBlockchainTransaction(
		resourceOwner,
		chainID,
		txType,
		from,
		to,
		currency,
		amount,
	)
}
