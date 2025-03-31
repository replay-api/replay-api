package billing_entities

import common "github.com/replay-api/replay-api/pkg/domain"

type Payment struct {
	common.BaseEntity
	PayableID                string            `json:"payable_id" bson:"payable_id"`
	Reference                string            `json:"reference" bson:"reference"`
	Amount                   float64           `json:"amount" bson:"amount"`
	Currency                 string            `json:"currency" bson:"currency"`
	Option                   PaymentOptionType `json:"option" bson:"option"`
	Status                   PaymentStatus     `json:"status" bson:"status"`
	Provider                 PaymentProvider   `json:"provider" bson:"provider"`
	PaymentProviderReference string            `json:"payment_provider_reference" bson:"payment_provider_reference"`
	Description              string            `json:"description" bson:"description"`
}

type PaymentStatus string

const (
	PaymentStatusReceived   PaymentStatus = "Received"
	PaymentStatusProcessing PaymentStatus = "Processing"
	PaymentStatusSucceeded  PaymentStatus = "Succeeded" // settled
	PaymentStatusFailed     PaymentStatus = "Failed"
	PaymentStatusInvoiced   PaymentStatus = "Invoiced"
	PaymentStatusInvalid    PaymentStatus = "Invalid"
	PaymentStatusRefunded   PaymentStatus = "Refunded"
)
