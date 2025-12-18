package custody_out

import (
	"context"
	"time"

	custody_vo "github.com/replay-api/replay-api/pkg/domain/custody/value-objects"
)

// HSMProvider defines the interface for Hardware Security Module operations
// Supports AWS CloudHSM, Azure HSM, Thales Luna, YubiHSM
type HSMProvider interface {
	// Key Storage
	StoreKeyShare(ctx context.Context, req *StoreKeyShareRequest) (*StoredKeyShare, error)
	RetrieveKeyShare(ctx context.Context, keyID string, shareIndex uint8) (*RetrievedKeyShare, error)
	DeleteKeyShare(ctx context.Context, keyID string, shareIndex uint8) error

	// Cryptographic Operations (performed inside HSM)
	SignInsideHSM(ctx context.Context, req *HSMSignRequest) (*HSMSignResult, error)
	DeriveChildKey(ctx context.Context, req *DeriveKeyRequest) (*DerivedKey, error)

	// Key Management
	ListKeys(ctx context.Context, filter *KeyFilter) ([]*HSMKeyInfo, error)
	GetKeyInfo(ctx context.Context, keyID string) (*HSMKeyInfo, error)
	RotateWrappingKey(ctx context.Context) error

	// Audit
	GetAuditLogs(ctx context.Context, filter *AuditFilter) ([]*HSMAuditLog, error)

	// Health
	HealthCheck(ctx context.Context) (*HSMHealthStatus, error)
	GetProviderInfo() HSMProviderInfo
}

// StoreKeyShareRequest represents a request to store a key share in HSM
type StoreKeyShareRequest struct {
	KeyID          string
	ShareIndex     uint8
	EncryptedShare []byte // Pre-encrypted share data
	Metadata       KeyShareStorageMetadata
	Policy         KeyUsagePolicy
}

type KeyShareStorageMetadata struct {
	WalletID    string
	Curve       custody_vo.KeyCurve
	Scheme      custody_vo.MPCScheme
	Purpose     custody_vo.KeyPurpose
	CreatedBy   string
	Tags        map[string]string
}

type KeyUsagePolicy struct {
	AllowedOperations []string // "sign", "derive", "export"
	RequireMFA        bool
	MaxSignsPerDay    int
	ExpiresAt         *time.Time
}

// StoredKeyShare represents a stored key share
type StoredKeyShare struct {
	KeyID       string
	ShareIndex  uint8
	HSMKeyID    string // Internal HSM key identifier
	Version     uint32
	StoredAt    time.Time
	ExpiresAt   *time.Time
}

// RetrievedKeyShare represents a retrieved key share for MPC operations
type RetrievedKeyShare struct {
	KeyID          string
	ShareIndex     uint8
	EncryptedShare []byte // Encrypted with session key
	SessionKeyID   string // Key used for encryption
	RetrievedAt    time.Time
}

// HSMSignRequest represents a signing request to be performed inside HSM
// Used for non-MPC operations or HSM-based signing
type HSMSignRequest struct {
	KeyID       string
	MessageHash []byte
	Algorithm   string // "ECDSA_SHA256", "EdDSA_Ed25519"
}

type HSMSignResult struct {
	Signature  []byte
	R          []byte
	S          []byte
	RecoveryID *uint8
	SignedAt   time.Time
}

// DeriveKeyRequest for HD key derivation
type DeriveKeyRequest struct {
	ParentKeyID    string
	DerivationPath string // BIP32/BIP44 path
	ChildIndex     uint32
	Hardened       bool
}

type DerivedKey struct {
	KeyID         string
	PublicKey     []byte
	ChainCode     []byte
	DerivationPath string
	DerivedAt     time.Time
}

// KeyFilter for listing keys
type KeyFilter struct {
	WalletID  *string
	Curve     *custody_vo.KeyCurve
	Purpose   *custody_vo.KeyPurpose
	IsActive  *bool
	CreatedAfter *time.Time
	Limit     int
	Offset    int
}

// HSMKeyInfo contains HSM key information
type HSMKeyInfo struct {
	KeyID         string
	HSMKeyID      string
	ShareIndex    uint8
	Curve         custody_vo.KeyCurve
	Scheme        custody_vo.MPCScheme
	Purpose       custody_vo.KeyPurpose
	IsActive      bool
	CreatedAt     time.Time
	LastUsedAt    *time.Time
	UsageCount    int64
	Policy        KeyUsagePolicy
}

// AuditFilter for retrieving audit logs
type AuditFilter struct {
	KeyID       *string
	Operation   *string
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
}

// HSMAuditLog represents an audit entry
type HSMAuditLog struct {
	ID          string
	KeyID       string
	Operation   string
	PerformedBy string
	IPAddress   string
	Timestamp   time.Time
	Success     bool
	ErrorMsg    string
	Metadata    map[string]string
}

// HSMHealthStatus represents HSM health information
type HSMHealthStatus struct {
	Provider    custody_vo.HSMProvider
	Status      string // "healthy", "degraded", "unavailable"
	Latency     time.Duration
	Partition   string
	KeyCount    int
	CheckedAt   time.Time
}

// HSMProviderInfo contains provider information
type HSMProviderInfo struct {
	Provider         custody_vo.HSMProvider
	Version          string
	FIPS140Level     int // 2 or 3
	SupportedCurves  []custody_vo.KeyCurve
	MaxKeysPerPartition int
	Features         []string
}

// SecureEnclaveProvider defines interface for TEE-based operations
// Supports AWS Nitro Enclaves, Azure SGX, Intel SGX
type SecureEnclaveProvider interface {
	// Enclave Operations
	ExecuteInEnclave(ctx context.Context, req *EnclaveRequest) (*EnclaveResponse, error)

	// Attestation
	GetAttestation(ctx context.Context) (*AttestationDocument, error)
	VerifyAttestation(ctx context.Context, doc *AttestationDocument) error

	// Key Operations (inside enclave)
	GenerateKeyInEnclave(ctx context.Context, req *EnclaveKeyGenRequest) (*EnclaveKeyResult, error)
	SignInEnclave(ctx context.Context, req *EnclaveSignRequest) (*EnclaveSignResult, error)

	// Health
	HealthCheck(ctx context.Context) (*EnclaveHealthStatus, error)
}

type EnclaveRequest struct {
	Operation string
	Payload   []byte
	Encrypted bool
}

type EnclaveResponse struct {
	Result    []byte
	Encrypted bool
}

type AttestationDocument struct {
	Document      []byte
	PCRValues     map[int]string // Platform Configuration Registers
	Timestamp     time.Time
	EnclaveID     string
	SignerCert    []byte
}

type EnclaveKeyGenRequest struct {
	KeyID   string
	Curve   custody_vo.KeyCurve
	Purpose custody_vo.KeyPurpose
}

type EnclaveKeyResult struct {
	KeyID       string
	PublicKey   []byte
	Attestation *AttestationDocument
}

type EnclaveSignRequest struct {
	KeyID       string
	MessageHash []byte
}

type EnclaveSignResult struct {
	Signature   []byte
	Attestation *AttestationDocument
}

type EnclaveHealthStatus struct {
	Provider    custody_vo.EnclaveProvider
	Status      string
	EnclaveID   string
	PCRValues   map[int]string
	CheckedAt   time.Time
}
