package billing_services

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

type PlanQueryService struct {
	common.BaseQueryService[billing_entities.Plan]
}

func NewPlanQueryService(eventReader billing_out.PlanReader) billing_in.PlanReader {
	queryableFields := map[string]bool{
		"ID":                   true,
		"Name":                 true,
		"Description":          true,
		"Kind":                 true,
		"CustomerType":         true,
		"Prices":               true,
		"OperationLimits":      true,
		"IsFree":               true,
		"IsAvailable":          true,
		"IsLegacy":             true,
		"IsActive":             true,
		"DisplayPriorityScore": true,
		"Regions":              true,
		"Languages":            true,
		"EffectiveDate":        true,
		"ExpirationDate":       true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
	}

	readableFields := map[string]bool{
		"ID":                   true,
		"Name":                 true,
		"Description":          true,
		"Kind":                 true,
		"CustomerType":         true,
		"Prices":               true,
		"OperationLimits":      true,
		"IsFree":               true,
		"IsAvailable":          true,
		"IsLegacy":             true,
		"IsActive":             true,
		"DisplayPriorityScore": true,
		"Regions":              true,
		"Languages":            true,
		"EffectiveDate":        true,
		"ExpirationDate":       true,
		"CreatedAt":            true,
		"UpdatedAt":            true,
	}

	return &common.BaseQueryService[billing_entities.Plan]{
		Reader:          eventReader, // PlanReader embeds common.Searchable[billing_entities.Plan]
		QueryableFields: queryableFields,
		ReadableFields:  readableFields,
		MaxPageSize:     100,
		Audience:        common.UserAudienceIDKey,
	}
}
