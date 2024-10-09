package common

import "context"

type Searchable[T any] interface {
	Search(ctx context.Context, s Search) ([]T, error)
	Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error)
}
