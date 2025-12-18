package custody_entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// Helper to create a test resource owner
func testResourceOwner() common.ResourceOwner {
	return common.ResourceOwner{
		TenantID: uuid.New(),
		ClientID: uuid.New(),
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
	}
}

// TestWalletType_Constants verifies wallet type constants
func TestWalletType_Constants(t *testing.T) {
	types := []WalletType{
		WalletTypePersonal,
		WalletTypeBusiness,
		WalletTypeOperations,
		WalletTypeTreasury,
		WalletTypeEscrow,
	}

	seen := make(map[WalletType]bool)
	for _, wt := range types {
		if wt == "" {
			t.Error("WalletType should not be empty")
		}
		if seen[wt] {
			t.Errorf("Duplicate WalletType: %s", wt)
		}
		seen[wt] = true
	}
}

// TestWalletStatus_Constants verifies wallet status constants
func TestWalletStatus_Constants(t *testing.T) {
	statuses := []WalletStatus{
		WalletStatusPending,
		WalletStatusActive,
		WalletStatusSuspended,
		WalletStatusRecovering,
		WalletStatusArchived,
	}

	seen := make(map[WalletStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("WalletStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate WalletStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestSecurityLevel_Constants verifies security level constants
func TestSecurityLevel_Constants(t *testing.T) {
	levels := []SecurityLevel{
		SecurityLevelBasic,
		SecurityLevelStandard,
		SecurityLevelHigh,
		SecurityLevelCritical,
	}

	seen := make(map[SecurityLevel]bool)
	for _, level := range levels {
		if level == "" {
			t.Error("SecurityLevel should not be empty")
		}
		if seen[level] {
			t.Errorf("Duplicate SecurityLevel: %s", level)
		}
		seen[level] = true
	}
}

// TestKYCStatus_Constants verifies KYC status constants
func TestKYCStatus_Constants(t *testing.T) {
	statuses := []KYCStatus{
		KYCStatusNone,
		KYCStatusPending,
		KYCStatusBasic,
		KYCStatusVerified,
		KYCStatusEnhanced,
	}

	seen := make(map[KYCStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("KYCStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate KYCStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestKeyShareStatus_Constants verifies key share status constants
func TestKeyShareStatus_Constants(t *testing.T) {
	statuses := []KeyShareStatus{
		KeyShareStatusActive,
		KeyShareStatusInactive,
		KeyShareStatusRotating,
		KeyShareStatusCompromised,
	}

	seen := make(map[KeyShareStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("KeyShareStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate KeyShareStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestGuardianType_Constants verifies guardian type constants
func TestGuardianType_Constants(t *testing.T) {
	types := []GuardianType{
		GuardianTypeEmail,
		GuardianTypePhone,
		GuardianTypeWallet,
		GuardianTypeHardware,
		GuardianTypeInstitution,
	}

	seen := make(map[GuardianType]bool)
	for _, gt := range types {
		if gt == "" {
			t.Error("GuardianType should not be empty")
		}
		if seen[gt] {
			t.Errorf("Duplicate GuardianType: %s", gt)
		}
		seen[gt] = true
	}
}

// TestGuardianStatus_Constants verifies guardian status constants
func TestGuardianStatus_Constants(t *testing.T) {
	statuses := []GuardianStatus{
		GuardianStatusPending,
		GuardianStatusActive,
		GuardianStatusInactive,
		GuardianStatusRemoved,
	}

	seen := make(map[GuardianStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("GuardianStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate GuardianStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestRecoveryStatus_Constants verifies recovery status constants
func TestRecoveryStatus_Constants(t *testing.T) {
	statuses := []RecoveryStatus{
		RecoveryStatusPending,
		RecoveryStatusApproved,
		RecoveryStatusExecuted,
		RecoveryStatusCancelled,
		RecoveryStatusExpired,
	}

	seen := make(map[RecoveryStatus]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("RecoveryStatus should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate RecoveryStatus: %s", status)
		}
		seen[status] = true
	}
}

// TestRiskLevel_Constants verifies risk level constants
func TestRiskLevel_Constants(t *testing.T) {
	levels := []RiskLevel{
		RiskLevelLow,
		RiskLevelMedium,
		RiskLevelHigh,
		RiskLevelCritical,
	}

	seen := make(map[RiskLevel]bool)
	for _, level := range levels {
		if level == "" {
			t.Error("RiskLevel should not be empty")
		}
		if seen[level] {
			t.Errorf("Duplicate RiskLevel: %s", level)
		}
		seen[level] = true
	}
}

// TestNewSmartWallet verifies wallet creation
func TestNewSmartWallet(t *testing.T) {
	owner := testResourceOwner()
	ownerID := uuid.New()
	walletName := "Test Wallet"
	walletType := WalletTypePersonal
	primaryChain := custody_vo.ChainSolanaMainnet

	wallet := NewSmartWallet(owner, ownerID, walletName, walletType, primaryChain)

	if wallet == nil {
		t.Fatal("Expected non-nil wallet")
	}
	if wallet.OwnerID != ownerID {
		t.Error("OwnerID mismatch")
	}
	if wallet.WalletName != walletName {
		t.Errorf("WalletName = %s, want %s", wallet.WalletName, walletName)
	}
	if wallet.WalletType != walletType {
		t.Errorf("WalletType = %s, want %s", wallet.WalletType, walletType)
	}
	if wallet.PrimaryChain != primaryChain {
		t.Errorf("PrimaryChain = %s, want %s", wallet.PrimaryChain, primaryChain)
	}
	if wallet.Status != WalletStatusPending {
		t.Errorf("Status = %s, want %s", wallet.Status, WalletStatusPending)
	}
	if wallet.SecurityLevel != SecurityLevelStandard {
		t.Errorf("SecurityLevel = %s, want %s", wallet.SecurityLevel, SecurityLevelStandard)
	}
	if wallet.KYCStatus != KYCStatusNone {
		t.Errorf("KYCStatus = %s, want %s", wallet.KYCStatus, KYCStatusNone)
	}
}

// TestNewSmartWallet_DefaultLimits verifies default transaction limits
func TestNewSmartWallet_DefaultLimits(t *testing.T) {
	owner := testResourceOwner()
	wallet := NewSmartWallet(owner, uuid.New(), "Test", WalletTypePersonal, custody_vo.ChainSolanaMainnet)

	if wallet.Limits.DailyLimit != 10000_00 {
		t.Errorf("DailyLimit = %d, want 1000000", wallet.Limits.DailyLimit)
	}
	if wallet.Limits.WeeklyLimit != 50000_00 {
		t.Errorf("WeeklyLimit = %d, want 5000000", wallet.Limits.WeeklyLimit)
	}
	if wallet.Limits.MonthlyLimit != 100000_00 {
		t.Errorf("MonthlyLimit = %d, want 10000000", wallet.Limits.MonthlyLimit)
	}
	if wallet.Limits.SingleTxLimit != 5000_00 {
		t.Errorf("SingleTxLimit = %d, want 500000", wallet.Limits.SingleTxLimit)
	}
}

// TestSmartWallet_SetMPCKeyConfig verifies MPC key configuration
func TestSmartWallet_SetMPCKeyConfig(t *testing.T) {
	wallet := createTestWallet()

	keyID := "mpc-key-001"
	config := MPCKeyConfiguration{
		Scheme:         custody_vo.MPCSchemeCMP,
		Curve:          custody_vo.CurveSecp256k1,
		Threshold:      custody_vo.Threshold2of3,
		KeyGeneratedAt: time.Now(),
	}
	shares := []KeyShareInfo{
		{ShareID: custody_vo.NewKeyShareID(keyID, 0), ShareIndex: 0, Location: custody_vo.LocationHSM, Status: KeyShareStatusActive},
		{ShareID: custody_vo.NewKeyShareID(keyID, 1), ShareIndex: 1, Location: custody_vo.LocationSecureEnclave, Status: KeyShareStatusActive},
		{ShareID: custody_vo.NewKeyShareID(keyID, 2), ShareIndex: 2, Location: custody_vo.LocationUserDevice, Status: KeyShareStatusActive},
	}

	wallet.SetMPCKeyConfig(keyID, config, shares)

	if wallet.MasterKeyID != keyID {
		t.Errorf("MasterKeyID = %s, want %s", wallet.MasterKeyID, keyID)
	}
	if len(wallet.KeyShares) != 3 {
		t.Errorf("KeyShares count = %d, want 3", len(wallet.KeyShares))
	}
}

// TestSmartWallet_AddChainAddress verifies chain address addition
func TestSmartWallet_AddChainAddress(t *testing.T) {
	wallet := createTestWallet()
	chainID := custody_vo.ChainPolygon
	address := "0x1234567890123456789012345678901234567890"

	wallet.AddChainAddress(chainID, address)

	if len(wallet.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(wallet.Addresses))
	}
	if _, ok := wallet.Addresses[chainID]; !ok {
		t.Error("Chain address not found")
	}
}

// TestSmartWallet_Activate verifies wallet activation
func TestSmartWallet_Activate(t *testing.T) {
	wallet := createTestWallet()
	setupWalletForActivation(wallet)

	err := _ = wallet.Activate()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if wallet.Status != WalletStatusActive {
		t.Errorf("Status = %s, want %s", wallet.Status, WalletStatusActive)
	}
	if wallet.ActivatedAt == nil {
		t.Error("ActivatedAt should be set")
	}
}

// TestSmartWallet_Activate_NoKey verifies activation fails without key
func TestSmartWallet_Activate_NoKey(t *testing.T) {
	wallet := createTestWallet()
	// Add address but no MPC key
	wallet.AddChainAddress(custody_vo.ChainSolanaMainnet, "test-address")

	err := _ = wallet.Activate()

	if err == nil {
		t.Error("Expected error for activation without MPC key")
	}
}

// TestSmartWallet_Activate_NoAddress verifies activation fails without address
func TestSmartWallet_Activate_NoAddress(t *testing.T) {
	wallet := createTestWallet()
	wallet.MasterKeyID = "test-key"

	err := _ = wallet.Activate()

	if err == nil {
		t.Error("Expected error for activation without address")
	}
}

// TestSmartWallet_Suspend verifies wallet suspension
func TestSmartWallet_Suspend(t *testing.T) {
	wallet := createTestWallet()
	setupWalletForActivation(wallet)
	_ = wallet.Activate()

	wallet.Suspend("Suspicious activity")

	if wallet.Status != WalletStatusSuspended {
		t.Errorf("Status = %s, want %s", wallet.Status, WalletStatusSuspended)
	}
	if wallet.SuspendedAt == nil {
		t.Error("SuspendedAt should be set")
	}
	if wallet.SuspendReason != "Suspicious activity" {
		t.Errorf("SuspendReason = %s, want 'Suspicious activity'", wallet.SuspendReason)
	}
}

// TestSmartWallet_GetAddress verifies address retrieval
func TestSmartWallet_GetAddress(t *testing.T) {
	wallet := createTestWallet()
	chainID := custody_vo.ChainPolygon
	expectedAddr := "0x1234567890123456789012345678901234567890"
	wallet.AddChainAddress(chainID, expectedAddr)

	addr, err := wallet.GetAddress(chainID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if addr != expectedAddr {
		t.Errorf("GetAddress = %s, want %s", addr, expectedAddr)
	}
}

// TestSmartWallet_GetAddress_MultipleChains verifies addresses for multiple chains
func TestSmartWallet_GetAddress_MultipleChains(t *testing.T) {
	wallet := createTestWallet()
	polygonAddr := "0xpolygon1234567890123456789012345678901234"
	baseAddr := "0xbase567890123456789012345678901234567890"
	wallet.AddChainAddress(custody_vo.ChainPolygon, polygonAddr)
	wallet.AddChainAddress(custody_vo.ChainBase, baseAddr)

	addr1, err1 := wallet.GetAddress(custody_vo.ChainPolygon)
	addr2, err2 := wallet.GetAddress(custody_vo.ChainBase)

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected error: %v, %v", err1, err2)
	}
	if addr1 != polygonAddr {
		t.Errorf("GetAddress(Polygon) = %s, want %s", addr1, polygonAddr)
	}
	if addr2 != baseAddr {
		t.Errorf("GetAddress(Base) = %s, want %s", addr2, baseAddr)
	}
}

// TestSmartWallet_GetAddress_NotFound verifies error for missing chain
func TestSmartWallet_GetAddress_NotFound(t *testing.T) {
	wallet := createTestWallet()

	_, err := wallet.GetAddress(custody_vo.ChainPolygon)

	if err == nil {
		t.Error("Expected error for missing chain address")
	}
}

// TestSmartWallet_CanSpend verifies spending limit checks
func TestSmartWallet_CanSpend(t *testing.T) {
	wallet := createTestWallet()
	setupWalletForActivation(wallet)
	_ = wallet.Activate()

	tests := []struct {
		name      string
		amount    uint64
		expectErr bool
	}{
		{"Within limit", 1000_00, false},
		{"At single tx limit", 5000_00, false},
		{"Exceeds single tx limit", 5001_00, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wallet.CanSpend(tt.amount)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestSmartWallet_CanSpend_Inactive verifies spending blocked for inactive wallet
func TestSmartWallet_CanSpend_Inactive(t *testing.T) {
	wallet := createTestWallet()
	// Wallet is in Pending status

	err := wallet.CanSpend(100)

	if err == nil {
		t.Error("Expected error for inactive wallet")
	}
}

// TestSmartWallet_RecordSpend verifies spending tracking
func TestSmartWallet_RecordSpend(t *testing.T) {
	wallet := createTestWallet()

	wallet.RecordSpend(1000)

	if wallet.Limits.DailyUsed != 1000 {
		t.Errorf("DailyUsed = %d, want 1000", wallet.Limits.DailyUsed)
	}
	if wallet.Limits.WeeklyUsed != 1000 {
		t.Errorf("WeeklyUsed = %d, want 1000", wallet.Limits.WeeklyUsed)
	}
	if wallet.Limits.MonthlyUsed != 1000 {
		t.Errorf("MonthlyUsed = %d, want 1000", wallet.Limits.MonthlyUsed)
	}
	if wallet.LastActivityAt == nil {
		t.Error("LastActivityAt should be set")
	}
}

// TestSmartWallet_InitiateRecovery verifies recovery initiation
func TestSmartWallet_InitiateRecovery(t *testing.T) {
	wallet := createWalletWithRecovery()

	initiatorID := uuid.New()
	newOwnerKey := []byte("new-owner-public-key-bytes")
	err := _ = wallet.InitiateRecovery(initiatorID, newOwnerKey)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if wallet.Status != WalletStatusRecovering {
		t.Errorf("Status = %s, want %s", wallet.Status, WalletStatusRecovering)
	}
	if wallet.PendingRecovery == nil {
		t.Error("PendingRecovery should be set")
	}
	if wallet.PendingRecovery.Status != RecoveryStatusPending {
		t.Errorf("Recovery status = %s, want %s", wallet.PendingRecovery.Status, RecoveryStatusPending)
	}
	if !wallet.IsFrozen {
		t.Error("Wallet should be frozen during recovery")
	}
}

// TestSmartWallet_InitiateRecovery_NotEnabled verifies error when recovery disabled
func TestSmartWallet_InitiateRecovery_NotEnabled(t *testing.T) {
	wallet := createTestWallet()
	wallet.RecoveryConfig.IsEnabled = false

	err := _ = wallet.InitiateRecovery(uuid.New(), []byte("new-key"))

	if err == nil {
		t.Error("Expected error for recovery not enabled")
	}
}

// TestSmartWallet_InitiateRecovery_AlreadyInProgress verifies duplicate recovery rejection
func TestSmartWallet_InitiateRecovery_AlreadyInProgress(t *testing.T) {
	wallet := createWalletWithRecovery()
	_ = wallet.InitiateRecovery(uuid.New(), []byte("key1"))

	err := _ = wallet.InitiateRecovery(uuid.New(), []byte("key2"))

	if err == nil {
		t.Error("Expected error for recovery already in progress")
	}
}

// TestSmartWallet_ApproveRecovery verifies guardian approval
func TestSmartWallet_ApproveRecovery(t *testing.T) {
	wallet := createWalletWithRecovery()
	_ = wallet.InitiateRecovery(uuid.New(), []byte("new-key"))
	guardianID := uuid.New()

	err := wallet.ApproveRecovery(guardianID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// InitiateRecovery already adds initiator as first approver, so we expect 2
	if wallet.PendingRecovery.ApprovalCount != 2 {
		t.Errorf("ApprovalCount = %d, want 2", wallet.PendingRecovery.ApprovalCount)
	}
}

// TestSmartWallet_ApproveRecovery_ThresholdMet verifies approval status change
func TestSmartWallet_ApproveRecovery_ThresholdMet(t *testing.T) {
	wallet := createWalletWithRecovery()
	_ = _ = wallet.InitiateRecovery(uuid.New(), []byte("new-key"))

	// Approve by another guardian (threshold is 2, initiator already counts as 1)
	_ = wallet.ApproveRecovery(uuid.New())

	if wallet.PendingRecovery.Status != RecoveryStatusApproved {
		t.Errorf("Recovery status = %s, want %s", wallet.PendingRecovery.Status, RecoveryStatusApproved)
	}
}

// TestSmartWallet_ApproveRecovery_DuplicateApproval verifies duplicate approval rejection
func TestSmartWallet_ApproveRecovery_DuplicateApproval(t *testing.T) {
	wallet := createWalletWithRecovery()
	initiatorID := uuid.New()
	_ = wallet.InitiateRecovery(initiatorID, []byte("new-key"))

	// Try to approve with the same ID that initiated
	err := wallet.ApproveRecovery(initiatorID)

	if err == nil {
		t.Error("Expected error for duplicate approval")
	}
}

// TestSmartWallet_ApproveRecovery_NoPending verifies error without pending recovery
func TestSmartWallet_ApproveRecovery_NoPending(t *testing.T) {
	wallet := createWalletWithRecovery()

	err := wallet.ApproveRecovery(uuid.New())

	if err == nil {
		t.Error("Expected error for no pending recovery")
	}
}

// Helper functions

func createTestWallet() *SmartWallet {
	owner := testResourceOwner()
	return NewSmartWallet(owner, uuid.New(), "Test Wallet", WalletTypePersonal, custody_vo.ChainSolanaMainnet)
}

func setupWalletForActivation(wallet *SmartWallet) {
	wallet.MasterKeyID = "test-key-001"
	wallet.AddChainAddress(custody_vo.ChainSolanaMainnet, "test-solana-address")
}

func createWalletWithRecovery() *SmartWallet {
	wallet := createTestWallet()
	setupWalletForActivation(wallet)
	_ = _ = wallet.Activate()

	wallet.RecoveryConfig = &WalletRecoveryConfig{
		IsEnabled:         true,
		GuardianThreshold: 2,
		RecoveryDelay:     24 * time.Hour,
	}

	return wallet
}
