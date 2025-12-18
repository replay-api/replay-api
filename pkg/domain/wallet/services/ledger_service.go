package wallet_services

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// LedgerRepository defines persistence for ledger operations
type LedgerRepository interface {
	// Accounts
	CreateAccount(ctx context.Context, account *wallet_entities.LedgerAccount) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*wallet_entities.LedgerAccount, error)
	GetAccountByCode(ctx context.Context, code string) (*wallet_entities.LedgerAccount, error)
	GetAccountByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerAccount, error)
	UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance, available, held *big.Float, version int) error
	
	// Journals
	CreateJournal(ctx context.Context, journal *wallet_entities.JournalEntry) error
	GetJournalByID(ctx context.Context, id uuid.UUID) (*wallet_entities.JournalEntry, error)
	GetLastJournalHash(ctx context.Context) (string, error)
	UpdateJournalStatus(ctx context.Context, id uuid.UUID, status wallet_entities.JournalStatus) error
	
	// Wallets (using LedgerWallet for ledger integration)
	CreateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error
	GetWalletByUserID(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerWallet, error)
	UpdateWallet(ctx context.Context, wallet *wallet_entities.LedgerWallet) error
	
	// Reporting
	GetJournalsByDateRange(ctx context.Context, from, to time.Time) ([]wallet_entities.JournalEntry, error)
	GetAccountBalances(ctx context.Context) ([]wallet_entities.LedgerAccount, error)
}

// LedgerService provides double-entry accounting operations
type LedgerService struct {
	mu         sync.Mutex
	repo       LedgerRepository
	auditTrail billing_in.AuditTrailCommand
	
	// System accounts (cached)
	systemAccounts map[string]*wallet_entities.LedgerAccount
}

// NewLedgerService creates a new ledger service
func NewLedgerService(repo LedgerRepository, auditTrail billing_in.AuditTrailCommand) *LedgerService {
	return &LedgerService{
		repo:           repo,
		auditTrail:     auditTrail,
		systemAccounts: make(map[string]*wallet_entities.LedgerAccount),
	}
}

// InitializeSystemAccounts creates the standard chart of accounts
func (s *LedgerService) InitializeSystemAccounts(ctx context.Context) error {
	for _, acct := range wallet_entities.StandardChartOfAccounts {
		existing, err := s.repo.GetAccountByCode(ctx, acct.Code)
		if err == nil && existing != nil {
			s.systemAccounts[acct.Code] = existing
			continue
		}

		now := time.Now().UTC()
		account := &wallet_entities.LedgerAccount{
			ID:               uuid.New(),
			Code:             acct.Code,
			Name:             acct.Name,
			Type:             acct.AccountType,
			Currency:         "USD",
			Balance:          big.NewFloat(0),
			AvailableBalance: big.NewFloat(0),
			HeldBalance:      big.NewFloat(0),
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
			Version:          1,
		}

		if err := s.repo.CreateAccount(ctx, account); err != nil {
			return fmt.Errorf("failed to create system account %s: %w", acct.Code, err)
		}

		s.systemAccounts[acct.Code] = account
	}

	slog.Info("System accounts initialized", "count", len(s.systemAccounts))
	return nil
}

// GetOrCreateUserWallet gets or creates a user wallet with ledger account
func (s *LedgerService) GetOrCreateUserWallet(ctx context.Context, userID uuid.UUID, currency string) (*wallet_entities.LedgerWallet, error) {
	// Check if wallet exists
	wallet, err := s.repo.GetWalletByUserID(ctx, userID, currency)
	if err == nil && wallet != nil {
		return wallet, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after lock
	wallet, err = s.repo.GetWalletByUserID(ctx, userID, currency)
	if err == nil && wallet != nil {
		return wallet, nil
	}

	// Create ledger account for user
	now := time.Now().UTC()
	ledgerAccount := &wallet_entities.LedgerAccount{
		ID:               uuid.New(),
		Code:             fmt.Sprintf("2001-%s", userID.String()[:8]),
		Name:             fmt.Sprintf("User Wallet - %s", userID.String()[:8]),
		Type:             wallet_entities.AccountTypeLiability,
		Currency:         currency,
		Balance:          big.NewFloat(0),
		AvailableBalance: big.NewFloat(0),
		HeldBalance:      big.NewFloat(0),
		UserID:           &userID,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
		Version:          1,
	}

	if err := s.repo.CreateAccount(ctx, ledgerAccount); err != nil {
		return nil, fmt.Errorf("failed to create user ledger account: %w", err)
	}

	// Create wallet
	wallet = wallet_entities.NewLedgerWallet(userID, ledgerAccount.ID, currency)
	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to create user wallet: %w", err)
	}

	slog.InfoContext(ctx, "User wallet created",
		"user_id", userID,
		"wallet_id", wallet.ID,
		"currency", currency,
	)

	return wallet, nil
}

// Deposit processes a deposit transaction
func (s *LedgerService) Deposit(ctx context.Context, req DepositRequest) (*wallet_entities.JournalEntry, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	// Get user wallet (creates if not exists)
	wallet, err := s.GetOrCreateUserWallet(ctx, req.UserID, req.Currency)
	if err != nil {
		return nil, err
	}

	// Get accounts
	cashAccount := s.systemAccounts["1001"] // Operating Cash
	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user account: %w", err)
	}

	amount := big.NewFloat(req.Amount)

	// Build journal entry
	// Debit: Cash (Asset increases)
	// Credit: User Wallet (Liability increases - we owe user money)
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypeDeposit,
		fmt.Sprintf("DEP-%s-%d", req.UserID.String()[:8], time.Now().Unix()),
		fmt.Sprintf("Deposit via %s", req.PaymentMethod),
		req.Currency,
		req.UserID,
		resourceOwner,
	)

	if err := journal.AddDebit(cashAccount.ID, cashAccount.Code, amount, "Cash received from deposit"); err != nil {
		return nil, fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(userAccount.ID, userAccount.Code, amount, "Credit to user wallet"); err != nil {
		return nil, fmt.Errorf("failed to add credit entry: %w", err)
	}

	if req.ExternalRef != "" {
		journal.ExternalRef = req.ExternalRef
		journal.Metadata["payment_provider"] = req.PaymentProvider
		journal.Metadata["payment_method"] = req.PaymentMethod
	}

	// Validate
	if err := journal.Validate(); err != nil {
		return nil, fmt.Errorf("journal validation failed: %w", err)
	}

	// Get last hash and compute
	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update account balances
	userAccount.Balance = new(big.Float).Add(userAccount.Balance, amount)
	userAccount.AvailableBalance = new(big.Float).Add(userAccount.AvailableBalance, amount)
	userAccount.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateAccountBalance(ctx, userAccount.ID, userAccount.Balance, userAccount.AvailableBalance, userAccount.HeldBalance, userAccount.Version); err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	// Update wallet
	wallet.Balance = new(big.Float).Add(wallet.Balance, amount)
	wallet.AvailableBalance = new(big.Float).Add(wallet.AvailableBalance, amount)
	wallet.TotalDeposits = new(big.Float).Add(wallet.TotalDeposits, amount)
	wallet.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	// Mark journal as posted
	if err := journal.MarkApproved(req.UserID); err != nil {
		return nil, fmt.Errorf("failed to approve journal: %w", err)
	}
	if err := journal.MarkPosted(); err != nil {
		return nil, fmt.Errorf("failed to post journal: %w", err)
	}

	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	// Audit trail
	if s.auditTrail != nil {
		if err := s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
			EventType:     billing_entities.AuditEventDeposit,
			UserID:        req.UserID,
			TargetType:    "wallet",
			TargetID:      wallet.ID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			BalanceBefore: floatFromBig(new(big.Float).Sub(wallet.Balance, amount)),
			BalanceAfter:  floatFromBig(wallet.Balance),
			TransactionID: journal.ID,
			ExternalRef:   req.ExternalRef,
			Description:   journal.Description,
		}); err != nil {
			slog.WarnContext(ctx, "Failed to record audit trail for deposit", "error", err)
		}
	}

	slog.InfoContext(ctx, "Deposit processed",
		"journal_id", journal.ID,
		"user_id", req.UserID,
		"amount", req.Amount,
		"currency", req.Currency,
	)

	return journal, nil
}

// Withdraw processes a withdrawal transaction
func (s *LedgerService) Withdraw(ctx context.Context, req WithdrawRequest) (*wallet_entities.JournalEntry, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	wallet, err := s.repo.GetWalletByUserID(ctx, req.UserID, req.Currency)
	if err != nil || wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	amount := big.NewFloat(req.Amount)

	// Check sufficient balance
	if !wallet.CanWithdraw(amount) {
		return nil, fmt.Errorf("insufficient available balance: %.2f available, %.2f requested",
			floatFromBig(wallet.AvailableBalance), req.Amount)
	}

	// Get accounts
	cashAccount := s.systemAccounts["1001"]
	_ = s.systemAccounts["2003"] // pendingWithdrawals - reserved for future two-phase withdrawal
	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user account: %w", err)
	}

	// Calculate fee
	fee := big.NewFloat(req.Fee)
	netAmount := new(big.Float).Sub(amount, fee)

	// Build journal entry
	// Debit: User Wallet (Liability decreases)
	// Credit: Cash (Asset decreases)
	// Credit: Platform Fee Revenue (if fee > 0)
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypeWithdrawal,
		fmt.Sprintf("WIT-%s-%d", req.UserID.String()[:8], time.Now().Unix()),
		fmt.Sprintf("Withdrawal to %s", req.RecipientAddress[:10]),
		req.Currency,
		req.UserID,
		resourceOwner,
	)

	if err := journal.AddDebit(userAccount.ID, userAccount.Code, amount, "Withdrawal from user wallet"); err != nil {
		return nil, fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(cashAccount.ID, cashAccount.Code, netAmount, "Cash disbursement"); err != nil {
		return nil, fmt.Errorf("failed to add credit entry: %w", err)
	}

	if fee.Cmp(big.NewFloat(0)) > 0 {
		feeAccount := s.systemAccounts["4001"]
		if err := journal.AddCredit(feeAccount.ID, feeAccount.Code, fee, "Withdrawal fee"); err != nil {
			return nil, fmt.Errorf("failed to add fee credit entry: %w", err)
		}
	}

	journal.Metadata["recipient_address"] = req.RecipientAddress
	journal.Metadata["withdrawal_method"] = req.Method
	journal.Metadata["fee"] = floatFromBig(fee)
	journal.Metadata["net_amount"] = floatFromBig(netAmount)

	// Validate
	if err := journal.Validate(); err != nil {
		return nil, fmt.Errorf("journal validation failed: %w", err)
	}

	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update balances
	userAccount.Balance = new(big.Float).Sub(userAccount.Balance, amount)
	userAccount.AvailableBalance = new(big.Float).Sub(userAccount.AvailableBalance, amount)
	userAccount.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateAccountBalance(ctx, userAccount.ID, userAccount.Balance, userAccount.AvailableBalance, userAccount.HeldBalance, userAccount.Version); err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}

	wallet.Balance = new(big.Float).Sub(wallet.Balance, amount)
	wallet.AvailableBalance = new(big.Float).Sub(wallet.AvailableBalance, amount)
	wallet.TotalWithdrawals = new(big.Float).Add(wallet.TotalWithdrawals, netAmount)
	wallet.TotalFees = new(big.Float).Add(wallet.TotalFees, fee)
	wallet.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	if err := journal.MarkApproved(req.UserID); err != nil {
		return nil, fmt.Errorf("failed to approve journal: %w", err)
	}
	if err := journal.MarkPosted(); err != nil {
		return nil, fmt.Errorf("failed to post journal: %w", err)
	}

	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return nil, fmt.Errorf("failed to create journal: %w", err)
	}

	// Audit trail
	if s.auditTrail != nil {
		if err := s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
			EventType:     billing_entities.AuditEventWithdrawal,
			UserID:        req.UserID,
			TargetType:    "wallet",
			TargetID:      wallet.ID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			BalanceBefore: floatFromBig(new(big.Float).Add(wallet.Balance, amount)),
			BalanceAfter:  floatFromBig(wallet.Balance),
			TransactionID: journal.ID,
			Description:   journal.Description,
			Metadata: map[string]interface{}{
				"fee":               floatFromBig(fee),
				"net_amount":        floatFromBig(netAmount),
				"recipient_address": req.RecipientAddress,
			},
		}); err != nil {
			slog.WarnContext(ctx, "Failed to record audit trail for withdrawal", "error", err)
		}
	}

	slog.InfoContext(ctx, "Withdrawal processed",
		"journal_id", journal.ID,
		"user_id", req.UserID,
		"amount", req.Amount,
		"fee", floatFromBig(fee),
		"net", floatFromBig(netAmount),
	)

	return journal, nil
}

// HoldFunds places a hold on user funds
func (s *LedgerService) HoldFunds(ctx context.Context, userID uuid.UUID, amount float64, currency string, reference uuid.UUID, reason string) error {
	resourceOwner := common.GetResourceOwner(ctx)

	wallet, err := s.repo.GetWalletByUserID(ctx, userID, currency)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	holdAmount := big.NewFloat(amount)

	if wallet.AvailableBalance.Cmp(holdAmount) < 0 {
		return fmt.Errorf("insufficient available balance for hold")
	}

	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return err
	}
	holdAccount := s.systemAccounts["2005"]

	// Build journal for hold
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypeHold,
		fmt.Sprintf("HOLD-%s", reference.String()[:8]),
		reason,
		currency,
		userID,
		resourceOwner,
	)

	if err := journal.AddDebit(userAccount.ID, userAccount.Code, holdAmount, "Hold placed on user funds"); err != nil {
		return fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(holdAccount.ID, holdAccount.Code, holdAmount, "Held user funds"); err != nil {
		return fmt.Errorf("failed to add credit entry: %w", err)
	}
	journal.Metadata["reference_id"] = reference.String()

	if err := journal.Validate(); err != nil {
		return err
	}

	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update balances - move from available to held
	userAccount.AvailableBalance = new(big.Float).Sub(userAccount.AvailableBalance, holdAmount)
	userAccount.HeldBalance = new(big.Float).Add(userAccount.HeldBalance, holdAmount)
	
	if err := s.repo.UpdateAccountBalance(ctx, userAccount.ID, userAccount.Balance, userAccount.AvailableBalance, userAccount.HeldBalance, userAccount.Version); err != nil {
		return err
	}

	wallet.AvailableBalance = new(big.Float).Sub(wallet.AvailableBalance, holdAmount)
	wallet.HeldBalance = new(big.Float).Add(wallet.HeldBalance, holdAmount)
	
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return err
	}

	if err := journal.MarkApproved(userID); err != nil {
		return fmt.Errorf("failed to approve journal: %w", err)
	}
	if err := journal.MarkPosted(); err != nil {
		return fmt.Errorf("failed to post journal: %w", err)
	}
	
	return s.repo.CreateJournal(ctx, journal)
}

// ReleaseFunds releases a hold on user funds
func (s *LedgerService) ReleaseFunds(ctx context.Context, userID uuid.UUID, amount float64, currency string, reference uuid.UUID, reason string) error {
	resourceOwner := common.GetResourceOwner(ctx)

	wallet, err := s.repo.GetWalletByUserID(ctx, userID, currency)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	releaseAmount := big.NewFloat(amount)

	if wallet.HeldBalance.Cmp(releaseAmount) < 0 {
		return fmt.Errorf("insufficient held balance for release")
	}

	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return err
	}
	holdAccount := s.systemAccounts["2005"]

	// Build reversal journal
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypeRelease,
		fmt.Sprintf("REL-%s", reference.String()[:8]),
		reason,
		currency,
		userID,
		resourceOwner,
	)

	if err := journal.AddDebit(holdAccount.ID, holdAccount.Code, releaseAmount, "Release held funds"); err != nil {
		return fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(userAccount.ID, userAccount.Code, releaseAmount, "Funds released to user"); err != nil {
		return fmt.Errorf("failed to add credit entry: %w", err)
	}
	journal.Metadata["reference_id"] = reference.String()

	if err := journal.Validate(); err != nil {
		return err
	}

	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	userAccount.AvailableBalance = new(big.Float).Add(userAccount.AvailableBalance, releaseAmount)
	userAccount.HeldBalance = new(big.Float).Sub(userAccount.HeldBalance, releaseAmount)
	
	if err := s.repo.UpdateAccountBalance(ctx, userAccount.ID, userAccount.Balance, userAccount.AvailableBalance, userAccount.HeldBalance, userAccount.Version); err != nil {
		return err
	}

	wallet.AvailableBalance = new(big.Float).Add(wallet.AvailableBalance, releaseAmount)
	wallet.HeldBalance = new(big.Float).Sub(wallet.HeldBalance, releaseAmount)
	
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return err
	}

	if err := journal.MarkApproved(userID); err != nil {
		return fmt.Errorf("failed to approve journal: %w", err)
	}
	if err := journal.MarkPosted(); err != nil {
		return fmt.Errorf("failed to post journal: %w", err)
	}
	
	return s.repo.CreateJournal(ctx, journal)
}

// GenerateTrialBalance generates a trial balance report
func (s *LedgerService) GenerateTrialBalance(ctx context.Context) (*wallet_entities.TrialBalance, error) {
	accounts, err := s.repo.GetAccountBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}

	tb := &wallet_entities.TrialBalance{
		AsOfDate:     time.Now().UTC(),
		Accounts:     make([]wallet_entities.TrialBalanceAccount, 0),
		TotalDebits:  big.NewFloat(0),
		TotalCredits: big.NewFloat(0),
		Currency:     "USD",
		GeneratedAt:  time.Now().UTC(),
	}

	for _, acct := range accounts {
		var debitBalance, creditBalance *big.Float

		switch acct.Type {
		case wallet_entities.AccountTypeAsset, wallet_entities.AccountTypeExpense:
			debitBalance = acct.Balance
			creditBalance = big.NewFloat(0)
			tb.TotalDebits = new(big.Float).Add(tb.TotalDebits, acct.Balance)
		case wallet_entities.AccountTypeLiability, wallet_entities.AccountTypeEquity, wallet_entities.AccountTypeRevenue:
			debitBalance = big.NewFloat(0)
			creditBalance = acct.Balance
			tb.TotalCredits = new(big.Float).Add(tb.TotalCredits, acct.Balance)
		}

		tb.Accounts = append(tb.Accounts, wallet_entities.TrialBalanceAccount{
			AccountCode:   acct.Code,
			AccountName:   acct.Name,
			AccountType:   acct.Type,
			DebitBalance:  debitBalance,
			CreditBalance: creditBalance,
		})
	}

	// Check if balanced
	diff := new(big.Float).Sub(tb.TotalDebits, tb.TotalCredits)
	tolerance := big.NewFloat(0.01)
	tb.IsBalanced = diff.Abs(diff).Cmp(tolerance) <= 0

	return tb, nil
}

// Request types

// DepositRequest contains deposit parameters
type DepositRequest struct {
	UserID          uuid.UUID
	Amount          float64
	Currency        string
	PaymentMethod   string
	PaymentProvider string
	ExternalRef     string
}

// WithdrawRequest contains withdrawal parameters
type WithdrawRequest struct {
	UserID           uuid.UUID
	Amount           float64
	Fee              float64
	Currency         string
	Method           string
	RecipientAddress string
	ExternalRef      string
}

func floatFromBig(f *big.Float) float64 {
	if f == nil {
		return 0
	}
	r, _ := f.Float64()
	return r
}

// RecordDeposit records a deposit transaction in the ledger
// Used by TransactionCoordinator for saga-based deposit processing
func (s *LedgerService) RecordDeposit(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	paymentID uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	// Convert to deposit request
	req := DepositRequest{
		UserID:        resourceOwner.UserID,
		Amount:        amount.ToFloat64(),
		Currency:      string(currency),
		PaymentMethod: metadata.OperationType,
		ExternalRef:   paymentID.String(),
	}
	
	journal, err := s.Deposit(ctx, req)
	if err != nil {
		return uuid.Nil, err
	}
	
	return journal.ID, nil
}

// RecordWithdrawal records a withdrawal transaction in the ledger
func (s *LedgerService) RecordWithdrawal(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	recipientAddress string,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	req := WithdrawRequest{
		UserID:           resourceOwner.UserID,
		Amount:           amount.ToFloat64(),
		Fee:              0, // Fee calculated by Withdraw method or passed in metadata
		Currency:         string(currency),
		Method:           metadata.OperationType,
		RecipientAddress: recipientAddress,
	}
	
	journal, err := s.Withdraw(ctx, req)
	if err != nil {
		return uuid.Nil, err
	}
	
	return journal.ID, nil
}

// RecordRefund records a refund/reversal transaction in the ledger
func (s *LedgerService) RecordRefund(
	ctx context.Context,
	originalTxID uuid.UUID,
	reason string,
) (uuid.UUID, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	// Get the original journal entry
	originalJournal, err := s.repo.GetJournalByID(ctx, originalTxID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get original transaction: %w", err)
	}
	
	// Create reversal
	reversal, err := originalJournal.CreateReversal(reason, resourceOwner.UserID, resourceOwner)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create reversal: %w", err)
	}
	
	// Get last hash and compute
	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	reversal.ComputeHash(lastHash)
	
	// Save reversal
	if err := s.repo.CreateJournal(ctx, reversal); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save reversal: %w", err)
	}
	
	// Log audit trail
	if s.auditTrail != nil {
		if err := s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
			EventType:     billing_entities.AuditEventRefund,
			UserID:        resourceOwner.UserID,
			TargetType:    "journal_entry",
			TargetID:      reversal.ID,
			TransactionID: originalTxID,
			Description:   fmt.Sprintf("Refund recorded for %s: %s", originalTxID, reason),
		}); err != nil {
			slog.WarnContext(ctx, "Failed to record audit trail for refund", "error", err)
		}
	}
	
	return reversal.ID, nil
}

// RecordEntryFee records a match entry fee in the ledger
func (s *LedgerService) RecordEntryFee(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	// Get wallet
	wallet, err := s.repo.GetWalletByUserID(ctx, resourceOwner.UserID, string(currency))
	if err != nil {
		return uuid.Nil, fmt.Errorf("wallet not found: %w", err)
	}
	
	amountBig := big.NewFloat(amount.ToFloat64())
	if !wallet.CanWithdraw(amountBig) {
		return uuid.Nil, fmt.Errorf("insufficient balance for entry fee")
	}
	
	// Get accounts
	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get user account: %w", err)
	}
	
	prizePoolAccount := s.systemAccounts["2002"] // Prize Pool Escrow
	
	// Build reference
	refID := uuid.New().String()[:8]
	if matchID != nil {
		refID = matchID.String()[:8]
	}
	
	// Build journal entry
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypeEntryFee,
		fmt.Sprintf("ENTRY-%s-%d", refID, time.Now().Unix()),
		"Match entry fee",
		string(currency),
		resourceOwner.UserID,
		resourceOwner,
	)
	
	if err := journal.AddDebit(userAccount.ID, userAccount.Code, amountBig, "Entry fee deducted"); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(prizePoolAccount.ID, prizePoolAccount.Code, amountBig, "Added to prize pool"); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add credit entry: %w", err)
	}
	
	if matchID != nil {
		journal.Metadata["match_id"] = matchID.String()
	}
	if tournamentID != nil {
		journal.Metadata["tournament_id"] = tournamentID.String()
	}
	
	if err := journal.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("journal validation failed: %w", err)
	}
	
	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)
	
	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save journal: %w", err)
	}
	
	// Update wallet balance
	wallet.AvailableBalance = new(big.Float).Sub(wallet.AvailableBalance, amountBig)
	wallet.Balance = new(big.Float).Sub(wallet.Balance, amountBig)
	wallet.TotalLosses = new(big.Float).Add(wallet.TotalLosses, amountBig)
	wallet.UpdatedAt = time.Now().UTC()
	wallet.Version++
	
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update wallet: %w", err)
	}
	
	return journal.ID, nil
}

// RecordPrizeWinning records prize money won by a player
func (s *LedgerService) RecordPrizeWinning(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	amount wallet_vo.Amount,
	matchID *uuid.UUID,
	tournamentID *uuid.UUID,
	metadata wallet_entities.LedgerMetadata,
) (uuid.UUID, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	// Get or create wallet
	wallet, err := s.GetOrCreateUserWallet(ctx, resourceOwner.UserID, string(currency))
	if err != nil {
		return uuid.Nil, err
	}
	
	amountBig := big.NewFloat(amount.ToFloat64())
	
	// Get accounts
	userAccount, err := s.repo.GetAccountByID(ctx, wallet.LedgerAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get user account: %w", err)
	}
	
	prizePoolAccount := s.systemAccounts["2002"] // Prize Pool Escrow
	
	// Build reference
	prizeRefID := uuid.New().String()[:8]
	if matchID != nil {
		prizeRefID = matchID.String()[:8]
	}
	
	// Build journal entry
	journal := wallet_entities.NewJournalEntry(
		wallet_entities.TxTypePrizeDistribution,
		fmt.Sprintf("PRIZE-%s-%d", prizeRefID, time.Now().Unix()),
		"Prize winnings",
		string(currency),
		resourceOwner.UserID,
		resourceOwner,
	)
	
	if err := journal.AddDebit(prizePoolAccount.ID, prizePoolAccount.Code, amountBig, "Prize pool distribution"); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add debit entry: %w", err)
	}
	if err := journal.AddCredit(userAccount.ID, userAccount.Code, amountBig, "Prize winnings credited"); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add credit entry: %w", err)
	}
	
	if matchID != nil {
		journal.Metadata["match_id"] = matchID.String()
	}
	if tournamentID != nil {
		journal.Metadata["tournament_id"] = tournamentID.String()
	}
	
	if err := journal.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("journal validation failed: %w", err)
	}
	
	lastHash, _ := s.repo.GetLastJournalHash(ctx)
	journal.ComputeHash(lastHash)
	
	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save journal: %w", err)
	}
	
	// Update wallet balance
	wallet.Balance = new(big.Float).Add(wallet.Balance, amountBig)
	wallet.AvailableBalance = new(big.Float).Add(wallet.AvailableBalance, amountBig)
	wallet.TotalWinnings = new(big.Float).Add(wallet.TotalWinnings, amountBig)
	wallet.UpdatedAt = time.Now().UTC()
	wallet.Version++
	
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return uuid.Nil, fmt.Errorf("failed to update wallet: %w", err)
	}
	
	// Log audit trail
	if s.auditTrail != nil {
		if err := s.auditTrail.RecordFinancialEvent(ctx, billing_in.RecordFinancialEventRequest{
			EventType:     billing_entities.AuditEventPrizeDistribution,
			UserID:        resourceOwner.UserID,
			TargetType:    "journal_entry",
			TargetID:      journal.ID,
			Amount:        amount.ToFloat64(),
			Currency:      string(currency),
			TransactionID: func() uuid.UUID { if matchID != nil { return *matchID }; return uuid.Nil }(),
			Description:   fmt.Sprintf("Prize of %.2f %s credited for match %s", amount.ToFloat64(), currency, prizeRefID),
		}); err != nil {
			slog.WarnContext(ctx, "Failed to record audit trail for prize distribution", "error", err)
		}
	}
	
	return journal.ID, nil
}
