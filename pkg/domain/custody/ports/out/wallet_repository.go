package custody_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	custody_entities "github.com/replay-api/replay-api/pkg/domain/custody/entities"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// SmartWalletRepository defines the interface for smart wallet persistence
type SmartWalletRepository interface {
	// CRUD Operations
	Create(ctx context.Context, wallet *custody_entities.SmartWallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*custody_entities.SmartWallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*custody_entities.SmartWallet, error)
	GetByAddress(ctx context.Context, chainID custody_vo.ChainID, address string) (*custody_entities.SmartWallet, error)
	Update(ctx context.Context, wallet *custody_entities.SmartWallet) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query Operations
	List(ctx context.Context, filter *WalletFilter) (*WalletListResult, error)
	GetByMPCKeyID(ctx context.Context, keyID string) (*custody_entities.SmartWallet, error)
	GetPendingRecoveries(ctx context.Context) ([]*custody_entities.SmartWallet, error)
	GetFrozenWallets(ctx context.Context) ([]*custody_entities.SmartWallet, error)

	// Guardian Operations
	AddGuardian(ctx context.Context, walletID uuid.UUID, guardian *custody_entities.Guardian) error
	RemoveGuardian(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error
	GetGuardians(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.Guardian, error)
	GetGuardianByAddress(ctx context.Context, walletID uuid.UUID, address string) (*custody_entities.Guardian, error)

	// Session Key Operations
	AddSessionKey(ctx context.Context, walletID uuid.UUID, sessionKey *custody_entities.SessionKey) error
	RevokeSessionKey(ctx context.Context, walletID uuid.UUID, keyAddress string) error
	GetActiveSessionKeys(ctx context.Context, walletID uuid.UUID) ([]*custody_entities.SessionKey, error)

	// Recovery Operations
	SetPendingRecovery(ctx context.Context, walletID uuid.UUID, recovery *custody_entities.PendingRecovery) error
	ClearPendingRecovery(ctx context.Context, walletID uuid.UUID) error
	AddRecoveryApproval(ctx context.Context, walletID uuid.UUID, guardianID uuid.UUID) error
}

// WalletFilter for querying wallets
type WalletFilter struct {
	UserID       *uuid.UUID
	TenantID     *uuid.UUID
	ChainID      *custody_vo.ChainID
	WalletType   *custody_entities.WalletType
	Status       *custody_entities.WalletStatus
	IsFrozen     *bool
	HasRecovery  *bool
	KYCStatus    *custody_entities.KYCStatus
	CreatedAfter *time.Time
	CreatedBefore *time.Time
	Limit        int
	Offset       int
	OrderBy      string
	OrderDesc    bool
}

// WalletListResult contains paginated wallet results
type WalletListResult struct {
	Wallets    []*custody_entities.SmartWallet
	TotalCount int64
	Limit      int
	Offset     int
}

// TransactionRepository for custody transaction records
type TransactionRepository interface {
	// CRUD
	Create(ctx context.Context, tx *CustodyTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*CustodyTransaction, error)
	GetByHash(ctx context.Context, chainID custody_vo.ChainID, hash string) (*CustodyTransaction, error)
	Update(ctx context.Context, tx *CustodyTransaction) error

	// Queries
	ListByWallet(ctx context.Context, walletID uuid.UUID, filter *TxFilter) (*TxListResult, error)
	GetPendingTransactions(ctx context.Context) ([]*CustodyTransaction, error)
	GetFailedTransactions(ctx context.Context, since time.Time) ([]*CustodyTransaction, error)

	// Aggregations
	GetDailySpending(ctx context.Context, walletID uuid.UUID, date time.Time) (*SpendingAggregate, error)
	GetWeeklySpending(ctx context.Context, walletID uuid.UUID, weekStart time.Time) (*SpendingAggregate, error)
	GetMonthlySpending(ctx context.Context, walletID uuid.UUID, month time.Time) (*SpendingAggregate, error)
}

// CustodyTransaction represents a custody system transaction
type CustodyTransaction struct {
	ID               uuid.UUID
	WalletID         uuid.UUID
	ChainID          custody_vo.ChainID
	TxHash           string
	TxType           TxType
	Status           TxRecordStatus
	From             string
	To               string
	Value            string // Decimal string
	TokenAddress     *string
	GasUsed          *uint64
	GasCost          *string
	SigningSessionID *string
	UserOpHash       *string // For ERC-4337
	Metadata         map[string]string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ConfirmedAt      *time.Time
	FailedAt         *time.Time
	FailureReason    *string
}

type TxType string

const (
	TxTypeTransfer         TxType = "Transfer"
	TxTypeTokenTransfer    TxType = "TokenTransfer"
	TxTypeContractCall     TxType = "ContractCall"
	TxTypeWalletDeploy     TxType = "WalletDeploy"
	TxTypeRecoveryExecute  TxType = "RecoveryExecute"
	TxTypeGuardianAdd      TxType = "GuardianAdd"
	TxTypeGuardianRemove   TxType = "GuardianRemove"
	TxTypeSessionKeyAdd    TxType = "SessionKeyAdd"
	TxTypeSessionKeyRevoke TxType = "SessionKeyRevoke"
)

type TxRecordStatus string

const (
	TxRecordStatusPending   TxRecordStatus = "Pending"
	TxRecordStatusSigned    TxRecordStatus = "Signed"
	TxRecordStatusSubmitted TxRecordStatus = "Submitted"
	TxRecordStatusConfirmed TxRecordStatus = "Confirmed"
	TxRecordStatusFailed    TxRecordStatus = "Failed"
)

type TxFilter struct {
	TxType       *TxType
	Status       *TxRecordStatus
	ChainID      *custody_vo.ChainID
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

type TxListResult struct {
	Transactions []*CustodyTransaction
	TotalCount   int64
	Limit        int
	Offset       int
}

type SpendingAggregate struct {
	WalletID     uuid.UUID
	Period       string // "daily", "weekly", "monthly"
	PeriodStart  time.Time
	TotalNative  string // Decimal string
	TotalUSD     string // Decimal string
	TxCount      int64
	ByToken      map[string]string // tokenAddress -> amount
}

// SigningSessionRepository for MPC signing session tracking
type SigningSessionRepository interface {
	Create(ctx context.Context, session *SigningSessionRecord) error
	GetByID(ctx context.Context, sessionID string) (*SigningSessionRecord, error)
	Update(ctx context.Context, session *SigningSessionRecord) error
	ListActive(ctx context.Context) ([]*SigningSessionRecord, error)
	ListByKey(ctx context.Context, keyID string) ([]*SigningSessionRecord, error)
}

type SigningSessionRecord struct {
	SessionID     string
	KeyID         string
	WalletID      uuid.UUID
	MessageHash   []byte
	MessageType   custody_vo.MessageType
	ChainID       custody_vo.ChainID
	State         custody_vo.SigningSessionState
	RequestedBy   string
	ApprovedBy    []string
	Signature     []byte
	RecoveryID    *uint8
	Error         *string
	Metadata      map[string]string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
	ExpiresAt     time.Time
}

// KeyRepository for MPC key metadata persistence
type KeyRepository interface {
	Create(ctx context.Context, key *KeyRecord) error
	GetByID(ctx context.Context, keyID string) (*KeyRecord, error)
	GetByWallet(ctx context.Context, walletID uuid.UUID) ([]*KeyRecord, error)
	Update(ctx context.Context, key *KeyRecord) error
	Deactivate(ctx context.Context, keyID string) error
	ListActive(ctx context.Context) ([]*KeyRecord, error)
}

type KeyRecord struct {
	KeyID          string
	WalletID       uuid.UUID
	PublicKey      []byte
	PublicKeyHex   string
	Curve          custody_vo.KeyCurve
	Scheme         custody_vo.MPCScheme
	Threshold      custody_vo.ThresholdConfig
	Purpose        custody_vo.KeyPurpose
	EVMAddress     string
	SolanaAddress  string
	ShareMetadata  []custody_vo.KeyShareMetadata
	IsActive       bool
	CreatedAt      time.Time
	LastUsedAt     *time.Time
	ExpiresAt      *time.Time
}
