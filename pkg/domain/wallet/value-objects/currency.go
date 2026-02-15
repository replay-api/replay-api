package wallet_vo

import "fmt"

// Currency represents supported currencies
type Currency string

const (
	CurrencyUSD  Currency = "USD"  // Fiat USD (internal accounting)
	CurrencyUSDC Currency = "USDC" // USD Coin (ERC-20)
	CurrencyUSDT Currency = "USDT" // Tether USD (ERC-20)
)

// ChainID represents a supported blockchain network
type ChainID int

const (
	ChainIDNone            ChainID = 0     // Off-chain / fiat
	ChainIDEthereumMainnet ChainID = 1     // Ethereum Mainnet
	ChainIDPolygonMainnet  ChainID = 137   // Polygon PoS Mainnet
	ChainIDPolygonAmoy     ChainID = 80002 // Polygon Amoy Testnet
	ChainIDArbitrumOne     ChainID = 42161 // Arbitrum One
	ChainIDBaseMainnet     ChainID = 8453  // Base Mainnet
)

// AllSupportedChains returns all chains we support for transactions
func AllSupportedChains() []ChainID {
	return []ChainID{
		ChainIDPolygonMainnet,
		ChainIDEthereumMainnet,
		ChainIDArbitrumOne,
		ChainIDBaseMainnet,
	}
}

// IsValid checks if the chain ID is supported
func (c ChainID) IsValid() bool {
	switch c {
	case ChainIDNone, ChainIDEthereumMainnet, ChainIDPolygonMainnet,
		ChainIDPolygonAmoy, ChainIDArbitrumOne, ChainIDBaseMainnet:
		return true
	default:
		return false
	}
}

// String returns a human-readable chain name
func (c ChainID) String() string {
	switch c {
	case ChainIDNone:
		return "off-chain"
	case ChainIDEthereumMainnet:
		return "Ethereum"
	case ChainIDPolygonMainnet:
		return "Polygon"
	case ChainIDPolygonAmoy:
		return "Polygon Amoy (Testnet)"
	case ChainIDArbitrumOne:
		return "Arbitrum"
	case ChainIDBaseMainnet:
		return "Base"
	default:
		return fmt.Sprintf("chain-%d", int(c))
	}
}

// ParseChainID parses an int into a ChainID
func ParseChainID(id int) (ChainID, error) {
	c := ChainID(id)
	if !c.IsValid() {
		return ChainIDNone, fmt.Errorf("unsupported chain ID: %d", id)
	}
	return c, nil
}

// PaymentMethod represents how a deposit/withdrawal is funded
type PaymentMethod string

const (
	PaymentMethodCrypto     PaymentMethod = "crypto"      // On-chain stablecoin transfer
	PaymentMethodCreditCard PaymentMethod = "credit_card"  // Stripe credit card
	PaymentMethodPIX        PaymentMethod = "pix"          // PIX (Brazil)
	PaymentMethodBankWire   PaymentMethod = "bank_transfer" // Bank wire
)

// IsValid checks if the payment method is supported
func (p PaymentMethod) IsValid() bool {
	switch p {
	case PaymentMethodCrypto, PaymentMethodCreditCard, PaymentMethodPIX, PaymentMethodBankWire:
		return true
	default:
		return false
	}
}

// IsFiat returns true if this is a fiat payment method
func (p PaymentMethod) IsFiat() bool {
	return p == PaymentMethodCreditCard || p == PaymentMethodPIX || p == PaymentMethodBankWire
}

// ChainContractAddress returns the ERC-20 contract address for a currency on a specific chain
func ChainContractAddress(currency Currency, chain ChainID) (string, error) {
	type key struct {
		currency Currency
		chain    ChainID
	}
	contracts := map[key]string{
		// Polygon Mainnet
		{CurrencyUSDC, ChainIDPolygonMainnet}: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359",
		{CurrencyUSDT, ChainIDPolygonMainnet}: "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
		// Ethereum Mainnet
		{CurrencyUSDC, ChainIDEthereumMainnet}: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		{CurrencyUSDT, ChainIDEthereumMainnet}: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		// Arbitrum One
		{CurrencyUSDC, ChainIDArbitrumOne}: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
		{CurrencyUSDT, ChainIDArbitrumOne}: "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9",
		// Base Mainnet
		{CurrencyUSDC, ChainIDBaseMainnet}: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		// Polygon Amoy Testnet
		{CurrencyUSDC, ChainIDPolygonAmoy}: "0x41E94Eb71898E8A6e13eF9F4D2cC53123F62C57C",
	}

	addr, ok := contracts[key{currency, chain}]
	if !ok {
		return "", fmt.Errorf("no contract for %s on %s", currency, chain)
	}
	return addr, nil
}

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
// Defaults to Polygon Mainnet. Use ChainContractAddress for other chains.
func (c Currency) ContractAddress() (string, error) {
	return ChainContractAddress(c, ChainIDPolygonMainnet)
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
