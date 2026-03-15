package wallet_vo

import (
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// BTCAddressType represents the format/type of a Bitcoin address
type BTCAddressType string

const (
	BTCAddressTypeP2PKH   BTCAddressType = "P2PKH"   // Legacy (starts with 1)
	BTCAddressTypeP2SH    BTCAddressType = "P2SH"    // Pay-to-Script-Hash (starts with 3)
	BTCAddressTypeBech32  BTCAddressType = "Bech32"   // SegWit v0 (starts with bc1q)
	BTCAddressTypeTaproot BTCAddressType = "Taproot"  // SegWit v1 / Taproot (starts with bc1p)
	BTCAddressTypeUnknown BTCAddressType = "Unknown"
)

// BTCAddress represents a validated Bitcoin address
type BTCAddress struct {
	address     string
	addressType BTCAddressType
	isTestnet   bool
}

var (
	// Mainnet address patterns
	p2pkhMainnetRegex  = regexp.MustCompile(`^1[a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	p2shMainnetRegex   = regexp.MustCompile(`^3[a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	bech32MainnetRegex = regexp.MustCompile(`^bc1q[a-z0-9]{38,58}$`)
	taprootMainnetRegex = regexp.MustCompile(`^bc1p[a-z0-9]{58}$`)

	// Testnet address patterns
	p2pkhTestnetRegex  = regexp.MustCompile(`^[mn][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	p2shTestnetRegex   = regexp.MustCompile(`^2[a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	bech32TestnetRegex = regexp.MustCompile(`^tb1q[a-z0-9]{38,58}$`)
	taprootTestnetRegex = regexp.MustCompile(`^tb1p[a-z0-9]{58}$`)

	// Signet patterns (same prefix as testnet)
	bech32SignetRegex  = regexp.MustCompile(`^tb1q[a-z0-9]{38,58}$`)
	taprootSignetRegex = regexp.MustCompile(`^tb1p[a-z0-9]{58}$`)
)

// NewBTCAddress creates a new validated Bitcoin address
func NewBTCAddress(address string) (BTCAddress, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return BTCAddress{}, fmt.Errorf("bitcoin address cannot be empty")
	}

	addrType, isTestnet, err := classifyBTCAddress(address)
	if err != nil {
		return BTCAddress{}, err
	}

	return BTCAddress{
		address:     address,
		addressType: addrType,
		isTestnet:   isTestnet,
	}, nil
}

// classifyBTCAddress determines the type and network of a Bitcoin address
func classifyBTCAddress(address string) (BTCAddressType, bool, error) {
	// Mainnet
	if taprootMainnetRegex.MatchString(address) {
		return BTCAddressTypeTaproot, false, nil
	}
	if bech32MainnetRegex.MatchString(address) {
		return BTCAddressTypeBech32, false, nil
	}
	if p2pkhMainnetRegex.MatchString(address) {
		return BTCAddressTypeP2PKH, false, nil
	}
	if p2shMainnetRegex.MatchString(address) {
		return BTCAddressTypeP2SH, false, nil
	}

	// Testnet / Signet
	if taprootTestnetRegex.MatchString(address) || taprootSignetRegex.MatchString(address) {
		return BTCAddressTypeTaproot, true, nil
	}
	if bech32TestnetRegex.MatchString(address) || bech32SignetRegex.MatchString(address) {
		return BTCAddressTypeBech32, true, nil
	}
	if p2pkhTestnetRegex.MatchString(address) {
		return BTCAddressTypeP2PKH, true, nil
	}
	if p2shTestnetRegex.MatchString(address) {
		return BTCAddressTypeP2SH, true, nil
	}

	return BTCAddressTypeUnknown, false, fmt.Errorf("invalid Bitcoin address format: %s", address)
}

// String returns the address string
func (a BTCAddress) String() string {
	return a.address
}

// Type returns the address format type
func (a BTCAddress) Type() BTCAddressType {
	return a.addressType
}

// IsMainnet returns true if this is a mainnet address
func (a BTCAddress) IsMainnet() bool {
	return !a.isTestnet
}

// IsTestnet returns true if this is a testnet/signet address
func (a BTCAddress) IsTestnet() bool {
	return a.isTestnet
}

// IsValid returns true if the address is valid
func (a BTCAddress) IsValid() bool {
	return a.address != "" && a.addressType != BTCAddressTypeUnknown
}

// IsTaproot returns true if this is a Taproot (P2TR) address
func (a BTCAddress) IsTaproot() bool {
	return a.addressType == BTCAddressTypeTaproot
}

// IsSegWit returns true if this is a SegWit address (bech32 or taproot)
func (a BTCAddress) IsSegWit() bool {
	return a.addressType == BTCAddressTypeBech32 || a.addressType == BTCAddressTypeTaproot
}

// IsEmpty returns true if the address is empty/unset
func (a BTCAddress) IsEmpty() bool {
	return a.address == ""
}

// Equals checks if two addresses are equal
func (a BTCAddress) Equals(other BTCAddress) bool {
	return a.address == other.address
}

// MarshalJSON implements json.Marshaler
func (a BTCAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, a.address)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (a *BTCAddress) UnmarshalJSON(data []byte) error {
	address := strings.Trim(string(data), `"`)
	if address == "" {
		*a = BTCAddress{}
		return nil
	}
	parsed, err := NewBTCAddress(address)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// MarshalBSON implements bson.Marshaler
func (a BTCAddress) MarshalBSON() ([]byte, error) {
	return bson.Marshal(a.address)
}

// UnmarshalBSON implements bson.Unmarshaler
func (a *BTCAddress) UnmarshalBSON(data []byte) error {
	var address string
	if err := bson.Unmarshal(data, &address); err != nil {
		return err
	}
	if address == "" {
		*a = BTCAddress{}
		return nil
	}
	parsed, err := NewBTCAddress(address)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// MarshalBSONValue implements bson.ValueMarshaler
func (a BTCAddress) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(a.address)
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler
func (a *BTCAddress) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t != bsontype.String {
		return fmt.Errorf("invalid BSON type for BTCAddress: %s", t)
	}
	var address string
	rv := bson.RawValue{Type: t, Value: data}
	if err := rv.Unmarshal(&address); err != nil {
		return err
	}
	if address == "" {
		*a = BTCAddress{}
		return nil
	}
	parsed, err := NewBTCAddress(address)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}
