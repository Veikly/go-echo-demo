package request

import httppagination "go-echo-demo/delivery/http/pagination"

// TaskPageQuery task 分页查询请求参数。
// 嵌入 BasePageQuery 获得 scene/cursor/direction/limit。
type TaskPageQuery struct {
	httppagination.BasePageQuery
	Status        *int   `query:"status"`         // task 状态，对应 enums.TaskStatus
	Title         string `query:"title"`          // 精确匹配 title（场景 task.by_status_title 使用）
	CreatedAfter  string `query:"created_after"`  // RFC3339，筛选 created_at > 该值（场景 task.by_created_at 使用）
	CreatedBefore string `query:"created_before"` // RFC3339，筛选 created_at < 该值（场景 task.by_created_at 使用）
	Desc          *bool  `query:"desc"`           // 排序方向，默认 true（降序）
}
