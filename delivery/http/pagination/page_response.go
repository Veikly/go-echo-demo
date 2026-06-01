package pagination

// PageResponse HTTP 响应结构（统一格式）。
type PageResponse[DTO any] struct {
	Items      []DTO  `json:"items"`
	NextCursor string `json:"nextCursor"`
	HasMore    bool   `json:"hasMore"`
	TotalCount *int64 `json:"totalCount,omitempty"`
}
