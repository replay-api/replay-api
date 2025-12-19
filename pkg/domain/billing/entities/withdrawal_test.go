package billing_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// WithdrawalStatus Constants Tests
// =============================================================================

func TestWithdrawalStatus_ValuesExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(WithdrawalStatus("pending"), WithdrawalStatusPending)
	assert.Equal(WithdrawalStatus("reviewing"), WithdrawalStatusReviewing)
	assert.Equal(WithdrawalStatus("approved"), WithdrawalStatusApproved)
	assert.Equal(WithdrawalStatus("processing"), WithdrawalStatusProcessing)
	assert.Equal(WithdrawalStatus("completed"), WithdrawalStatusCompleted)
	assert.Equal(WithdrawalStatus("rejected"), WithdrawalStatusRejected)
	assert.Equal(WithdrawalStatus("failed"), WithdrawalStatusFailed)
	assert.Equal(WithdrawalStatus("canceled"), WithdrawalStatusCanceled)
}

// =============================================================================
// WithdrawalMethod Constants Tests
// =============================================================================

func TestWithdrawalMethod_ValuesExist(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(WithdrawalMethod("bank_transfer"), WithdrawalMethodBankTransfer)
	assert.Equal(WithdrawalMethod("pix"), WithdrawalMethodPIX)
	assert.Equal(WithdrawalMethod("paypal"), WithdrawalMethodPayPal)
	assert.Equal(WithdrawalMethod("crypto"), WithdrawalMethodCrypto)
	assert.Equal(WithdrawalMethod("wire"), WithdrawalMethodWire)
}

// =============================================================================
// NewWithdrawal Tests
// =============================================================================

func TestNewWithdrawal_CreatesValidWithdrawal(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	userID := uuid.New()
	walletID := uuid.New()

	bankDetails := BankDetails{
		PIXKey:     "test@example.com",
		PIXKeyType: "email",
	}

	withdrawal := NewWithdrawal(
		userID,
		walletID,
		100.00,
		"BRL",
		WithdrawalMethodPIX,
		bankDetails,
		2.50, // fee
		rxn,
	)

	assert.NotNil(withdrawal)
	assert.NotEqual(uuid.Nil, withdrawal.ID)
	assert.Equal(userID, withdrawal.UserID)
	assert.Equal(walletID, withdrawal.WalletID)
	assert.Equal(100.00, withdrawal.Amount)
	assert.Equal("BRL", withdrawal.Currency)
	assert.Equal(WithdrawalMethodPIX, withdrawal.Method)
	assert.Equal(WithdrawalStatusPending, withdrawal.Status)
	assert.Equal(2.50, withdrawal.Fee)
	assert.Equal(97.50, withdrawal.NetAmount) // amount - fee
	assert.Equal(rxn, withdrawal.ResourceOwner)
}

func TestNewWithdrawal_CalculatesNetAmountCorrectly(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	testCases := []struct {
		amount    float64
		fee       float64
		netAmount float64
	}{
		{100.00, 5.00, 95.00},
		{500.00, 10.00, 490.00},
		{1000.00, 0.00, 1000.00},
		{50.00, 2.50, 47.50},
	}

	for _, tc := range testCases {
		withdrawal := NewWithdrawal(
			uuid.New(),
			uuid.New(),
			tc.amount,
			"USD",
			WithdrawalMethodBankTransfer,
			BankDetails{AccountNumber: "123456"},
			tc.fee,
			rxn,
		)
		assert.Equal(tc.netAmount, withdrawal.NetAmount)
	}
}

func TestNewWithdrawal_InitializesWithHistory(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodPayPal,
		BankDetails{PayPalEmail: "user@paypal.com"},
		0.00,
		rxn,
	)

	assert.Len(withdrawal.History, 1)
	assert.Equal(WithdrawalStatusPending, withdrawal.History[0].Status)
	assert.Equal("Withdrawal request created", withdrawal.History[0].Reason)
	assert.Nil(withdrawal.History[0].UpdatedBy)
}

// =============================================================================
// GetID Tests
// =============================================================================

func TestWithdrawal_GetID_ReturnsCorrectID(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	assert.Equal(withdrawal.ID, withdrawal.GetID())
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestValidate_SucceedsWithValidData(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"BRL",
		WithdrawalMethodPIX,
		BankDetails{PIXKey: "12345678900"},
		2.00,
		rxn,
	)

	err := withdrawal.Validate()
	assert.NoError(err)
}

func TestValidate_FailsWithNilUserID(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.Nil,
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "user_id is required")
}

func TestValidate_FailsWithZeroAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		0,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "amount must be positive")
}

func TestValidate_FailsWithNegativeAmount(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		-50.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "amount must be positive")
}

func TestValidate_FailsWithEmptyCurrency(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "currency is required")
}

func TestValidate_FailsWithEmptyMethod(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		"",
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "withdrawal method is required")
}

// =============================================================================
// ValidateBankDetails Tests (by method)
// =============================================================================

func TestValidateBankDetails_PIX_FailsWithoutPIXKey(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"BRL",
		WithdrawalMethodPIX,
		BankDetails{}, // No PIX key
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "PIX key is required")
}

func TestValidateBankDetails_PayPal_FailsWithoutEmail(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodPayPal,
		BankDetails{}, // No PayPal email
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "PayPal email is required")
}

func TestValidateBankDetails_Crypto_FailsWithoutAddress(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodCrypto,
		BankDetails{}, // No crypto address
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "crypto address is required")
}

func TestValidateBankDetails_BankTransfer_FailsWithoutAccountNumber(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{}, // No account number
		0,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "account number is required")
}

func TestValidateBankDetails_Wire_FailsWithoutAccountNumber(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		1000.00,
		"USD",
		WithdrawalMethodWire,
		BankDetails{}, // No account number
		25.00,
		rxn,
	)

	err := withdrawal.Validate()
	assert.Error(err)
	assert.Contains(err.Error(), "account number is required")
}

// =============================================================================
// Approve Tests
// =============================================================================

func TestApprove_SetsApprovedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	reviewerID := uuid.New()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Approve(reviewerID)

	assert.Equal(WithdrawalStatusApproved, withdrawal.Status)
	assert.NotNil(withdrawal.ReviewedBy)
	assert.Equal(reviewerID, *withdrawal.ReviewedBy)
	assert.NotNil(withdrawal.ReviewedAt)
}

func TestApprove_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	reviewerID := uuid.New()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Approve(reviewerID)

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusApproved, withdrawal.History[1].Status)
	assert.Equal("Withdrawal approved", withdrawal.History[1].Reason)
	assert.NotNil(withdrawal.History[1].UpdatedBy)
	assert.Equal(reviewerID, *withdrawal.History[1].UpdatedBy)
}

// =============================================================================
// Reject Tests
// =============================================================================

func TestReject_SetsRejectedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	reviewerID := uuid.New()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Reject(reviewerID, "Suspicious activity detected")

	assert.Equal(WithdrawalStatusRejected, withdrawal.Status)
	assert.Equal("Suspicious activity detected", withdrawal.RejectionReason)
	assert.NotNil(withdrawal.ReviewedBy)
	assert.Equal(reviewerID, *withdrawal.ReviewedBy)
	assert.NotNil(withdrawal.ReviewedAt)
}

func TestReject_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	reviewerID := uuid.New()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Reject(reviewerID, "Insufficient KYC")

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusRejected, withdrawal.History[1].Status)
	assert.Equal("Insufficient KYC", withdrawal.History[1].Reason)
}

// =============================================================================
// MarkProcessing Tests
// =============================================================================

func TestMarkProcessing_SetsProcessingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.Approve(uuid.New())

	withdrawal.MarkProcessing()

	assert.Equal(WithdrawalStatusProcessing, withdrawal.Status)
	assert.NotNil(withdrawal.ProcessedAt)
}

func TestMarkProcessing_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.MarkProcessing()

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusProcessing, withdrawal.History[1].Status)
	assert.Equal("Payment processing started", withdrawal.History[1].Reason)
}

// =============================================================================
// Complete Tests
// =============================================================================

func TestComplete_SetsCompletedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Complete("PROV-REF-12345")

	assert.Equal(WithdrawalStatusCompleted, withdrawal.Status)
	assert.Equal("PROV-REF-12345", withdrawal.ProviderReference)
	assert.NotNil(withdrawal.CompletedAt)
}

func TestComplete_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Complete("REF")

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusCompleted, withdrawal.History[1].Status)
	assert.Equal("Withdrawal completed", withdrawal.History[1].Reason)
}

// =============================================================================
// Fail Tests
// =============================================================================

func TestFail_SetsFailedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Fail("Bank rejected transaction")

	assert.Equal(WithdrawalStatusFailed, withdrawal.Status)
	assert.Equal("Bank rejected transaction", withdrawal.RejectionReason)
}

func TestFail_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Fail("Timeout")

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusFailed, withdrawal.History[1].Status)
	assert.Equal("Timeout", withdrawal.History[1].Reason)
}

// =============================================================================
// Cancel Tests
// =============================================================================

func TestCancel_SetsCanceledStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Cancel()

	assert.Equal(WithdrawalStatusCanceled, withdrawal.Status)
}

func TestCancel_AddsHistoryEntry(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	withdrawal.Cancel()

	assert.Len(withdrawal.History, 2)
	assert.Equal(WithdrawalStatusCanceled, withdrawal.History[1].Status)
	assert.Equal("Canceled by user", withdrawal.History[1].Reason)
}

// =============================================================================
// IsCancelable Tests
// =============================================================================

func TestIsCancelable_ReturnsTrueForPendingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	assert.True(withdrawal.IsCancelable())
}

func TestIsCancelable_ReturnsTrueForReviewingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.Status = WithdrawalStatusReviewing

	assert.True(withdrawal.IsCancelable())
}

func TestIsCancelable_ReturnsFalseForApprovedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.Approve(uuid.New())

	assert.False(withdrawal.IsCancelable())
}

func TestIsCancelable_ReturnsFalseForProcessingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.MarkProcessing()

	assert.False(withdrawal.IsCancelable())
}

func TestIsCancelable_ReturnsFalseForCompletedStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.Complete("ref")

	assert.False(withdrawal.IsCancelable())
}

// =============================================================================
// IsPending Tests
// =============================================================================

func TestIsPending_ReturnsTrueForPendingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)

	assert.True(withdrawal.IsPending())
}

func TestIsPending_ReturnsFalseForNonPendingStatus(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{AccountNumber: "123"},
		0,
		rxn,
	)
	withdrawal.Approve(uuid.New())

	assert.False(withdrawal.IsPending())
}

// =============================================================================
// BankDetails Struct Tests
// =============================================================================

func TestBankDetails_AllFieldsAccessible(t *testing.T) {
	assert := assert.New(t)

	details := BankDetails{
		AccountHolder: "John Doe",
		BankName:      "Test Bank",
		BankCode:      "001",
		AccountNumber: "123456789",
		RoutingNumber: "987654321",
		IBAN:          "BR12345678901234567890123456789012",
		SWIFT:         "TESTBRSP",
		PIXKey:        "john@example.com",
		PIXKeyType:    "email",
		PayPalEmail:   "john@paypal.com",
		CryptoAddress: "0x1234567890abcdef",
		CryptoNetwork: "ethereum",
	}

	assert.Equal("John Doe", details.AccountHolder)
	assert.Equal("Test Bank", details.BankName)
	assert.Equal("001", details.BankCode)
	assert.Equal("123456789", details.AccountNumber)
	assert.Equal("987654321", details.RoutingNumber)
	assert.Equal("BR12345678901234567890123456789012", details.IBAN)
	assert.Equal("TESTBRSP", details.SWIFT)
	assert.Equal("john@example.com", details.PIXKey)
	assert.Equal("email", details.PIXKeyType)
	assert.Equal("john@paypal.com", details.PayPalEmail)
	assert.Equal("0x1234567890abcdef", details.CryptoAddress)
	assert.Equal("ethereum", details.CryptoNetwork)
}

// =============================================================================
// WithdrawalHistory Struct Tests
// =============================================================================

func TestWithdrawalHistory_StructFields(t *testing.T) {
	assert := assert.New(t)
	updatedBy := uuid.New()

	history := WithdrawalHistory{
		Status:    WithdrawalStatusApproved,
		Reason:    "Approved by admin",
		UpdatedBy: &updatedBy,
		Timestamp: time.Now(),
	}

	assert.Equal(WithdrawalStatusApproved, history.Status)
	assert.Equal("Approved by admin", history.Reason)
	assert.Equal(&updatedBy, history.UpdatedBy)
	assert.WithinDuration(time.Now(), history.Timestamp, time.Second)
}

// =============================================================================
// Business Scenario Tests - E-Sports Platform Specific
// =============================================================================

func TestScenario_TournamentWinnerWithdrawal(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	playerID := uuid.New()
	walletID := uuid.New()
	adminID := uuid.New()

	// Tournament winner withdraws prize money via PIX (Brazil)
	withdrawal := NewWithdrawal(
		playerID,
		walletID,
		5000.00, // Prize money
		"BRL",
		WithdrawalMethodPIX,
		BankDetails{
			PIXKey:     "12345678900", // CPF
			PIXKeyType: "cpf",
		},
		25.00, // Platform fee
		rxn,
	)

	assert.NoError(withdrawal.Validate())
	assert.Equal(4975.00, withdrawal.NetAmount)
	assert.True(withdrawal.IsPending())

	// Admin reviews and approves
	withdrawal.Approve(adminID)
	assert.Equal(WithdrawalStatusApproved, withdrawal.Status)
	assert.False(withdrawal.IsCancelable())

	// Process the withdrawal
	withdrawal.MarkProcessing()
	assert.Equal(WithdrawalStatusProcessing, withdrawal.Status)

	// Complete with provider reference
	withdrawal.Complete("PIX-TXN-2024-12345")
	assert.Equal(WithdrawalStatusCompleted, withdrawal.Status)
	assert.Equal("PIX-TXN-2024-12345", withdrawal.ProviderReference)
	assert.Len(withdrawal.History, 4) // pending -> approved -> processing -> completed
}

func TestScenario_HighValueWithdrawalRejection(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	complianceOfficerID := uuid.New()

	// High-value withdrawal requiring additional KYC
	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		10000.00,
		"USD",
		WithdrawalMethodWire,
		BankDetails{
			AccountHolder: "Pro Gamer Inc",
			BankName:      "Chase Bank",
			AccountNumber: "123456789",
			RoutingNumber: "021000021",
			SWIFT:         "CHASUS33",
		},
		50.00,
		rxn,
	)

	assert.NoError(withdrawal.Validate())

	// Compliance officer rejects due to incomplete KYC
	withdrawal.Reject(complianceOfficerID, "Incomplete KYC documentation for high-value transaction")

	assert.Equal(WithdrawalStatusRejected, withdrawal.Status)
	assert.Equal("Incomplete KYC documentation for high-value transaction", withdrawal.RejectionReason)
	assert.False(withdrawal.IsCancelable())
}

func TestScenario_CryptoWithdrawalFailed(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	// Crypto withdrawal that fails at the network level
	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		500.00,
		"USDT",
		WithdrawalMethodCrypto,
		BankDetails{
			CryptoAddress: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			CryptoNetwork: "ethereum",
		},
		5.00, // Gas fee
		rxn,
	)

	assert.NoError(withdrawal.Validate())

	withdrawal.Approve(uuid.New())
	withdrawal.MarkProcessing()

	// Transaction fails on-chain
	withdrawal.Fail("Transaction reverted: insufficient gas")

	assert.Equal(WithdrawalStatusFailed, withdrawal.Status)
	assert.Contains(withdrawal.RejectionReason, "gas")
}

func TestScenario_UserCancelsBeforeProcessing(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()

	// User initiates withdrawal but changes mind
	withdrawal := NewWithdrawal(
		uuid.New(),
		uuid.New(),
		100.00,
		"EUR",
		WithdrawalMethodPayPal,
		BankDetails{PayPalEmail: "player@email.com"},
		2.00,
		rxn,
	)

	assert.True(withdrawal.IsCancelable())

	withdrawal.Cancel()

	assert.Equal(WithdrawalStatusCanceled, withdrawal.Status)
	assert.False(withdrawal.IsCancelable())
	assert.False(withdrawal.IsPending())
}

// =============================================================================
// Full Lifecycle Test
// =============================================================================

func TestWithdrawal_FullLifecycle(t *testing.T) {
	assert := assert.New(t)
	rxn := testBillingResourceOwner()
	userID := uuid.New()
	walletID := uuid.New()
	adminID := uuid.New()

	// 1. Create withdrawal
	withdrawal := NewWithdrawal(
		userID,
		walletID,
		250.00,
		"USD",
		WithdrawalMethodBankTransfer,
		BankDetails{
			AccountHolder: "Test User",
			BankName:      "Bank of America",
			AccountNumber: "1234567890",
			RoutingNumber: "026009593",
		},
		5.00,
		rxn,
	)

	assert.Equal(WithdrawalStatusPending, withdrawal.Status)
	assert.Equal(245.00, withdrawal.NetAmount)
	assert.Len(withdrawal.History, 1)

	// 2. Validate
	err := withdrawal.Validate()
	assert.NoError(err)

	// 3. Approve
	withdrawal.Approve(adminID)
	assert.Equal(WithdrawalStatusApproved, withdrawal.Status)
	assert.Len(withdrawal.History, 2)

	// 4. Start processing
	withdrawal.MarkProcessing()
	assert.Equal(WithdrawalStatusProcessing, withdrawal.Status)
	assert.NotNil(withdrawal.ProcessedAt)
	assert.Len(withdrawal.History, 3)

	// 5. Complete
	withdrawal.Complete("ACH-2024-001234")
	assert.Equal(WithdrawalStatusCompleted, withdrawal.Status)
	assert.Equal("ACH-2024-001234", withdrawal.ProviderReference)
	assert.NotNil(withdrawal.CompletedAt)
	assert.Len(withdrawal.History, 4)

	// Verify timestamps are properly set
	assert.True(withdrawal.ReviewedAt.Before(time.Now()))
	assert.True(withdrawal.ProcessedAt.Before(time.Now()))
	assert.True(withdrawal.CompletedAt.Before(time.Now()))
}

