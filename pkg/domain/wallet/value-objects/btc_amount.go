package wallet_vo

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// BtcAmount represents a Bitcoin amount with satoshi precision (8 decimal places)
// Uses satoshis (int64) internally to avoid floating-point errors
// CRITICAL: All Bitcoin financial calculations MUST use this type — never raw float64
type BtcAmount struct {
	satoshis int64 // Amount in satoshis (e.g., 0.001 BTC = 100_000 satoshis)
}

const (
	satoshisPerBTC = 100_000_000 // 1 BTC = 100,000,000 satoshis

	// MaxBtcAmountSatoshis is the maximum representable BTC amount (~21 million BTC)
	// This matches Bitcoin's theoretical supply cap
	MaxBtcAmountSatoshis int64 = 2_100_000_000_000_000 // 21,000,000.00000000 BTC

	// MinBtcAmountSatoshis is the minimum non-zero amount (1 satoshi = 0.00000001 BTC)
	MinBtcAmountSatoshis int64 = 1
)

// NewBtcAmount creates a new BtcAmount from BTC (as float)
// Clamps to MaxBtcAmountSatoshis if exceeded
func NewBtcAmount(btc float64) BtcAmount {
	sats := int64(math.Round(btc * float64(satoshisPerBTC)))
	if sats > MaxBtcAmountSatoshis || sats < -MaxBtcAmountSatoshis {
		if sats > 0 {
			sats = MaxBtcAmountSatoshis
		} else {
			sats = -MaxBtcAmountSatoshis
		}
	}
	return BtcAmount{satoshis: sats}
}

// NewBtcAmountSafe creates a new BtcAmount with error checking
func NewBtcAmountSafe(btc float64) (BtcAmount, error) {
	if math.IsNaN(btc) || math.IsInf(btc, 0) {
		return BtcAmount{}, fmt.Errorf("invalid BTC amount: %v", btc)
	}
	sats := int64(math.Round(btc * float64(satoshisPerBTC)))
	if sats > MaxBtcAmountSatoshis || sats < -MaxBtcAmountSatoshis {
		return BtcAmount{}, fmt.Errorf("BTC amount %f exceeds maximum allowed (%.8f BTC)", btc, float64(MaxBtcAmountSatoshis)/float64(satoshisPerBTC))
	}
	return BtcAmount{satoshis: sats}, nil
}

// NewBtcAmountFromSatoshis creates a new BtcAmount from satoshis (exact)
func NewBtcAmountFromSatoshis(sats int64) BtcAmount {
	return BtcAmount{satoshis: sats}
}

// NewBtcAmountFromString creates a new BtcAmount from a string representation
func NewBtcAmountFromString(s string) (BtcAmount, error) {
	btc, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return BtcAmount{}, fmt.Errorf("invalid BTC amount string: %s", s)
	}
	return NewBtcAmount(btc), nil
}

// ToBTC returns the amount in BTC (as float)
func (a BtcAmount) ToBTC() float64 {
	return float64(a.satoshis) / float64(satoshisPerBTC)
}

// Satoshis returns the amount in satoshis (exact)
func (a BtcAmount) Satoshis() int64 {
	return a.satoshis
}

// String returns a formatted string (e.g., "₿0.00100000")
func (a BtcAmount) String() string {
	return fmt.Sprintf("₿%.8f", a.ToBTC())
}

// ToFloat64 returns the amount as a float64 (BTC)
func (a BtcAmount) ToFloat64() float64 {
	return a.ToBTC()
}

// ToUSD converts this BTC amount to USD cents given a price per BTC
func (a BtcAmount) ToUSD(btcPriceUSD float64) Amount {
	usdValue := a.ToBTC() * btcPriceUSD
	return NewAmount(usdValue)
}

// Add adds two BTC amounts
func (a BtcAmount) Add(other BtcAmount) BtcAmount {
	return BtcAmount{satoshis: a.satoshis + other.satoshis}
}

// Subtract subtracts another BTC amount
func (a BtcAmount) Subtract(other BtcAmount) BtcAmount {
	return BtcAmount{satoshis: a.satoshis - other.satoshis}
}

// Multiply multiplies by a factor
func (a BtcAmount) Multiply(factor float64) BtcAmount {
	result := float64(a.satoshis) * factor
	return BtcAmount{satoshis: int64(math.Round(result))}
}

// Divide divides by a divisor (returns zero if divisor is zero)
func (a BtcAmount) Divide(divisor float64) BtcAmount {
	if divisor == 0 {
		return BtcAmount{satoshis: 0}
	}
	result := float64(a.satoshis) / divisor
	return BtcAmount{satoshis: int64(math.Round(result))}
}

// Percentage calculates a percentage of the BTC amount
func (a BtcAmount) Percentage(percent float64) BtcAmount {
	return a.Multiply(percent / 100.0)
}

// IsZero checks if the amount is zero
func (a BtcAmount) IsZero() bool {
	return a.satoshis == 0
}

// IsNegative checks if the amount is negative
func (a BtcAmount) IsNegative() bool {
	return a.satoshis < 0
}

// IsPositive checks if the amount is positive
func (a BtcAmount) IsPositive() bool {
	return a.satoshis > 0
}

// GreaterThan checks if this amount is greater than another
func (a BtcAmount) GreaterThan(other BtcAmount) bool {
	return a.satoshis > other.satoshis
}

// GreaterThanOrEqual checks if this amount is >= another
func (a BtcAmount) GreaterThanOrEqual(other BtcAmount) bool {
	return a.satoshis >= other.satoshis
}

// LessThan checks if this amount is less than another
func (a BtcAmount) LessThan(other BtcAmount) bool {
	return a.satoshis < other.satoshis
}

// LessThanOrEqual checks if this amount is <= another
func (a BtcAmount) LessThanOrEqual(other BtcAmount) bool {
	return a.satoshis <= other.satoshis
}

// Equals checks if two BTC amounts are equal
func (a BtcAmount) Equals(other BtcAmount) bool {
	return a.satoshis == other.satoshis
}

// Abs returns the absolute value
func (a BtcAmount) Abs() BtcAmount {
	if a.satoshis < 0 {
		return BtcAmount{satoshis: -a.satoshis}
	}
	return a
}

// MarshalJSON implements json.Marshaler — serializes as "0.00100000" string
func (a BtcAmount) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%.8f"`, a.ToBTC())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (a *BtcAmount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := NewBtcAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*a = NewBtcAmount(f)
	return nil
}

// MarshalBSON implements bson.Marshaler
func (a BtcAmount) MarshalBSON() ([]byte, error) {
	return bson.Marshal(bsonBtcAmount{Satoshis: a.satoshis})
}

// UnmarshalBSON implements bson.Unmarshaler
func (a *BtcAmount) UnmarshalBSON(data []byte) error {
	var ba bsonBtcAmount
	if err := bson.Unmarshal(data, &ba); err == nil && ba.Satoshis != 0 {
		a.satoshis = ba.Satoshis
		return nil
	}

	var i int64
	if err := bson.Unmarshal(data, &i); err == nil {
		a.satoshis = i
		return nil
	}

	var f float64
	if err := bson.Unmarshal(data, &f); err == nil {
		*a = NewBtcAmount(f)
		return nil
	}

	var s string
	if err := bson.Unmarshal(data, &s); err == nil {
		parsed, err := NewBtcAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	}

	return fmt.Errorf("cannot unmarshal BtcAmount from BSON")
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler for flexible deserialization
func (a *BtcAmount) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	rv := bson.RawValue{Type: t, Value: data}
	switch t {
	case bsontype.Int64:
		var i int64
		if err := rv.Unmarshal(&i); err != nil {
			return err
		}
		a.satoshis = i
		return nil

	case bsontype.Int32:
		var i int32
		if err := rv.Unmarshal(&i); err != nil {
			return err
		}
		a.satoshis = int64(i)
		return nil

	case bsontype.Double:
		var f float64
		if err := rv.Unmarshal(&f); err != nil {
			return err
		}
		*a = NewBtcAmount(f)
		return nil

	case bsontype.String:
		var s string
		if err := rv.Unmarshal(&s); err != nil {
			return err
		}
		parsed, err := NewBtcAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil

	case bsontype.EmbeddedDocument:
		var ba bsonBtcAmount
		if err := rv.Unmarshal(&ba); err != nil {
			return err
		}
		a.satoshis = ba.Satoshis
		return nil

	default:
		return fmt.Errorf("cannot unmarshal BtcAmount from BSON type %s", t)
	}
}

// bsonBtcAmount is used for BSON marshaling/unmarshaling
type bsonBtcAmount struct {
	Satoshis int64 `bson:"satoshis"`
}

// FromUSD converts a USD amount to BTC given a price per BTC
func FromUSD(usd Amount, btcPriceUSD float64) BtcAmount {
	if btcPriceUSD <= 0 {
		return BtcAmount{}
	}
	btcValue := usd.Dollars() / btcPriceUSD
	return NewBtcAmount(btcValue)
}
