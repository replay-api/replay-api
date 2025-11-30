package wallet_services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_vo "github.com/replay-api/replay-api/pkg/domain/wallet/value-objects"
)

// ReconciliationService verifies that wallet balances match ledger calculations
// Critical for financial integrity - detects bugs, fraud, or data corruption
type ReconciliationService struct {
	walletRepo wallet_out.WalletRepository
	ledgerRepo wallet_out.LedgerRepository
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService(
	walletRepo wallet_out.WalletRepository,
	ledgerRepo wallet_out.LedgerRepository,
) *ReconciliationService {
	return &ReconciliationService{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
	}
}

// ReconciliationResult contains the result of a wallet balance reconciliation
type ReconciliationResult struct {
	WalletID         uuid.UUID                                   `json:"wallet_id"`
	UserID           uuid.UUID                                   `json:"user_id"`
	ReconciliationID uuid.UUID                                   `json:"reconciliation_id"`
	Timestamp        time.Time                                   `json:"timestamp"`
	Status           ReconciliationStatus                        `json:"status"`
	Discrepancies    []BalanceDiscrepancy                        `json:"discrepancies,omitempty"`
	CurrencyResults  map[wallet_vo.Currency]CurrencyReconciliation `json:"currency_results"`
	TotalChecked     int                                         `json:"total_currencies_checked"`
	TotalMatched     int                                         `json:"total_matched"`
	TotalMismatched  int                                         `json:"total_mismatched"`
	Notes            string                                      `json:"notes,omitempty"`
}

// ReconciliationStatus indicates the outcome of reconciliation
type ReconciliationStatus string

const (
	ReconciliationStatusMatched       ReconciliationStatus = "Matched"       // All balances match
	ReconciliationStatusMismatched    ReconciliationStatus = "Mismatched"    // Discrepancies found
	ReconciliationStatusPartialMatch  ReconciliationStatus = "PartialMatch"  // Some match, some don't
	ReconciliationStatusError         ReconciliationStatus = "Error"         // Reconciliation failed
	ReconciliationStatusManualReview  ReconciliationStatus = "ManualReview"  // Requires human review
)

// CurrencyReconciliation contains reconciliation details for a single currency
type CurrencyReconciliation struct {
	Currency        wallet_vo.Currency `json:"currency"`
	WalletBalance   wallet_vo.Amount   `json:"wallet_balance"`
	LedgerBalance   wallet_vo.Amount   `json:"ledger_balance"`
	Difference      wallet_vo.Amount   `json:"difference"`
	Matched         bool               `json:"matched"`
	LastLedgerEntry *time.Time         `json:"last_ledger_entry,omitempty"`
}

// BalanceDiscrepancy describes a balance mismatch between wallet and ledger
type BalanceDiscrepancy struct {
	Currency       wallet_vo.Currency `json:"currency"`
	WalletBalance  wallet_vo.Amount   `json:"wallet_balance"`
	LedgerBalance  wallet_vo.Amount   `json:"ledger_balance"`
	Difference     wallet_vo.Amount   `json:"difference"`
	DifferenceUSD  wallet_vo.Amount   `json:"difference_usd"`  // For prioritization
	Severity       DiscrepancySeverity `json:"severity"`
	DetectedAt     time.Time          `json:"detected_at"`
	RequiresAction bool               `json:"requires_action"`
}

// DiscrepancySeverity indicates how serious a discrepancy is
type DiscrepancySeverity string

const (
	SeverityLow      DiscrepancySeverity = "Low"      // < $1 difference
	SeverityMedium   DiscrepancySeverity = "Medium"   // $1 - $100 difference
	SeverityHigh     DiscrepancySeverity = "High"     // $100 - $1000 difference
	SeverityCritical DiscrepancySeverity = "Critical" // > $1000 difference
)

// ReconcileWallet verifies that a wallet's balances match the ledger
// This is the source of truth - ledger balance is always correct
func (s *ReconciliationService) ReconcileWallet(
	ctx context.Context,
	walletID uuid.UUID,
) (*ReconciliationResult, error) {
	// Fetch wallet
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	result := &ReconciliationResult{
		WalletID:         walletID,
		UserID:           wallet.ResourceOwner.UserID,
		ReconciliationID: uuid.New(),
		Timestamp:        time.Now().UTC(),
		CurrencyResults:  make(map[wallet_vo.Currency]CurrencyReconciliation),
		Discrepancies:    []BalanceDiscrepancy{},
	}

	// Check each currency in the wallet
	for currency, walletBalance := range wallet.Balances {
		currencyResult := s.reconcileCurrency(ctx, walletID, currency, walletBalance)
		result.CurrencyResults[currency] = currencyResult
		result.TotalChecked++

		if currencyResult.Matched {
			result.TotalMatched++
		} else {
			result.TotalMismatched++

			// Create discrepancy record
			discrepancy := BalanceDiscrepancy{
				Currency:       currency,
				WalletBalance:  walletBalance,
				LedgerBalance:  currencyResult.LedgerBalance,
				Difference:     currencyResult.Difference,
				DifferenceUSD:  currencyResult.Difference, // TODO: Convert to USD for prioritization
				Severity:       s.calculateSeverity(currencyResult.Difference),
				DetectedAt:     time.Now().UTC(),
				RequiresAction: !currencyResult.Difference.IsZero(),
			}
			result.Discrepancies = append(result.Discrepancies, discrepancy)
		}
	}

	// Determine overall status
	result.Status = s.determineStatus(result)

	// Log results
	s.logReconciliationResult(ctx, result)

	return result, nil
}

// reconcileCurrency verifies balance for a single currency
func (s *ReconciliationService) reconcileCurrency(
	ctx context.Context,
	walletID uuid.UUID,
	currency wallet_vo.Currency,
	walletBalance wallet_vo.Amount,
) CurrencyReconciliation {
	// Calculate balance from ledger (source of truth)
	ledgerBalance, err := s.ledgerRepo.CalculateBalance(ctx, walletID, currency)
	if err != nil {
		slog.ErrorContext(ctx, "failed to calculate ledger balance",
			"wallet_id", walletID,
			"currency", currency,
			"error", err)

		// Return error state
		return CurrencyReconciliation{
			Currency:      currency,
			WalletBalance: walletBalance,
			LedgerBalance: wallet_vo.NewAmount(0),
			Difference:    walletBalance,
			Matched:       false,
		}
	}

	// Calculate difference
	difference := walletBalance.Subtract(ledgerBalance)

	return CurrencyReconciliation{
		Currency:      currency,
		WalletBalance: walletBalance,
		LedgerBalance: ledgerBalance,
		Difference:    difference,
		Matched:       difference.IsZero(),
	}
}

// calculateSeverity determines how serious a discrepancy is based on amount
func (s *ReconciliationService) calculateSeverity(difference wallet_vo.Amount) DiscrepancySeverity {
	absDiff := difference
	if difference.IsNegative() {
		absDiff = wallet_vo.NewAmount(0).Subtract(difference)
	}

	// Convert cents to dollars for threshold comparison
	cents := absDiff.ToCents()

	if cents < 100 { // < $1
		return SeverityLow
	} else if cents < 10000 { // < $100
		return SeverityMedium
	} else if cents < 100000 { // < $1000
		return SeverityHigh
	}

	return SeverityCritical
}

// determineStatus determines overall reconciliation status
func (s *ReconciliationService) determineStatus(result *ReconciliationResult) ReconciliationStatus {
	if result.TotalMismatched == 0 {
		return ReconciliationStatusMatched
	}

	if result.TotalMatched == 0 {
		return ReconciliationStatusMismatched
	}

	// Check severity - critical discrepancies require manual review
	for _, disc := range result.Discrepancies {
		if disc.Severity == SeverityCritical || disc.Severity == SeverityHigh {
			return ReconciliationStatusManualReview
		}
	}

	return ReconciliationStatusPartialMatch
}

// AutoCorrectWallet automatically corrects wallet balance to match ledger
// DANGEROUS: Only use when you're certain ledger is correct
// Should require manual approval for large discrepancies
func (s *ReconciliationService) AutoCorrectWallet(
	ctx context.Context,
	walletID uuid.UUID,
	approverID uuid.UUID,
) error {
	result, err := s.ReconcileWallet(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to reconcile wallet: %w", err)
	}

	if result.Status == ReconciliationStatusMatched {
		slog.InfoContext(ctx, "wallet already reconciled, no correction needed",
			"wallet_id", walletID)
		return nil
	}

	// Fetch wallet
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to find wallet: %w", err)
	}

	// Check if critical discrepancies require manual approval
	hasCritical := false
	for _, disc := range result.Discrepancies {
		if disc.Severity == SeverityCritical {
			hasCritical = true
			break
		}
	}

	if hasCritical {
		return fmt.Errorf("critical discrepancies detected, manual approval required")
	}

	// Correct each mismatched currency
	for currency, currencyResult := range result.CurrencyResults {
		if !currencyResult.Matched {
			// Update wallet balance to match ledger
			wallet.Balances[currency] = currencyResult.LedgerBalance

			slog.WarnContext(ctx, "auto-correcting wallet balance",
				"wallet_id", walletID,
				"currency", currency,
				"old_balance", currencyResult.WalletBalance.String(),
				"new_balance", currencyResult.LedgerBalance.String(),
				"difference", currencyResult.Difference.String(),
				"approver_id", approverID)
		}
	}

	// Save corrected wallet
	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	slog.InfoContext(ctx, "wallet auto-corrected successfully",
		"wallet_id", walletID,
		"currencies_corrected", result.TotalMismatched,
		"approver_id", approverID)

	return nil
}

// ReconcileAllWallets runs reconciliation for all active wallets
// Should be run daily as a cron job
func (s *ReconciliationService) ReconcileAllWallets(
	ctx context.Context,
) (*BatchReconciliationReport, error) {
	report := &BatchReconciliationReport{
		ReportID:  uuid.New(),
		StartedAt: time.Now().UTC(),
		Results:   []ReconciliationResult{},
	}

	// TODO: Implement batch reconciliation
	// This requires a method to list all wallets, which isn't in the current interface
	// For now, this is a placeholder for future implementation

	report.CompletedAt = time.Now().UTC()
	report.Status = "Completed"

	return report, nil
}

// VerifyLedgerIntegrity checks that all ledger entries balance correctly
// For double-entry accounting: SUM(debits) must equal SUM(credits)
func (s *ReconciliationService) VerifyLedgerIntegrity(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) (*LedgerIntegrityReport, error) {
	// TODO: Implement ledger integrity verification
	// This requires querying all entries in date range and verifying:
	// 1. Every transaction has matching debit/credit entries
	// 2. Total debits = total credits across all entries
	// 3. No orphaned entries

	report := &LedgerIntegrityReport{
		ReportID:  uuid.New(),
		FromDate:  fromDate,
		ToDate:    toDate,
		Timestamp: time.Now().UTC(),
	}

	return report, nil
}

// logReconciliationResult logs the reconciliation outcome
func (s *ReconciliationService) logReconciliationResult(ctx context.Context, result *ReconciliationResult) {
	if result.Status == ReconciliationStatusMatched {
		slog.InfoContext(ctx, "wallet reconciliation: all balances match",
			"wallet_id", result.WalletID,
			"currencies_checked", result.TotalChecked)
	} else {
		slog.WarnContext(ctx, "wallet reconciliation: discrepancies found",
			"wallet_id", result.WalletID,
			"status", result.Status,
			"total_mismatched", result.TotalMismatched,
			"discrepancies", result.Discrepancies)
	}
}

// BatchReconciliationReport contains results from reconciling multiple wallets
type BatchReconciliationReport struct {
	ReportID         uuid.UUID              `json:"report_id"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      time.Time              `json:"completed_at"`
	Status           string                 `json:"status"`
	Results          []ReconciliationResult `json:"results"`
	TotalWallets     int                    `json:"total_wallets"`
	TotalMatched     int                    `json:"total_matched"`
	TotalMismatched  int                    `json:"total_mismatched"`
	TotalErrors      int                    `json:"total_errors"`
	CriticalIssues   int                    `json:"critical_issues"`
}

// LedgerIntegrityReport contains results from ledger integrity verification
type LedgerIntegrityReport struct {
	ReportID             uuid.UUID `json:"report_id"`
	FromDate             time.Time `json:"from_date"`
	ToDate               time.Time `json:"to_date"`
	Timestamp            time.Time `json:"timestamp"`
	TotalTransactions    int64     `json:"total_transactions"`
	TotalDebits          int64     `json:"total_debits"`
	TotalCredits         int64     `json:"total_credits"`
	DebitsMatchCredits   bool      `json:"debits_match_credits"`
	UnbalancedTxCount    int       `json:"unbalanced_transactions"`
	OrphanedEntries      int       `json:"orphaned_entries"`
	IntegrityStatus      string    `json:"integrity_status"`
}
