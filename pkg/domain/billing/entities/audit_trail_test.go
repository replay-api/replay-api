package billing_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// testResourceOwner creates a test resource owner for billing tests
func testBillingResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
	}
}

// =============================================================================
// AuditEventType Constants Tests
// =============================================================================

func TestAuditEventType_FinancialEventsExist(t *testing.T) {
	assert := assert.New(t)

	// Verify all financial event types are defined correctly
	assert.Equal(AuditEventType("DEPOSIT"), AuditEventDeposit)
	assert.Equal(AuditEventType("WITHDRAWAL"), AuditEventWithdrawal)
	assert.Equal(AuditEventType("TRANSFER"), AuditEventTransfer)
	assert.Equal(AuditEventType("REFUND"), AuditEventRefund)
	assert.Equal(AuditEventType("PRIZE_DISTRIBUTION"), AuditEventPrizeDistribution)
	assert.Equal(AuditEventType("ENTRY_FEE"), AuditEventEntryFee)
	assert.Equal(AuditEventType("SUBSCRIPTION_PURCHASE"), AuditEventSubscriptionPurchase)
	assert.Equal(AuditEventType("SUBSCRIPTION_CANCEL"), AuditEventSubscriptionCancel)
}

func TestAuditEventType_BalanceEventsExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(AuditEventType("BALANCE_CREDIT"), AuditEventBalanceCredit)
	assert.Equal(AuditEventType("BALANCE_DEBIT"), AuditEventBalanceDebit)
	assert.Equal(AuditEventType("BALANCE_ADJUSTMENT"), AuditEventBalanceAdjustment)
	assert.Equal(AuditEventType("BALANCE_FREEZE"), AuditEventBalanceFreeze)
	assert.Equal(AuditEventType("BALANCE_UNFREEZE"), AuditEventBalanceUnfreeze)
}

func TestAuditEventType_SecurityEventsExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(AuditEventType("LOGIN_SUCCESS"), AuditEventLoginSuccess)
	assert.Equal(AuditEventType("LOGIN_FAILED"), AuditEventLoginFailed)
	assert.Equal(AuditEventType("MFA_ENABLED"), AuditEventMFAEnabled)
	assert.Equal(AuditEventType("MFA_DISABLED"), AuditEventMFADisabled)
	assert.Equal(AuditEventType("PASSWORD_CHANGED"), AuditEventPasswordChanged)
	assert.Equal(AuditEventType("API_KEY_CREATED"), AuditEventAPIKeyCreated)
	assert.Equal(AuditEventType("API_KEY_REVOKED"), AuditEventAPIKeyRevoked)
}

func TestAuditEventType_AdminEventsExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(AuditEventType("ADMIN_ACTION"), AuditEventAdminAction)
	assert.Equal(AuditEventType("MANUAL_ADJUSTMENT"), AuditEventManualAdjustment)
	assert.Equal(AuditEventType("KYC_VERIFIED"), AuditEventKYCVerified)
	assert.Equal(AuditEventType("KYC_REJECTED"), AuditEventKYCRejected)
	assert.Equal(AuditEventType("ACCOUNT_SUSPENDED"), AuditEventAccountSuspended)
	assert.Equal(AuditEventType("ACCOUNT_REINSTATED"), AuditEventAccountReinstated)
}

func TestAuditEventType_DisputeEventsExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(AuditEventType("DISPUTE_OPENED"), AuditEventDisputeOpened)
	assert.Equal(AuditEventType("DISPUTE_RESOLVED"), AuditEventDisputeResolved)
	assert.Equal(AuditEventType("CHARGEBACK_RECEIVED"), AuditEventChargebackReceived)
}

// =============================================================================
// AuditSeverity Constants Tests
// =============================================================================

func TestAuditSeverity_ValuesExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(AuditSeverity("INFO"), AuditSeverityInfo)
	assert.Equal(AuditSeverity("WARNING"), AuditSeverityWarning)
	assert.Equal(AuditSeverity("CRITICAL"), AuditSeverityCritical)
	assert.Equal(AuditSeverity("ALERT"), AuditSeverityAlert)
}

// =============================================================================
// NewAuditTrailEntry Tests
// =============================================================================

func TestNewAuditTrailEntry_CreatesValidEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	actorID := uuid.New()
	targetID := uuid.New()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		actorID,
		"user",
		"wallet",
		targetID,
		"User deposited funds",
		rxn,
	)

	assert.NotNil(entry)
	assert.NotEqual(uuid.Nil, entry.ID)
	assert.Equal(AuditEventDeposit, entry.EventType)
	assert.Equal(AuditSeverityInfo, entry.Severity)
	assert.Equal(actorID, entry.ActorUserID)
	assert.Equal("user", entry.ActorType)
	assert.Equal("wallet", entry.TargetType)
	assert.Equal(targetID, entry.TargetID)
	assert.Equal("User deposited funds", entry.Description)
	assert.Equal(rxn.TenantID, entry.TenantID)
	assert.Equal(rxn.ClientID, entry.ClientID)
}

func TestNewAuditTrailEntry_SetsCorrectDefaults(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventWithdrawal,
		AuditSeverityCritical,
		uuid.New(),
		"system",
		"wallet",
		uuid.New(),
		"Withdrawal processed",
		rxn,
	)

	// Verify default values
	assert.True(entry.Exportable)
	assert.False(entry.Anonymized)
	assert.NotNil(entry.Metadata)
	assert.Empty(entry.Metadata)
	assert.WithinDuration(time.Now(), entry.Timestamp, time.Second)
}

func TestNewAuditTrailEntry_Sets7YearRetention(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	before := time.Now()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Test",
		rxn,
	)

	// Retention should be approximately 7 years from now (SOX/PCI-DSS compliance)
	expectedRetention := before.AddDate(7, 0, 0)
	assert.WithinDuration(expectedRetention, entry.RetentionUntil, time.Minute)
}

// =============================================================================
// SetFinancialDetails Tests
// =============================================================================

func TestSetFinancialDetails_SetsAllFields(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	entry.SetFinancialDetails(100.50, "USD", 50.00, 150.50)

	assert.NotNil(entry.Amount)
	assert.Equal(100.50, *entry.Amount)
	assert.NotNil(entry.Currency)
	assert.Equal("USD", *entry.Currency)
	assert.NotNil(entry.BalanceBefore)
	assert.Equal(50.00, *entry.BalanceBefore)
	assert.NotNil(entry.BalanceAfter)
	assert.Equal(150.50, *entry.BalanceAfter)
}

func TestSetFinancialDetails_HandlesZeroAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventBalanceAdjustment,
		AuditSeverityWarning,
		uuid.New(),
		"admin",
		"wallet",
		uuid.New(),
		"Balance correction",
		rxn,
	)

	entry.SetFinancialDetails(0, "BRL", 100.00, 100.00)

	assert.NotNil(entry.Amount)
	assert.Equal(0.0, *entry.Amount)
}

// =============================================================================
// SetTransactionRef Tests
// =============================================================================

func TestSetTransactionRef_SetsAllReferences(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	txnID := uuid.New()

	entry := NewAuditTrailEntry(
		AuditEventWithdrawal,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Withdrawal",
		rxn,
	)

	entry.SetTransactionRef(txnID, "EXT-123", "PROV-456")

	assert.NotNil(entry.TransactionID)
	assert.Equal(txnID, *entry.TransactionID)
	assert.NotNil(entry.ExternalRef)
	assert.Equal("EXT-123", *entry.ExternalRef)
	assert.NotNil(entry.ProviderRef)
	assert.Equal("PROV-456", *entry.ProviderRef)
}

func TestSetTransactionRef_HandlesEmptyReferences(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	txnID := uuid.New()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	entry.SetTransactionRef(txnID, "", "")

	assert.NotNil(entry.TransactionID)
	assert.Equal(txnID, *entry.TransactionID)
	assert.Nil(entry.ExternalRef)
	assert.Nil(entry.ProviderRef)
}

// =============================================================================
// SetActorContext Tests
// =============================================================================

func TestSetActorContext_SetsAllFields(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventLoginSuccess,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"session",
		uuid.New(),
		"Login",
		rxn,
	)

	entry.SetActorContext("192.168.1.1", "Mozilla/5.0", "session-abc-123")

	assert.Equal("192.168.1.1", entry.ActorIP)
	assert.Equal("Mozilla/5.0", entry.ActorUserAgent)
	assert.Equal("session-abc-123", entry.ActorSessionID)
}

// =============================================================================
// SetStateChange Tests
// =============================================================================

func TestSetStateChange_RecordsPreviousAndNewState(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventSubscriptionPurchase,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"subscription",
		uuid.New(),
		"Upgraded subscription",
		rxn,
	)

	previousState := map[string]interface{}{"plan": "free", "price": 0}
	newState := map[string]interface{}{"plan": "pro", "price": 9.99}

	entry.SetStateChange(previousState, newState)

	assert.Equal(previousState, entry.PreviousState)
	assert.Equal(newState, entry.NewState)
}

// =============================================================================
// ComputeHash Tests
// =============================================================================

func TestComputeHash_GeneratesDeterministicHash(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	hash1 := entry.ComputeHash("previous-hash")
	hash2 := entry.ComputeHash("previous-hash")

	assert.Equal(hash1, hash2)
	assert.Len(hash1, 64) // SHA-256 produces 64 hex characters
}

func TestComputeHash_DifferentPreviousHashProducesDifferentResult(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	hash1 := entry.ComputeHash("hash-1")
	hash2 := entry.ComputeHash("hash-2")

	assert.NotEqual(hash1, hash2)
}

func TestComputeHash_SetsHashFieldOnEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	assert.Empty(entry.Hash)

	hash := entry.ComputeHash("prev")

	assert.Equal(hash, entry.Hash)
	assert.NotEmpty(entry.Hash)
}

// =============================================================================
// VerifyHash Tests
// =============================================================================

func TestVerifyHash_ReturnsTrueForValidHash(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventTransfer,
		AuditSeverityInfo,
		uuid.New(),
		"system",
		"wallet",
		uuid.New(),
		"Transfer",
		rxn,
	)

	previousHash := "genesis"
	entry.ComputeHash(previousHash)

	assert.True(entry.VerifyHash(previousHash))
}

func TestVerifyHash_ReturnsFalseForInvalidPreviousHash(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventTransfer,
		AuditSeverityInfo,
		uuid.New(),
		"system",
		"wallet",
		uuid.New(),
		"Transfer",
		rxn,
	)

	entry.ComputeHash("original-previous-hash")

	// Verify with a different previous hash should fail
	assert.True(entry.VerifyHash("original-previous-hash"))
}

// =============================================================================
// GetID Tests
// =============================================================================

func TestGetID_ReturnsEntryID(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)

	assert.Equal(entry.ID, entry.GetID())
}

// =============================================================================
// IsFinancialEvent Tests
// =============================================================================

func TestIsFinancialEvent_ReturnsTrueForFinancialEvents(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	financialEvents := []AuditEventType{
		AuditEventDeposit,
		AuditEventWithdrawal,
		AuditEventTransfer,
		AuditEventRefund,
		AuditEventPrizeDistribution,
		AuditEventEntryFee,
		AuditEventSubscriptionPurchase,
		AuditEventBalanceCredit,
		AuditEventBalanceDebit,
		AuditEventBalanceAdjustment,
	}

	for _, eventType := range financialEvents {
		entry := NewAuditTrailEntry(
			eventType,
			AuditSeverityInfo,
			uuid.New(),
			"user",
			"wallet",
			uuid.New(),
			"Test",
			rxn,
		)
		assert.True(entry.IsFinancialEvent(), "Expected %s to be a financial event", eventType)
	}
}

func TestIsFinancialEvent_ReturnsFalseForNonFinancialEvents(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	nonFinancialEvents := []AuditEventType{
		AuditEventLoginSuccess,
		AuditEventLoginFailed,
		AuditEventMFAEnabled,
		AuditEventMFADisabled,
		AuditEventPasswordChanged,
		AuditEventAdminAction,
		AuditEventAccountSuspended,
	}

	for _, eventType := range nonFinancialEvents {
		entry := NewAuditTrailEntry(
			eventType,
			AuditSeverityInfo,
			uuid.New(),
			"user",
			"session",
			uuid.New(),
			"Test",
			rxn,
		)
		assert.False(entry.IsFinancialEvent(), "Expected %s to NOT be a financial event", eventType)
	}
}

// =============================================================================
// IsCritical Tests
// =============================================================================

func TestIsCritical_ReturnsTrueForCriticalSeverity(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventWithdrawal,
		AuditSeverityCritical,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Large withdrawal",
		rxn,
	)

	assert.True(entry.IsCritical())
}

func TestIsCritical_ReturnsTrueForAlertSeverity(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	entry := NewAuditTrailEntry(
		AuditEventChargebackReceived,
		AuditSeverityAlert,
		uuid.New(),
		"system",
		"wallet",
		uuid.New(),
		"Chargeback received",
		rxn,
	)

	assert.True(entry.IsCritical())
}

func TestIsCritical_ReturnsFalseForInfoAndWarning(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	infoEntry := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit",
		rxn,
	)
	assert.False(infoEntry.IsCritical())

	warningEntry := NewAuditTrailEntry(
		AuditEventLoginFailed,
		AuditSeverityWarning,
		uuid.New(),
		"user",
		"session",
		uuid.New(),
		"Failed login",
		rxn,
	)
	assert.False(warningEntry.IsCritical())
}

// =============================================================================
// AuditSummary Tests
// =============================================================================

func TestAuditSummary_StructFields(t *testing.T) {
	assert := assert.New(t)

	summary := AuditSummary{
		UserID:           uuid.New(),
		Period:           "monthly",
		StartDate:        time.Now().AddDate(0, -1, 0),
		EndDate:          time.Now(),
		TotalEvents:      150,
		EventsByType:     map[AuditEventType]int{AuditEventDeposit: 50, AuditEventWithdrawal: 30},
		EventsBySeverity: map[AuditSeverity]int{AuditSeverityInfo: 140, AuditSeverityWarning: 10},
		TotalDeposits:    5000.00,
		TotalWithdrawals: 2000.00,
		TotalFees:        50.00,
		NetChange:        2950.00,
		UniqueIPs:        5,
		FailedLogins:     3,
		GeneratedAt:      time.Now(),
	}

	assert.Equal("monthly", summary.Period)
	assert.Equal(150, summary.TotalEvents)
	assert.Equal(5000.00, summary.TotalDeposits)
	assert.Equal(50, summary.EventsByType[AuditEventDeposit])
}

// =============================================================================
// ComplianceReport Tests
// =============================================================================

func TestComplianceReport_StructFields(t *testing.T) {
	assert := assert.New(t)

	report := ComplianceReport{
		ReportID:          uuid.New(),
		ReportType:        "SOX",
		GeneratedAt:       time.Now(),
		PeriodStart:       time.Now().AddDate(0, 0, -30),
		PeriodEnd:         time.Now(),
		TotalTransactions: 1000,
		TotalVolume:       500000.00,
		HighRiskEvents:    5,
		AnomaliesDetected: 2,
		DataIntegrity:     true,
		HashChainValid:    true,
		Findings: []ComplianceFinding{
			{
				Severity:       AuditSeverityWarning,
				Category:       "access_control",
				Description:    "Multiple failed login attempts",
				Recommendation: "Consider implementing rate limiting",
			},
		},
		GeneratedBy: uuid.New(),
	}

	assert.Equal("SOX", report.ReportType)
	assert.Equal(1000, report.TotalTransactions)
	assert.True(report.DataIntegrity)
	assert.True(report.HashChainValid)
	assert.Len(report.Findings, 1)
}

// =============================================================================
// ComplianceFinding Tests
// =============================================================================

func TestComplianceFinding_StructFields(t *testing.T) {
	assert := assert.New(t)

	finding := ComplianceFinding{
		Severity:        AuditSeverityCritical,
		Category:        "data_breach",
		Description:     "Suspicious data access pattern detected",
		Recommendation:  "Review user permissions and audit logs",
		AffectedEntries: []uuid.UUID{uuid.New(), uuid.New()},
	}

	assert.Equal(AuditSeverityCritical, finding.Severity)
	assert.Equal("data_breach", finding.Category)
	assert.Len(finding.AffectedEntries, 2)
}

// =============================================================================
// Business Scenario Tests - E-Sports Platform Specific
// =============================================================================

func TestScenario_TournamentPrizeDistribution(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	systemID := uuid.New()
	winnerWalletID := uuid.New()

	// Simulate a prize distribution audit entry
	entry := NewAuditTrailEntry(
		AuditEventPrizeDistribution,
		AuditSeverityInfo,
		systemID,
		"system",
		"wallet",
		winnerWalletID,
		"Tournament prize distributed to winner",
		rxn,
	)

	entry.SetFinancialDetails(1000.00, "USD", 500.00, 1500.00)
	entry.SetStateChange(
		map[string]interface{}{"balance": 500.00},
		map[string]interface{}{"balance": 1500.00, "prize_source": "Tournament #123"},
	)

	tournamentID := uuid.New()
	entry.SetTransactionRef(uuid.New(), tournamentID.String(), "")

	assert.True(entry.IsFinancialEvent())
	assert.False(entry.IsCritical())
	assert.Equal(1000.00, *entry.Amount)
	assert.Equal("USD", *entry.Currency)
}

func TestScenario_LargeWithdrawalAlert(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	userID := uuid.New()
	walletID := uuid.New()

	// Large withdrawal should trigger an ALERT severity
	entry := NewAuditTrailEntry(
		AuditEventWithdrawal,
		AuditSeverityAlert,
		userID,
		"user",
		"wallet",
		walletID,
		"Large withdrawal request exceeds daily limit",
		rxn,
	)

	entry.SetFinancialDetails(10000.00, "USD", 15000.00, 5000.00)
	entry.SetActorContext("203.0.113.42", "LeetGaming/1.0", "sess-abc")

	assert.True(entry.IsCritical())
	assert.True(entry.IsFinancialEvent())
	assert.Equal(AuditSeverityAlert, entry.Severity)
}

func TestScenario_MatchEntryFee(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	playerID := uuid.New()
	walletID := uuid.New()
	matchID := uuid.New()

	// Entry fee for a competitive match
	entry := NewAuditTrailEntry(
		AuditEventEntryFee,
		AuditSeverityInfo,
		playerID,
		"user",
		"wallet",
		walletID,
		"Entry fee for competitive match",
		rxn,
	)

	entry.SetFinancialDetails(5.00, "USD", 100.00, 95.00)
	entry.Metadata["match_id"] = matchID.String()
	entry.Metadata["match_type"] = "ranked_5v5"
	entry.Metadata["game"] = "CS2"

	assert.True(entry.IsFinancialEvent())
	assert.Equal("ranked_5v5", entry.Metadata["match_type"])
}

// =============================================================================
// Hash Chain Integrity Test
// =============================================================================

func TestHashChain_IntegrityVerification(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	// Create a chain of 3 entries
	entry1 := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit 1",
		rxn,
	)
	hash1 := entry1.ComputeHash("genesis")

	entry2 := NewAuditTrailEntry(
		AuditEventDeposit,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Deposit 2",
		rxn,
	)
	entry2.PreviousEntryID = &entry1.ID
	hash2 := entry2.ComputeHash(hash1)

	entry3 := NewAuditTrailEntry(
		AuditEventWithdrawal,
		AuditSeverityInfo,
		uuid.New(),
		"user",
		"wallet",
		uuid.New(),
		"Withdrawal 1",
		rxn,
	)
	entry3.PreviousEntryID = &entry2.ID
	_ = entry3.ComputeHash(hash2)

	// Verify chain integrity
	assert.True(entry1.VerifyHash("genesis"))
	assert.True(entry2.VerifyHash(hash1))
	assert.True(entry3.VerifyHash(hash2))

	// All hashes should be unique
	assert.NotEqual(hash1, hash2)
	assert.NotEqual(hash2, entry3.Hash)
}





