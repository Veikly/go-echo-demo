package pagination

import (
	"context"
	"go-echo-demo/internal/domain/pagination"
)

type QueryUseCase[T, DTO any] struct {
	repo        pagination.Repository[T]
	registry    *pagination.Registry
	toDTO       func(T) DTO
	injectRules func(ctx context.Context, q pagination.PageQuery) (pagination.PageQuery, error)
}

type QueryUseCaseConfig[T, DTO any] struct {
	Repo        pagination.Repository[T]
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

// ExecuteInput Use Case 的输入参数（由 Delivery 层传入）。
type ExecuteInput struct {
	Scene          pagination.SceneID
	Params         pagination.SceneParams
	Cursor         string
	Limit          int
	WithTotalCount bool
}

// Execute Use Case 执行逻辑。
func (uc *QueryUseCase[T, DTO]) Execute(ctx context.Context, input ExecuteInput) (pagination.PageResult[DTO], error) {
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
