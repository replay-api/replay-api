package common

import (
	"context"

	"github.com/google/uuid"
)

type Searchable[T any] interface {
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Search(ctx context.Context, s Search) ([]T, error)
	Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error)
}
