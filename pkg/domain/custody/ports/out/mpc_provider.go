package custody_out

import (
	"context"
	"time"

	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// MPCProvider defines the interface for MPC operations
// Implementations can use different MPC providers (Fireblocks, Fordefi, custom TSS)
type MPCProvider interface {
	// Key Generation
	InitiateKeyGeneration(ctx context.Context, req *KeyGenRequest) (*KeyGenSession, error)
	GetKeyGenStatus(ctx context.Context, sessionID string) (*KeyGenSession, error)
	FinalizeKeyGeneration(ctx context.Context, sessionID string) (*GeneratedKey, error)

	// Signing Operations
	InitiateSigning(ctx context.Context, req *SigningRequest) (*SigningSession, error)
	GetSigningStatus(ctx context.Context, sessionID string) (*SigningSession, error)
	GetSignature(ctx context.Context, sessionID string) (*SignatureResult, error)

	// Key Management
	GetPublicKey(ctx context.Context, keyID string) (*PublicKeyInfo, error)
	RefreshKeyShares(ctx context.Context, keyID string) error
	RevokeKey(ctx context.Context, keyID string) error

	// Provider Info
	GetProviderInfo() ProviderInfo
	HealthCheck(ctx context.Context) error
}

// KeyGenRequest represents a request to generate new MPC keys
type KeyGenRequest struct {
	KeyID          string
	WalletID       string
	Curve          custody_vo.KeyCurve
	Scheme         custody_vo.MPCScheme
	Threshold      custody_vo.ThresholdConfig
	Metadata       map[string]string
	ExpiresAt      *time.Time
}

// KeyGenSession represents an ongoing key generation ceremony
type KeyGenSession struct {
	SessionID      string
	KeyID          string
	State          KeyGenState
	Participants   []ParticipantStatus
	PublicKey      []byte // Available after completion
	CreatedAt      time.Time
	CompletedAt    *time.Time
	Error          string
}

type KeyGenState string

const (
	KeyGenStateInitiated  KeyGenState = "Initiated"
	KeyGenStateRound1     KeyGenState = "Round1"
	KeyGenStateRound2     KeyGenState = "Round2"
	KeyGenStateRound3     KeyGenState = "Round3"
	KeyGenStateCompleted  KeyGenState = "Completed"
	KeyGenStateFailed     KeyGenState = "Failed"
)

type ParticipantStatus struct {
	ShareIndex  uint8
	Location    custody_vo.KeyShareLocation
	Status      string
	LastUpdated time.Time
}

// GeneratedKey represents a successfully generated MPC key
type GeneratedKey struct {
	KeyID           string
	PublicKey       []byte
	PublicKeyHex    string
	Curve           custody_vo.KeyCurve
	Scheme          custody_vo.MPCScheme
	Threshold       custody_vo.ThresholdConfig
	ShareMetadata   []custody_vo.KeyShareMetadata
	EVMAddress      string // Derived Ethereum address
	SolanaAddress   string // Derived Solana address
	CreatedAt       time.Time
}

// SigningRequest represents a request to sign a message
type SigningRequest struct {
	SessionID     string
	KeyID         string
	MessageHash   []byte // 32-byte hash
	MessageType   custody_vo.MessageType
	ChainID       custody_vo.ChainID
	Metadata      map[string]string
	ExpiresAt     time.Time
}

// SigningSession represents an ongoing signing session
type SigningSession struct {
	SessionID    string
	KeyID        string
	State        SigningState
	MessageHash  []byte
	Signature    []byte // Available after completion
	RecoveryID   *uint8 // For ECDSA (v value)
	CreatedAt    time.Time
	CompletedAt  *time.Time
	Error        string
}

type SigningState string

const (
	SigningStateInitiated SigningState = "Initiated"
	SigningStateRound1    SigningState = "Round1"
	SigningStateRound2    SigningState = "Round2"
	SigningStateCompleted SigningState = "Completed"
	SigningStateFailed    SigningState = "Failed"
	SigningStateExpired   SigningState = "Expired"
)

// SignatureResult contains the final signature
type SignatureResult struct {
	SessionID   string
	Signature   []byte
	RecoveryID  *uint8
	R           []byte
	S           []byte
	CompletedAt time.Time
}

// PublicKeyInfo contains public key information
type PublicKeyInfo struct {
	KeyID         string
	PublicKey     []byte
	PublicKeyHex  string
	Curve         custody_vo.KeyCurve
	Scheme        custody_vo.MPCScheme
	EVMAddress    string
	SolanaAddress string
	IsActive      bool
	CreatedAt     time.Time
}

// ProviderInfo contains MPC provider information
type ProviderInfo struct {
	Name             string
	Version          string
	SupportedCurves  []custody_vo.KeyCurve
	SupportedSchemes []custody_vo.MPCScheme
	Features         []string
}
