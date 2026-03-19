package query_controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	shared "github.com/resource-ownership/go-common/pkg/common"
)

// SubscriptionQueryController handles subscription query endpoints
type SubscriptionQueryController struct {
	subscriptionReader billing_out.SubscriptionReader
	planReader         billing_out.PlanReader
}

// NewSubscriptionQueryController creates a new subscription query controller
func NewSubscriptionQueryController(c container.Container) *SubscriptionQueryController {
	var subscriptionReader billing_out.SubscriptionReader
	var planReader billing_out.PlanReader

	if err := c.Resolve(&subscriptionReader); err != nil {
		slog.Error("Failed to resolve SubscriptionReader", "error", err)
	}

	if err := c.Resolve(&planReader); err != nil {
		slog.Error("Failed to resolve PlanReader", "error", err)
	}

	return &SubscriptionQueryController{
		subscriptionReader: subscriptionReader,
		planReader:         planReader,
	}
}

// SubscriptionResponse is the API response for subscription data
type SubscriptionResponse struct {
	ID            uuid.UUID                             `json:"id"`
	PlanID        uuid.UUID                             `json:"plan_id"`
	Plan          *PlanResponse                         `json:"plan,omitempty"`
	BillingPeriod billing_entities.BillingPeriodType    `json:"billing_period"`
	Status        billing_entities.SubscriptionStatus   `json:"status"`
	StartAt       int64                                 `json:"start_at"`
	EndAt         *int64                                `json:"end_at,omitempty"`
	IsFree        bool                                  `json:"is_free"`
	IsPro         bool                                  `json:"is_pro"`
	IsElite       bool                                  `json:"is_elite"`
	Features      []string                              `json:"features"`
	CreatedAt     int64                                 `json:"created_at"`
	UpdatedAt     int64                                 `json:"updated_at"`
}

// PriceInfo represents pricing for a specific billing period
type PriceInfo struct {
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	TotalDiscount float64 `json:"total_discount,omitempty"`
	YearlyTotal   float64 `json:"yearly_total,omitempty"` // Total amount for yearly billing
}

// PlanResponse is the API response for plan data
type PlanResponse struct {
	ID                   uuid.UUID                       `json:"id"`
	Name                 string                          `json:"name"`
	Description          string                          `json:"description"`
	Kind                 billing_entities.PlanKindType   `json:"kind"`
	IsFree               bool                            `json:"is_free"`
	IsAvailable          bool                            `json:"is_available"`
	PriceAmount          float64                         `json:"price_amount"`          // Monthly price (deprecated, use prices)
	PriceCurrency        string                          `json:"price_currency"`        // Default currency
	BillingInterval      string                          `json:"billing_interval"`      // Default billing interval
	Prices               map[string]PriceInfo            `json:"prices"`                // Pricing for default/requested currency
	AllPrices            map[string][]PriceInfo          `json:"all_prices"`            // All currency prices per billing period
	Regions              []string                        `json:"regions"`               // Regions this plan is available in (empty = all)
	Languages            []string                        `json:"languages,omitempty"`   // Supported languages
	Features             []string                        `json:"features"`
	DisplayPriorityScore int                             `json:"display_priority_score"`
}

// regionToCurrency maps region codes to their default currency.
// Used to filter multi-currency prices for a specific region.
var regionToCurrency = map[string]string{
	"NA":    "USD",
	"BR":    "BRL",
	"EU":    "EUR",
	"LATAM": "MXN",
	"ASIA":  "CNY",
}

// findPriceForCurrency returns the price matching the given currency from a price list.
// Falls back to USD, then the first available price.
func findPriceForCurrency(prices []billing_entities.Price, currency string) *billing_entities.Price {
	for i, p := range prices {
		if strings.EqualFold(p.Currency, currency) {
			return &prices[i]
		}
	}
	// Fallback: try USD
	for i, p := range prices {
		if strings.EqualFold(p.Currency, "USD") {
			return &prices[i]
		}
	}
	if len(prices) > 0 {
		return &prices[0]
	}
	return nil
}

// GetCurrentSubscriptionHandler handles GET /subscriptions/current
// SECURITY: This endpoint requires authentication and only returns the authenticated user's subscription
func (c *SubscriptionQueryController) GetCurrentSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SECURITY: Verify authentication (critical for subscription data)
	resourceOwner := requireAuthentication(w, r)
	if resourceOwner == nil {
		return // Response already written
	}

	if c.subscriptionReader == nil {
		slog.ErrorContext(ctx, "SubscriptionReader not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"error":"Subscription service unavailable"}`))
		return
	}

	subscription, err := c.subscriptionReader.GetCurrentSubscription(ctx, *resourceOwner)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get current subscription", "error", err, "user_id", resourceOwner.UserID)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	// If no subscription found, return a default free response
	if subscription == nil {
		response := SubscriptionResponse{
			Status:   billing_entities.SubscriptionStatusActive,
			IsFree:   true,
			IsPro:    false,
			IsElite:  false,
			Features: []string{"basic_matchmaking", "queue_access"},
		}
		common.WriteSuccess(w, response)
		return
	}

	// Get plan details if available
	var planResponse *PlanResponse
	if c.planReader != nil {
		plan, err := c.planReader.GetByID(ctx, subscription.PlanID)
		if err != nil {
			slog.WarnContext(ctx, "Failed to get plan details", "error", err, "plan_id", subscription.PlanID)
		} else if plan != nil {
			planResponse = &PlanResponse{
				ID:                   plan.ID,
				Name:                 plan.Name,
				Description:          plan.Description,
				Kind:                 plan.Kind,
				IsFree:               plan.IsFree,
				IsAvailable:          plan.IsAvailable,
				DisplayPriorityScore: plan.DisplayPriorityScore,
				Features:             getPlanFeatures(plan),
			}

			// Get price info if available (use monthly as default)
			if prices, ok := plan.Prices[billing_entities.BillingPeriodMonthly]; ok && len(prices) > 0 {
				planResponse.PriceAmount = prices[0].Amount
				planResponse.PriceCurrency = prices[0].Currency
				planResponse.BillingInterval = string(billing_entities.BillingPeriodMonthly)
			}
		}
	}

	// Determine plan tier
	isPro := false
	isElite := false
	features := []string{"basic_matchmaking", "queue_access"}

	if planResponse != nil {
		features = planResponse.Features
		switch planResponse.Kind {
		case billing_entities.PlanKindTypePro:
			isPro = true
		case billing_entities.PlanKindTypeTeam, billing_entities.PlanKindTypeBusiness, billing_entities.PlanKindTypeCustom:
			isPro = true
			isElite = true
		}
	}

	var endAtMs *int64
	if subscription.EndAt != nil {
		ts := subscription.EndAt.UnixMilli()
		endAtMs = &ts
	}

	response := SubscriptionResponse{
		ID:            subscription.ID,
		PlanID:        subscription.PlanID,
		Plan:          planResponse,
		BillingPeriod: subscription.BillingPeriod,
		Status:        subscription.Status,
		StartAt:       subscription.StartAt.UnixMilli(),
		EndAt:         endAtMs,
		IsFree:        subscription.IsFree,
		IsPro:         isPro,
		IsElite:       isElite,
		Features:      features,
		CreatedAt:     subscription.CreatedAt.UnixMilli(),
		UpdatedAt:     subscription.UpdatedAt.UnixMilli(),
	}

	common.WriteSuccess(w, response)
}

// getPlanFeatures extracts feature names from a plan
// Priority: 1) Direct Features field, 2) OperationLimits keys, 3) Default fallback
func getPlanFeatures(plan *billing_entities.Plan) []string {
	// First, check if plan has explicit Features defined
	if len(plan.Features) > 0 {
		return plan.Features
	}

	// Second, extract from OperationLimits keys
	if len(plan.OperationLimits) > 0 {
		features := make([]string, 0, len(plan.OperationLimits))
		for key := range plan.OperationLimits {
			features = append(features, string(key))
		}
		return features
	}

	// Default features based on plan kind
	switch plan.Kind {
	case billing_entities.PlanKindTypePro:
		return []string{"priority_matchmaking", "queue_access", "analytics", "advanced_stats"}
	case billing_entities.PlanKindTypeBusiness, billing_entities.PlanKindTypeTeam:
		return []string{"unlimited_matchmaking", "priority_queue", "analytics", "premium_support", "team_management"}
	default:
		return []string{"basic_matchmaking", "queue_access"}
	}
}

// GetSubscriptionByIDHandler handles GET /subscriptions/{subscription_id}
// SECURITY: Only returns subscription if user owns it or is admin
func (c *SubscriptionQueryController) GetSubscriptionByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SECURITY: Verify authentication
	resourceOwner := requireAuthentication(w, r)
	if resourceOwner == nil {
		return
	}

	vars := mux.Vars(r)
	subscriptionIDStr := vars["subscription_id"]
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"error":"Invalid subscription ID"}`))
		return
	}

	if c.subscriptionReader == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"error":"Subscription service unavailable"}`))
		return
	}

	// Use GetByID to retrieve subscription
	subscription, err := c.subscriptionReader.GetByID(ctx, subscriptionID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get subscription by ID", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	if subscription == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":"Subscription not found"}`))
		return
	}

	// Security check - verify user owns this subscription or is admin
	if subscription.ResourceOwner.UserID != resourceOwner.UserID && !shared.IsAdmin(ctx) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"error":"Access denied"}`))
		return
	}

	// Get plan details
	var planResponse *PlanResponse
	if c.planReader != nil {
		plan, _ := c.planReader.GetByID(ctx, subscription.PlanID)
		if plan != nil {
			planResponse = &PlanResponse{
				ID:          plan.ID,
				Name:        plan.Name,
				Description: plan.Description,
				Kind:        plan.Kind,
				IsFree:      plan.IsFree,
				IsAvailable: plan.IsAvailable,
				Features:    getPlanFeatures(plan),
			}
		}
	}

	isPro := planResponse != nil && (planResponse.Kind == billing_entities.PlanKindTypePro ||
		planResponse.Kind == billing_entities.PlanKindTypeTeam ||
		planResponse.Kind == billing_entities.PlanKindTypeBusiness)
	isElite := planResponse != nil && (planResponse.Kind == billing_entities.PlanKindTypeTeam ||
		planResponse.Kind == billing_entities.PlanKindTypeBusiness ||
		planResponse.Kind == billing_entities.PlanKindTypeCustom)

	features := []string{"basic_matchmaking", "queue_access"}
	if planResponse != nil {
		features = planResponse.Features
	}

	var endAtMs *int64
	if subscription.EndAt != nil {
		ts := subscription.EndAt.UnixMilli()
		endAtMs = &ts
	}

	response := SubscriptionResponse{
		ID:            subscription.ID,
		PlanID:        subscription.PlanID,
		Plan:          planResponse,
		BillingPeriod: subscription.BillingPeriod,
		Status:        subscription.Status,
		StartAt:       subscription.StartAt.UnixMilli(),
		EndAt:         endAtMs,
		IsFree:        subscription.IsFree,
		IsPro:         isPro,
		IsElite:       isElite,
		Features:      features,
		CreatedAt:     subscription.CreatedAt.UnixMilli(),
		UpdatedAt:     subscription.UpdatedAt.UnixMilli(),
	}

	common.WriteSuccess(w, response)
}

// PlanQueryController handles plan query endpoints
type PlanQueryController struct {
	planReader billing_out.PlanReader
}

// NewPlanQueryController creates a new plan query controller
func NewPlanQueryController(c container.Container) *PlanQueryController {
	var planReader billing_out.PlanReader

	if err := c.Resolve(&planReader); err != nil {
		slog.Error("Failed to resolve PlanReader", "error", err)
	}

	return &PlanQueryController{
		planReader: planReader,
	}
}

// ListAvailablePlansHandler handles GET /plans
// Returns all available plans (public endpoint)
// Query params:
//   - region: filter prices by region (NA, BR, EU, LATAM, ASIA). Returns all currencies if omitted.
//   - currency: filter prices by currency code (USD, BRL, EUR, MXN, CNY). Takes precedence over region.
func (c *PlanQueryController) ListAvailablePlansHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if c.planReader == nil {
		slog.ErrorContext(ctx, "PlanReader not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"error":"Plan service unavailable"}`))
		return
	}

	plans, err := c.planReader.GetAvailablePlans(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get available plans", "error", err)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	// Determine target currency from query params
	requestedCurrency := ""
	if currency := r.URL.Query().Get("currency"); currency != "" {
		requestedCurrency = strings.ToUpper(currency)
	} else if region := r.URL.Query().Get("region"); region != "" {
		if cur, ok := regionToCurrency[strings.ToUpper(region)]; ok {
			requestedCurrency = cur
		}
	}

	// Convert to response format
	response := make([]PlanResponse, 0, len(plans))
	for _, plan := range plans {
		pr := PlanResponse{
			ID:                   plan.ID,
			Name:                 plan.Name,
			Description:          plan.Description,
			Kind:                 plan.Kind,
			IsFree:               plan.IsFree,
			IsAvailable:          plan.IsAvailable,
			DisplayPriorityScore: plan.DisplayPriorityScore,
			Features:             getPlanFeatures(plan),
			Prices:               make(map[string]PriceInfo),
			AllPrices:            make(map[string][]PriceInfo),
			Regions:              plan.Regions,
			Languages:            plan.Languages,
		}

		// Build all_prices (every currency per period) and prices (filtered or default)
		for period, prices := range plan.Prices {
			// Build all_prices array for this period
			allPricesForPeriod := make([]PriceInfo, 0, len(prices))
			for _, p := range prices {
				pi := PriceInfo{
					Amount:        p.Amount,
					Currency:      p.Currency,
					TotalDiscount: p.TotalDiscount,
				}
				if period == billing_entities.BillingPeriodYearly {
					pi.YearlyTotal = p.Amount * 12
				}
				allPricesForPeriod = append(allPricesForPeriod, pi)
			}
			pr.AllPrices[string(period)] = allPricesForPeriod

			// Build filtered prices map (single currency per period)
			var selectedPrice *billing_entities.Price
			if requestedCurrency != "" {
				selectedPrice = findPriceForCurrency(prices, requestedCurrency)
			} else if len(prices) > 0 {
				selectedPrice = findPriceForCurrency(prices, "USD")
			}

			if selectedPrice != nil {
				priceInfo := PriceInfo{
					Amount:        selectedPrice.Amount,
					Currency:      selectedPrice.Currency,
					TotalDiscount: selectedPrice.TotalDiscount,
				}
				if period == billing_entities.BillingPeriodYearly {
					priceInfo.YearlyTotal = selectedPrice.Amount * 12
				}
				pr.Prices[string(period)] = priceInfo
			}
		}

		// Set default price info (monthly as default for backwards compatibility)
		if prices, ok := plan.Prices[billing_entities.BillingPeriodMonthly]; ok && len(prices) > 0 {
			defaultPrice := findPriceForCurrency(prices, func() string {
				if requestedCurrency != "" {
					return requestedCurrency
				}
				return "USD"
			}())
			if defaultPrice != nil {
				pr.PriceAmount = defaultPrice.Amount
				pr.PriceCurrency = defaultPrice.Currency
				pr.BillingInterval = string(billing_entities.BillingPeriodMonthly)
			}
		}

		response = append(response, pr)
	}

	common.WriteSuccess(w, response)
}

// GetPlanByIDHandler handles GET /plans/{plan_id}
func (c *PlanQueryController) GetPlanByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	planIDStr := vars["plan_id"]
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"error":"Invalid plan ID"}`))
		return
	}

	if c.planReader == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"error":"Plan service unavailable"}`))
		return
	}

	plan, err := c.planReader.GetByID(ctx, planID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get plan", "error", err, "plan_id", planID)
		common.WriteErrorFromDomainError(w, err)
		return
	}

	if plan == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":"Plan not found"}`))
		return
	}

	response := PlanResponse{
		ID:                   plan.ID,
		Name:                 plan.Name,
		Description:          plan.Description,
		Kind:                 plan.Kind,
		IsFree:               plan.IsFree,
		IsAvailable:          plan.IsAvailable,
		DisplayPriorityScore: plan.DisplayPriorityScore,
		Features:             getPlanFeatures(plan),
		Prices:               make(map[string]PriceInfo),
		AllPrices:            make(map[string][]PriceInfo),
		Regions:              plan.Regions,
		Languages:            plan.Languages,
	}

	// Determine target currency from query params
	requestedCurrency := ""
	if currency := r.URL.Query().Get("currency"); currency != "" {
		requestedCurrency = strings.ToUpper(currency)
	} else if region := r.URL.Query().Get("region"); region != "" {
		if cur, ok := regionToCurrency[strings.ToUpper(region)]; ok {
			requestedCurrency = cur
		}
	}

	// Build all_prices and filtered prices
	for period, prices := range plan.Prices {
		allPricesForPeriod := make([]PriceInfo, 0, len(prices))
		for _, p := range prices {
			pi := PriceInfo{
				Amount:        p.Amount,
				Currency:      p.Currency,
				TotalDiscount: p.TotalDiscount,
			}
			if period == billing_entities.BillingPeriodYearly {
				pi.YearlyTotal = p.Amount * 12
			}
			allPricesForPeriod = append(allPricesForPeriod, pi)
		}
		response.AllPrices[string(period)] = allPricesForPeriod

		var selectedPrice *billing_entities.Price
		if requestedCurrency != "" {
			selectedPrice = findPriceForCurrency(prices, requestedCurrency)
		} else if len(prices) > 0 {
			selectedPrice = findPriceForCurrency(prices, "USD")
		}

		if selectedPrice != nil {
			priceInfo := PriceInfo{
				Amount:        selectedPrice.Amount,
				Currency:      selectedPrice.Currency,
				TotalDiscount: selectedPrice.TotalDiscount,
			}
			if period == billing_entities.BillingPeriodYearly {
				priceInfo.YearlyTotal = selectedPrice.Amount * 12
			}
			response.Prices[string(period)] = priceInfo
		}
	}

	// Get price info if available (use monthly as default)
	if prices, ok := plan.Prices[billing_entities.BillingPeriodMonthly]; ok && len(prices) > 0 {
		defaultPrice := findPriceForCurrency(prices, func() string {
			if requestedCurrency != "" {
				return requestedCurrency
			}
			return "USD"
		}())
		if defaultPrice != nil {
			response.PriceAmount = defaultPrice.Amount
			response.PriceCurrency = defaultPrice.Currency
			response.BillingInterval = string(billing_entities.BillingPeriodMonthly)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}
