package wallet_vo

import "fmt"

// Currency represents supported currencies
type Currency string

const (
	CurrencyUSD  Currency = "USD"  // Fiat USD (internal accounting)
	CurrencyUSDC Currency = "USDC" // USD Coin (ERC-20)
	CurrencyUSDT Currency = "USDT" // Tether USD (ERC-20)
)

// AllCurrencies returns all supported currencies
func AllCurrencies() []Currency {
	return []Currency{CurrencyUSD, CurrencyUSDC, CurrencyUSDT}
}

// ParseCurrency parses a string into a Currency
func ParseCurrency(s string) (Currency, error) {
	c := Currency(s)
	if !c.IsValid() {
		return "", fmt.Errorf("invalid currency: %s", s)
	}
	return c, nil
}

// IsValid checks if the currency is supported
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyUSDC, CurrencyUSDT:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (c Currency) String() string {
	return string(c)
}

// Symbol returns the currency symbol
func (c Currency) Symbol() string {
	switch c {
	case CurrencyUSD, CurrencyUSDC, CurrencyUSDT:
		return "$"
	default:
		return ""
	}
}

// IsStablecoin checks if the currency is a blockchain stablecoin
func (c Currency) IsStablecoin() bool {
	return c == CurrencyUSDC || c == CurrencyUSDT
}

// ContractAddress returns the ERC-20 contract address for blockchain currencies
// (Polygon Mumbai testnet addresses)
func (c Currency) ContractAddress() (string, error) {
	switch c {
	case CurrencyUSDC:
		return "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174", nil // Polygon USDC
	case CurrencyUSDT:
		return "0xc2132D05D31c914a87C6611C10748AEb04B58e8F", nil // Polygon USDT
	case CurrencyUSD:
		return "", fmt.Errorf("USD is not a blockchain currency")
	default:
		return "", fmt.Errorf("unknown currency: %s", c)
	}
}

// Decimals returns the decimal places for the currency
func (c Currency) Decimals() int {
	switch c {
	case CurrencyUSDC, CurrencyUSDT:
		return 6 // USDC and USDT use 6 decimals
	case CurrencyUSD:
		return 2 // Fiat USD uses 2 decimals
	default:
		return 2
	}
}
