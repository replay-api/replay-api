package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SortDirectionKey int

const (
	AscendingIDKey  SortDirectionKey = 1
	DescendingIDKey SortDirectionKey = -1
)

type SearchOperator string

const (
	EqualsOperator             SearchOperator = "eq"       // Exact match (default)
	NotEqualsOperator          SearchOperator = "ne"       // Not equal
	GreaterThanOperator        SearchOperator = "gt"       // Greater than
	LessThanOperator           SearchOperator = "lt"       // Less than
	GreaterThanOrEqualOperator SearchOperator = "gte"      // Greater than or equal
	LessThanOrEqualOperator    SearchOperator = "lte"      // Less than or equal
	ContainsOperator           SearchOperator = "contains" // Case-insensitive substring match
	StartsWithOperator         SearchOperator = "startswith"
	EndsWithOperator           SearchOperator = "endswith"
	InOperator                 SearchOperator = "in"  // Match any value in a list
	NotInOperator              SearchOperator = "nin" // Not in a list
)

const DefaultPageSize uint = 50

type IntendedAudienceKey string

const (
	TenantAudienceIDKey            IntendedAudienceKey = "TenantAudience"
	ClientApplicationAudienceIDKey IntendedAudienceKey = "ClientAudience"
	GroupAudienceIDKey             IntendedAudienceKey = "GroupAudience"
	UserAudienceIDKey              IntendedAudienceKey = "UserAudience"
)

type AggregationTypeKey string

type SearchableValue struct {
	Field    string
	Values   []interface{}
	Operator SearchOperator
}

type SearchableDateRange struct {
	Field string
	Min   *time.Time
	Max   *time.Time
}

type SearchableDurationRange struct {
	Field string
	Min   *time.Duration // opcionais
	Max   *time.Duration
}

type SortableField struct {
	Field     string
	Direction SortDirectionKey
}

type SearchParameter struct {
	ValueParams       []SearchableValue         `json:"values" bson:"value_params"`
	DateParams        []SearchableDateRange     `json:"date" bson:"date_params"`
	DurationParams    []SearchableDurationRange `json:"time" bson:"duration_params"`
	AggregationParams []SearchAggregation       `json:"aggregate" bson:"aggregation_params"`
	Operator          SearchOperator            `json:"operator" bson:"operator"` // if not provided, default to InOperator
}

type SearchAggregation struct {
	Params    []SearchParameter  `json:"params" bson:"params"`
	ParamType AggregationTypeKey `json:"type" bson:"type"` // REVIEW: redundant w/ "Operator" or use nested?
}

type SearchResultOptions struct {
	Skip       uint     `json:"skip" bson:"skip"`        // default = 0
	Limit      uint     `json:"limit" bson:"limit"`      // default = 50
	PickFields []string `json:"pick" bson:"pick_fields"` // if not informed, pick all
	OmitFields []string `json:"omit" bson:"omit_fields"` // if not informed, doesnt omit any
}

type SearchVisibilityOptions struct {
	RequestSource    ResourceOwner       `json:"-" bson:"request_source"`
	IntendedAudience IntendedAudienceKey `json:"-" bson:"intended_audience"` // Default: User
}

type Search struct {
	SearchParams      []SearchAggregation     `json:"search_params" bson:"search_params"`
	ResultOptions     SearchResultOptions     `json:"result_options" bson:"result_options"`
	SortOptions       []SortableField         `json:"sort_options" bson:"sort_options"`
	VisibilityOptions SearchVisibilityOptions `json:"-" bson:"visibility_options"`
}

func NewSearchByID(ctx context.Context, id uuid.UUID, audienceLevel IntendedAudienceKey) Search {
	v := []SearchableValue{
		{
			Field: "ID",
			Values: []interface{}{
				id,
			},
		},
	}

	p := []SearchParameter{
		{
			ValueParams: v,
		},
	}

	pipe := []SearchAggregation{
		{
			Params: p,
		},
	}

	visibility := SearchVisibilityOptions{
		RequestSource:    GetResourceOwner(ctx),
		IntendedAudience: audienceLevel,
	}

	result := NewSearchResultOptions(0, 1)

	return Search{
		SearchParams:      pipe,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

func NewSearchResultOptions(skip uint, limit uint) SearchResultOptions {
	return SearchResultOptions{
		Skip:  skip,
		Limit: limit,
	}
}

func NewSearchByAggregation(ctx context.Context, aggregationParams []SearchAggregation, resultOptions SearchResultOptions, audienceLevel IntendedAudienceKey) Search {
	visibility := SearchVisibilityOptions{
		RequestSource:    GetResourceOwner(ctx),
		IntendedAudience: audienceLevel,
	}

	if resultOptions.Limit == 0 {
		resultOptions.Limit = DefaultPageSize
	}

	return Search{
		SearchParams:      aggregationParams,
		ResultOptions:     resultOptions,
		VisibilityOptions: visibility,
	}
}

func NewSearchByValues(ctx context.Context, valueParams []SearchableValue, resultOptions SearchResultOptions, audienceLevel IntendedAudienceKey) Search {
	params := []SearchAggregation{
		{
			Params: []SearchParameter{
				{
					ValueParams: valueParams,
				},
			},
		},
	}

	return NewSearchByAggregation(ctx, params, resultOptions, audienceLevel)
}

func NewSearchByRange(ctx context.Context, dateParams []SearchableDateRange, resultOptions SearchResultOptions, audienceLevel IntendedAudienceKey) Search {
	params := []SearchAggregation{
		{
			Params: []SearchParameter{
				{
					DateParams: dateParams,
				},
			},
		},
	}

	return NewSearchByAggregation(ctx, params, resultOptions, audienceLevel)
}

func NewSearch(ctx context.Context, audienceLevel IntendedAudienceKey) Search {
	visibility := SearchVisibilityOptions{
		RequestSource:    GetResourceOwner(ctx),
		IntendedAudience: audienceLevel,
	}

	result := SearchResultOptions{
		Skip:  0,
		Limit: DefaultPageSize,
	}

	return Search{
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}

const (
	MaxRecursiveDepth = 10
)

func ValidateSearchParameters(searchParams []SearchAggregation, queryableFields map[string]bool) error {
	for _, param := range searchParams {
		for _, valueParam := range param.Params {
			err := ValidateValueParams(valueParam, queryableFields)
			if err != nil {
				return err
			}
		}

		for _, dateParam := range param.Params {
			err := ValidateDateParams(dateParam, queryableFields)
			if err != nil {
				return err
			}
		}

		for _, durationParam := range param.Params {
			err := ValidateDurationParams(durationParam, queryableFields)
			if err != nil {
				return err
			}
		}

		for _, aggregationParam := range param.Params {
			for index, aggregation := range aggregationParam.AggregationParams {
				if index >= MaxRecursiveDepth {
					return fmt.Errorf("maximum AggregationParams recursive depth %d reached", MaxRecursiveDepth)
				}

				err := ValidateSearchParameters([]SearchAggregation{aggregation}, queryableFields)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func ValidateDurationParams(durationParam SearchParameter, queryableFields map[string]bool) error {
	for _, duration := range durationParam.DurationParams {
		field := duration.Field
		if strings.HasSuffix(field, ".*") {
			prefix := strings.TrimSuffix(field, ".*")
			allowed := false
			for qField, allowedValue := range queryableFields {
				if strings.HasPrefix(qField, prefix) && allowedValue {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("filtering on DurationParams fields matching '%s.*' is not permitted", prefix)
			}
		} else {
			allowed, exists := queryableFields[field]
			if !exists || !allowed {
				return fmt.Errorf("filtering on DurationParams field '%s' is not permitted", field)
			}
		}
	}
	return nil
}

func ValidateDateParams(dateParam SearchParameter, queryableFields map[string]bool) error {
	for _, date := range dateParam.DateParams {
		field := date.Field
		if strings.HasSuffix(field, ".*") {
			prefix := strings.TrimSuffix(field, ".*")
			allowed := false
			for qField, allowedValue := range queryableFields {
				if strings.HasPrefix(qField, prefix) && allowedValue {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("filtering on DateParams fields matching '%s.*' is not permitted", prefix)
			}
		} else {
			allowed, exists := queryableFields[field]
			if !exists || !allowed {
				return fmt.Errorf("filtering on DateParams field '%s' is not permitted", field)
			}
		}
	}
	return nil
}

func ValidateValueParams(valueParam SearchParameter, queryableFields map[string]bool) error {
	for _, value := range valueParam.ValueParams {
		field := value.Field
		if strings.HasSuffix(field, ".*") {
			prefix := strings.TrimSuffix(field, ".*")
			allowed := false
			for qField, allowedValue := range queryableFields {
				if strings.HasPrefix(qField, prefix) && allowedValue {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("filtering on ValueParams fields matching '%s.*' is not permitted", prefix)
			}
		} else {
			allowed, exists := queryableFields[field]
			if !exists || !allowed {
				return fmt.Errorf("filtering on ValueParams field '%s' is not permitted", field)
			}
		}
	}
	return nil
}

func ValidateResultOptions(resultOptions SearchResultOptions, returnableFields map[string]bool) error {
	for _, field := range resultOptions.PickFields {
		if _, allowed := returnableFields[field]; !allowed {
			return fmt.Errorf("returning field '%s' is not permitted (1)", field)
		}

		if allowed, exists := returnableFields[field]; !exists || !allowed {
			if strings.HasSuffix(field, ".*") {
				prefix := strings.TrimSuffix(field, ".*")
				allowed = false
				for f := range returnableFields {
					if strings.HasPrefix(f, prefix) {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				return fmt.Errorf("returning field '%s' is strictly forbidden (2)", field)
			}
		}
	}

	for _, field := range resultOptions.OmitFields {
		if _, allowed := returnableFields[field]; !allowed {
			return fmt.Errorf("omitting field '%s' is not permitted", field)
		}
	}

	if resultOptions.Limit == 0 {
		return fmt.Errorf("limit must be a positive integer")
	}

	return nil
}
