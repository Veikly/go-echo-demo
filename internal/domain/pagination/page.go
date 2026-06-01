package pagination

import "time"

const defaultLimit = 20

type FilterOp string

const (
	FilterOpEq            FilterOp = "=="
	FilterOpNe            FilterOp = "!="
	FilterOpGt            FilterOp = ">"
	FilterOpGte           FilterOp = ">="
	FilterOpLt            FilterOp = "<"
	FilterOpLte           FilterOp = "<="
	FilterOpIn            FilterOp = "in"
	FilterOpNotIn         FilterOp = "not-in"
	FilterOpArrayContains FilterOp = "array-contains"
)

type LogicOp string

const (
	LogicAnd LogicOp = "AND"
	LogicOr  LogicOp = "OR"
)

// 支持嵌套的过滤条件树
// 叶节点：Field + Op + Value 非空，Children 为空。
// 组合节点：Children 非空，Logic 指定 AND / OR，Field/Op/Value 忽略。
type FilterCriteria struct {
	Field    string
	Value    any
	Op       FilterOp
	Logic    LogicOp
	Children []FilterCriteria
}

type SortField struct {
	Field      string
	Descending bool
}

// PageQuery 分页查询参数
type PageQuery struct {
	Cursor            *CursorData // 解码后的游标，nil 表示首页
	Limit             int
	Filters           []FilterCriteria
	SortBy            []SortField
	IncludeTotalCount bool // 是否需要返回查询总数
}

type PageResult[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
	TotalCount *int64 // 仅 IncludeTotalCount=true 时非 nil
}

// 查询构建器 链式Builder
type QueryBuilder struct {
	q PageQuery
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{q: PageQuery{Limit: defaultLimit}}
}

func (b *QueryBuilder) Where(field string, op FilterOp, value any) *QueryBuilder {
	b.q.Filters = append(b.q.Filters, FilterCriteria{Field: field, Op: op, Value: value})
	return b
}

func (b *QueryBuilder) WhereOr(criteria ...FilterCriteria) *QueryBuilder {
	b.q.Filters = append(b.q.Filters, FilterCriteria{
		Logic:    LogicOr,
		Children: criteria,
	})
	return b
}

func (b *QueryBuilder) OrderBy(field string, desc bool) *QueryBuilder {
	b.q.SortBy = append(b.q.SortBy, SortField{Field: field, Descending: desc})
	return b
}

func (b *QueryBuilder) WithTotalCount() *QueryBuilder {
	b.q.IncludeTotalCount = true
	return b
}

func (b *QueryBuilder) Build() PageQuery {
	return b.q
}

// ApplyPaging 将游标/分页参数叠加到已有 PageQuery，同时完成 cursor 解码。
// 解码失败时返回 error，由 Use Case 层向上传递。
func ApplyPaging(base PageQuery, rawCursor string, limit int) (PageQuery, error) {
	q := base
	if limit > 0 {
		q.Limit = limit
	}
	if rawCursor != "" {
		data, err := DecodeCursor(rawCursor)
		if err != nil {
			return PageQuery{}, err
		}
		q.Cursor = &data
	}
	return q, nil
}

// Timestamp helper（Firestore 存储时间常用 time.Time）
func Now() time.Time { return time.Now().UTC() }
