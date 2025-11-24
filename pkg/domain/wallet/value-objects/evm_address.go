package wallet_vo

import (
	"fmt"
	"regexp"
	"strings"
)

// EVMAddress represents an Ethereum Virtual Machine compatible address
type EVMAddress struct {
	address string
}

var (
	// evmAddressRegex validates Ethereum address format (0x + 40 hex chars)
	evmAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// NewEVMAddress creates a new EVM address with validation
func NewEVMAddress(address string) (EVMAddress, error) {
	// Normalize address
	address = strings.TrimSpace(address)

	if !evmAddressRegex.MatchString(address) {
		return EVMAddress{}, fmt.Errorf("invalid EVM address format: %s (expected 0x + 40 hex characters)", address)
	}

	// Convert to checksum format (simple version - in production use ethereum package)
	checksummed := checksumAddress(address)

	return EVMAddress{address: checksummed}, nil
}

// String returns the string representation of the address
func (e EVMAddress) String() string {
	return e.address
}

// IsValid checks if the address is valid
func (e EVMAddress) IsValid() bool {
	return evmAddressRegex.MatchString(e.address)
}

// Equals checks if two addresses are equal
func (e EVMAddress) Equals(other EVMAddress) bool {
	return strings.EqualFold(e.address, other.address)
}

// IsZero checks if this is the zero address
func (e EVMAddress) IsZero() bool {
	return e.address == "0x0000000000000000000000000000000000000000"
}

// checksumAddress converts an address to EIP-55 checksum format (simplified)
// In production, use github.com/ethereum/go-ethereum/common for proper checksumming
func checksumAddress(address string) string {
	// For now, return lowercase (proper implementation would use Keccak256)
	return strings.ToLower(address)
}

// MarshalJSON implements json.Marshaler
func (e EVMAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.address)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (e *EVMAddress) UnmarshalJSON(data []byte) error {
	// Remove quotes
	address := strings.Trim(string(data), `"`)

	parsed, err := NewEVMAddress(address)
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}
