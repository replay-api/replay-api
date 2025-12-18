package blockchain_services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	blockchain_entities "github.com/replay-api/replay-api/pkg/domain/blockchain/entities"
	blockchain_in "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/in"
	blockchain_out "github.com/replay-api/replay-api/pkg/domain/blockchain/ports/out"
	blockchain_vo "github.com/replay-api/replay-api/pkg/domain/blockchain/value-objects"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// BlockchainServiceImpl implements the BlockchainService interface
type BlockchainServiceImpl struct {
	chainManager      blockchain_out.MultiChainClient
	vaultContract     blockchain_out.VaultContract
	ledgerContract    blockchain_out.LedgerContract
	txRepo            blockchain_out.TransactionRepository
	prizePoolRepo     blockchain_out.PrizePoolRepository
	syncStateRepo     blockchain_out.SyncStateRepository

	// Contract addresses per chain
	vaultAddresses  map[blockchain_vo.ChainID]wallet_vo.EVMAddress
	ledgerAddresses map[blockchain_vo.ChainID]wallet_vo.EVMAddress
	tokenAddresses  map[blockchain_vo.ChainID]map[wallet_vo.Currency]wallet_vo.EVMAddress

	// Configuration
	escrowDuration    time.Duration
	platformContribution wallet_vo.Amount
	defaultChainID    blockchain_vo.ChainID
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService(
	chainManager blockchain_out.MultiChainClient,
	vaultContract blockchain_out.VaultContract,
	ledgerContract blockchain_out.LedgerContract,
	txRepo blockchain_out.TransactionRepository,
	prizePoolRepo blockchain_out.PrizePoolRepository,
	syncStateRepo blockchain_out.SyncStateRepository,
) *BlockchainServiceImpl {
	return &BlockchainServiceImpl{
		chainManager:         chainManager,
		vaultContract:        vaultContract,
		ledgerContract:       ledgerContract,
		txRepo:               txRepo,
		prizePoolRepo:        prizePoolRepo,
		syncStateRepo:        syncStateRepo,
		vaultAddresses:       make(map[blockchain_vo.ChainID]wallet_vo.EVMAddress),
		ledgerAddresses:      make(map[blockchain_vo.ChainID]wallet_vo.EVMAddress),
		tokenAddresses:       make(map[blockchain_vo.ChainID]map[wallet_vo.Currency]wallet_vo.EVMAddress),
		escrowDuration:       72 * time.Hour,
		platformContribution: wallet_vo.NewAmount(50), // $0.50
		defaultChainID:       blockchain_vo.PrimaryChain(),
	}
}

// SetContractAddress sets a contract address for a chain
func (s *BlockchainServiceImpl) SetVaultAddress(chainID blockchain_vo.ChainID, addr wallet_vo.EVMAddress) {
	s.vaultAddresses[chainID] = addr
}

func (s *BlockchainServiceImpl) SetLedgerAddress(chainID blockchain_vo.ChainID, addr wallet_vo.EVMAddress) {
	s.ledgerAddresses[chainID] = addr
}

func (s *BlockchainServiceImpl) SetTokenAddress(chainID blockchain_vo.ChainID, currency wallet_vo.Currency, addr wallet_vo.EVMAddress) {
	if s.tokenAddresses[chainID] == nil {
		s.tokenAddresses[chainID] = make(map[wallet_vo.Currency]wallet_vo.EVMAddress)
	}
	s.tokenAddresses[chainID][currency] = addr
}

// CreatePrizePool creates a new on-chain prize pool
func (s *BlockchainServiceImpl) CreatePrizePool(ctx context.Context, cmd blockchain_in.CreatePrizePoolCommand) (*blockchain_entities.OnChainPrizePool, error) {
	// Validate chain
	chainID := cmd.ChainID
	if chainID == 0 {
		chainID = s.defaultChainID
	}

	vaultAddr, ok := s.vaultAddresses[chainID]
	if !ok {
		return nil, fmt.Errorf("vault not deployed on chain %d", chainID)
	}

	// Get resource owner from context
	resourceOwner := common.GetResourceOwner(ctx)

	// Create pool entity
	pool := blockchain_entities.NewOnChainPrizePool(
		resourceOwner,
		cmd.MatchID,
		chainID,
		vaultAddr,
		cmd.TokenAddress,
		cmd.Currency,
		cmd.EntryFee,
		cmd.PlatformFeePercent,
	)

	// Call contract to create pool
	entryFeeBig := big.NewInt(int64(cmd.EntryFee.Cents()))
	txHash, err := s.vaultContract.CreatePrizePool(ctx, cmd.MatchID, cmd.TokenAddress, entryFeeBig, cmd.PlatformFeePercent)
	if err != nil {
		return nil, fmt.Errorf("failed to create prize pool on-chain: %w", err)
	}

	// Update pool with tx info
	pool.MarkCreated(txHash, s.platformContribution)

	// Create blockchain transaction record
	tx := blockchain_entities.NewBlockchainTransaction(
		resourceOwner,
		chainID,
		blockchain_entities.TxTypeContractCall,
		wallet_vo.EVMAddress{}, // Operator address
		vaultAddr,
		cmd.Currency,
		wallet_vo.NewAmount(0),
	)
	tx.SetTxHash(txHash)
	tx.SetRelatedMatch(cmd.MatchID)
	tx.Metadata["operation"] = "CreatePrizePool"

	// Save to cache
	if err := s.prizePoolRepo.Save(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to save prize pool: %w", err)
	}
	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return pool, nil
}

// JoinPrizePool adds a player to a prize pool
func (s *BlockchainServiceImpl) JoinPrizePool(ctx context.Context, cmd blockchain_in.JoinPrizePoolCommand) error {
	// Get existing pool
	pool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.MatchID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	if pool.Status != blockchain_entities.PoolStatusAccumulating {
		return fmt.Errorf("pool is not accepting entries: status=%s", pool.Status)
	}

	// Call contract to deposit entry fee
	txHash, err := s.vaultContract.DepositEntryFee(ctx, cmd.MatchID, cmd.PlayerAddress)
	if err != nil {
		return fmt.Errorf("failed to deposit entry fee: %w", err)
	}

	// Update pool
	if err := pool.AddParticipant(cmd.PlayerAddress, pool.EntryFeePerPlayer); err != nil {
		return err
	}

	// Record transaction
	resourceOwner := common.GetResourceOwner(ctx)
	tx := blockchain_entities.NewBlockchainTransaction(
		resourceOwner,
		pool.ChainID,
		blockchain_entities.TxTypeEntryFee,
		cmd.PlayerAddress,
		pool.ContractAddr,
		pool.Currency,
		pool.EntryFeePerPlayer,
	)
	tx.SetTxHash(txHash)
	tx.SetRelatedMatch(cmd.MatchID)
	tx.SetWallet(cmd.PlayerWalletID)

	// Record in ledger
	if err := s.recordLedgerEntry(ctx, uuid.New(), cmd.PlayerAddress, pool.TokenAddress,
		big.NewInt(-int64(pool.EntryFeePerPlayer.Cents())), "ENTRY_FEE", &cmd.MatchID); err != nil {
		// Log but don't fail - ledger is secondary
		fmt.Printf("Warning: failed to record ledger entry: %v\n", err)
	}

	// Save updates
	if err := s.prizePoolRepo.Save(ctx, pool); err != nil {
		return fmt.Errorf("failed to update prize pool: %w", err)
	}
	if err := s.txRepo.Save(ctx, tx); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}

// LockPrizePool locks the prize pool when match starts
func (s *BlockchainServiceImpl) LockPrizePool(ctx context.Context, matchID uuid.UUID) error {
	pool, err := s.prizePoolRepo.FindByMatchID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Call contract
	txHash, err := s.vaultContract.LockPrizePool(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to lock prize pool: %w", err)
	}

	// Update pool
	if err := pool.Lock(txHash); err != nil {
		return err
	}

	return s.prizePoolRepo.Save(ctx, pool)
}

// DistributePrizes distributes prizes to winners
func (s *BlockchainServiceImpl) DistributePrizes(ctx context.Context, cmd blockchain_in.DistributePrizesCommand) error {
	pool, err := s.prizePoolRepo.FindByMatchID(ctx, cmd.MatchID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Check escrow
	if !pool.IsEscrowComplete() {
		return fmt.Errorf("escrow period not complete")
	}

	// Prepare winner data for contract
	addresses := make([]wallet_vo.EVMAddress, len(cmd.Winners))
	shares := make([]uint16, len(cmd.Winners))
	for i, w := range cmd.Winners {
		addresses[i] = w.Address
		shares[i] = w.ShareBPS
	}

	// Call contract
	txHash, err := s.vaultContract.DistributePrizes(ctx, cmd.MatchID, addresses, shares)
	if err != nil {
		return fmt.Errorf("failed to distribute prizes: %w", err)
	}

	// Calculate winner amounts
	distributable := pool.GetDistributableAmount()
	winners := make([]blockchain_entities.PrizeWinner, len(cmd.Winners))
	for i, w := range cmd.Winners {
		prizeAmount := wallet_vo.NewAmountFromCents(distributable.Cents() * int64(w.ShareBPS) / 10000)
		winners[i] = blockchain_entities.PrizeWinner{
			Address:  w.Address,
			Rank:     w.Rank,
			Amount:   prizeAmount,
			ShareBPS: w.ShareBPS,
			IsMVP:    w.IsMVP,
		}

		// Record prize in ledger
		if err := s.recordLedgerEntry(ctx, uuid.New(), w.Address, pool.TokenAddress,
			big.NewInt(int64(prizeAmount.Cents())), "PRIZE", &cmd.MatchID); err != nil {
			fmt.Printf("Warning: failed to record prize ledger entry: %v\n", err)
		}
	}

	// Calculate platform fee
	platformFee := wallet_vo.NewAmountFromCents(pool.TotalAmount.Cents() * int64(pool.PlatformFeePercent) / 10000)

	// Update pool
	if err := pool.Distribute(txHash, winners, platformFee); err != nil {
		return err
	}

	return s.prizePoolRepo.Save(ctx, pool)
}

// CancelPrizePool cancels a prize pool and refunds participants
func (s *BlockchainServiceImpl) CancelPrizePool(ctx context.Context, matchID uuid.UUID) error {
	pool, err := s.prizePoolRepo.FindByMatchID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("prize pool not found: %w", err)
	}

	// Call contract
	_, err = s.vaultContract.CancelPrizePool(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to cancel prize pool: %w", err)
	}

	// Record refunds in ledger
	for _, participant := range pool.Participants {
		contribution := pool.Contributions[participant.String()]
		if err := s.recordLedgerEntry(ctx, uuid.New(), participant, pool.TokenAddress,
			big.NewInt(int64(contribution.Cents())), "REFUND", &matchID); err != nil {
			fmt.Printf("Warning: failed to record refund ledger entry: %v\n", err)
		}
	}

	// Update pool
	if err := pool.Cancel(); err != nil {
		return err
	}

	return s.prizePoolRepo.Save(ctx, pool)
}

// DepositToVault deposits tokens to user's vault balance
func (s *BlockchainServiceImpl) DepositToVault(ctx context.Context, cmd blockchain_in.DepositCommand) (*blockchain_entities.BlockchainTransaction, error) {
	chainID := cmd.ChainID
	if chainID == 0 {
		chainID = s.defaultChainID
	}

	amountBig := big.NewInt(int64(cmd.Amount.Cents()))
	txHash, err := s.vaultContract.Deposit(ctx, cmd.UserAddress, cmd.TokenAddress, amountBig)
	if err != nil {
		return nil, fmt.Errorf("failed to deposit: %w", err)
	}

	// Create transaction record
	resourceOwner := common.GetResourceOwner(ctx)
	tx := blockchain_entities.NewBlockchainTransaction(
		resourceOwner,
		chainID,
		blockchain_entities.TxTypeDeposit,
		cmd.UserAddress,
		s.vaultAddresses[chainID],
		cmd.Currency,
		cmd.Amount,
	)
	tx.SetTxHash(txHash)
	tx.SetWallet(cmd.WalletID)

	// Record in ledger
	if err := s.recordLedgerEntry(ctx, uuid.New(), cmd.UserAddress, cmd.TokenAddress,
		amountBig, "DEPOSIT", nil); err != nil {
		fmt.Printf("Warning: failed to record deposit ledger entry: %v\n", err)
	}

	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return tx, nil
}

// WithdrawFromVault withdraws tokens from user's vault balance
func (s *BlockchainServiceImpl) WithdrawFromVault(ctx context.Context, cmd blockchain_in.WithdrawCommand) (*blockchain_entities.BlockchainTransaction, error) {
	chainID := cmd.ChainID
	if chainID == 0 {
		chainID = s.defaultChainID
	}

	amountBig := big.NewInt(int64(cmd.Amount.Cents()))
	txHash, err := s.vaultContract.Withdraw(ctx, cmd.UserAddress, cmd.TokenAddress, amountBig)
	if err != nil {
		return nil, fmt.Errorf("failed to withdraw: %w", err)
	}

	// Create transaction record
	resourceOwner := common.GetResourceOwner(ctx)
	tx := blockchain_entities.NewBlockchainTransaction(
		resourceOwner,
		chainID,
		blockchain_entities.TxTypeWithdrawal,
		s.vaultAddresses[chainID],
		cmd.UserAddress,
		cmd.Currency,
		cmd.Amount,
	)
	tx.SetTxHash(txHash)
	tx.SetWallet(cmd.WalletID)

	// Record in ledger (negative amount for withdrawal)
	if err := s.recordLedgerEntry(ctx, uuid.New(), cmd.UserAddress, cmd.TokenAddress,
		new(big.Int).Neg(amountBig), "WITHDRAWAL", nil); err != nil {
		fmt.Printf("Warning: failed to record withdrawal ledger entry: %v\n", err)
	}

	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return tx, nil
}

// RecordTransaction records a transaction to the on-chain ledger
func (s *BlockchainServiceImpl) RecordTransaction(ctx context.Context, cmd blockchain_in.RecordLedgerEntryCommand) error {
	return s.recordLedgerEntry(ctx, cmd.TransactionID, cmd.Account, cmd.Token, cmd.Amount, cmd.Category, cmd.MatchID)
}

func (s *BlockchainServiceImpl) recordLedgerEntry(ctx context.Context, txID uuid.UUID, account, token wallet_vo.EVMAddress, amount *big.Int, category string, matchID *uuid.UUID) error {
	_, err := s.ledgerContract.RecordEntry(ctx, txID, account, token, amount, category, matchID)
	return err
}

// GetLedgerBalance gets the on-chain ledger balance for an account
func (s *BlockchainServiceImpl) GetLedgerBalance(ctx context.Context, account wallet_vo.EVMAddress, token wallet_vo.EVMAddress) (*big.Int, error) {
	return s.ledgerContract.GetAccountBalance(ctx, account, token)
}

// SyncPrizePool syncs a prize pool from chain to cache
func (s *BlockchainServiceImpl) SyncPrizePool(ctx context.Context, matchID uuid.UUID) error {
	// Get on-chain data
	onChainPool, err := s.vaultContract.GetPrizePoolInfo(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get on-chain pool: %w", err)
	}

	// Get cached pool
	cachedPool, err := s.prizePoolRepo.FindByMatchID(ctx, matchID)
	if err != nil {
		// Create new if not exists
		resourceOwner := common.GetResourceOwner(ctx)
		cachedPool = blockchain_entities.NewOnChainPrizePool(
			resourceOwner,
			matchID,
			onChainPool.ChainID,
			onChainPool.ContractAddr,
			onChainPool.TokenAddress,
			onChainPool.Currency,
			onChainPool.EntryFeePerPlayer,
			onChainPool.PlatformFeePercent,
		)
	}

	// Update from on-chain data
	cachedPool.Status = onChainPool.Status
	cachedPool.TotalAmount = onChainPool.TotalAmount
	cachedPool.Participants = onChainPool.Participants
	cachedPool.EscrowEndTime = onChainPool.EscrowEndTime

	// Get current block for sync state
	client, err := s.chainManager.GetClient(cachedPool.ChainID)
	if err != nil {
		return err
	}
	blockNum, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return err
	}

	cachedPool.UpdateSyncState(blockNum, true)

	return s.prizePoolRepo.Save(ctx, cachedPool)
}

// SyncAllPendingPools syncs all pending prize pools
func (s *BlockchainServiceImpl) SyncAllPendingPools(ctx context.Context) error {
	pools, err := s.prizePoolRepo.FindPendingDistribution(ctx)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		if err := s.SyncPrizePool(ctx, pool.MatchID); err != nil {
			fmt.Printf("Failed to sync pool %s: %v\n", pool.MatchID, err)
		}
	}

	return nil
}

// GetPrizePool gets a prize pool from cache
func (s *BlockchainServiceImpl) GetPrizePool(ctx context.Context, matchID uuid.UUID) (*blockchain_entities.OnChainPrizePool, error) {
	return s.prizePoolRepo.FindByMatchID(ctx, matchID)
}

// GetTransaction gets a transaction from cache
func (s *BlockchainServiceImpl) GetTransaction(ctx context.Context, txID uuid.UUID) (*blockchain_entities.BlockchainTransaction, error) {
	return s.txRepo.FindByID(ctx, txID)
}

// GetTransactionsByWallet gets transactions for a wallet
func (s *BlockchainServiceImpl) GetTransactionsByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*blockchain_entities.BlockchainTransaction, int64, error) {
	return s.txRepo.FindByWallet(ctx, walletID, limit, offset)
}
