package wallet_vo

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// Amount represents a monetary amount with fixed-point precision
// Uses cents (int64) internally to avoid floating-point errors
type Amount struct {
	cents int64 // Amount in cents (e.g., $10.50 = 1050 cents)
}

const (
	centsPerDollar = 100
)

// NewAmount creates a new Amount from dollars (as float)
func NewAmount(dollars float64) Amount {
	cents := int64(math.Round(dollars * float64(centsPerDollar)))
	return Amount{cents: cents}
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
func (a Amount) Divide(divisor float64) Amount {
	if divisor == 0 {
		return Amount{cents: 0}
	}
	result := float64(a.cents) / divisor
	return Amount{cents: int64(math.Round(result))}
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
