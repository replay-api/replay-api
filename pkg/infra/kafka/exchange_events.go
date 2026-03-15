package kafka

// Exchange/Bitcoin topics
const (
	TopicExchangeQuoteCreated    = "exchange.quote.created"
	TopicExchangeOrderCreated    = "exchange.order.created"
	TopicExchangeOrderExecuting  = "exchange.order.executing"
	TopicExchangeOrderFilled     = "exchange.order.filled"
	TopicExchangeOrderFailed     = "exchange.order.failed"
	TopicExchangeOrderCancelled  = "exchange.order.cancelled"
	TopicExchangePriceUpdated    = "exchange.price.updated"
	TopicBitcoinWithdrawPending  = "bitcoin.withdrawal.pending"
	TopicBitcoinWithdrawConfirmed = "bitcoin.withdrawal.confirmed"
	TopicBitcoinWithdrawFailed   = "bitcoin.withdrawal.failed"
	TopicBitcoinDepositDetected  = "bitcoin.deposit.detected"
	TopicBitcoinDepositConfirmed = "bitcoin.deposit.confirmed"
	TopicLightningInvoiceCreated = "lightning.invoice.created"
	TopicLightningPaymentSent    = "lightning.payment.sent"
	TopicLightningPaymentReceived = "lightning.payment.received"
	TopicExchangeDLQ             = "exchange.dlq"
)

// Exchange event types
const (
	EventTypeQuoteCreated     = "QUOTE_CREATED"
	EventTypeOrderCreated     = "ORDER_CREATED"
	EventTypeOrderExecuting   = "ORDER_EXECUTING"
	EventTypeOrderFilled      = "ORDER_FILLED"
	EventTypeOrderFailed      = "ORDER_FAILED"
	EventTypeOrderCancelled   = "ORDER_CANCELLED"
	EventTypePriceUpdated     = "PRICE_UPDATED"
	EventTypeBTCWithdrawInit  = "BTC_WITHDRAW_INIT"
	EventTypeBTCWithdrawDone  = "BTC_WITHDRAW_DONE"
	EventTypeBTCWithdrawFail  = "BTC_WITHDRAW_FAIL"
	EventTypeBTCDepositDetect = "BTC_DEPOSIT_DETECT"
	EventTypeBTCDepositConfirm = "BTC_DEPOSIT_CONFIRM"
	EventTypeLNInvoiceCreated = "LN_INVOICE_CREATED"
	EventTypeLNPaymentSent    = "LN_PAYMENT_SENT"
	EventTypeLNPaymentReceived = "LN_PAYMENT_RECEIVED"
)
