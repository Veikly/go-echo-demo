package adapters

import (
	"context"
	"go-echo-demo/internal/domain/pagination"
)

type Repository[T any] interface {
	Query(ctx context.Context, p pagination.PageQuery) (pagination.PageResult[T], error)
	Get(ctx context.Context, id string) (T, error)
}
