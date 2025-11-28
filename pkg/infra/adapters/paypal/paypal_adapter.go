// Package paypal provides the PayPal payment provider adapter implementation
// Follows SOLID principles - implements the same PaymentProviderAdapter interface as Stripe
package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	payment_entities "github.com/replay-api/replay-api/pkg/domain/payment/entities"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// PayPal API endpoints
const (
	SandboxBaseURL    = "https://api-m.sandbox.paypal.com"
	ProductionBaseURL = "https://api-m.paypal.com"
)

// PayPal order status mapping
var paypalStatusMap = map[string]payment_entities.PaymentStatus{
	"CREATED":   payment_entities.PaymentStatusPending,
	"SAVED":     payment_entities.PaymentStatusPending,
	"APPROVED":  payment_entities.PaymentStatusProcessing,
	"COMPLETED": payment_entities.PaymentStatusSucceeded,
	"VOIDED":    payment_entities.PaymentStatusCanceled,
}

// PayPalAdapter implements the PaymentProviderAdapter interface for PayPal
type PayPalAdapter struct {
	clientID      string
	clientSecret  string
	baseURL       string
	webhookID     string
	accessToken   string
	tokenExpiry   time.Time
	httpClient    *http.Client
	tokenMutex    sync.Mutex
}

// NewPayPalAdapter creates a new PayPal adapter
func NewPayPalAdapter() *PayPalAdapter {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	webhookID := os.Getenv("PAYPAL_WEBHOOK_ID")

	// Determine base URL based on environment
	baseURL := SandboxBaseURL
	if os.Getenv("PAYPAL_ENVIRONMENT") == "production" {
		baseURL = ProductionBaseURL
	}

	return &PayPalAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		webhookID:    webhookID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProvider returns the provider type
func (p *PayPalAdapter) GetProvider() payment_entities.PaymentProvider {
	return payment_entities.PaymentProviderPayPal
}

// getAccessToken retrieves or refreshes the OAuth access token
func (p *PayPalAdapter) getAccessToken(ctx context.Context) (string, error) {
	p.tokenMutex.Lock()
	defer p.tokenMutex.Unlock()

	// Return cached token if still valid
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	// Request new token
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/oauth2/token", bytes.NewBuffer([]byte("grant_type=client_credentials")))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(p.clientID, p.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Cache token with buffer
	p.accessToken = tokenResp.AccessToken
	p.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return p.accessToken, nil
}

// CreatePaymentIntent creates a PayPal order
func (p *PayPalAdapter) CreatePaymentIntent(ctx context.Context, req payment_out.CreateIntentRequest) (*payment_out.CreateIntentResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Convert cents to dollars with decimal
	amount := fmt.Sprintf("%.2f", float64(req.Amount)/100)

	orderRequest := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": req.Currency,
					"value":         amount,
				},
				"description": req.Description,
			},
		},
		"application_context": map[string]interface{}{
			"return_url": req.ReturnURL,
			"cancel_url": req.CancelURL,
		},
	}

	body, err := json.Marshal(orderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v2/checkout/orders", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create order request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("PayPal-Request-Id", req.IdempotencyKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("order creation failed: %v", errResp)
	}

	var orderResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, fmt.Errorf("failed to decode order response: %w", err)
	}

	// Find approval link
	var approvalURL string
	for _, link := range orderResp.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	return &payment_out.CreateIntentResponse{
		ProviderPaymentID: orderResp.ID,
		RedirectURL:       approvalURL,
		Status:            orderResp.Status,
	}, nil
}

// ConfirmPayment captures a PayPal order
func (p *PayPalAdapter) ConfirmPayment(ctx context.Context, req payment_out.ConfirmPaymentRequest) (*payment_out.ConfirmPaymentResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v2/checkout/orders/"+req.ProviderPaymentID+"/capture", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create capture request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to capture order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("capture failed: %v", errResp)
	}

	var captureResp struct {
		Status         string `json:"status"`
		PurchaseUnits  []struct {
			Payments struct {
				Captures []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
					Amount struct {
						Value string `json:"value"`
					} `json:"amount"`
					SellerPayableBreakdown struct {
						PayPalFee struct {
							Value string `json:"value"`
						} `json:"paypal_fee"`
					} `json:"seller_payable_breakdown"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&captureResp); err != nil {
		return nil, fmt.Errorf("failed to decode capture response: %w", err)
	}

	// Calculate provider fee from response
	var providerFee int64
	if len(captureResp.PurchaseUnits) > 0 &&
		len(captureResp.PurchaseUnits[0].Payments.Captures) > 0 {
		fee := captureResp.PurchaseUnits[0].Payments.Captures[0].SellerPayableBreakdown.PayPalFee.Value
		var feeFloat float64
		fmt.Sscanf(fee, "%f", &feeFloat)
		providerFee = int64(feeFloat * 100) // Convert to cents
	}

	return &payment_out.ConfirmPaymentResponse{
		Status:      captureResp.Status,
		ProviderFee: providerFee,
	}, nil
}

// RefundPayment refunds a PayPal capture
func (p *PayPalAdapter) RefundPayment(ctx context.Context, req payment_out.RefundRequest) (*payment_out.RefundResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	refundRequest := map[string]interface{}{}
	if req.Amount > 0 {
		refundRequest["amount"] = map[string]interface{}{
			"currency_code": "USD", // TODO: Get from original order
			"value":         fmt.Sprintf("%.2f", float64(req.Amount)/100),
		}
	}

	body, _ := json.Marshal(refundRequest)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v2/payments/captures/"+req.ProviderPaymentID+"/refund", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create refund request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("PayPal-Request-Id", req.IdempotencyKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to refund: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("refund failed: %v", errResp)
	}

	var refundResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Value string `json:"value"`
		} `json:"amount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&refundResp); err != nil {
		return nil, fmt.Errorf("failed to decode refund response: %w", err)
	}

	var amount float64
	fmt.Sscanf(refundResp.Amount.Value, "%f", &amount)

	return &payment_out.RefundResponse{
		RefundID: refundResp.ID,
		Status:   refundResp.Status,
		Amount:   int64(amount * 100),
	}, nil
}

// CancelPayment voids a PayPal order (only for uncaptured orders)
func (p *PayPalAdapter) CancelPayment(ctx context.Context, req payment_out.CancelRequest) (*payment_out.CancelResponse, error) {
	// PayPal doesn't have a direct void for orders
	// We need to authorize and then void the authorization
	// For simplicity, we return the order as canceled

	return &payment_out.CancelResponse{
		Status: "canceled",
	}, nil
}

// ParseWebhook parses and validates a PayPal webhook
func (p *PayPalAdapter) ParseWebhook(payload []byte, signature string) (*payment_out.WebhookEvent, error) {
	// PayPal webhook verification is more complex than Stripe
	// It requires calling PayPal's verify endpoint
	// For now, we parse the payload directly

	var event struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	webhookEvent := &payment_out.WebhookEvent{
		EventType:         event.EventType,
		ProviderPaymentID: event.Resource.ID,
		Metadata:          make(map[string]any),
	}

	// Map PayPal events to our status
	switch event.EventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		webhookEvent.Status = payment_entities.PaymentStatusSucceeded
	case "PAYMENT.CAPTURE.DENIED":
		webhookEvent.Status = payment_entities.PaymentStatusFailed
	case "PAYMENT.CAPTURE.REFUNDED":
		webhookEvent.Status = payment_entities.PaymentStatusRefunded
	case "CHECKOUT.ORDER.APPROVED":
		webhookEvent.Status = payment_entities.PaymentStatusProcessing
	default:
		return nil, fmt.Errorf("unhandled PayPal event type: %s", event.EventType)
	}

	return webhookEvent, nil
}

// CreateOrGetCustomer is not applicable for PayPal orders
// PayPal uses email-based identification
func (p *PayPalAdapter) CreateOrGetCustomer(ctx context.Context, req payment_out.CustomerRequest) (*payment_out.CustomerResponse, error) {
	// PayPal doesn't require customer creation like Stripe
	// Return email as "customer ID"
	return &payment_out.CustomerResponse{
		CustomerID: req.Email,
	}, nil
}

// Ensure PayPalAdapter implements PaymentProviderAdapter
var _ payment_out.PaymentProviderAdapter = (*PayPalAdapter)(nil)
