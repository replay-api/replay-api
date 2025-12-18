package custody_vo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// MPCScheme represents the MPC protocol used
type MPCScheme string

const (
	// Threshold Signature Schemes
	MPCSchemeGG20     MPCScheme = "GG20"      // Gennaro-Goldfeder 2020 (ECDSA)
	MPCSchemeCMP      MPCScheme = "CMP"       // Canetti-Makriyannis-Peled (faster ECDSA)
	MPCSchemeFROST    MPCScheme = "FROST"     // Flexible Round-Optimized Schnorr (EdDSA)
	MPCSchemeLindell  MPCScheme = "Lindell17" // Lindell 2017 (2-party ECDSA)

	// For Solana (Ed25519)
	MPCSchemeFROSTEd25519 MPCScheme = "FROST-Ed25519"
)

// KeyCurve represents the elliptic curve for keys
type KeyCurve string

const (
	CurveSecp256k1 KeyCurve = "secp256k1" // Ethereum, Bitcoin
	CurveEd25519   KeyCurve = "ed25519"   // Solana, Cosmos
	CurveP256      KeyCurve = "P-256"     // NIST P-256
)

// ThresholdConfig defines the t-of-n threshold for MPC
type ThresholdConfig struct {
	Threshold   uint8 `json:"threshold"`    // t - minimum signers required
	TotalShares uint8 `json:"total_shares"` // n - total key shares
}

// Common threshold configurations
var (
	Threshold2of3 = ThresholdConfig{Threshold: 2, TotalShares: 3}
	Threshold3of5 = ThresholdConfig{Threshold: 3, TotalShares: 5}
	Threshold4of7 = ThresholdConfig{Threshold: 4, TotalShares: 7}
)

// KeyShareID uniquely identifies a key share
type KeyShareID string

func NewKeyShareID(keyID string, shareIndex uint8) KeyShareID {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", keyID, shareIndex)))
	return KeyShareID(hex.EncodeToString(hash[:16]))
}

// KeyShareLocation represents where a key share is stored
type KeyShareLocation string

const (
	LocationHSM          KeyShareLocation = "HSM"           // Hardware Security Module
	LocationSecureEnclave KeyShareLocation = "SecureEnclave" // AWS Nitro/Azure Confidential
	LocationKMS          KeyShareLocation = "KMS"           // Cloud KMS (wrapped)
	LocationUserDevice   KeyShareLocation = "UserDevice"    // User's device (for recovery)
	LocationColdStorage  KeyShareLocation = "ColdStorage"   // Air-gapped cold storage
)

// KeyShareMetadata contains metadata about a key share
type KeyShareMetadata struct {
	ShareID       KeyShareID       `json:"share_id"`
	ShareIndex    uint8            `json:"share_index"`
	Location      KeyShareLocation `json:"location"`
	ProviderID    string           `json:"provider_id"`    // HSM ID, enclave ID, etc.
	EncryptionKey string           `json:"encryption_key"` // KMS key ARN for wrapped shares
	CreatedAt     time.Time        `json:"created_at"`
	LastUsedAt    *time.Time       `json:"last_used_at,omitempty"`
	IsActive      bool             `json:"is_active"`
}

// MPCKeyGenRequest represents a request to generate new MPC keys
type MPCKeyGenRequest struct {
	KeyID           string          `json:"key_id"`
	Curve           KeyCurve        `json:"curve"`
	Scheme          MPCScheme       `json:"scheme"`
	Threshold       ThresholdConfig `json:"threshold"`
	ShareLocations  []KeyShareLocation `json:"share_locations"`
	InitiatorID     string          `json:"initiator_id"`
	Purpose         KeyPurpose      `json:"purpose"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
}

// KeyPurpose defines the intended use of a key
type KeyPurpose string

const (
	KeyPurposeTransaction KeyPurpose = "Transaction" // Normal transactions
	KeyPurposeRecovery    KeyPurpose = "Recovery"    // Social recovery
	KeyPurposeAdmin       KeyPurpose = "Admin"       // Administrative operations
	KeyPurposeRotation    KeyPurpose = "Rotation"    // Key rotation ceremonies
)

// SigningSessionID uniquely identifies an MPC signing session
type SigningSessionID string

// SigningSessionState represents the state of an MPC signing session
type SigningSessionState string

const (
	SigningStateInitiated   SigningSessionState = "Initiated"
	SigningStateRound1      SigningSessionState = "Round1"
	SigningStateRound2      SigningSessionState = "Round2"
	SigningStateRound3      SigningSessionState = "Round3"
	SigningStateCompleted   SigningSessionState = "Completed"
	SigningStateFailed      SigningSessionState = "Failed"
	SigningStateExpired     SigningSessionState = "Expired"
)

// SigningRequest represents a request to sign a message/transaction
type SigningRequest struct {
	SessionID     SigningSessionID `json:"session_id"`
	KeyID         string           `json:"key_id"`
	MessageHash   []byte           `json:"message_hash"`   // 32-byte hash to sign
	MessageType   MessageType      `json:"message_type"`
	ChainID       ChainID          `json:"chain_id"`
	Participants  []KeyShareID     `json:"participants"`   // Which shares are participating
	RequestedBy   string           `json:"requested_by"`
	ApprovedBy    []string         `json:"approved_by"`
	ExpiresAt     time.Time        `json:"expires_at"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

// MessageType indicates what kind of message is being signed
type MessageType string

const (
	MessageTypeTransaction   MessageType = "Transaction"
	MessageTypeTypedData     MessageType = "TypedData"     // EIP-712
	MessageTypePersonalSign  MessageType = "PersonalSign"  // EIP-191
	MessageTypeSolanaMessage MessageType = "SolanaMessage"
)

// SigningResult contains the result of an MPC signing operation
type SigningResult struct {
	SessionID   SigningSessionID `json:"session_id"`
	Signature   []byte           `json:"signature"`
	RecoveryID  *uint8           `json:"recovery_id,omitempty"` // For ECDSA (v value)
	PublicKey   []byte           `json:"public_key"`
	CompletedAt time.Time        `json:"completed_at"`
}

// HSMConfig represents HSM provider configuration
type HSMConfig struct {
	Provider    HSMProvider `json:"provider"`
	ClusterID   string      `json:"cluster_id"`
	Region      string      `json:"region"`
	KeyAlias    string      `json:"key_alias"`
	Credentials HSMCredentials `json:"credentials"`
}

type HSMProvider string

const (
	HSMProviderAWSCloudHSM HSMProvider = "AWS_CloudHSM"
	HSMProviderAzureHSM    HSMProvider = "Azure_HSM"
	HSMProviderGoogleHSM   HSMProvider = "Google_CloudHSM"
	HSMProviderThales      HSMProvider = "Thales_Luna"
	HSMProviderYubico      HSMProvider = "YubiHSM"
)

type HSMCredentials struct {
	Type         string `json:"type"`          // "certificate", "password", "iam"
	CertPath     string `json:"cert_path,omitempty"`
	PartitionPIN string `json:"partition_pin,omitempty"`
	IAMRole      string `json:"iam_role,omitempty"`
}

// SecureEnclaveConfig for TEE-based key storage
type SecureEnclaveConfig struct {
	Provider     EnclaveProvider `json:"provider"`
	AttestationURL string        `json:"attestation_url"`
	EnclaveID    string          `json:"enclave_id"`
	PCRValues    map[int]string  `json:"pcr_values"` // Platform Configuration Registers
}

type EnclaveProvider string

const (
	EnclaveProviderAWSNitro     EnclaveProvider = "AWS_Nitro"
	EnclaveProviderAzureSGX     EnclaveProvider = "Azure_SGX"
	EnclaveProviderGoogleSEV    EnclaveProvider = "Google_SEV"
	EnclaveProviderIntelSGX     EnclaveProvider = "Intel_SGX"
)
