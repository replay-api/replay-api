package billing_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

// AuditTrailRepository defines the storage interface for audit entries
// Must support immutability and chain verification for compliance
type AuditTrailRepository interface {
	// Create persists a new audit entry (must be append-only)
	Create(ctx context.Context, entry *billing_entities.AuditTrailEntry) error

	// GetByID retrieves a specific audit entry
	GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.AuditTrailEntry, error)

	// GetLatestForTarget gets the most recent audit entry for a target
	GetLatestForTarget(ctx context.Context, targetType string, targetID uuid.UUID) (*billing_entities.AuditTrailEntry, error)

	// GetByUser retrieves audit entries for a specific user
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]billing_entities.AuditTrailEntry, error)

	// GetByTarget retrieves audit entries for a specific target entity
	GetByTarget(ctx context.Context, targetType string, targetID uuid.UUID, limit, offset int) ([]billing_entities.AuditTrailEntry, error)

	// GetByEventType retrieves entries of a specific event type
	GetByEventType(ctx context.Context, eventType billing_entities.AuditEventType, from, to time.Time, limit, offset int) ([]billing_entities.AuditTrailEntry, error)

	// GetBySeverity retrieves entries at or above a severity level
	GetBySeverity(ctx context.Context, severity billing_entities.AuditSeverity, from, to time.Time, limit, offset int) ([]billing_entities.AuditTrailEntry, error)

	// GetChainForVerification retrieves entries for hash chain verification
	GetChainForVerification(ctx context.Context, targetType string, targetID uuid.UUID, from, to time.Time) ([]billing_entities.AuditTrailEntry, error)

	// GetForComplianceReport retrieves entries for compliance reporting
	GetForComplianceReport(ctx context.Context, from, to time.Time) ([]billing_entities.AuditTrailEntry, error)

	// CountByType counts entries by event type within a period
	CountByType(ctx context.Context, eventType billing_entities.AuditEventType, from, to time.Time) (int64, error)

	// GetFinancialSummary retrieves aggregated financial audit data
	GetFinancialSummary(ctx context.Context, userID uuid.UUID, from, to time.Time) (*billing_entities.AuditSummary, error)

	// ArchiveOldEntries moves entries past retention to cold storage
	ArchiveOldEntries(ctx context.Context, before time.Time) (int64, error)
}

// AuditAlertSender sends critical audit alerts to monitoring systems
type AuditAlertSender interface {
	// SendCriticalAlert sends an immediate alert for critical events
	SendCriticalAlert(ctx context.Context, entry *billing_entities.AuditTrailEntry) error

	// SendDailyDigest sends a daily summary of audit events
	SendDailyDigest(ctx context.Context, summary *billing_entities.AuditSummary) error

	// SendComplianceReport sends a compliance report to stakeholders
	SendComplianceReport(ctx context.Context, report *billing_entities.ComplianceReport) error
}

