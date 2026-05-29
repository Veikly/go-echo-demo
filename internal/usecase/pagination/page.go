package pagination

import (
	"context"
	"go-echo-demo/internal/domain/pagination"
)

type QueryUseCase[T, DTO any] struct {
	repo     pagination.Repository[T]
	registry *pagination.Registry
	toDTO    func(T) DTO
	// injectRules 注入不可绕过的业务规则（如租户隔离）。
	// 每次 Execute 前调用，返回追加了规则后的 PageQuery。
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
	Scene  pagination.SceneID
	Params pagination.SceneParams
	Cursor string
	Dir    pagination.CursorDir
	Limit  int
}

// Execute Use Case 执行逻辑。
func (uc *QueryUseCase[T, DTO]) Execute(ctx context.Context, input ExecuteInput) (pagination.PageResult[DTO], error) {
	// 1. 从 SceneRegistry 构建基础 PageQuery
	baseQuery, err := uc.registry.Build(input.Scene, input.Params)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	// 2. 叠加分页参数
	q, err := pagination.ApplyPaging(baseQuery, input.Cursor, input.Dir, input.Limit)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	// 3. 注入不可绕过的业务规则（租户隔离、权限过滤等）
	q, err = uc.injectRules(ctx, q)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	// 4. 执行查询
	raw, err := uc.repo.Query(ctx, q)
	if err != nil {
		return pagination.PageResult[DTO]{}, err
	}

	// 5. 领域实体 → DTO
	dtos := make([]DTO, len(raw.Items))
	for i, item := range raw.Items {
		dtos[i] = uc.toDTO(item)
	}

	return pagination.PageResult[DTO]{
		Items:      dtos,
		NextCursor: raw.NextCursor,
		PrevCursor: raw.PrevCursor,
		HasMore:    raw.HasMore,
		TotalCount: raw.TotalCount,
	}, nil
}
