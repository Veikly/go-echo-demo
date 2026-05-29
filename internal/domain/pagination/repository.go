package pagination

import "context"

type Repository[T any] interface {
	Query(ctx context.Context, p PageQuery) (PageResult[T], error)
	Get(ctx context.Context, id string) (T, error)
}
