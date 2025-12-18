package custody_entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// SmartWallet represents a custodial smart wallet with MPC key management
// This is the core entity for banking-grade wallet infrastructure
type SmartWallet struct {
	common.BaseEntity

	// Identity
	UserID        uuid.UUID  `json:"user_id" bson:"user_id"`               // Platform user ID
	OwnerID       uuid.UUID  `json:"owner_id" bson:"owner_id"`             // Alias for backward compatibility
	WalletName    string     `json:"wallet_name" bson:"wallet_name"`
	Label         string     `json:"label" bson:"label"`                   // Human-readable label
	WalletType    WalletType `json:"wallet_type" bson:"wallet_type"`
	PublicKey     string     `json:"public_key" bson:"public_key"`         // Master public key

	// Multi-chain addresses (derived from MPC master key)
	Addresses     map[custody_vo.ChainID]string `json:"addresses" bson:"addresses"`
	PrimaryChain  custody_vo.ChainID            `json:"primary_chain" bson:"primary_chain"`

	// MPC Key Management
	MasterKeyID   string                    `json:"master_key_id" bson:"master_key_id"`
	KeyConfig     MPCKeyConfiguration       `json:"key_config" bson:"key_config"`
	KeyShares     []KeyShareInfo            `json:"key_shares" bson:"key_shares"`

	// Account Abstraction (EVM chains)
	AAConfig      *AccountAbstractionConfig `json:"aa_config,omitempty" bson:"aa_config,omitempty"`

	// Social Recovery
	RecoveryConfig  *WalletRecoveryConfig `json:"recovery_config,omitempty" bson:"recovery_config,omitempty"`
	PendingRecovery *PendingRecovery      `json:"pending_recovery,omitempty" bson:"pending_recovery,omitempty"`
	IsFrozen        bool                  `json:"is_frozen" bson:"is_frozen"`

	// Security & Compliance
	SecurityLevel SecurityLevel             `json:"security_level" bson:"security_level"`
	RiskProfile   RiskProfile               `json:"risk_profile" bson:"risk_profile"`
	Limits        TransactionLimits         `json:"limits" bson:"limits"`
	KYCStatus     KYCStatus                 `json:"kyc_status" bson:"kyc_status"`

	// Status
	Status        WalletStatus              `json:"status" bson:"status"`
	ActivatedAt   *time.Time                `json:"activated_at,omitempty" bson:"activated_at,omitempty"`
	LastActivityAt *time.Time               `json:"last_activity_at,omitempty" bson:"last_activity_at,omitempty"`
	SuspendedAt   *time.Time                `json:"suspended_at,omitempty" bson:"suspended_at,omitempty"`
	SuspendReason string                    `json:"suspend_reason,omitempty" bson:"suspend_reason,omitempty"`

	// Metadata
	Tags          []string                  `json:"tags,omitempty" bson:"tags,omitempty"`
	Metadata      map[string]interface{}    `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// WalletType defines the wallet classification
type WalletType string

const (
	WalletTypePersonal   WalletType = "Personal"   // Individual user wallet
	WalletTypeBusiness   WalletType = "Business"   // Business/org wallet
	WalletTypeOperations WalletType = "Operations" // Platform operations (hot wallet)
	WalletTypeTreasury   WalletType = "Treasury"   // Cold storage treasury
	WalletTypeEscrow     WalletType = "Escrow"     // Prize pool escrow
)

// WalletStatus represents the wallet state
type WalletStatus string

const (
	WalletStatusPending    WalletStatus = "Pending"    // Awaiting key generation
	WalletStatusActive     WalletStatus = "Active"     // Fully operational
	WalletStatusSuspended  WalletStatus = "Suspended"  // Temporarily frozen
	WalletStatusRecovering WalletStatus = "Recovering" // In recovery process
	WalletStatusArchived   WalletStatus = "Archived"   // Deactivated
)

// SecurityLevel determines signing requirements
type SecurityLevel string

const (
	SecurityLevelBasic    SecurityLevel = "Basic"    // Single approval
	SecurityLevelStandard SecurityLevel = "Standard" // 2-of-3 MPC
	SecurityLevelHigh     SecurityLevel = "High"     // 3-of-5 MPC + time delay
	SecurityLevelCritical SecurityLevel = "Critical" // Multi-party + HSM + time delay
)

// KYCStatus represents know-your-customer verification status
type KYCStatus string

const (
	KYCStatusNone       KYCStatus = "None"
	KYCStatusPending    KYCStatus = "Pending"
	KYCStatusBasic      KYCStatus = "Basic"      // Email/Phone verified
	KYCStatusVerified   KYCStatus = "Verified"   // ID verified
	KYCStatusEnhanced   KYCStatus = "Enhanced"   // Full KYC + AML
)

// ChainAddressInfo holds detailed address info for a specific chain (used for extended metadata)
type ChainAddressInfo struct {
	ChainID         custody_vo.ChainID `json:"chain_id" bson:"chain_id"`
	Address         string             `json:"address" bson:"address"`
	PublicKey       []byte             `json:"public_key" bson:"public_key"`
	DerivationPath  string             `json:"derivation_path" bson:"derivation_path"`
	IsSmartContract bool               `json:"is_smart_contract" bson:"is_smart_contract"`
	ContractAddress *string            `json:"contract_address,omitempty" bson:"contract_address,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	LastUsedAt      *time.Time         `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
}

// MPCKeyConfiguration holds MPC key setup
type MPCKeyConfiguration struct {
	Scheme         custody_vo.MPCScheme         `json:"scheme" bson:"scheme"`
	Curve          custody_vo.KeyCurve          `json:"curve" bson:"curve"`
	Threshold      custody_vo.ThresholdConfig   `json:"threshold" bson:"threshold"`
	KeyGeneratedAt time.Time                    `json:"key_generated_at" bson:"key_generated_at"`
	LastRotatedAt  *time.Time                   `json:"last_rotated_at,omitempty" bson:"last_rotated_at,omitempty"`
	NextRotationAt *time.Time                   `json:"next_rotation_at,omitempty" bson:"next_rotation_at,omitempty"`
}

// KeyShareInfo describes a key share without exposing sensitive data
type KeyShareInfo struct {
	ShareID     custody_vo.KeyShareID       `json:"share_id" bson:"share_id"`
	ShareIndex  uint8                       `json:"share_index" bson:"share_index"`
	Location    custody_vo.KeyShareLocation `json:"location" bson:"location"`
	ProviderID  string                      `json:"provider_id" bson:"provider_id"`
	Status      KeyShareStatus              `json:"status" bson:"status"`
	LastHealthCheck *time.Time              `json:"last_health_check,omitempty" bson:"last_health_check,omitempty"`
}

type KeyShareStatus string

const (
	KeyShareStatusActive   KeyShareStatus = "Active"
	KeyShareStatusInactive KeyShareStatus = "Inactive"
	KeyShareStatusRotating KeyShareStatus = "Rotating"
	KeyShareStatusCompromised KeyShareStatus = "Compromised"
)

// AccountAbstractionConfig for EVM smart contract wallets (ERC-4337)
type AccountAbstractionConfig struct {
	IsDeployed      map[custody_vo.ChainID]bool `json:"is_deployed" bson:"is_deployed"` // Per-chain deployment status
	FactoryAddress  string                      `json:"factory_address" bson:"factory_address"`
	EntryPointAddr  string                      `json:"entry_point_address" bson:"entry_point_address"`
	ImplementationAddr string                   `json:"implementation_address" bson:"implementation_address"`
	Salt            []byte                      `json:"salt" bson:"salt"`
	InitCode        []byte                      `json:"init_code,omitempty" bson:"init_code,omitempty"`
	Nonce           uint64                      `json:"nonce" bson:"nonce"`
	DeployedAt      *time.Time                  `json:"deployed_at,omitempty" bson:"deployed_at,omitempty"`

	// Paymaster for gas sponsorship
	PaymasterEnabled bool   `json:"paymaster_enabled" bson:"paymaster_enabled"`
	PaymasterAddress string `json:"paymaster_address,omitempty" bson:"paymaster_address,omitempty"`

	// Session keys for delegated signing
	SessionKeys     []SessionKeyConfig `json:"session_keys,omitempty" bson:"session_keys,omitempty"`
}

// SessionKeyConfig for temporary delegated permissions
type SessionKeyConfig struct {
	KeyID         string    `json:"key_id" bson:"key_id"`
	PublicKey     []byte    `json:"public_key" bson:"public_key"`
	Permissions   []string  `json:"permissions" bson:"permissions"` // Contract addresses or methods
	SpendingLimit uint64    `json:"spending_limit" bson:"spending_limit"`
	ValidFrom     time.Time `json:"valid_from" bson:"valid_from"`
	ValidUntil    time.Time `json:"valid_until" bson:"valid_until"`
	UsageCount    uint64    `json:"usage_count" bson:"usage_count"`
	MaxUsage      uint64    `json:"max_usage" bson:"max_usage"`
}

// SessionKey represents an active session key with wallet association
type SessionKey struct {
	WalletID    uuid.UUID        `json:"wallet_id" bson:"wallet_id"`
	KeyAddress  string           `json:"key_address" bson:"key_address"`
	Label       string           `json:"label" bson:"label"`
	Config      SessionKeyConfig `json:"config" bson:"config"`
	IsActive    bool             `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time        `json:"created_at" bson:"created_at"`
	RevokedAt   *time.Time       `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
}

// WalletRecoveryConfig for social recovery
type WalletRecoveryConfig struct {
	IsEnabled         bool              `json:"is_enabled" bson:"is_enabled"`
	GuardianThreshold uint8             `json:"guardian_threshold" bson:"guardian_threshold"` // Guardians needed
	RecoveryDelay     time.Duration     `json:"recovery_delay" bson:"recovery_delay"`         // Time lock
	LastRecoveryAt    *time.Time        `json:"last_recovery_at,omitempty" bson:"last_recovery_at,omitempty"`
}

// RecoveryConfig alias for backward compatibility
type RecoveryConfig = WalletRecoveryConfig

// Guardian represents a social recovery guardian
type Guardian struct {
	ID           uuid.UUID      `json:"id" bson:"id"`
	WalletID     uuid.UUID      `json:"wallet_id" bson:"wallet_id"`
	GuardianType GuardianType   `json:"guardian_type" bson:"guardian_type"`
	Address      string         `json:"address" bson:"address"`         // Wallet address or derived address
	Email        string         `json:"email,omitempty" bson:"email,omitempty"`
	Phone        string         `json:"phone,omitempty" bson:"phone,omitempty"`
	Label        string         `json:"label,omitempty" bson:"label,omitempty"`
	Weight       uint8          `json:"weight" bson:"weight"` // Voting weight for recovery
	PublicKey    []byte         `json:"public_key,omitempty" bson:"public_key,omitempty"`
	IsActive     bool           `json:"is_active" bson:"is_active"`
	AddedAt      time.Time      `json:"added_at" bson:"added_at"`
	ConfirmedAt  *time.Time     `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	LastActiveAt *time.Time     `json:"last_active_at,omitempty" bson:"last_active_at,omitempty"`
	Status       GuardianStatus `json:"status" bson:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

type GuardianType string

const (
	GuardianTypeEmail      GuardianType = "Email"
	GuardianTypePhone      GuardianType = "Phone"
	GuardianTypeWallet     GuardianType = "Wallet"     // Another wallet address
	GuardianTypeHardware   GuardianType = "Hardware"   // Hardware key (YubiKey)
	GuardianTypeInstitution GuardianType = "Institution" // Trusted institution
)

type GuardianStatus string

const (
	GuardianStatusPending   GuardianStatus = "Pending"
	GuardianStatusActive    GuardianStatus = "Active"
	GuardianStatusInactive  GuardianStatus = "Inactive"
	GuardianStatusRemoved   GuardianStatus = "Removed"
)

// PendingRecovery represents an in-progress recovery
type PendingRecovery struct {
	ID               uuid.UUID      `json:"id" bson:"id"`
	NewOwnerKey      []byte         `json:"new_owner_key" bson:"new_owner_key"`
	NewEVMAddress    string         `json:"new_evm_address,omitempty" bson:"new_evm_address,omitempty"`
	NewSolanaAddress string         `json:"new_solana_address,omitempty" bson:"new_solana_address,omitempty"`
	InitiatedAt      time.Time      `json:"initiated_at" bson:"initiated_at"`
	ExecutableAt     time.Time      `json:"executable_at" bson:"executable_at"`
	InitiatedBy      uuid.UUID      `json:"initiated_by" bson:"initiated_by"`
	ApprovalCount    uint8          `json:"approval_count" bson:"approval_count"`
	Approvers        []uuid.UUID    `json:"approvers" bson:"approvers"`
	Executed         bool           `json:"executed" bson:"executed"`
	Reason           string         `json:"reason,omitempty" bson:"reason,omitempty"`
	ExpiresAt        time.Time      `json:"expires_at" bson:"expires_at"`
	Status           RecoveryStatus `json:"status" bson:"status"`
}

type RecoveryStatus string

const (
	RecoveryStatusPending   RecoveryStatus = "Pending"
	RecoveryStatusApproved  RecoveryStatus = "Approved"
	RecoveryStatusExecuted  RecoveryStatus = "Executed"
	RecoveryStatusCancelled RecoveryStatus = "Cancelled"
	RecoveryStatusExpired   RecoveryStatus = "Expired"
)

// TransactionLimits defines spending limits
type TransactionLimits struct {
	DailyLimit        uint64            `json:"daily_limit" bson:"daily_limit"`
	WeeklyLimit       uint64            `json:"weekly_limit" bson:"weekly_limit"`
	MonthlyLimit      uint64            `json:"monthly_limit" bson:"monthly_limit"`
	SingleTxLimit     uint64            `json:"single_tx_limit" bson:"single_tx_limit"`
	DailyUsed         uint64            `json:"daily_used" bson:"daily_used"`
	WeeklyUsed        uint64            `json:"weekly_used" bson:"weekly_used"`
	MonthlyUsed       uint64            `json:"monthly_used" bson:"monthly_used"`
	LastResetDaily    time.Time         `json:"last_reset_daily" bson:"last_reset_daily"`
	LastResetWeekly   time.Time         `json:"last_reset_weekly" bson:"last_reset_weekly"`
	LastResetMonthly  time.Time         `json:"last_reset_monthly" bson:"last_reset_monthly"`
	WhitelistedAddrs  []string          `json:"whitelisted_addresses,omitempty" bson:"whitelisted_addresses,omitempty"`
}

// RiskProfile for AML/fraud scoring
type RiskProfile struct {
	Score           float64   `json:"score" bson:"score"`           // 0.0-1.0
	Level           RiskLevel `json:"level" bson:"level"`
	LastAssessedAt  time.Time `json:"last_assessed_at" bson:"last_assessed_at"`
	Flags           []string  `json:"flags,omitempty" bson:"flags,omitempty"`
}

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "Low"
	RiskLevelMedium   RiskLevel = "Medium"
	RiskLevelHigh     RiskLevel = "High"
	RiskLevelCritical RiskLevel = "Critical"
)

// NewSmartWallet creates a new smart wallet
func NewSmartWallet(
	resourceOwner common.ResourceOwner,
	ownerID uuid.UUID,
	walletName string,
	walletType WalletType,
	primaryChain custody_vo.ChainID,
) *SmartWallet {
	baseEntity := common.NewPrivateEntity(resourceOwner)

	return &SmartWallet{
		BaseEntity:    baseEntity,
		UserID:        ownerID,
		OwnerID:       ownerID,
		WalletName:    walletName,
		Label:         walletName,
		WalletType:    walletType,
		PrimaryChain:  primaryChain,
		Addresses:     make(map[custody_vo.ChainID]string),
		Status:        WalletStatusPending,
		SecurityLevel: SecurityLevelStandard,
		KYCStatus:     KYCStatusNone,
		RecoveryConfig: &WalletRecoveryConfig{
			IsEnabled:         false,
			GuardianThreshold: 2,
			RecoveryDelay:     24 * time.Hour,
		},
		AAConfig: &AccountAbstractionConfig{
			IsDeployed: make(map[custody_vo.ChainID]bool),
		},
		Limits: TransactionLimits{
			DailyLimit:    10000_00,  // $10,000 default
			WeeklyLimit:   50000_00,
			MonthlyLimit:  100000_00,
			SingleTxLimit: 5000_00,
		},
		RiskProfile: RiskProfile{
			Score:          0.0,
			Level:          RiskLevelLow,
			LastAssessedAt: time.Now(),
		},
		Metadata: make(map[string]interface{}),
	}
}

// SetMPCKeyConfig sets the MPC key configuration after key generation
func (w *SmartWallet) SetMPCKeyConfig(keyID string, config MPCKeyConfiguration, shares []KeyShareInfo) {
	w.MasterKeyID = keyID
	w.KeyConfig = config
	w.KeyShares = shares
	w.UpdatedAt = time.Now()
}

// AddChainAddress adds an address for a specific chain
func (w *SmartWallet) AddChainAddress(chainID custody_vo.ChainID, address string) {
	w.Addresses[chainID] = address
	w.UpdatedAt = time.Now()
}

// Activate activates the wallet after key generation
func (w *SmartWallet) Activate() error {
	if w.MasterKeyID == "" {
		return fmt.Errorf("cannot activate wallet without MPC key")
	}
	if len(w.Addresses) == 0 {
		return fmt.Errorf("cannot activate wallet without any chain addresses")
	}

	now := time.Now()
	w.Status = WalletStatusActive
	w.ActivatedAt = &now
	w.UpdatedAt = now
	return nil
}

// Suspend suspends the wallet
func (w *SmartWallet) Suspend(reason string) {
	now := time.Now()
	w.Status = WalletStatusSuspended
	w.SuspendedAt = &now
	w.SuspendReason = reason
	w.UpdatedAt = now
}

// GetAddress returns the address for a specific chain
func (w *SmartWallet) GetAddress(chainID custody_vo.ChainID) (string, error) {
	if addr, ok := w.Addresses[chainID]; ok && addr != "" {
		return addr, nil
	}
	return "", fmt.Errorf("no address for chain %s", chainID)
}

// CanSpend checks if a transaction amount is within limits
func (w *SmartWallet) CanSpend(amount uint64) error {
	if w.Status != WalletStatusActive {
		return fmt.Errorf("wallet is not active: %s", w.Status)
	}

	if amount > w.Limits.SingleTxLimit {
		return fmt.Errorf("amount %d exceeds single transaction limit %d", amount, w.Limits.SingleTxLimit)
	}

	if w.Limits.DailyUsed+amount > w.Limits.DailyLimit {
		return fmt.Errorf("amount would exceed daily limit")
	}

	return nil
}

// RecordSpend records a spending transaction
func (w *SmartWallet) RecordSpend(amount uint64) {
	now := time.Now()

	// Reset daily if new day
	if now.YearDay() != w.Limits.LastResetDaily.YearDay() || now.Year() != w.Limits.LastResetDaily.Year() {
		w.Limits.DailyUsed = 0
		w.Limits.LastResetDaily = now
	}

	w.Limits.DailyUsed += amount
	w.Limits.WeeklyUsed += amount
	w.Limits.MonthlyUsed += amount
	w.LastActivityAt = &now
	w.UpdatedAt = now
}

// InitiateRecovery starts a social recovery process
func (w *SmartWallet) InitiateRecovery(initiatorID uuid.UUID, newOwnerKey []byte) error {
	if w.RecoveryConfig == nil || !w.RecoveryConfig.IsEnabled {
		return fmt.Errorf("social recovery is not enabled")
	}

	if w.PendingRecovery != nil && !w.PendingRecovery.Executed {
		return fmt.Errorf("recovery already in progress")
	}

	now := time.Now()
	w.PendingRecovery = &PendingRecovery{
		ID:            uuid.New(),
		InitiatedAt:   now,
		InitiatedBy:   initiatorID,
		NewOwnerKey:   newOwnerKey,
		Approvers:     []uuid.UUID{initiatorID},
		ApprovalCount: 1,
		ExecutableAt:  now.Add(w.RecoveryConfig.RecoveryDelay),
		ExpiresAt:     now.Add(7 * 24 * time.Hour), // 7 day expiry
		Status:        RecoveryStatusPending,
		Executed:      false,
	}

	w.Status = WalletStatusRecovering
	w.IsFrozen = true
	w.UpdatedAt = now
	return nil
}

// ApproveRecovery records a guardian approval
func (w *SmartWallet) ApproveRecovery(guardianID uuid.UUID) error {
	if w.PendingRecovery == nil {
		return fmt.Errorf("no pending recovery")
	}

	pr := w.PendingRecovery
	if pr.Executed {
		return fmt.Errorf("recovery already executed")
	}

	// Check not already approved
	for _, approver := range pr.Approvers {
		if approver == guardianID {
			return fmt.Errorf("guardian already approved")
		}
	}

	pr.Approvers = append(pr.Approvers, guardianID)
	pr.ApprovalCount++

	// Check if threshold met
	if pr.ApprovalCount >= w.RecoveryConfig.GuardianThreshold {
		pr.Status = RecoveryStatusApproved
	}

	w.UpdatedAt = time.Now()
	return nil
}
