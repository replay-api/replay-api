package billing_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type PayableType string

const (
	PayableTypeSubscription PayableType = "Subscription" // default, direction -> in
	PayableTypeOrder        PayableType = "Order"        // shopping cart, direction -> in
	PayableTypePayout       PayableType = "Payout"       // paying players/tournament, direction -> OUT

	PayableTypeDeposit  PayableType = "Deposit"  // deposit to wallet, direction -> in
	PayableTypeWithdraw PayableType = "Withdraw" // withdraw from wallet, direction -> OUT
	PayableTypeRefund   PayableType = "Refund"   // refund for transactions, direction -> OUT
	PayableTypeTransfer PayableType = "Transfer" // transfer between accounts, direction -> between accounts

	// PayableTypeGiftCard PayableType = "GiftCard" // gift card transactions, direction -> in/out
	// PayableTypeDonation PayableType = "Donation" // donation transactions, direction -> in
	// PayableTypeTax      PayableType = "Tax"      // tax transactions, direction -> OUT
	// PayableTypeFee      PayableType = "Fee"      // fee transactions, direction -> in
	// PayableTypeReward   PayableType = "Reward"   // reward transactions, direction -> OUT

	// PayableTypeBettingPayout PayableType = "BettingPayout" // betting earnings, direction -> OUT
	// PayableTypeInvoice      PayableType = "Invoice"      // new invoice type, direction -> in
)

type PayableStatus string

const (
	PayableStatusAwaitingPayment PayableStatus = "AwaitingPayment"
	PayableStatusExpired         PayableStatus = "Expired"
	PayableStatusProcessing      PayableStatus = "Processing"
	PayableStatusPaid            PayableStatus = "Paid"
	PayableStatusFailed          PayableStatus = "Failed"
	PayableStatusRefundRequested PayableStatus = "RefundRequested"
	PayableStatusRefunded        PayableStatus = "Refunded"
	PayableStatusCanceled        PayableStatus = "Canceled"
	PayableStatusInvalid         PayableStatus = "Invalid"
	PayableStatusDisputed        PayableStatus = "Disputed"
)

type PaymentOptionType string

const (
	PaymentOptionCreditCard    PaymentOptionType = "CreditCard"
	PaymentOptionCrypto        PaymentOptionType = "Crypto"
	PaymentOptionWireTransfer  PaymentOptionType = "Wire"
	PaymentOptionBrazilianPix  PaymentOptionType = "PIX"
	PaymentOptionRemittance    PaymentOptionType = "Remittance"
	PaymentOptionWalletBalance PaymentOptionType = "Wallet"
	PaymentOptionCash          PaymentOptionType = "Cash"
	PaymentOptionBankTransfer  PaymentOptionType = "BankTransfer"
	PaymentOptionCheck         PaymentOptionType = "Check"
	PaymentOptionDebitCard     PaymentOptionType = "DebitCard"
)

type PaymentProvider string

const (
	PaymentProviderStripe    PaymentProvider = "Stripe"
	PaymentProviderPayPal    PaymentProvider = "PayPal"
	PaymentProviderApplePay  PaymentProvider = "ApplePay"
	PaymentProviderGooglePay PaymentProvider = "GooglePay"
	PaymentProviderWeChat    PaymentProvider = "WeChat"
	PaymentProviderWhatsApp  PaymentProvider = "WhatsApp"
	PaymentProviderMercado   PaymentProvider = "MercadoPago"
	PaymentProviderPagSeg    PaymentProvider = "PagSeguro"
	PaymentProviderAlipay    PaymentProvider = "Alipay"
	PaymentProviderPaytm     PaymentProvider = "Paytm"
	PaymentProviderAdyen     PaymentProvider = "Adyen"
	PaymentProviderPayFort   PaymentProvider = "PayFort"
	PaymentProviderPayStack  PaymentProvider = "PayStack"
	PaymentProviderRazorPay  PaymentProvider = "RazorPay"
	PaymentProviderDoku      PaymentProvider = "Doku"
	PaymentProviderMomo      PaymentProvider = "Momo"
	PaymentProviderTrueMoney PaymentProvider = "TrueMoney"
	PaymentProviderBaaS      PaymentProvider = "BaaS"
	PaymentProviderRemitter  PaymentProvider = "Remitter"
)

type Payable struct {
	common.BaseEntity
	Amount         float64           `json:"amount" bson:"amount"`
	Currency       string            `json:"currency" bson:"currency"`
	ExpirationDate time.Time         `json:"expiration" bson:"expiration"`
	PaymentDate    *time.Time        `json:"payment_date" bson:"payment_date"`
	ProcessedDate  *time.Time        `json:"processed_date" bson:"processed_date"`
	PaymentID      *uuid.UUID        `json:"payment_id" bson:"payment_id"`
	Provider       PaymentProvider   `json:"provider" bson:"provider"`
	Description    string            `json:"description" bson:"description"`
	Status         PayableStatus     `json:"status" bson:"status"`
	Type           PayableType       `json:"type" bson:"type"`
	Reference      string            `json:"reference" bson:"reference"`
	VoucherID      *uuid.UUID        `json:"voucher_id" bson:"voucher_id"`
	TotalDiscount  float64           `json:"total_discount" bson:"total_discount"`
	NetTotal       float64           `json:"net_total" bson:"net_total"`
	GrossTotal     float64           `json:"gross_total" bson:"gross_total"`
	ShippingCost   float64           `json:"shipping_cost" bson:"shipping_cost"`
	Discounts      []Discount        `json:"discounts" bson:"discounts"`
	Option         PaymentOptionType `json:"option" bson:"option"`
	History        []PayableHistory  `json:"history" bson:"history"`
}

type Discount struct {
	Amount float64 `json:"amount" bson:"amount"`
	Reason string  `json:"reason" bson:"reason"`
}

type PayableHistory struct {
	Date   time.Time     `json:"date" bson:"date"`
	Status PayableStatus `json:"status" bson:"status"`
	Reason string        `json:"reason" bson:"reason"`
}
