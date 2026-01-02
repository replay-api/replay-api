package billing_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

// AuditTrailService implements bank-grade audit trail management
type AuditTrailService struct {
	repo        billing_out.AuditTrailRepository
	alertSender billing_out.AuditAlertSender
}

// NewAuditTrailService creates a new audit trail service
func NewAuditTrailService(
	repo billing_out.AuditTrailRepository,
	alertSender billing_out.AuditAlertSender,
) *AuditTrailService {
	return &AuditTrailService{
		repo:        repo,
		alertSender: alertSender,
	}
}

// RecordFinancialEvent creates an audit entry for a financial transaction
func (s *AuditTrailService) RecordFinancialEvent(ctx context.Context, req billing_in.RecordFinancialEventRequest) error {
	resourceOwner := common.GetResourceOwner(ctx)

	// Determine severity based on amount thresholds
	severity := s.determineFinancialSeverity(req.Amount, req.EventType)

	entry := billing_entities.NewAuditTrailEntry(
		req.EventType,
		severity,
		req.UserID,
		"user",
		req.TargetType,
		req.TargetID,
		req.Description,
		resourceOwner,
	)

	entry.SetFinancialDetails(req.Amount, req.Currency, req.BalanceBefore, req.BalanceAfter)
	entry.SetTransactionRef(req.TransactionID, req.ExternalRef, req.ProviderRef)
	entry.SetActorContext(req.IP, req.UserAgent, req.SessionID)
	entry.Metadata = req.Metadata

	// Get previous entry for hash chain
	previousEntry, err := s.repo.GetLatestForTarget(ctx, req.TargetType, req.TargetID)
	previousHash := ""
	if err == nil && previousEntry != nil {
		previousHash = previousEntry.Hash
		entry.PreviousEntryID = &previousEntry.ID
	}

	// Compute hash for chain integrity
	entry.ComputeHash(previousHash)

	// Persist the entry
	if err := s.repo.Create(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "Failed to create audit entry",
			"error", err,
			"event_type", req.EventType,
			"user_id", req.UserID,
		)
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	slog.InfoContext(ctx, "Financial audit event recorded",
		"audit_id", entry.ID,
		"event_type", req.EventType,
		"amount", req.Amount,
		"currency", req.Currency,
	)

	// Send alert for critical events
	if entry.IsCritical() && s.alertSender != nil {
		go func() {
			if err := s.alertSender.SendCriticalAlert(context.Background(), entry); err != nil {
				slog.Error("Failed to send critical alert", "error", err, "audit_id", entry.ID)
			}
		}()
	}

	return nil
}

// RecordSecurityEvent creates an audit entry for a security event
func (s *AuditTrailService) RecordSecurityEvent(ctx context.Context, req billing_in.RecordSecurityEventRequest) error {
	resourceOwner := common.GetResourceOwner(ctx)

	entry := billing_entities.NewAuditTrailEntry(
		req.EventType,
		req.Severity,
		req.UserID,
		"user",
		req.TargetType,
		req.TargetID,
		req.Description,
		resourceOwner,
	)

	entry.SetActorContext(req.IP, req.UserAgent, req.SessionID)
	entry.Metadata = req.Metadata

	// Get previous entry for hash chain
	previousEntry, err := s.repo.GetLatestForTarget(ctx, req.TargetType, req.TargetID)
	previousHash := ""
	if err == nil && previousEntry != nil {
		previousHash = previousEntry.Hash
		entry.PreviousEntryID = &previousEntry.ID
	}

	entry.ComputeHash(previousHash)

	if err := s.repo.Create(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "Failed to create security audit entry",
			"error", err,
			"event_type", req.EventType,
		)
		return fmt.Errorf("failed to create security audit entry: %w", err)
	}

	slog.InfoContext(ctx, "Security audit event recorded",
		"audit_id", entry.ID,
		"event_type", req.EventType,
		"severity", req.Severity,
	)

	// Alert on critical security events
	if req.Severity == billing_entities.AuditSeverityCritical || req.Severity == billing_entities.AuditSeverityAlert {
		if s.alertSender != nil {
			go func() {
				if err := s.alertSender.SendCriticalAlert(context.Background(), entry); err != nil {
					slog.Error("Failed to send security alert", "error", err)
				}
			}()
		}
	}

	return nil
}

// RecordAdminAction creates an audit entry for an admin action
func (s *AuditTrailService) RecordAdminAction(ctx context.Context, req billing_in.RecordAdminActionRequest) error {
	resourceOwner := common.GetResourceOwner(ctx)

	entry := billing_entities.NewAuditTrailEntry(
		billing_entities.AuditEventAdminAction,
		billing_entities.AuditSeverityWarning,
		req.AdminUserID,
		"admin",
		req.TargetType,
		req.TargetID,
		fmt.Sprintf("Admin action: %s - %s", req.Action, req.Reason),
		resourceOwner,
	)

	entry.TargetUserID = &req.TargetUserID
	entry.SetActorContext(req.IP, req.UserAgent, "")
	entry.SetStateChange(req.PreviousState, req.NewState)

	// Get previous entry for hash chain
	previousEntry, err := s.repo.GetLatestForTarget(ctx, req.TargetType, req.TargetID)
	previousHash := ""
	if err == nil && previousEntry != nil {
		previousHash = previousEntry.Hash
		entry.PreviousEntryID = &previousEntry.ID
	}

	entry.ComputeHash(previousHash)

	if err := s.repo.Create(ctx, entry); err != nil {
		return fmt.Errorf("failed to create admin audit entry: %w", err)
	}

	slog.WarnContext(ctx, "Admin action audit recorded",
		"audit_id", entry.ID,
		"admin_id", req.AdminUserID,
		"action", req.Action,
		"target_user", req.TargetUserID,
	)

	return nil
}

// VerifyChainIntegrity checks the hash chain for tampering
func (s *AuditTrailService) VerifyChainIntegrity(ctx context.Context, targetType string, targetID uuid.UUID, from, to time.Time) (*billing_in.ChainIntegrityResult, error) {
	entries, err := s.repo.GetChainForVerification(ctx, targetType, targetID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain for verification: %w", err)
	}

	result := &billing_in.ChainIntegrityResult{
		Valid:          true,
		EntriesChecked: len(entries),
		VerifiedAt:     time.Now().UTC(),
	}

	if len(entries) == 0 {
		result.Message = "No entries found in the specified range"
		return result, nil
	}

	// Verify chain integrity
	previousHash := ""
	for i, entry := range entries {
		expectedHash := entry.ComputeHash(previousHash)
		if entry.Hash != expectedHash {
			result.Valid = false
			result.FirstBrokenAt = &entry.ID
			result.Message = fmt.Sprintf("Chain integrity broken at entry %d (ID: %s)", i, entry.ID)

			slog.ErrorContext(ctx, "CRITICAL: Audit chain integrity violation detected",
				"entry_id", entry.ID,
				"position", i,
				"expected_hash", expectedHash,
				"actual_hash", entry.Hash,
			)

			return result, nil
		}
		previousHash = entry.Hash
	}

	result.Message = fmt.Sprintf("All %d entries verified successfully", len(entries))
	return result, nil
}

// GetUserAuditHistory retrieves audit history for a user
func (s *AuditTrailService) GetUserAuditHistory(ctx context.Context, userID uuid.UUID, filters billing_in.AuditFilters) (*billing_in.AuditHistoryResult, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	// Authorization: users can only view their own audit history unless admin
	if userID != resourceOwner.UserID && !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden("cannot view another user's audit history")
	}

	limit := filters.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entries, err := s.repo.GetByUser(ctx, userID, limit+1, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user audit history: %w", err)
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	return &billing_in.AuditHistoryResult{
		Entries:    entries,
		TotalCount: int64(len(entries)),
		HasMore:    hasMore,
	}, nil
}

// GetTransactionAudit retrieves the full audit trail for a transaction
func (s *AuditTrailService) GetTransactionAudit(ctx context.Context, transactionID uuid.UUID) ([]billing_entities.AuditTrailEntry, error) {
	return s.repo.GetByTarget(ctx, "transaction", transactionID, 100, 0)
}

// GenerateComplianceReport generates a compliance report for a period
func (s *AuditTrailService) GenerateComplianceReport(ctx context.Context, reportType string, from, to time.Time) (*billing_entities.ComplianceReport, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	// Only admins can generate compliance reports
	if !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden("only administrators can generate compliance reports")
	}

	entries, err := s.repo.GetForComplianceReport(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries for compliance report: %w", err)
	}

	report := &billing_entities.ComplianceReport{
		ReportID:          uuid.New(),
		ReportType:        reportType,
		GeneratedAt:       time.Now().UTC(),
		PeriodStart:       from,
		PeriodEnd:         to,
		TotalTransactions: len(entries),
		DataIntegrity:     true,
		HashChainValid:    true,
		GeneratedBy:       resourceOwner.UserID,
		Findings:          []billing_entities.ComplianceFinding{},
	}

	// Calculate totals and check for anomalies
	var totalVolume float64
	highRiskCount := 0
	previousHash := ""

	for _, entry := range entries {
		// Sum financial volume
		if entry.Amount != nil {
			totalVolume += *entry.Amount
		}

		// Count high-risk events
		if entry.Severity == billing_entities.AuditSeverityCritical || entry.Severity == billing_entities.AuditSeverityAlert {
			highRiskCount++
		}

		// Verify hash chain
		if !entry.VerifyHash(previousHash) {
			report.HashChainValid = false
			report.DataIntegrity = false
			report.Findings = append(report.Findings, billing_entities.ComplianceFinding{
				Severity:        billing_entities.AuditSeverityCritical,
				Category:        "DATA_INTEGRITY",
				Description:     fmt.Sprintf("Hash chain broken at entry %s", entry.ID),
				Recommendation:  "Investigate potential data tampering",
				AffectedEntries: []uuid.UUID{entry.ID},
			})
		}
		previousHash = entry.Hash
	}

	report.TotalVolume = totalVolume
	report.HighRiskEvents = highRiskCount

	// Check for anomalies
	report.AnomaliesDetected = s.detectAnomalies(entries, &report.Findings)

	slog.InfoContext(ctx, "Compliance report generated",
		"report_id", report.ReportID,
		"report_type", reportType,
		"total_transactions", report.TotalTransactions,
		"findings", len(report.Findings),
	)

	return report, nil
}

// GetAuditSummary gets aggregated audit statistics
func (s *AuditTrailService) GetAuditSummary(ctx context.Context, userID uuid.UUID, period string) (*billing_entities.AuditSummary, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	if userID != resourceOwner.UserID && !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden("cannot view another user's audit summary")
	}

	now := time.Now().UTC()
	var from, to time.Time

	switch period {
	case "daily":
		from = now.Truncate(24 * time.Hour)
		to = from.Add(24 * time.Hour)
	case "weekly":
		from = now.AddDate(0, 0, -int(now.Weekday()))
		to = from.AddDate(0, 0, 7)
	case "monthly":
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
	default:
		from = now.AddDate(0, 0, -30)
		to = now
	}

	return s.repo.GetFinancialSummary(ctx, userID, from, to)
}

// SearchAudit performs a filtered search across audit entries
func (s *AuditTrailService) SearchAudit(ctx context.Context, query billing_in.AuditSearchQuery) (*billing_in.AuditSearchResult, error) {
	if !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden("only administrators can search audit logs")
	}

	startTime := time.Now()

	limit := query.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	// Build filters based on query
	filters := billing_in.AuditFilters{
		EventTypes:  query.EventTypes,
		Severities:  query.Severities,
		FromDate:    query.FromDate,
		ToDate:      query.ToDate,
		Limit:       limit + 1,
		Offset:      query.Offset,
	}

	// Use combined search if available
	entries := make([]billing_entities.AuditTrailEntry, 0)
	for _, userID := range query.UserIDs {
		userEntries, err := s.repo.GetByUser(ctx, userID, filters.Limit, filters.Offset)
		if err != nil {
			continue
		}
		entries = append(entries, userEntries...)
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	return &billing_in.AuditSearchResult{
		Entries:    entries,
		TotalCount: int64(len(entries)),
		HasMore:    hasMore,
		SearchTime: time.Since(startTime).Milliseconds(),
	}, nil
}

// ExportAudit exports audit entries for external compliance systems
func (s *AuditTrailService) ExportAudit(ctx context.Context, from, to time.Time, format string) ([]byte, error) {
	if !common.IsAdmin(ctx) {
		return nil, common.NewErrForbidden("only administrators can export audit logs")
	}

	entries, err := s.repo.GetForComplianceReport(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries for export: %w", err)
	}

	switch format {
	case "json":
		return json.MarshalIndent(entries, "", "  ")
	case "csv":
		return s.exportToCSV(entries)
	default:
		return json.Marshal(entries)
	}
}

// determineFinancialSeverity calculates severity based on transaction details
func (s *AuditTrailService) determineFinancialSeverity(amount float64, eventType billing_entities.AuditEventType) billing_entities.AuditSeverity {
	// Large transaction thresholds
	if amount >= 10000 {
		return billing_entities.AuditSeverityAlert
	}
	if amount >= 1000 {
		return billing_entities.AuditSeverityCritical
	}
	if amount >= 100 {
		return billing_entities.AuditSeverityWarning
	}

	// Specific event types that are always higher severity
	criticalEvents := map[billing_entities.AuditEventType]bool{
		billing_entities.AuditEventWithdrawal:        true,
		billing_entities.AuditEventRefund:            true,
		billing_entities.AuditEventBalanceAdjustment: true,
		billing_entities.AuditEventManualAdjustment:  true,
	}

	if criticalEvents[eventType] {
		return billing_entities.AuditSeverityWarning
	}

	return billing_entities.AuditSeverityInfo
}

// detectAnomalies checks for suspicious patterns in audit entries
func (s *AuditTrailService) detectAnomalies(entries []billing_entities.AuditTrailEntry, findings *[]billing_entities.ComplianceFinding) int {
	anomalyCount := 0

	// Track rapid successive transactions per user
	userTransactions := make(map[uuid.UUID][]time.Time)
	userAmounts := make(map[uuid.UUID]float64)

	for _, entry := range entries {
		if entry.Amount != nil && entry.IsFinancialEvent() {
			userTransactions[entry.ActorUserID] = append(userTransactions[entry.ActorUserID], entry.Timestamp)
			userAmounts[entry.ActorUserID] += *entry.Amount
		}
	}

	// Check for velocity anomalies (too many transactions in short time)
	for userID, timestamps := range userTransactions {
		if len(timestamps) > 10 {
			// Check if 10+ transactions in 1 hour
			first := timestamps[0]
			for i := 10; i < len(timestamps); i++ {
				if timestamps[i].Sub(first) < time.Hour {
					*findings = append(*findings, billing_entities.ComplianceFinding{
						Severity:       billing_entities.AuditSeverityWarning,
						Category:       "VELOCITY_ANOMALY",
						Description:    fmt.Sprintf("User %s had %d transactions within 1 hour", userID, i+1),
						Recommendation: "Review for potential automated fraud or abuse",
					})
					anomalyCount++
					break
				}
				first = timestamps[i-10]
			}
		}

		// Check for high volume
		if userAmounts[userID] > 50000 {
			*findings = append(*findings, billing_entities.ComplianceFinding{
				Severity:       billing_entities.AuditSeverityCritical,
				Category:       "HIGH_VOLUME",
				Description:    fmt.Sprintf("User %s transacted $%.2f in the reporting period", userID, userAmounts[userID]),
				Recommendation: "Verify KYC status and source of funds",
			})
			anomalyCount++
		}
	}

	return anomalyCount
}

// exportToCSV converts entries to CSV format
func (s *AuditTrailService) exportToCSV(entries []billing_entities.AuditTrailEntry) ([]byte, error) {
	var result []byte
	header := "ID,EventType,Severity,Timestamp,ActorUserID,TargetType,TargetID,Amount,Currency,Description\n"
	result = append(result, []byte(header)...)

	for _, entry := range entries {
		amount := ""
		currency := ""
		if entry.Amount != nil {
			amount = fmt.Sprintf("%.2f", *entry.Amount)
		}
		if entry.Currency != nil {
			currency = *entry.Currency
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,\"%s\"\n",
			entry.ID,
			entry.EventType,
			entry.Severity,
			entry.Timestamp.Format(time.RFC3339),
			entry.ActorUserID,
			entry.TargetType,
			entry.TargetID,
			amount,
			currency,
			entry.Description,
		)
		result = append(result, []byte(line)...)
	}

	return result, nil
}

// Ensure interface compliance
var _ billing_in.AuditTrailCommand = (*AuditTrailService)(nil)
var _ billing_in.AuditTrailQuery = (*AuditTrailService)(nil)

// NoOpAuditTrailService provides a no-op implementation for basic functionality
type NoOpAuditTrailService struct{}

// NewNoOpAuditTrailService creates a new no-op audit trail service
func NewNoOpAuditTrailService() *NoOpAuditTrailService {
	return &NoOpAuditTrailService{}
}

// RecordFinancialEvent is a no-op implementation
func (s *NoOpAuditTrailService) RecordFinancialEvent(ctx context.Context, req billing_in.RecordFinancialEventRequest) error {
	return nil
}

// RecordSecurityEvent is a no-op implementation
func (s *NoOpAuditTrailService) RecordSecurityEvent(ctx context.Context, req billing_in.RecordSecurityEventRequest) error {
	return nil
}

// RecordAdminAction is a no-op implementation
func (s *NoOpAuditTrailService) RecordAdminAction(ctx context.Context, req billing_in.RecordAdminActionRequest) error {
	return nil
}

// VerifyChainIntegrity is a no-op implementation
func (s *NoOpAuditTrailService) VerifyChainIntegrity(ctx context.Context, targetType string, targetID uuid.UUID, from, to time.Time) (*billing_in.ChainIntegrityResult, error) {
	return &billing_in.ChainIntegrityResult{
		Valid:          true,
		EntriesChecked: 0,
		VerifiedAt:     time.Now(),
		Message:        "No-op audit trail - integrity check skipped",
	}, nil
}

// Ensure interface compliance for no-op service
var _ billing_in.AuditTrailCommand = (*NoOpAuditTrailService)(nil)

