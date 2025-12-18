package custody_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// RecoveryService defines the interface for social recovery operations
type RecoveryService interface {
	// Guardian Management
	AddGuardian(ctx context.Context, req *AddGuardianRequest) (*GuardianResult, error)
	RemoveGuardian(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error
	GetGuardians(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.Guardian, error)
	SetGuardianThreshold(ctx context.Context, walletID uuid.UUID, threshold uint8) error

	// Recovery Process
	InitiateRecovery(ctx context.Context, req *InitiateRecoveryRequest) (*RecoveryInitResult, error)
	ApproveRecovery(ctx context.Context, req *ApproveRecoveryRequest) (*RecoveryApprovalResult, error)
	ExecuteRecovery(ctx context.Context, walletID uuid.UUID) (*RecoveryExecutionResult, error)
	CancelRecovery(ctx context.Context, walletID uuid.UUID) error

	// Recovery Status
	GetRecoveryStatus(ctx context.Context, walletID uuid.UUID) (*RecoveryStatus, error)
	GetPendingRecoveries(ctx context.Context, guardianAddress string) ([]*PendingRecoveryInfo, error)

	// Recovery Configuration
	SetRecoveryDelay(ctx context.Context, walletID uuid.UUID, delay time.Duration) error
	GetRecoveryConfig(ctx context.Context, walletID uuid.UUID) (*RecoveryConfig, error)
}

// AddGuardianRequest for adding a guardian
type AddGuardianRequest struct {
	WalletID      uuid.UUID
	GuardianType  custody_entities.GuardianType
	Address       string // Wallet address for wallet guardian
	Email         string // Email for email guardian
	Phone         string // Phone for phone guardian
	Label         string
	Weight        uint8 // For weighted threshold
	Metadata      map[string]string
}

// GuardianResult contains guardian addition result
type GuardianResult struct {
	GuardianID    uuid.UUID
	WalletID      uuid.UUID
	GuardianType  custody_entities.GuardianType
	Address       string
	Label         string
	Weight        uint8
	TxHash        *string // If on-chain registration needed
	CreatedAt     time.Time
}

// InitiateRecoveryRequest for starting recovery
type InitiateRecoveryRequest struct {
	WalletID       uuid.UUID
	InitiatorID    uuid.UUID // Guardian initiating
	NewOwnerKey    []byte    // New MPC public key
	NewEVMAddress  string
	NewSolanaAddress string
	Reason         string
	Evidence       []RecoveryEvidence
}

type RecoveryEvidence struct {
	Type        string // "kyc_document", "notarized_letter", "video_verification"
	DocumentID  string
	Hash        []byte
	UploadedAt  time.Time
}

// RecoveryInitResult contains recovery initiation result
type RecoveryInitResult struct {
	RecoveryID    uuid.UUID
	WalletID      uuid.UUID
	NewOwnerKey   []byte
	ExecutableAt  time.Time
	RequiredApprovals uint8
	CurrentApprovals  uint8
	TxHash        *string // If on-chain initiation
	InitiatedAt   time.Time
}

// ApproveRecoveryRequest for guardian approval
type ApproveRecoveryRequest struct {
	WalletID     uuid.UUID
	GuardianID   uuid.UUID
	Signature    []byte
	Message      string
}

// RecoveryApprovalResult contains approval result
type RecoveryApprovalResult struct {
	WalletID          uuid.UUID
	GuardianID        uuid.UUID
	ApprovalCount     uint8
	RequiredApprovals uint8
	IsReady           bool // Can execute now
	TxHash            *string
	ApprovedAt        time.Time
}

// RecoveryExecutionResult contains recovery execution result
type RecoveryExecutionResult struct {
	WalletID       uuid.UUID
	OldOwnerKey    []byte
	NewOwnerKey    []byte
	OldEVMAddress  string
	NewEVMAddress  string
	OldSolanaAddress string
	NewSolanaAddress string
	TxHashes       map[custody_vo.ChainID]string // Per-chain tx hashes
	ExecutedAt     time.Time
}

// RecoveryStatus contains current recovery status
type RecoveryStatus struct {
	WalletID          uuid.UUID
	HasPendingRecovery bool
	NewOwnerKey       []byte
	NewEVMAddress     string
	NewSolanaAddress  string
	InitiatedAt       *time.Time
	ExecutableAt      *time.Time
	TimeRemaining     *time.Duration
	RequiredApprovals uint8
	CurrentApprovals  uint8
	Approvers         []ApproverInfo
	CanExecute        bool
	CanCancel         bool
}

type ApproverInfo struct {
	GuardianID   uuid.UUID
	GuardianType custody_entities.GuardianType
	Label        string
	Approved     bool
	ApprovedAt   *time.Time
}

// PendingRecoveryInfo for guardian dashboard
type PendingRecoveryInfo struct {
	WalletID          uuid.UUID
	WalletLabel       string
	UserID            uuid.UUID
	NewOwnerKey       []byte
	InitiatedAt       time.Time
	ExecutableAt      time.Time
	RequiredApprovals uint8
	CurrentApprovals  uint8
	MyApprovalStatus  string // "pending", "approved", "not_guardian"
	Reason            string
}

// RecoveryConfig contains recovery configuration
type RecoveryConfig struct {
	WalletID          uuid.UUID
	RecoveryDelay     time.Duration
	GuardianThreshold uint8
	TotalGuardians    uint8
	Guardians         []GuardianInfo
}

type GuardianInfo struct {
	GuardianID   uuid.UUID
	GuardianType custody_entities.GuardianType
	Address      string
	Label        string
	Weight       uint8
	IsActive     bool
	AddedAt      time.Time
}

// EmergencyService defines the interface for emergency wallet operations
type EmergencyService interface {
	// Emergency Actions
	EmergencyFreeze(ctx context.Context, req *EmergencyFreezeRequest) (*EmergencyResult, error)
	EmergencyUnfreeze(ctx context.Context, req *EmergencyUnfreezeRequest) (*EmergencyResult, error)

	// Emergency Transfer (guardian-initiated)
	EmergencyTransfer(ctx context.Context, req *EmergencyTransferRequest) (*EmergencyTransferResult, error)

	// Status
	GetEmergencyStatus(ctx context.Context, walletID uuid.UUID) (*EmergencyStatus, error)
}

// EmergencyFreezeRequest for emergency freeze
type EmergencyFreezeRequest struct {
	WalletID    uuid.UUID
	InitiatorID uuid.UUID // Guardian or owner
	Reason      string
	Duration    *time.Duration // Optional auto-unfreeze
}

// EmergencyUnfreezeRequest for emergency unfreeze
type EmergencyUnfreezeRequest struct {
	WalletID   uuid.UUID
	Signatures []GuardianSignature
}

type GuardianSignature struct {
	GuardianID uuid.UUID
	Signature  []byte
	Message    []byte
}

// EmergencyResult contains emergency action result
type EmergencyResult struct {
	WalletID     uuid.UUID
	Action       string // "freeze", "unfreeze"
	TxHashes     map[custody_vo.ChainID]string
	ExecutedAt   time.Time
	UnfreezesAt  *time.Time
}

// EmergencyTransferRequest for guardian-initiated emergency transfer
type EmergencyTransferRequest struct {
	WalletID    uuid.UUID
	ChainID     custody_vo.ChainID
	To          string // Safe address (must be whitelisted)
	AssetType   string // "native" or "token"
	TokenAddress *string
	Amount      string // "all" or specific amount
	Signatures  []GuardianSignature
	Reason      string
}

// EmergencyTransferResult contains emergency transfer result
type EmergencyTransferResult struct {
	WalletID     uuid.UUID
	ChainID      custody_vo.ChainID
	TxHash       string
	From         string
	To           string
	Amount       string
	TokenAddress *string
	ExecutedAt   time.Time
}

// EmergencyStatus contains emergency status
type EmergencyStatus struct {
	WalletID      uuid.UUID
	IsFrozen      bool
	FrozenAt      *time.Time
	FrozenBy      *uuid.UUID
	FreezeReason  *string
	UnfreezesAt   *time.Time
	CanUnfreeze   bool
	RequiredSigs  uint8
	CurrentSigs   uint8
}
