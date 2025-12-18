package blockchain_vo

import (
	"encoding/json"
	"testing"
)

// TestNewTxHash_ValidHex verifies valid hex string parsing
func TestNewTxHash_ValidHex(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"With 0x prefix", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Without prefix", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Uppercase hex", "0xABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890"},
		{"Mixed case", "0xAbCdEf1234567890abcdef1234567890ABCDEF1234567890abcdef1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewTxHash(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if hash.IsZero() {
				t.Error("Expected non-zero hash")
			}
		})
	}
}

// TestNewTxHash_InvalidInput verifies error handling for invalid inputs
func TestNewTxHash_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Too short", "0x1234"},
		{"Too long", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
		{"Invalid characters", "0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG"},
		{"Empty string", ""},
		{"Just prefix", "0x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTxHash(tt.input)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestNewTxHashFromBytes verifies byte array parsing
func TestNewTxHashFromBytes(t *testing.T) {
	validBytes := make([]byte, 32)
	for i := range validBytes {
		validBytes[i] = byte(i)
	}

	hash, err := NewTxHashFromBytes(validBytes)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if hash.IsZero() {
		t.Error("Expected non-zero hash")
	}

	// Verify bytes match
	resultBytes := hash.Bytes()
	for i, b := range resultBytes {
		if b != validBytes[i] {
			t.Errorf("Byte mismatch at %d: got %d, want %d", i, b, validBytes[i])
		}
	}
}

// TestNewTxHashFromBytes_InvalidLength verifies length validation
func TestNewTxHashFromBytes_InvalidLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Too short", 16},
		{"Too long", 64},
		{"Empty", 0},
		{"One byte", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := make([]byte, tt.length)
			_, err := NewTxHashFromBytes(bytes)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestTxHash_String verifies hex string output
func TestTxHash_String(t *testing.T) {
	input := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	hash, _ := NewTxHash(input)

	result := hash.String()

	if result != input {
		t.Errorf("String() = %s, want %s", result, input)
	}
}

// TestTxHash_IsZero verifies zero hash detection
func TestTxHash_IsZero(t *testing.T) {
	zeroHash := TxHash{}
	if !zeroHash.IsZero() {
		t.Error("Expected zero hash to return true")
	}

	nonZeroHash, _ := NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if nonZeroHash.IsZero() {
		t.Error("Expected non-zero hash to return false")
	}
}

// TestTxHash_Equals verifies hash equality comparison
func TestTxHash_Equals(t *testing.T) {
	hash1, _ := NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	hash2, _ := NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	hash3, _ := NewTxHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	if !hash1.Equals(hash2) {
		t.Error("Expected equal hashes to return true")
	}

	if hash1.Equals(hash3) {
		t.Error("Expected different hashes to return false")
	}
}

// TestTxHash_MarshalJSON verifies JSON serialization
func TestTxHash_MarshalJSON(t *testing.T) {
	hash, _ := NewTxHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	data, err := json.Marshal(hash)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"`
	if string(data) != expected {
		t.Errorf("MarshalJSON = %s, want %s", string(data), expected)
	}
}

// TestTxHash_UnmarshalJSON verifies JSON deserialization
func TestTxHash_UnmarshalJSON(t *testing.T) {
	jsonData := `"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"`

	var hash TxHash
	err := json.Unmarshal([]byte(jsonData), &hash)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	expected := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	if hash.String() != expected {
		t.Errorf("UnmarshalJSON result = %s, want %s", hash.String(), expected)
	}
}

// TestTxHash_UnmarshalJSON_Invalid verifies JSON error handling
func TestTxHash_UnmarshalJSON_Invalid(t *testing.T) {
	invalidJSON := `"0xinvalid"`

	var hash TxHash
	err := json.Unmarshal([]byte(invalidJSON), &hash)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestNewBlockHash verifies BlockHash alias works
func TestNewBlockHash(t *testing.T) {
	input := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	hash, err := NewBlockHash(input)
	if err != nil {
		t.Fatalf("NewBlockHash failed: %v", err)
	}

	if hash.String() != input {
		t.Errorf("BlockHash String() = %s, want %s", hash.String(), input)
	}
}

// TestTxHash_Deterministic verifies same input produces same output
func TestTxHash_Deterministic(t *testing.T) {
	input := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	hash1, _ := NewTxHash(input)
	hash2, _ := NewTxHash(input)

	if !hash1.Equals(hash2) {
		t.Error("Same input should produce identical hashes")
	}
}
