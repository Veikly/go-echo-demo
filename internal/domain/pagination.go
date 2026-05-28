package domain

import "encoding/json"

type PageCursor struct {
	V          int             `json:"v"`           // cursor version
	Resource   string          `json:"resource"`    // 数据源
	Profile    string          `json:"profile"`     // 查询能力
	Sort       string          `json:"sort"`        // 排序字段和类型
	FilterHash string          `json:"filter_hash"` // 业务筛选规则哈希值
	Direction  string          `json:"direction"`   // 排序方向
	Exp        *int64          `json:"exp"`         // 过期时间 避免cursor长期有效
	Data       json.RawMessage `json:"data"`        // 具体的分页位置值 不同类型data也不同 { "value": "2026-05-20T10:00:00Z","id": "post_020"}
}

type PageRequest struct {
	Limit      int
	Profile    string
	Sort       string
	Cursor     *PageCursor
	FilterHash string
}

type PageResult[T any] struct {
	Items      []T
	Limit      int
	Profile    string
	Sort       string
	HasNext    bool
	NextCursor *PageCursor
}
