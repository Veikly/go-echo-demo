package pagination

import (
	"context"
	"go-echo-demo/internal/adapters"
	"go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/usecase/usecaseio"
)

type QueryUseCase[T, DTO any] struct {
	repo        adapters.Repository[T]
	registry    *pagination.Registry
	toDTO       func(T) DTO
	injectRules func(ctx context.Context, q pagination.PageQuery) (pagination.PageQuery, error)
}

type QueryUseCaseConfig[T, DTO any] struct {
	Repo        adapters.Repository[T]
	Registry    *pagination.Registry
	ToDTO       func(T) DTO
	InjectRules func(ctx context.Context, q pagination.PageQuery) (pagination.PageQuery, error)
}

func NewQueryUseCase[T, DTO any](cfg QueryUseCaseConfig[T, DTO]) *QueryUseCase[T, DTO] {
	injectRules := cfg.InjectRules
	if injectRules == nil {
		injectRules = func(_ context.Context, q pagination.PageQuery) (pagination.PageQuery, error) { return q, nil }
	}
	return &QueryUseCase[T, DTO]{
		repo:        cfg.Repo,
		registry:    cfg.Registry,
		toDTO:       cfg.ToDTO,
		injectRules: injectRules,
	}
}

// Execute Use Case 执行逻辑。
func (uc *QueryUseCase[T, DTO]) Execute(ctx context.Context, input usecaseio.ExecuteInput) (pagination.PageResult[DTO], error) {
	baseQuery, err := uc.registry.Build(input.Scene, input.Params)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	q, err := pagination.ApplyPaging(baseQuery, input.Cursor, input.Limit)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}
	q.IncludeTotalCount = input.WithTotalCount

	q, err = uc.injectRules(ctx, q)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	raw, err := uc.repo.Query(ctx, q)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	dtos := make([]DTO, len(raw.Items))
	for i, item := range raw.Items {
		dtos[i] = uc.toDTO(item)
	}

	return pagination.PageResult[DTO]{
		Items:      dtos,
		NextCursor: raw.NextCursor,
		HasMore:    raw.HasMore,
		TotalCount: raw.TotalCount,
	}, nil
}
