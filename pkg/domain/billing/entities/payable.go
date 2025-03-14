package billing_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type PayableType string

const (
	PayableTypeSubscription PayableType = "Subscription"
	PayableTypePayout       PayableType = "Payout"
	PayableTypeShoppingCart PayableType = "ShoppingCart"
	PayableTypeRefund       PayableType = "Refund"
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

type PaymentOption string

const (
	PaymentOptionCreditCard   PaymentOption = "CreditCard"
	PaymentOptionCrypto       PaymentOption = "Crypto"
	PaymentOptionWireTransfer PaymentOption = "Wire"
	PaymentOptionBrazilianPix PaymentOption = "PIX"
	PaymentOptionRemittance   PaymentOption = "Remittance"
)

type PaymentProvider string

const (
	PaymentProviderStripe    PaymentProvider = "Stripe"
	PaymentProviderPayPal    PaymentProvider = "PayPal"
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
	Amount         float64          `json:"amount" bson:"amount"`
	Currency       string           `json:"currency" bson:"currency"`
	ExpirationDate time.Time        `json:"expiration" bson:"expiration"`
	PaymentDate    *time.Time       `json:"payment_date" bson:"payment_date"`
	ProcessedDate  *time.Time       `json:"processed_date" bson:"processed_date"`
	PaymentID      *uuid.UUID       `json:"payment_id" bson:"payment_id"`
	Provider       PaymentProvider  `json:"provider" bson:"provider"`
	Description    string           `json:"description" bson:"description"`
	Status         PayableStatus    `json:"status" bson:"status"`
	Type           PayableType      `json:"type" bson:"type"`
	Reference      string           `json:"reference" bson:"reference"`
	VoucherID      *uuid.UUID       `json:"voucher_id" bson:"voucher_id"`
	TotalDiscount  float64          `json:"total_discount" bson:"total_discount"`
	NetTotal       float64          `json:"net_total" bson:"net_total"`
	GrossTotal     float64          `json:"gross_total" bson:"gross_total"`
	ShippingCost   float64          `json:"shipping_cost" bson:"shipping_cost"`
	Discounts      []Discount       `json:"discounts" bson:"discounts"`
	Option         PaymentOption    `json:"option" bson:"option"`
	History        []PayableHistory `json:"history" bson:"history"`
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
