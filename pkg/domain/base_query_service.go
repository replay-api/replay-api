package common

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
)

type BaseQueryService[T any] struct {
	Reader          Searchable[T]
	QueryableFields map[string]bool
	ReadableFields  map[string]bool
	MaxPageSize     uint
	Audience        IntendedAudienceKey
	name            string
}

func (service *BaseQueryService[T]) GetName() string {
	if service.name != "" {
		return service.name
	}

	service.name = reflect.TypeOf(service).Name()

	return service.name
}

// / GetByID returns a single entity by its ID using ClientApplicationAudienceIDKey as the intended audience.
func (service *BaseQueryService[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	params := []SearchAggregation{
		{
			Params: []SearchParameter{
				{
					ValueParams: []SearchableValue{
						{
							Field: "ID",
							Values: []interface{}{
								id,
							},
						},
					},
				},
			},
		},
	}

	visibility := SearchVisibilityOptions{
		RequestSource:    GetResourceOwner(ctx),
		IntendedAudience: ClientApplicationAudienceIDKey,
	}

	result := SearchResultOptions{
		Skip:  0,
		Limit: 1,
	}

	search := Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}

	entities, err := service.Reader.Search(ctx, search)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error searching. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	res := entities[0]

	return &res, nil
}

func (service *BaseQueryService[T]) Search(ctx context.Context, s Search) ([]T, error) {
	var omitFields []string
	var pickFields []string
	for fieldName, isReadable := range service.ReadableFields {
		if !isReadable {
			omitFields = append(omitFields, fieldName)
			continue
		}

		pickFields = append(pickFields, fieldName)
	}

	if len(omitFields) > 0 {
		slog.Info("Omitting fields", "fields", omitFields)
	}

	s.ResultOptions.OmitFields = omitFields
	s.ResultOptions.PickFields = pickFields

	entities, err := service.Reader.Search(ctx, s)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error filtering. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	return entities, nil
}

func (svc *BaseQueryService[T]) Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error) {
	err := ValidateSearchParameters(searchParams, svc.QueryableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating search parameters: %v", err)
	}

	err = ValidateResultOptions(resultOptions, svc.ReadableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating result options: %v", err)
	}

	intendedAud := GetIntendedAudience(ctx)

	if intendedAud == nil {
		intendedAud = &svc.Audience
	}

	s := NewSearchByAggregation(ctx, searchParams, resultOptions, *intendedAud)

	return &s, nil
}
