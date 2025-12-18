package billing_entities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

// AuditEventType categorizes audit events for compliance reporting
type AuditEventType string

const (
	// Financial Events - SOX/PCI-DSS Compliance
	AuditEventDeposit          AuditEventType = "DEPOSIT"
	AuditEventWithdrawal       AuditEventType = "WITHDRAWAL"
	AuditEventTransfer         AuditEventType = "TRANSFER"
	AuditEventRefund           AuditEventType = "REFUND"
	AuditEventPrizeDistribution AuditEventType = "PRIZE_DISTRIBUTION"
	AuditEventEntryFee         AuditEventType = "ENTRY_FEE"
	AuditEventSubscriptionPurchase AuditEventType = "SUBSCRIPTION_PURCHASE"
	AuditEventSubscriptionCancel   AuditEventType = "SUBSCRIPTION_CANCEL"

	// Balance Events
	AuditEventBalanceCredit    AuditEventType = "BALANCE_CREDIT"
	AuditEventBalanceDebit     AuditEventType = "BALANCE_DEBIT"
	AuditEventBalanceAdjustment AuditEventType = "BALANCE_ADJUSTMENT"
	AuditEventBalanceFreeze    AuditEventType = "BALANCE_FREEZE"
	AuditEventBalanceUnfreeze  AuditEventType = "BALANCE_UNFREEZE"

	// Security Events
	AuditEventLoginSuccess     AuditEventType = "LOGIN_SUCCESS"
	AuditEventLoginFailed      AuditEventType = "LOGIN_FAILED"
	AuditEventMFAEnabled       AuditEventType = "MFA_ENABLED"
	AuditEventMFADisabled      AuditEventType = "MFA_DISABLED"
	AuditEventPasswordChanged  AuditEventType = "PASSWORD_CHANGED"
	AuditEventAPIKeyCreated    AuditEventType = "API_KEY_CREATED"
	AuditEventAPIKeyRevoked    AuditEventType = "API_KEY_REVOKED"

	// Admin Events
	AuditEventAdminAction      AuditEventType = "ADMIN_ACTION"
	AuditEventManualAdjustment AuditEventType = "MANUAL_ADJUSTMENT"
	AuditEventKYCVerified      AuditEventType = "KYC_VERIFIED"
	AuditEventKYCRejected      AuditEventType = "KYC_REJECTED"
	AuditEventAccountSuspended AuditEventType = "ACCOUNT_SUSPENDED"
	AuditEventAccountReinstated AuditEventType = "ACCOUNT_REINSTATED"

	// Dispute Events
	AuditEventDisputeOpened    AuditEventType = "DISPUTE_OPENED"
	AuditEventDisputeResolved  AuditEventType = "DISPUTE_RESOLVED"
	AuditEventChargebackReceived AuditEventType = "CHARGEBACK_RECEIVED"
)

// AuditSeverity indicates the importance level for monitoring
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "INFO"
	AuditSeverityWarning  AuditSeverity = "WARNING"
	AuditSeverityCritical AuditSeverity = "CRITICAL"
	AuditSeverityAlert    AuditSeverity = "ALERT"
)

// AuditTrailEntry represents an immutable audit log entry
// Designed for SOX/PCI-DSS/GDPR compliance
type AuditTrailEntry struct {
	ID              uuid.UUID              `json:"id" bson:"_id"`
	EventType       AuditEventType         `json:"event_type" bson:"event_type"`
	Severity        AuditSeverity          `json:"severity" bson:"severity"`
	Timestamp       time.Time              `json:"timestamp" bson:"timestamp"`
	
	// Actor Information (who performed the action)
	ActorUserID     uuid.UUID              `json:"actor_user_id" bson:"actor_user_id"`
	ActorType       string                 `json:"actor_type" bson:"actor_type"` // user, admin, system, api
	ActorIP         string                 `json:"actor_ip" bson:"actor_ip"`
	ActorUserAgent  string                 `json:"actor_user_agent" bson:"actor_user_agent"`
	ActorSessionID  string                 `json:"actor_session_id,omitempty" bson:"actor_session_id,omitempty"`

	// Target Information (what was affected)
	TargetUserID    *uuid.UUID             `json:"target_user_id,omitempty" bson:"target_user_id,omitempty"`
	TargetType      string                 `json:"target_type" bson:"target_type"` // wallet, subscription, user, etc.
	TargetID        uuid.UUID              `json:"target_id" bson:"target_id"`

	// Financial Details (for monetary events)
	Amount          *float64               `json:"amount,omitempty" bson:"amount,omitempty"`
	Currency        *string                `json:"currency,omitempty" bson:"currency,omitempty"`
	BalanceBefore   *float64               `json:"balance_before,omitempty" bson:"balance_before,omitempty"`
	BalanceAfter    *float64               `json:"balance_after,omitempty" bson:"balance_after,omitempty"`

	// Transaction Linking
	TransactionID   *uuid.UUID             `json:"transaction_id,omitempty" bson:"transaction_id,omitempty"`
	ExternalRef     *string                `json:"external_ref,omitempty" bson:"external_ref,omitempty"`
	ProviderRef     *string                `json:"provider_ref,omitempty" bson:"provider_ref,omitempty"`

	// Event Details
	Description     string                 `json:"description" bson:"description"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	PreviousState   map[string]interface{} `json:"previous_state,omitempty" bson:"previous_state,omitempty"`
	NewState        map[string]interface{} `json:"new_state,omitempty" bson:"new_state,omitempty"`

	// Integrity & Chain
	PreviousEntryID *uuid.UUID             `json:"previous_entry_id,omitempty" bson:"previous_entry_id,omitempty"`
	Hash            string                 `json:"hash" bson:"hash"`

	// Compliance Fields
	RetentionUntil  time.Time              `json:"retention_until" bson:"retention_until"`
	Exportable      bool                   `json:"exportable" bson:"exportable"`
	Anonymized      bool                   `json:"anonymized" bson:"anonymized"`

	// Tenant/Organization
	TenantID        uuid.UUID              `json:"tenant_id" bson:"tenant_id"`
	ClientID        uuid.UUID              `json:"client_id" bson:"client_id"`
}

// NewAuditTrailEntry creates a new audit trail entry with proper defaults
func NewAuditTrailEntry(
	eventType AuditEventType,
	severity AuditSeverity,
	actorUserID uuid.UUID,
	actorType string,
	targetType string,
	targetID uuid.UUID,
	description string,
	rxn common.ResourceOwner,
) *AuditTrailEntry {
	now := time.Now().UTC()
	
	entry := &AuditTrailEntry{
		ID:             uuid.New(),
		EventType:      eventType,
		Severity:       severity,
		Timestamp:      now,
		ActorUserID:    actorUserID,
		ActorType:      actorType,
		TargetType:     targetType,
		TargetID:       targetID,
		Description:    description,
		RetentionUntil: now.AddDate(7, 0, 0), // 7 year retention for financial records
		Exportable:     true,
		Anonymized:     false,
		TenantID:       rxn.TenantID,
		ClientID:       rxn.ClientID,
		Metadata:       make(map[string]interface{}),
	}

	return entry
}

// SetFinancialDetails sets financial transaction details
func (a *AuditTrailEntry) SetFinancialDetails(amount float64, currency string, balanceBefore, balanceAfter float64) {
	a.Amount = &amount
	a.Currency = &currency
	a.BalanceBefore = &balanceBefore
	a.BalanceAfter = &balanceAfter
}

// SetTransactionRef sets transaction reference information
func (a *AuditTrailEntry) SetTransactionRef(transactionID uuid.UUID, externalRef, providerRef string) {
	a.TransactionID = &transactionID
	if externalRef != "" {
		a.ExternalRef = &externalRef
	}
	if providerRef != "" {
		a.ProviderRef = &providerRef
	}
}

// SetActorContext sets the actor's request context
func (a *AuditTrailEntry) SetActorContext(ip, userAgent, sessionID string) {
	a.ActorIP = ip
	a.ActorUserAgent = userAgent
	a.ActorSessionID = sessionID
}

// SetStateChange records before/after state for auditing
func (a *AuditTrailEntry) SetStateChange(previous, new map[string]interface{}) {
	a.PreviousState = previous
	a.NewState = new
}

// ComputeHash generates a SHA-256 hash of the entry for integrity verification
// This creates an immutable chain similar to blockchain principles
func (a *AuditTrailEntry) ComputeHash(previousHash string) string {
	data := struct {
		ID            uuid.UUID
		EventType     AuditEventType
		Timestamp     time.Time
		ActorUserID   uuid.UUID
		TargetID      uuid.UUID
		Amount        *float64
		Description   string
		PreviousHash  string
	}{
		ID:           a.ID,
		EventType:    a.EventType,
		Timestamp:    a.Timestamp,
		ActorUserID:  a.ActorUserID,
		TargetID:     a.TargetID,
		Amount:       a.Amount,
		Description:  a.Description,
		PreviousHash: previousHash,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	a.Hash = hex.EncodeToString(hash[:])
	return a.Hash
}

// VerifyHash checks if the entry's hash is valid
func (a *AuditTrailEntry) VerifyHash(previousHash string) bool {
	expectedHash := a.ComputeHash(previousHash)
	return a.Hash == expectedHash
}

// GetID returns the entry ID
func (a AuditTrailEntry) GetID() uuid.UUID {
	return a.ID
}

// IsFinancialEvent returns true if this is a monetary transaction
func (a *AuditTrailEntry) IsFinancialEvent() bool {
	financialEvents := map[AuditEventType]bool{
		AuditEventDeposit:            true,
		AuditEventWithdrawal:         true,
		AuditEventTransfer:           true,
		AuditEventRefund:             true,
		AuditEventPrizeDistribution:  true,
		AuditEventEntryFee:           true,
		AuditEventSubscriptionPurchase: true,
		AuditEventBalanceCredit:      true,
		AuditEventBalanceDebit:       true,
		AuditEventBalanceAdjustment:  true,
	}
	return financialEvents[a.EventType]
}

// IsCritical returns true if this event requires immediate attention
func (a *AuditTrailEntry) IsCritical() bool {
	return a.Severity == AuditSeverityCritical || a.Severity == AuditSeverityAlert
}

// AuditSummary provides aggregated audit statistics
type AuditSummary struct {
	UserID            uuid.UUID                   `json:"user_id"`
	Period            string                      `json:"period"` // daily, weekly, monthly
	StartDate         time.Time                   `json:"start_date"`
	EndDate           time.Time                   `json:"end_date"`
	TotalEvents       int                         `json:"total_events"`
	EventsByType      map[AuditEventType]int      `json:"events_by_type"`
	EventsBySeverity  map[AuditSeverity]int       `json:"events_by_severity"`
	TotalDeposits     float64                     `json:"total_deposits"`
	TotalWithdrawals  float64                     `json:"total_withdrawals"`
	TotalFees         float64                     `json:"total_fees"`
	NetChange         float64                     `json:"net_change"`
	UniqueIPs         int                         `json:"unique_ips"`
	FailedLogins      int                         `json:"failed_logins"`
	GeneratedAt       time.Time                   `json:"generated_at"`
}

// ComplianceReport represents a regulatory compliance report
type ComplianceReport struct {
	ReportID          uuid.UUID                   `json:"report_id"`
	ReportType        string                      `json:"report_type"` // SOX, PCI-DSS, GDPR, AML
	GeneratedAt       time.Time                   `json:"generated_at"`
	PeriodStart       time.Time                   `json:"period_start"`
	PeriodEnd         time.Time                   `json:"period_end"`
	TotalTransactions int                         `json:"total_transactions"`
	TotalVolume       float64                     `json:"total_volume"`
	HighRiskEvents    int                         `json:"high_risk_events"`
	AnomaliesDetected int                         `json:"anomalies_detected"`
	DataIntegrity     bool                        `json:"data_integrity"`
	HashChainValid    bool                        `json:"hash_chain_valid"`
	Findings          []ComplianceFinding         `json:"findings"`
	GeneratedBy       uuid.UUID                   `json:"generated_by"`
}

// ComplianceFinding represents an issue found during compliance check
type ComplianceFinding struct {
	Severity      AuditSeverity `json:"severity"`
	Category      string        `json:"category"`
	Description   string        `json:"description"`
	Recommendation string       `json:"recommendation"`
	AffectedEntries []uuid.UUID `json:"affected_entries"`
}

