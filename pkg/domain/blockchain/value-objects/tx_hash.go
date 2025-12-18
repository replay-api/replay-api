package blockchain_vo

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// TxHash represents an Ethereum transaction hash (32 bytes)
type TxHash struct {
	hash [32]byte
}

// NewTxHash creates a TxHash from a hex string (with or without 0x prefix)
func NewTxHash(hashStr string) (TxHash, error) {
	hashStr = strings.TrimPrefix(hashStr, "0x")

	if len(hashStr) != 64 {
		return TxHash{}, fmt.Errorf("invalid tx hash length: expected 64 hex chars, got %d", len(hashStr))
	}

	bytes, err := hex.DecodeString(hashStr)
	if err != nil {
		return TxHash{}, fmt.Errorf("invalid hex string: %w", err)
	}

	var hash [32]byte
	copy(hash[:], bytes)

	return TxHash{hash: hash}, nil
}

// NewTxHashFromBytes creates a TxHash from bytes
func NewTxHashFromBytes(bytes []byte) (TxHash, error) {
	if len(bytes) != 32 {
		return TxHash{}, fmt.Errorf("invalid tx hash length: expected 32 bytes, got %d", len(bytes))
	}

	var hash [32]byte
	copy(hash[:], bytes)

	return TxHash{hash: hash}, nil
}

// String returns the hex string representation with 0x prefix
func (t TxHash) String() string {
	return "0x" + hex.EncodeToString(t.hash[:])
}

// Bytes returns the raw bytes
func (t TxHash) Bytes() []byte {
	return t.hash[:]
}

// IsZero checks if this is the zero hash
func (t TxHash) IsZero() bool {
	for _, b := range t.hash {
		if b != 0 {
			return false
		}
	}
	return true
}

// Equals checks if two hashes are equal
func (t TxHash) Equals(other TxHash) bool {
	return t.hash == other.hash
}

// MarshalJSON implements json.Marshaler
func (t TxHash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TxHash) UnmarshalJSON(data []byte) error {
	hashStr := strings.Trim(string(data), `"`)
	parsed, err := NewTxHash(hashStr)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// BlockHash represents an Ethereum block hash (same structure as TxHash)
type BlockHash = TxHash

// NewBlockHash creates a BlockHash from a hex string
func NewBlockHash(hashStr string) (BlockHash, error) {
	return NewTxHash(hashStr)
}
