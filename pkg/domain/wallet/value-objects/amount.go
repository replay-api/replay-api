package wallet_vo

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Amount represents a monetary amount with fixed-point precision
// Uses cents (int64) internally to avoid floating-point errors
// CRITICAL: All financial calculations MUST use this type — never raw float64
type Amount struct {
	cents int64 // Amount in cents (e.g., $10.50 = 1050 cents)
}

const (
	centsPerDollar = 100

	// MaxAmountCents is the maximum representable amount (100 million dollars)
	// This prevents overflow in arithmetic operations
	MaxAmountCents int64 = 10_000_000_000 // $100,000,000.00

	// MinAmountCents is the minimum non-zero amount (1 cent)
	MinAmountCents int64 = 1
)

// NewAmount creates a new Amount from dollars (as float)
// Panics if the resulting amount exceeds MaxAmountCents (programming error)
func NewAmount(dollars float64) Amount {
	cents := int64(math.Round(dollars * float64(centsPerDollar)))
	if cents > MaxAmountCents || cents < -MaxAmountCents {
		// Clamp to max — callers should validate amounts before reaching here
		if cents > 0 {
			cents = MaxAmountCents
		} else {
			cents = -MaxAmountCents
		}
	}
	return Amount{cents: cents}
}

// NewAmountSafe creates a new Amount from dollars with error checking
func NewAmountSafe(dollars float64) (Amount, error) {
	if math.IsNaN(dollars) || math.IsInf(dollars, 0) {
		return Amount{}, fmt.Errorf("invalid amount: %v", dollars)
	}
	cents := int64(math.Round(dollars * float64(centsPerDollar)))
	if cents > MaxAmountCents || cents < -MaxAmountCents {
		return Amount{}, fmt.Errorf("amount %f exceeds maximum allowed ($%.2f)", dollars, float64(MaxAmountCents)/float64(centsPerDollar))
	}
	return Amount{cents: cents}, nil
}

// NewAmountFromCents creates a new Amount from cents (exact)
func NewAmountFromCents(cents int64) Amount {
	return Amount{cents: cents}
}

// NewAmountFromString creates a new Amount from a string representation
func NewAmountFromString(s string) (Amount, error) {
	dollars, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Amount{}, fmt.Errorf("invalid amount string: %s", s)
	}
	return NewAmount(dollars), nil
}

// Dollars returns the amount in dollars (as float)
func (a Amount) Dollars() float64 {
	return float64(a.cents) / float64(centsPerDollar)
}

// Cents returns the amount in cents (exact)
func (a Amount) Cents() int64 {
	return a.cents
}

// String returns a string representation (e.g., "$10.50")
func (a Amount) String() string {
	return fmt.Sprintf("$%.2f", a.Dollars())
}

// ToFloat returns the amount as a float64
func (a Amount) ToFloat() float64 {
	return a.Dollars()
}

// ToFloat64 returns the amount as a float64 (alias for ToFloat)
func (a Amount) ToFloat64() float64 {
	return a.Dollars()
}

// ToCents returns the amount in cents (alias for Cents)
func (a Amount) ToCents() int64 {
	return a.cents
}

// Add adds two amounts
func (a Amount) Add(other Amount) Amount {
	return Amount{cents: a.cents + other.cents}
}

// Subtract subtracts another amount
func (a Amount) Subtract(other Amount) Amount {
	return Amount{cents: a.cents - other.cents}
}

// Multiply multiplies by a factor
func (a Amount) Multiply(factor float64) Amount {
	result := float64(a.cents) * factor
	return Amount{cents: int64(math.Round(result))}
}

// Divide divides by a divisor
// Returns zero amount if divisor is zero (callers should validate)
// For safe division with error handling, use DivideSafe
func (a Amount) Divide(divisor float64) Amount {
	if divisor == 0 {
		return Amount{cents: 0} // Defensive: prevent panic, but caller should validate
	}
	result := float64(a.cents) / divisor
	return Amount{cents: int64(math.Round(result))}
}

// DivideSafe divides by a divisor with error handling for zero division
func (a Amount) DivideSafe(divisor float64) (Amount, error) {
	if divisor == 0 {
		return Amount{}, fmt.Errorf("division by zero")
	}
	if math.IsNaN(divisor) || math.IsInf(divisor, 0) {
		return Amount{}, fmt.Errorf("invalid divisor: %v", divisor)
	}
	result := float64(a.cents) / divisor
	return Amount{cents: int64(math.Round(result))}, nil
}

// IsZero checks if the amount is zero
func (a Amount) IsZero() bool {
	return a.cents == 0
}

// IsNegative checks if the amount is negative
func (a Amount) IsNegative() bool {
	return a.cents < 0
}

// IsPositive checks if the amount is positive
func (a Amount) IsPositive() bool {
	return a.cents > 0
}

// GreaterThan checks if this amount is greater than another
func (a Amount) GreaterThan(other Amount) bool {
	return a.cents > other.cents
}

// GreaterThanOrEqual checks if this amount is greater than or equal to another
func (a Amount) GreaterThanOrEqual(other Amount) bool {
	return a.cents >= other.cents
}

// LessThan checks if this amount is less than another
func (a Amount) LessThan(other Amount) bool {
	return a.cents < other.cents
}

// LessThanOrEqual checks if this amount is less than or equal to another
func (a Amount) LessThanOrEqual(other Amount) bool {
	return a.cents <= other.cents
}

// Equals checks if two amounts are equal
func (a Amount) Equals(other Amount) bool {
	return a.cents == other.cents
}

// Abs returns the absolute value
func (a Amount) Abs() Amount {
	if a.cents < 0 {
		return Amount{cents: -a.cents}
	}
	return a
}

// Percentage calculates a percentage of the amount
func (a Amount) Percentage(percent float64) Amount {
	return a.Multiply(percent / 100.0)
}

// MarshalJSON implements json.Marshaler
func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%.2f"`, a.Dollars())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (a *Amount) UnmarshalJSON(data []byte) error {
	// Try to parse as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := NewAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	}

	// Try to parse as number
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*a = NewAmount(f)
	return nil
}

// MarshalBSON implements bson.Marshaler
func (a Amount) MarshalBSON() ([]byte, error) {
	return bson.Marshal(bsonAmount{Cents: a.cents})
}

// UnmarshalBSON implements bson.Unmarshaler
func (a *Amount) UnmarshalBSON(data []byte) error {
	// Try to unmarshal as bsonAmount struct
	var ba bsonAmount
	if err := bson.Unmarshal(data, &ba); err == nil && ba.Cents != 0 {
		a.cents = ba.Cents
		return nil
	}

	// Try to unmarshal as float64
	var f float64
	if err := bson.Unmarshal(data, &f); err == nil {
		*a = NewAmount(f)
		return nil
	}

	// Try to unmarshal as int64
	var i int64
	if err := bson.Unmarshal(data, &i); err == nil {
		a.cents = i
		return nil
	}

	// Try to unmarshal as Decimal128
	var d primitive.Decimal128
	if err := bson.Unmarshal(data, &d); err == nil {
		bigInt, exp, err := d.BigInt()
		if err == nil && bigInt != nil {
			// Convert to dollars using the exponent
			dollars := float64(bigInt.Int64())
			for i := 0; i < -exp; i++ {
				dollars /= 10
			}
			for i := 0; i < exp; i++ {
				dollars *= 10
			}
			*a = NewAmount(dollars)
			return nil
		}
	}

	// Try to unmarshal as string
	var s string
	if err := bson.Unmarshal(data, &s); err == nil {
		parsed, err := NewAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	}

	return fmt.Errorf("cannot unmarshal Amount from BSON")
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler for Decimal128 support
func (a *Amount) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bsontype.Decimal128:
		var d primitive.Decimal128
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&d); err != nil {
			return err
		}
		bigInt, exp, err := d.BigInt()
		if err != nil {
			return err
		}
		// Convert to dollars using the exponent
		dollars := float64(bigInt.Int64())
		for i := 0; i < -exp; i++ {
			dollars /= 10
		}
		for i := 0; i < exp; i++ {
			dollars *= 10
		}
		*a = NewAmount(dollars)
		return nil

	case bsontype.Double:
		var f float64
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&f); err != nil {
			return err
		}
		*a = NewAmount(f)
		return nil

	case bsontype.Int64:
		var i int64
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&i); err != nil {
			return err
		}
		a.cents = i
		return nil

	case bsontype.Int32:
		var i int32
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&i); err != nil {
			return err
		}
		a.cents = int64(i)
		return nil

	case bsontype.String:
		var s string
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&s); err != nil {
			return err
		}
		parsed, err := NewAmountFromString(s)
		if err != nil {
			return err
		}
		*a = parsed
		return nil

	case bsontype.EmbeddedDocument:
		// Handle {cents: int64} format
		var ba bsonAmount
		rv := bson.RawValue{Type: t, Value: data}
		if err := rv.Unmarshal(&ba); err != nil {
			return err
		}
		a.cents = ba.Cents
		return nil

	default:
		return fmt.Errorf("cannot unmarshal Amount from BSON type %s", t)
	}
}

// bsonAmount is used for BSON marshaling/unmarshaling
type bsonAmount struct {
	Cents int64 `bson:"cents"`
}
