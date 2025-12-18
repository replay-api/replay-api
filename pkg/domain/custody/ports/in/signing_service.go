package custody_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// SigningService defines the interface for MPC signing operations
type SigningService interface {
	// Signing Operations
	RequestSigning(ctx context.Context, req *SigningRequest) (*SigningResult, error)
	GetSigningStatus(ctx context.Context, sessionID string) (*SigningStatus, error)
	CancelSigning(ctx context.Context, sessionID string) error

	// Typed Data Signing (EIP-712)
	SignTypedData(ctx context.Context, req *TypedDataSigningRequest) (*SigningResult, error)

	// Personal Sign (EIP-191)
	PersonalSign(ctx context.Context, req *PersonalSignRequest) (*SigningResult, error)

	// Solana-specific
	SignSolanaTransaction(ctx context.Context, req *SolanaSigningRequest) (*SigningResult, error)
	SignSolanaMessage(ctx context.Context, req *SolanaMessageRequest) (*SigningResult, error)

	// Transaction Signing
	SignTransaction(ctx context.Context, req *TxSigningRequest) (*SignedTxResult, error)
	SignUserOperation(ctx context.Context, req *UserOpSigningRequest) (*SignedUserOpResult, error)

	// Key Information
	GetSignerAddress(ctx context.Context, walletID uuid.UUID, chainID custody_vo.ChainID) (string, error)
	GetPublicKey(ctx context.Context, walletID uuid.UUID) (*PublicKeyInfo, error)
}

// SigningRequest for raw message signing
type SigningRequest struct {
	WalletID      uuid.UUID
	MessageHash   []byte // 32-byte hash
	MessageType   custody_vo.MessageType
	ChainID       custody_vo.ChainID
	RequestedBy   string
	Reason        string
	Metadata      map[string]string
	ExpiresIn     time.Duration
	IdempotencyKey string
}

// SigningResult contains the signing result
type SigningResult struct {
	SessionID   string
	Signature   []byte
	SignatureHex string
	R           []byte
	S           []byte
	V           *uint8 // Recovery ID for ECDSA
	PublicKey   []byte
	CompletedAt time.Time
}

// SigningStatus contains signing session status
type SigningStatus struct {
	SessionID    string
	State        custody_vo.SigningSessionState
	Signature    []byte
	Error        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
	ExpiresAt    time.Time
}

// TypedDataSigningRequest for EIP-712 typed data signing
type TypedDataSigningRequest struct {
	WalletID       uuid.UUID
	ChainID        custody_vo.ChainID
	Domain         TypedDataDomain
	PrimaryType    string
	Types          map[string][]TypedDataField
	Message        map[string]interface{}
	RequestedBy    string
	Reason         string
	IdempotencyKey string
}

type TypedDataDomain struct {
	Name              string
	Version           string
	ChainId           uint64
	VerifyingContract string
	Salt              []byte
}

type TypedDataField struct {
	Name string
	Type string
}

// PersonalSignRequest for EIP-191 personal sign
type PersonalSignRequest struct {
	WalletID       uuid.UUID
	ChainID        custody_vo.ChainID
	Message        []byte
	RequestedBy    string
	Reason         string
	IdempotencyKey string
}

// SolanaSigningRequest for Solana transaction signing
type SolanaSigningRequest struct {
	WalletID        uuid.UUID
	SerializedTx    []byte
	RecentBlockhash string
	RequestedBy     string
	Reason          string
	IdempotencyKey  string
}

// SolanaMessageRequest for Solana message signing
type SolanaMessageRequest struct {
	WalletID       uuid.UUID
	Message        []byte
	RequestedBy    string
	Reason         string
	IdempotencyKey string
}

// TxSigningRequest for EVM transaction signing
type TxSigningRequest struct {
	WalletID       uuid.UUID
	ChainID        custody_vo.ChainID
	To             string
	Value          string // Decimal string in wei
	Data           []byte
	Nonce          uint64
	GasLimit       uint64
	MaxFeePerGas   string
	MaxPriorityFee string
	RequestedBy    string
	Reason         string
	IdempotencyKey string
}

// SignedTxResult contains signed transaction
type SignedTxResult struct {
	SessionID     string
	RawTx         []byte
	RawTxHex      string
	TxHash        string
	Signature     []byte
	CompletedAt   time.Time
}

// UserOpSigningRequest for ERC-4337 UserOperation signing
type UserOpSigningRequest struct {
	WalletID           uuid.UUID
	ChainID            custody_vo.ChainID
	Sender             string
	Nonce              string
	InitCode           []byte
	CallData           []byte
	AccountGasLimits   [32]byte
	PreVerificationGas string
	GasFees            [32]byte
	PaymasterAndData   []byte
	RequestedBy        string
	Reason             string
	IdempotencyKey     string
}

// SignedUserOpResult contains signed UserOperation
type SignedUserOpResult struct {
	SessionID    string
	UserOpHash   string
	Signature    []byte
	SignatureHex string
	SignedUserOp *SignedUserOperation
	CompletedAt  time.Time
}

type SignedUserOperation struct {
	Sender               string
	Nonce                string
	InitCode             []byte
	CallData             []byte
	AccountGasLimits     [32]byte
	PreVerificationGas   string
	GasFees              [32]byte
	PaymasterAndData     []byte
	Signature            []byte
}

// PublicKeyInfo contains public key information
type PublicKeyInfo struct {
	KeyID         string
	PublicKey     []byte
	PublicKeyHex  string
	Curve         custody_vo.KeyCurve
	EVMAddress    string
	SolanaAddress string
}

// KeyGenerationService defines the interface for MPC key generation
type KeyGenerationService interface {
	// Key Generation
	GenerateKey(ctx context.Context, req *KeyGenRequest) (*KeyGenResult, error)
	GetKeyGenStatus(ctx context.Context, sessionID string) (*KeyGenStatus, error)

	// Key Management
	GetKeyInfo(ctx context.Context, keyID string) (*KeyInfo, error)
	RotateKey(ctx context.Context, keyID string) (*KeyRotationResult, error)
	RevokeKey(ctx context.Context, keyID string) error

	// Key Derivation (BIP32/BIP44)
	DeriveChildKey(ctx context.Context, req *DeriveKeyRequest) (*DerivedKeyResult, error)
}

// KeyGenRequest for key generation
type KeyGenRequest struct {
	WalletID   uuid.UUID
	Curve      custody_vo.KeyCurve
	Scheme     custody_vo.MPCScheme
	Threshold  custody_vo.ThresholdConfig
	Purpose    custody_vo.KeyPurpose
	ExpiresAt  *time.Time
}

// KeyGenResult contains key generation result
type KeyGenResult struct {
	SessionID     string
	KeyID         string
	PublicKey     []byte
	PublicKeyHex  string
	EVMAddress    string
	SolanaAddress string
	Curve         custody_vo.KeyCurve
	Scheme        custody_vo.MPCScheme
	Threshold     custody_vo.ThresholdConfig
	CreatedAt     time.Time
}

// KeyGenStatus contains key generation status
type KeyGenStatus struct {
	SessionID    string
	KeyID        string
	State        string
	PublicKey    []byte
	Error        *string
	CreatedAt    time.Time
	CompletedAt  *time.Time
}

// KeyInfo contains key information
type KeyInfo struct {
	KeyID          string
	PublicKey      []byte
	PublicKeyHex   string
	Curve          custody_vo.KeyCurve
	Scheme         custody_vo.MPCScheme
	Threshold      custody_vo.ThresholdConfig
	Purpose        custody_vo.KeyPurpose
	EVMAddress     string
	SolanaAddress  string
	IsActive       bool
	CreatedAt      time.Time
	LastUsedAt     *time.Time
	ExpiresAt      *time.Time
}

// KeyRotationResult contains key rotation result
type KeyRotationResult struct {
	OldKeyID      string
	NewKeyID      string
	NewPublicKey  []byte
	NewEVMAddress string
	NewSolanaAddress string
	RotatedAt     time.Time
}

// DeriveKeyRequest for BIP32/BIP44 key derivation
type DeriveKeyRequest struct {
	ParentKeyID    string
	DerivationPath string // e.g., "m/44'/60'/0'/0/0"
	ChainID        custody_vo.ChainID
}

// DerivedKeyResult contains derived key result
type DerivedKeyResult struct {
	KeyID          string
	ParentKeyID    string
	DerivationPath string
	PublicKey      []byte
	Address        string
	ChainID        custody_vo.ChainID
	CreatedAt      time.Time
}
