package custody_vo

import (
	"testing"
)

// TestNewKeyShareID verifies deterministic key share ID generation
func TestNewKeyShareID(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		shareIndex uint8
	}{
		{"First share", "key-001", 0},
		{"Second share", "key-001", 1},
		{"Third share", "key-001", 2},
		{"Different key", "key-002", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewKeyShareID(tt.keyID, tt.shareIndex)
			if id == "" {
				t.Error("Expected non-empty key share ID")
			}

			// Verify determinism - same inputs should produce same output
			id2 := NewKeyShareID(tt.keyID, tt.shareIndex)
			if id != id2 {
				t.Errorf("KeyShareID not deterministic: %s != %s", id, id2)
			}
		})
	}
}

// TestNewKeyShareID_Uniqueness verifies different inputs produce different IDs
func TestNewKeyShareID_Uniqueness(t *testing.T) {
	id1 := NewKeyShareID("key-001", 0)
	id2 := NewKeyShareID("key-001", 1)
	id3 := NewKeyShareID("key-002", 0)

	if id1 == id2 {
		t.Error("Different share indices should produce different IDs")
	}
	if id1 == id3 {
		t.Error("Different key IDs should produce different share IDs")
	}
}

// TestThresholdConfig_Presets verifies threshold presets
func TestThresholdConfig_Presets(t *testing.T) {
	tests := []struct {
		name      string
		config    ThresholdConfig
		threshold uint8
		total     uint8
	}{
		{"2-of-3", Threshold2of3, 2, 3},
		{"3-of-5", Threshold3of5, 3, 5},
		{"4-of-7", Threshold4of7, 4, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Threshold != tt.threshold {
				t.Errorf("Threshold = %d, want %d", tt.config.Threshold, tt.threshold)
			}
			if tt.config.TotalShares != tt.total {
				t.Errorf("TotalShares = %d, want %d", tt.config.TotalShares, tt.total)
			}
		})
	}
}

// TestMPCScheme_Constants verifies MPC scheme constants
func TestMPCScheme_Constants(t *testing.T) {
	schemes := []MPCScheme{
		MPCSchemeGG20,
		MPCSchemeCMP,
		MPCSchemeFROST,
		MPCSchemeLindell,
		MPCSchemeFROSTEd25519,
	}

	for _, scheme := range schemes {
		if scheme == "" {
			t.Error("MPC scheme should not be empty")
		}
	}

	// Verify specific values
	if MPCSchemeGG20 != "GG20" {
		t.Errorf("MPCSchemeGG20 = %s, want GG20", MPCSchemeGG20)
	}
	if MPCSchemeFROSTEd25519 != "FROST-Ed25519" {
		t.Errorf("MPCSchemeFROSTEd25519 = %s, want FROST-Ed25519", MPCSchemeFROSTEd25519)
	}
}

// TestKeyCurve_Constants verifies key curve constants
func TestKeyCurve_Constants(t *testing.T) {
	tests := []struct {
		curve    KeyCurve
		expected string
	}{
		{CurveSecp256k1, "secp256k1"},
		{CurveEd25519, "ed25519"},
		{CurveP256, "P-256"},
	}

	for _, tt := range tests {
		t.Run(string(tt.curve), func(t *testing.T) {
			if string(tt.curve) != tt.expected {
				t.Errorf("Curve = %s, want %s", tt.curve, tt.expected)
			}
		})
	}
}

// TestKeyShareLocation_Constants verifies key share location constants
func TestKeyShareLocation_Constants(t *testing.T) {
	locations := []KeyShareLocation{
		LocationHSM,
		LocationSecureEnclave,
		LocationKMS,
		LocationUserDevice,
		LocationColdStorage,
	}

	seen := make(map[KeyShareLocation]bool)
	for _, loc := range locations {
		if loc == "" {
			t.Error("KeyShareLocation should not be empty")
		}
		if seen[loc] {
			t.Errorf("Duplicate KeyShareLocation: %s", loc)
		}
		seen[loc] = true
	}
}

// TestKeyPurpose_Constants verifies key purpose constants
func TestKeyPurpose_Constants(t *testing.T) {
	purposes := []KeyPurpose{
		KeyPurposeTransaction,
		KeyPurposeRecovery,
		KeyPurposeAdmin,
		KeyPurposeRotation,
	}

	for _, purpose := range purposes {
		if purpose == "" {
			t.Error("KeyPurpose should not be empty")
		}
	}
}

// TestSigningSessionState_Constants verifies signing session states
func TestSigningSessionState_Constants(t *testing.T) {
	states := []SigningSessionState{
		SigningStateInitiated,
		SigningStateRound1,
		SigningStateRound2,
		SigningStateRound3,
		SigningStateCompleted,
		SigningStateFailed,
		SigningStateExpired,
	}

	seen := make(map[SigningSessionState]bool)
	for _, state := range states {
		if state == "" {
			t.Error("SigningSessionState should not be empty")
		}
		if seen[state] {
			t.Errorf("Duplicate SigningSessionState: %s", state)
		}
		seen[state] = true
	}
}

// TestMessageType_Constants verifies message type constants
func TestMessageType_Constants(t *testing.T) {
	types := []MessageType{
		MessageTypeTransaction,
		MessageTypeTypedData,
		MessageTypePersonalSign,
		MessageTypeSolanaMessage,
	}

	seen := make(map[MessageType]bool)
	for _, mt := range types {
		if mt == "" {
			t.Error("MessageType should not be empty")
		}
		if seen[mt] {
			t.Errorf("Duplicate MessageType: %s", mt)
		}
		seen[mt] = true
	}
}

// TestHSMProvider_Constants verifies HSM provider constants
func TestHSMProvider_Constants(t *testing.T) {
	providers := []HSMProvider{
		HSMProviderAWSCloudHSM,
		HSMProviderAzureHSM,
		HSMProviderGoogleHSM,
		HSMProviderThales,
		HSMProviderYubico,
	}

	for _, provider := range providers {
		if provider == "" {
			t.Error("HSMProvider should not be empty")
		}
	}
}

// TestEnclaveProvider_Constants verifies enclave provider constants
func TestEnclaveProvider_Constants(t *testing.T) {
	providers := []EnclaveProvider{
		EnclaveProviderAWSNitro,
		EnclaveProviderAzureSGX,
		EnclaveProviderGoogleSEV,
		EnclaveProviderIntelSGX,
	}

	for _, provider := range providers {
		if provider == "" {
			t.Error("EnclaveProvider should not be empty")
		}
	}
}

// TestMPCKeyGenRequest_Structure verifies request structure
func TestMPCKeyGenRequest_Structure(t *testing.T) {
	req := MPCKeyGenRequest{
		KeyID:          "test-key-001",
		Curve:          CurveSecp256k1,
		Scheme:         MPCSchemeCMP,
		Threshold:      Threshold2of3,
		ShareLocations: []KeyShareLocation{LocationHSM, LocationSecureEnclave, LocationUserDevice},
		InitiatorID:    "user-001",
		Purpose:        KeyPurposeTransaction,
	}

	if req.KeyID == "" {
		t.Error("KeyID should not be empty")
	}
	if len(req.ShareLocations) != 3 {
		t.Errorf("Expected 3 share locations, got %d", len(req.ShareLocations))
	}
	if req.Threshold.Threshold != 2 {
		t.Errorf("Expected threshold 2, got %d", req.Threshold.Threshold)
	}
}

// TestSigningRequest_Structure verifies signing request structure
func TestSigningRequest_Structure(t *testing.T) {
	hash := make([]byte, 32)
	req := SigningRequest{
		SessionID:    "session-001",
		KeyID:        "key-001",
		MessageHash:  hash,
		MessageType:  MessageTypeTransaction,
		ChainID:      ChainPolygon,
		Participants: []KeyShareID{NewKeyShareID("key-001", 0), NewKeyShareID("key-001", 1)},
		RequestedBy:  "user-001",
	}

	if len(req.MessageHash) != 32 {
		t.Errorf("MessageHash should be 32 bytes, got %d", len(req.MessageHash))
	}
	if len(req.Participants) < 2 {
		t.Error("Expected at least 2 participants for signing")
	}
}

// TestKeyShareMetadata_Structure verifies metadata structure
func TestKeyShareMetadata_Structure(t *testing.T) {
	meta := KeyShareMetadata{
		ShareID:       NewKeyShareID("key-001", 0),
		ShareIndex:    0,
		Location:      LocationHSM,
		ProviderID:    "hsm-cluster-001",
		EncryptionKey: "arn:aws:kms:us-east-1:...",
		IsActive:      true,
	}

	if meta.ShareID == "" {
		t.Error("ShareID should not be empty")
	}
	if !meta.IsActive {
		t.Error("Expected share to be active")
	}
}

// TestHSMConfig_Structure verifies HSM config structure
func TestHSMConfig_Structure(t *testing.T) {
	config := HSMConfig{
		Provider:  HSMProviderAWSCloudHSM,
		ClusterID: "cluster-001",
		Region:    "us-east-1",
		KeyAlias:  "leet-wallet-master",
		Credentials: HSMCredentials{
			Type:    "iam",
			IAMRole: "arn:aws:iam::123456789:role/hsm-access",
		},
	}

	if config.Provider != HSMProviderAWSCloudHSM {
		t.Errorf("Expected AWS CloudHSM provider, got %s", config.Provider)
	}
	if config.Credentials.Type != "iam" {
		t.Errorf("Expected IAM credential type, got %s", config.Credentials.Type)
	}
}

// TestSecureEnclaveConfig_Structure verifies enclave config structure
func TestSecureEnclaveConfig_Structure(t *testing.T) {
	config := SecureEnclaveConfig{
		Provider:       EnclaveProviderAWSNitro,
		AttestationURL: "https://attestation.us-east-1.aws.nitro-enclaves.amazonaws.com",
		EnclaveID:      "enclave-001",
		PCRValues: map[int]string{
			0: "abc123...",
			1: "def456...",
			2: "ghi789...",
		},
	}

	if config.Provider != EnclaveProviderAWSNitro {
		t.Errorf("Expected AWS Nitro provider, got %s", config.Provider)
	}
	if len(config.PCRValues) < 3 {
		t.Error("Expected at least 3 PCR values")
	}
}
