package common

import (
	"context"
	"fmt"
	"reflect"
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

func (service *BaseQueryService[T]) Search(ctx context.Context, s Search) ([]T, error) {
	gameEvents, err := service.Reader.Search(ctx, s)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error filtering. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	return gameEvents, nil
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

	s := NewSearchByAggregation(ctx, searchParams, resultOptions, svc.Audience)

	return &s, nil
}
