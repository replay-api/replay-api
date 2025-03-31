package billing_entities

import (
	"time"

	common "github.com/replay-api/replay-api/pkg/domain"
)

type BillingPeriodType string

const (
	BillingPeriodMonthly  BillingPeriodType = "monthly"
	BillingPeriodYearly   BillingPeriodType = "yearly"
	BillingPeriodLifetime BillingPeriodType = "lifetime"
)

type DisplayUnitType string

const (
	// doesnt uses unit
	DisplayUnitTypeCount      DisplayUnitType = "count"
	DisplayTypeBoolean        DisplayUnitType = "boolean"
	DisplayUnitTypePercentage DisplayUnitType = "percentage"
	DisplayUnitTypeTime       DisplayUnitType = "time"
	DisplayUnitTypeLength     DisplayUnitType = "length"

	// requires suffix/preffix

	DisplayUnitTypeCurrencyAmount   DisplayUnitType = "currency_amount"
	DisplayUnitCryptoCurrencyAmount DisplayUnitType = "crypto_currency_amount"
	DisplayUnitTypeBytes            DisplayUnitType = "bytes"
)

type DisplayUnit string

const (
	// time
	DisplayUnitSecond DisplayUnit = "s" // seconds
	DisplayUnitMinute DisplayUnit = "m" // minutes
	DisplayUnitHour   DisplayUnit = "h" // hours
	DisplayUnitDay    DisplayUnit = "d" // days
	DisplayUnitWeek   DisplayUnit = "w" // weeks
	DisplayUnitMonth  DisplayUnit = "M" // months
	DisplayUnitYear   DisplayUnit = "Y" // years

	// bytes
	DisplayUnitMB DisplayUnit = "MB" // Megabytes
	DisplayUnitGB DisplayUnit = "GB" // Gigabytes
	DisplayUnitTB DisplayUnit = "TB" // Terabytes

	// most used currencies
	DisplayUnitCurrencySymbolUSD DisplayUnit = "$"    // US Dollar
	DisplayUnitCurrencySymbolEUR DisplayUnit = "€"    // Euro
	DisplayUnitCurrencySymbolJPY DisplayUnit = "¥"    // Japanese Yen
	DisplayUnitCurrencySymbolGBP DisplayUnit = "£"    // British Pound
	DisplayUnitCurrencySymbolAUD DisplayUnit = "A$"   // Australian Dollar
	DisplayUnitCurrencySymbolCAD DisplayUnit = "C$"   // Canadian Dollar
	DisplayUnitCurrencySymbolCHF DisplayUnit = "Fr."  // Swiss Franc
	DisplayUnitCurrencySymbolCNY DisplayUnit = "¥"    // Chinese Yuan
	DisplayUnitCurrencySymbolHKD DisplayUnit = "HK$"  // Hong Kong Dollar
	DisplayUnitCurrencySymbolNZD DisplayUnit = "NZ$"  // New Zealand Dollar
	DisplayUnitCurrencySymbolSEK DisplayUnit = "kr"   // Swedish Krona
	DisplayUnitCurrencySymbolKRW DisplayUnit = "₩"    // South Korean Won
	DisplayUnitCurrencySymbolSGD DisplayUnit = "S$"   // Singapore Dollar
	DisplayUnitCurrencySymbolNOK DisplayUnit = "kr"   // Norwegian Krone
	DisplayUnitCurrencySymbolMXN DisplayUnit = "Mex$" // Mexican Peso
	DisplayUnitCurrencySymbolINR DisplayUnit = "₹"    // Indian Rupee
	DisplayUnitCurrencySymbolRUB DisplayUnit = "₽"    // Russian Ruble
	DisplayUnitCurrencySymbolZAR DisplayUnit = "R"    // South African Rand
	DisplayUnitCurrencySymbolTRY DisplayUnit = "₺"    // Turkish Lira
	DisplayUnitCurrencySymbolBRL DisplayUnit = "R$"   // Brazilian Real
	DisplayUnitCurrencySymbolTWD DisplayUnit = "NT$"  // New Taiwan Dollar
	DisplayUnitCurrencySymbolDKK DisplayUnit = "kr"   // Danish Krone
	DisplayUnitCurrencySymbolPLN DisplayUnit = "zł"   // Polish Zloty
	DisplayUnitCurrencySymbolTHB DisplayUnit = "฿"    // Thai Baht
	DisplayUnitCurrencySymbolIDR DisplayUnit = "Rp"   // Indonesian Rupiah
	DisplayUnitCurrencySymbolHUF DisplayUnit = "Ft"   // Hungarian Forint
	DisplayUnitCurrencySymbolCZK DisplayUnit = "Kč"   // Czech Koruna
	DisplayUnitCurrencySymbolILS DisplayUnit = "₪"    // Israeli Shekel
	DisplayUnitCurrencySymbolCLP DisplayUnit = "CLP$" // Chilean Peso
	DisplayUnitCurrencySymbolPHP DisplayUnit = "₱"    // Philippine Peso
	DisplayUnitCurrencySymbolAED DisplayUnit = "د.إ"  // UAE Dirham
	DisplayUnitCurrencySymbolCOP DisplayUnit = "COL$" // Colombian Peso
	DisplayUnitCurrencySymbolSAR DisplayUnit = "﷼"    // Saudi Riyal
	DisplayUnitCurrencySymbolMYR DisplayUnit = "RM"   // Malaysian Ringgit
	DisplayUnitCurrencySymbolRON DisplayUnit = "lei"  // Romanian Leu

	// most used crypto currencies
	DisplayUnitCryptoCurrencySymbolBTC   DisplayUnit = "₿"     // Bitcoin
	DisplayUnitCryptoCurrencySymbolETH   DisplayUnit = "Ξ"     // Ethereum
	DisplayUnitCryptoCurrencySymbolXRP   DisplayUnit = "XRP"   // Ripple
	DisplayUnitCryptoCurrencySymbolLTC   DisplayUnit = "Ł"     // Litecoin
	DisplayUnitCryptoCurrencySymbolBCH   DisplayUnit = "BCH"   // Bitcoin Cash
	DisplayUnitCryptoCurrencySymbolEOS   DisplayUnit = "EOS"   // EOS.IO
	DisplayUnitCryptoCurrencySymbolXLM   DisplayUnit = "XLM"   // Stellar
	DisplayUnitCryptoCurrencySymbolADA   DisplayUnit = "ADA"   // Cardano
	DisplayUnitCryptoCurrencySymbolXMR   DisplayUnit = "XMR"   // Monero
	DisplayUnitCryptoCurrencySymbolDASH  DisplayUnit = "DASH"  // Dash
	DisplayUnitCryptoCurrencySymbolNEO   DisplayUnit = "NEO"   // NEO
	DisplayUnitCryptoCurrencySymbolTRX   DisplayUnit = "TRX"   // TRON
	DisplayUnitCryptoCurrencySymbolXEM   DisplayUnit = "XEM"   // NEM
	DisplayUnitCryptoCurrencySymbolETC   DisplayUnit = "ETC"   // Ethereum Classic
	DisplayUnitCryptoCurrencySymbolVEN   DisplayUnit = "VEN"   // VeChain
	DisplayUnitCryptoCurrencySymbolQTUM  DisplayUnit = "QTUM"  // Qtum
	DisplayUnitCryptoCurrencySymbolICX   DisplayUnit = "ICX"   // ICON
	DisplayUnitCryptoCurrencySymbolLSK   DisplayUnit = "LSK"   // Lisk
	DisplayUnitCryptoCurrencySymbolOMG   DisplayUnit = "OMG"   // OmiseGO
	DisplayUnitCryptoCurrencySymbolZEC   DisplayUnit = "ZEC"   // Zcash
	DisplayUnitCryptoCurrencySymbolBCN   DisplayUnit = "BCN"   // Bytecoin
	DisplayUnitCryptoCurrencySymbolSTEEM DisplayUnit = "STEEM" // Steem
	DisplayUnitCryptoCurrencySymbolXVG   DisplayUnit = "XVG"   // Verge
	DisplayUnitCryptoCurrencySymbolSC    DisplayUnit = "SC"    // Siacoin
	DisplayUnitCryptoCurrencySymbolBCD   DisplayUnit = "BCD"   // Bitcoin Diamond
	DisplayUnitCryptoCurrencySymbolSTRAT DisplayUnit = "STRAT" // Stratis
	DisplayUnitCryptoCurrencySymbolDOGE  DisplayUnit = "DOGE"  // Dogecoin
)

type CurrencyUnit string

const (
	CurrencyUnitUSD CurrencyUnit = "USD" // United States Dollar
	CurrencyUnitEUR CurrencyUnit = "EUR" // Euro
	CurrencyUnitJPY CurrencyUnit = "JPY" // Japanese Yen
	CurrencyUnitGBP CurrencyUnit = "GBP" // British Pound Sterling
	CurrencyUnitAUD CurrencyUnit = "AUD" // Australian Dollar
	CurrencyUnitCAD CurrencyUnit = "CAD" // Canadian Dollar
	CurrencyUnitCHF CurrencyUnit = "CHF" // Swiss Franc
	CurrencyUnitCNY CurrencyUnit = "CNY" // Chinese Yuan
	CurrencyUnitHKD CurrencyUnit = "HKD" // Hong Kong Dollar
	CurrencyUnitNZD CurrencyUnit = "NZD" // New Zealand Dollar
	CurrencyUnitSEK CurrencyUnit = "SEK" // Swedish Krona
	CurrencyUnitKRW CurrencyUnit = "KRW" // South Korean Won
	CurrencyUnitSGD CurrencyUnit = "SGD" // Singapore Dollar
	CurrencyUnitNOK CurrencyUnit = "NOK" // Norwegian Krone
	CurrencyUnitMXN CurrencyUnit = "MXN" // Mexican Peso
	CurrencyUnitINR CurrencyUnit = "INR" // Indian Rupee
	CurrencyUnitRUB CurrencyUnit = "RUB" // Russian Ruble
	CurrencyUnitZAR CurrencyUnit = "ZAR" // South African Rand
	CurrencyUnitTRY CurrencyUnit = "TRY" // Turkish Lira
	CurrencyUnitBRL CurrencyUnit = "BRL" // Brazilian Real
	CurrencyUnitTWD CurrencyUnit = "TWD" // New Taiwan Dollar
	CurrencyUnitDKK CurrencyUnit = "DKK" // Danish Krone
	CurrencyUnitPLN CurrencyUnit = "PLN" // Polish Zloty
	CurrencyUnitTHB CurrencyUnit = "THB" // Thai Baht
	CurrencyUnitIDR CurrencyUnit = "IDR" // Indonesian Rupiah
	CurrencyUnitHUF CurrencyUnit = "HUF" // Hungarian Forint
	CurrencyUnitCZK CurrencyUnit = "CZK" // Czech Koruna
	CurrencyUnitILS CurrencyUnit = "ILS" // Israeli New Shekel
	CurrencyUnitCLP CurrencyUnit = "CLP" // Chilean Peso
	CurrencyUnitPHP CurrencyUnit = "PHP" // Philippine Peso
	CurrencyUnitAED CurrencyUnit = "AED" // United Arab Emirates Dirham
	CurrencyUnitCOP CurrencyUnit = "COP" // Colombian Peso
	CurrencyUnitSAR CurrencyUnit = "SAR" // Saudi Riyal
	CurrencyUnitMYR CurrencyUnit = "MYR" // Malaysian Ringgit
	CurrencyUnitRON CurrencyUnit = "RON" // Romanian Leu
)

type DisplayUnitCurrencyMapType map[CurrencyUnit]DisplayUnit

var DisplayUnitCurrencyMap = DisplayUnitCurrencyMapType{
	CurrencyUnitUSD: DisplayUnitCurrencySymbolUSD, // United States Dollar
	CurrencyUnitEUR: DisplayUnitCurrencySymbolEUR, // Euro
	CurrencyUnitJPY: DisplayUnitCurrencySymbolJPY, // Japanese Yen
	CurrencyUnitGBP: DisplayUnitCurrencySymbolGBP, // British Pound Sterling
	CurrencyUnitAUD: DisplayUnitCurrencySymbolAUD, // Australian Dollar
	CurrencyUnitCAD: DisplayUnitCurrencySymbolCAD, // Canadian Dollar
	CurrencyUnitCHF: DisplayUnitCurrencySymbolCHF, // Swiss Franc
	CurrencyUnitCNY: DisplayUnitCurrencySymbolCNY, // Chinese Yuan
	CurrencyUnitHKD: DisplayUnitCurrencySymbolHKD, // Hong Kong Dollar
	CurrencyUnitNZD: DisplayUnitCurrencySymbolNZD, // New Zealand Dollar
	CurrencyUnitSEK: DisplayUnitCurrencySymbolSEK, // Swedish Krona
	CurrencyUnitKRW: DisplayUnitCurrencySymbolKRW, // South Korean Won
	CurrencyUnitSGD: DisplayUnitCurrencySymbolSGD, // Singapore Dollar
	CurrencyUnitNOK: DisplayUnitCurrencySymbolNOK, // Norwegian Krone
	CurrencyUnitMXN: DisplayUnitCurrencySymbolMXN, // Mexican Peso
	CurrencyUnitINR: DisplayUnitCurrencySymbolINR, // Indian Rupee
	CurrencyUnitRUB: DisplayUnitCurrencySymbolRUB, // Russian Ruble
	CurrencyUnitZAR: DisplayUnitCurrencySymbolZAR, // South African Rand
	CurrencyUnitTRY: DisplayUnitCurrencySymbolTRY, // Turkish Lira
	CurrencyUnitBRL: DisplayUnitCurrencySymbolBRL, // Brazilian Real
	CurrencyUnitTWD: DisplayUnitCurrencySymbolTWD, // New Taiwan Dollar
	CurrencyUnitDKK: DisplayUnitCurrencySymbolDKK, // Danish Krone
	CurrencyUnitPLN: DisplayUnitCurrencySymbolPLN, // Polish Zloty
	CurrencyUnitTHB: DisplayUnitCurrencySymbolTHB, // Thai Baht
	CurrencyUnitIDR: DisplayUnitCurrencySymbolIDR, // Indonesian Rupiah
	CurrencyUnitHUF: DisplayUnitCurrencySymbolHUF, // Hungarian Forint
	CurrencyUnitCZK: DisplayUnitCurrencySymbolCZK, // Czech Koruna
	CurrencyUnitILS: DisplayUnitCurrencySymbolILS, // Israeli New Shekel
	CurrencyUnitCLP: DisplayUnitCurrencySymbolCLP, // Chilean Peso
	CurrencyUnitPHP: DisplayUnitCurrencySymbolPHP, // Philippine Peso
	CurrencyUnitAED: DisplayUnitCurrencySymbolAED, // United Arab Emirates Dirham
	CurrencyUnitCOP: DisplayUnitCurrencySymbolCOP, // Colombian Peso
	CurrencyUnitSAR: DisplayUnitCurrencySymbolSAR, // Saudi Riyal
	CurrencyUnitMYR: DisplayUnitCurrencySymbolMYR, // Malaysian Ringgit
	CurrencyUnitRON: DisplayUnitCurrencySymbolRON, // Romanian Leu
}

type CryptoCurrencyUnit string

const (
	CryptoCurrencyUnitBTC   CryptoCurrencyUnit = "BTC"   // Bitcoin
	CryptoCurrencyUnitETH   CryptoCurrencyUnit = "ETH"   // Ethereum
	CryptoCurrencyUnitXRP   CryptoCurrencyUnit = "XRP"   // Ripple
	CryptoCurrencyUnitLTC   CryptoCurrencyUnit = "LTC"   // Litecoin
	CryptoCurrencyUnitBCH   CryptoCurrencyUnit = "BCH"   // Bitcoin Cash
	CryptoCurrencyUnitEOS   CryptoCurrencyUnit = "EOS"   // EOS.IO
	CryptoCurrencyUnitXLM   CryptoCurrencyUnit = "XLM"   // Stellar
	CryptoCurrencyUnitADA   CryptoCurrencyUnit = "ADA"   // Cardano
	CryptoCurrencyUnitXMR   CryptoCurrencyUnit = "XMR"   // Monero
	CryptoCurrencyUnitDASH  CryptoCurrencyUnit = "DASH"  // Dash
	CryptoCurrencyUnitNEO   CryptoCurrencyUnit = "NEO"   // NEO
	CryptoCurrencyUnitTRX   CryptoCurrencyUnit = "TRX"   // TRON
	CryptoCurrencyUnitXEM   CryptoCurrencyUnit = "XEM"   // NEM
	CryptoCurrencyUnitETC   CryptoCurrencyUnit = "ETC"   // Ethereum Classic
	CryptoCurrencyUnitVEN   CryptoCurrencyUnit = "VEN"   // VeChain
	CryptoCurrencyUnitQTUM  CryptoCurrencyUnit = "QTUM"  // Qtum
	CryptoCurrencyUnitICX   CryptoCurrencyUnit = "ICX"   // ICON
	CryptoCurrencyUnitLSK   CryptoCurrencyUnit = "LSK"   // Lisk
	CryptoCurrencyUnitOMG   CryptoCurrencyUnit = "OMG"   // OmiseGO
	CryptoCurrencyUnitZEC   CryptoCurrencyUnit = "ZEC"   // Zcash
	CryptoCurrencyUnitBCN   CryptoCurrencyUnit = "BCN"   // Bytecoin
	CryptoCurrencyUnitSTEEM CryptoCurrencyUnit = "STEEM" // Steem
	CryptoCurrencyUnitXVG   CryptoCurrencyUnit = "XVG"   // Verge
	CryptoCurrencyUnitSC    CryptoCurrencyUnit = "SC"    // Siacoin
	CryptoCurrencyUnitBCD   CryptoCurrencyUnit = "BCD"   // Bitcoin Diamond
	CryptoCurrencyUnitSTRAT CryptoCurrencyUnit = "STRAT" // Stratis
	CryptoCurrencyUnitDOGE  CryptoCurrencyUnit = "DOGE"  // Dogecoin
)

type CryptoCurrencyUnitMapType map[CryptoCurrencyUnit]DisplayUnit

var CryptoCurrencyUnitMap = CryptoCurrencyUnitMapType{
	CryptoCurrencyUnitBTC:   DisplayUnitCryptoCurrencySymbolBTC,   // Bitcoin
	CryptoCurrencyUnitETH:   DisplayUnitCryptoCurrencySymbolETH,   // Ethereum
	CryptoCurrencyUnitXRP:   DisplayUnitCryptoCurrencySymbolXRP,   // Ripple
	CryptoCurrencyUnitLTC:   DisplayUnitCryptoCurrencySymbolLTC,   // Litecoin
	CryptoCurrencyUnitBCH:   DisplayUnitCryptoCurrencySymbolBCH,   // Bitcoin Cash
	CryptoCurrencyUnitEOS:   DisplayUnitCryptoCurrencySymbolEOS,   // EOS.IO
	CryptoCurrencyUnitXLM:   DisplayUnitCryptoCurrencySymbolXLM,   // Stellar
	CryptoCurrencyUnitADA:   DisplayUnitCryptoCurrencySymbolADA,   // Cardano
	CryptoCurrencyUnitXMR:   DisplayUnitCryptoCurrencySymbolXMR,   // Monero
	CryptoCurrencyUnitDASH:  DisplayUnitCryptoCurrencySymbolDASH,  // Dash
	CryptoCurrencyUnitNEO:   DisplayUnitCryptoCurrencySymbolNEO,   // NEO
	CryptoCurrencyUnitTRX:   DisplayUnitCryptoCurrencySymbolTRX,   // TRON
	CryptoCurrencyUnitXEM:   DisplayUnitCryptoCurrencySymbolXEM,   // NEM
	CryptoCurrencyUnitETC:   DisplayUnitCryptoCurrencySymbolETC,   // Ethereum Classic
	CryptoCurrencyUnitVEN:   DisplayUnitCryptoCurrencySymbolVEN,   // VeChain
	CryptoCurrencyUnitQTUM:  DisplayUnitCryptoCurrencySymbolQTUM,  // Qtum
	CryptoCurrencyUnitICX:   DisplayUnitCryptoCurrencySymbolICX,   // ICON
	CryptoCurrencyUnitLSK:   DisplayUnitCryptoCurrencySymbolLSK,   // Lisk
	CryptoCurrencyUnitOMG:   DisplayUnitCryptoCurrencySymbolOMG,   // OmiseGO
	CryptoCurrencyUnitZEC:   DisplayUnitCryptoCurrencySymbolZEC,   // Zcash
	CryptoCurrencyUnitBCN:   DisplayUnitCryptoCurrencySymbolBCN,   // Bytecoin
	CryptoCurrencyUnitSTEEM: DisplayUnitCryptoCurrencySymbolSTEEM, // Steem
	CryptoCurrencyUnitXVG:   DisplayUnitCryptoCurrencySymbolXVG,   // Verge
	CryptoCurrencyUnitSC:    DisplayUnitCryptoCurrencySymbolSC,    // Siacoin
	CryptoCurrencyUnitBCD:   DisplayUnitCryptoCurrencySymbolBCD,   // Bitcoin Diamond
	CryptoCurrencyUnitSTRAT: DisplayUnitCryptoCurrencySymbolSTRAT, // Stratis
	CryptoCurrencyUnitDOGE:  DisplayUnitCryptoCurrencySymbolDOGE,  // Dogecoin
}

type DisplayUnitPosition string

const (
	DisplayUnitPositionBefore DisplayUnitPosition = "before"
	DisplayUnitPositionAfter  DisplayUnitPosition = "after"
)

type PlanKindType string

const (
	PlanKindTypeFree     PlanKindType = "free"
	PlanKindTypeStarter  PlanKindType = "starter"
	PlanKindTypePro      PlanKindType = "pro"
	PlanKindTypeTeam     PlanKindType = "team"
	PlanKindTypeBusiness PlanKindType = "business"
	PlanKindTypeCustom   PlanKindType = "custom"
)

type PlanCustomerType string

const (
	PlanCustomerTypeIndividual PlanCustomerType = "individual"
	PlanCustomerTypeGroup      PlanCustomerType = "group"
	PlanCustomerTypeCompany    PlanCustomerType = "company"
)

type Plan struct {
	common.BaseEntity
	Name                 string                                `json:"name" bson:"name"`
	Description          string                                `json:"description" bson:"description"`
	Kind                 PlanKindType                          `json:"kind" bson:"kind"`
	CustomerType         PlanCustomerType                      `json:"customer_type" bson:"customer_type"`
	Prices               map[BillingPeriodType][]Price         `json:"prices" bson:"prices"`
	OperationLimits      map[BillableOperationKey]BillableItem `json:"operation_limits" bson:"operation_limits"`
	IsFree               bool                                  `json:"is_free" bson:"is_free"`
	IsAvailable          bool                                  `json:"is_available" bson:"is_available"`
	IsLegacy             bool                                  `json:"is_legacy" bson:"is_legacy"`
	IsActive             bool                                  `json:"is_active" bson:"is_active"`
	DisplayPriorityScore int                                   `json:"display_priority_score" bson:"display_priority_score"`
	Regions              []string                              `json:"regions" bson:"regions"`
	Languages            []string                              `json:"languages" bson:"languages"`
	EffectiveDate        time.Time                             `json:"effective_date" bson:"effective_date"`
	ExpirationDate       *time.Time                            `json:"expiration_date" bson:"expiration_date"`
}

type BillableItem struct {
	Name                string              `json:"name" bson:"name"`
	Description         string              `json:"description" bson:"description"`
	Limit               float64             `json:"limit" bson:"limit"`
	DisplayUnitType     DisplayUnitType     `json:"display_unit_type" bson:"display_unit_type"`
	DisplayUnit         DisplayUnit         `json:"display_unit" bson:"display_unit"`
	DisplayUnitPosition DisplayUnitPosition `json:"display_unit_position" bson:"display_unit_position"`
	PrefixString        string              `json:"prefix_string" bson:"prefix_string"`
	SuffixString        string              `json:"suffix_string" bson:"suffix_string"`
}

type Price struct {
	Amount        float64 `json:"amount" bson:"amount"`
	Currency      string  `json:"currency" bson:"currency"`
	TotalDiscount float64 `json:"total_discount" bson:"total_discount"`
}
