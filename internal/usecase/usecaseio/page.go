package usecaseio

import "go-echo-demo/internal/domain/pagination"

// ExecuteInput Use Case 的输入参数（由 Delivery 层传入）。
type ExecuteInput struct {
	Scene          pagination.SceneID
	Params         pagination.SceneParams
	Cursor         string
	Limit          int
	WithTotalCount bool
}
