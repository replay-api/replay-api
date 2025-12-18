package billing_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

// AuditTrailCommand defines operations for creating audit entries
type AuditTrailCommand interface {
	// RecordFinancialEvent creates an audit entry for a financial transaction
	RecordFinancialEvent(ctx context.Context, req RecordFinancialEventRequest) error

	// RecordSecurityEvent creates an audit entry for a security event
	RecordSecurityEvent(ctx context.Context, req RecordSecurityEventRequest) error

	// RecordAdminAction creates an audit entry for an admin action
	RecordAdminAction(ctx context.Context, req RecordAdminActionRequest) error

	// VerifyChainIntegrity checks the hash chain for tampering
	VerifyChainIntegrity(ctx context.Context, targetType string, targetID uuid.UUID, from, to time.Time) (*ChainIntegrityResult, error)
}

// AuditTrailQuery defines operations for retrieving audit data
type AuditTrailQuery interface {
	// GetUserAuditHistory retrieves audit history for a user
	GetUserAuditHistory(ctx context.Context, userID uuid.UUID, filters AuditFilters) (*AuditHistoryResult, error)

	// GetTransactionAudit retrieves the full audit trail for a transaction
	GetTransactionAudit(ctx context.Context, transactionID uuid.UUID) ([]billing_entities.AuditTrailEntry, error)

	// GenerateComplianceReport generates a compliance report for a period
	GenerateComplianceReport(ctx context.Context, reportType string, from, to time.Time) (*billing_entities.ComplianceReport, error)

	// GetAuditSummary gets aggregated audit statistics
	GetAuditSummary(ctx context.Context, userID uuid.UUID, period string) (*billing_entities.AuditSummary, error)

	// SearchAudit performs a filtered search across audit entries
	SearchAudit(ctx context.Context, query AuditSearchQuery) (*AuditSearchResult, error)

	// ExportAudit exports audit entries for external compliance systems
	ExportAudit(ctx context.Context, from, to time.Time, format string) ([]byte, error)
}

// RecordFinancialEventRequest contains data for recording a financial audit event
type RecordFinancialEventRequest struct {
	EventType     billing_entities.AuditEventType
	UserID        uuid.UUID
	TargetType    string
	TargetID      uuid.UUID
	Amount        float64
	Currency      string
	BalanceBefore float64
	BalanceAfter  float64
	TransactionID uuid.UUID
	ExternalRef   string
	ProviderRef   string
	Description   string
	Metadata      map[string]interface{}
	IP            string
	UserAgent     string
	SessionID     string
}

// RecordSecurityEventRequest contains data for recording a security audit event
type RecordSecurityEventRequest struct {
	EventType   billing_entities.AuditEventType
	Severity    billing_entities.AuditSeverity
	UserID      uuid.UUID
	TargetType  string
	TargetID    uuid.UUID
	Description string
	Metadata    map[string]interface{}
	IP          string
	UserAgent   string
	SessionID   string
}

// RecordAdminActionRequest contains data for recording an admin audit event
type RecordAdminActionRequest struct {
	AdminUserID   uuid.UUID
	TargetUserID  uuid.UUID
	TargetType    string
	TargetID      uuid.UUID
	Action        string
	Reason        string
	PreviousState map[string]interface{}
	NewState      map[string]interface{}
	IP            string
	UserAgent     string
}

// AuditFilters defines filtering options for audit queries
type AuditFilters struct {
	EventTypes  []billing_entities.AuditEventType
	Severities  []billing_entities.AuditSeverity
	FromDate    *time.Time
	ToDate      *time.Time
	TargetTypes []string
	Limit       int
	Offset      int
}

// AuditHistoryResult contains paginated audit history
type AuditHistoryResult struct {
	Entries    []billing_entities.AuditTrailEntry `json:"entries"`
	TotalCount int64                               `json:"total_count"`
	HasMore    bool                                `json:"has_more"`
}

// ChainIntegrityResult contains the result of a chain integrity check
type ChainIntegrityResult struct {
	Valid          bool      `json:"valid"`
	EntriesChecked int       `json:"entries_checked"`
	FirstBrokenAt  *uuid.UUID `json:"first_broken_at,omitempty"`
	VerifiedAt     time.Time  `json:"verified_at"`
	Message        string     `json:"message"`
}

// AuditSearchQuery defines a search query for audit entries
type AuditSearchQuery struct {
	Query       string                              `json:"query"`
	EventTypes  []billing_entities.AuditEventType   `json:"event_types,omitempty"`
	Severities  []billing_entities.AuditSeverity    `json:"severities,omitempty"`
	UserIDs     []uuid.UUID                         `json:"user_ids,omitempty"`
	FromDate    *time.Time                          `json:"from_date,omitempty"`
	ToDate      *time.Time                          `json:"to_date,omitempty"`
	MinAmount   *float64                            `json:"min_amount,omitempty"`
	MaxAmount   *float64                            `json:"max_amount,omitempty"`
	Currencies  []string                            `json:"currencies,omitempty"`
	Limit       int                                 `json:"limit"`
	Offset      int                                 `json:"offset"`
}

// AuditSearchResult contains search results
type AuditSearchResult struct {
	Entries    []billing_entities.AuditTrailEntry `json:"entries"`
	TotalCount int64                               `json:"total_count"`
	HasMore    bool                                `json:"has_more"`
	SearchTime int64                               `json:"search_time_ms"`
}

