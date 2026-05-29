package pagination

import (
	"context"
	"fmt"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain/pagination"

	"cloud.google.com/go/firestore"
)

// 将Firestore Snapshot映射到 领域模型
type Mapper[T any] func(*firestore.DocumentSnapshot) (T, error)

// FirestoreRepository[T] 通用 Firestore 分页仓储。
type FirestoreRepository[T any] struct {
	client     *firestore.Client
	collection string
	mapper     Mapper[T]
}

func NewFirestoreRepository[T any](
	client *firestore.Client,
	collection string,
	mapper Mapper[T],
) *FirestoreRepository[T] {
	return &FirestoreRepository[T]{
		client:     client,
		collection: collection,
		mapper:     mapper,
	}
}

func (r FirestoreRepository[T]) Query(ctx context.Context, q pagination.PageQuery) (pagination.PageResult[T], error) {
	// 参数规范化（防御性处理，Use Case 层应已处理，此处兜底）
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if q.Direction == "" {
		q.Direction = pagination.CursorForward
	}

	fsQuery, err := r.buildQuery(q)
	if err != nil {
		return pagination.PageResult[T]{}, err
	}

	docs, err := fsQuery.Limit(q.Limit + 1).Documents(ctx).GetAll()
	if err != nil {
		// todo 这里需要对错误转转换 不要返回原始gRPC错误码
		return pagination.PageResult[T]{}, err
	}

	// 反向翻页结果集需翻转，保证返回顺序始终与正向一致
	if q.Direction == pagination.CursorBackward {
		for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
			docs[i], docs[j] = docs[j], docs[i]
		}
	}

	hasMore := len(docs) > q.Limit
	if hasMore {
		// 如果有下一页 返回原始结果
		docs = docs[:q.Limit]
	}

	items := make([]T, 0, len(docs))
	for _, doc := range docs {
		item, err := r.mapper(doc)
		if err != nil {
			return pagination.PageResult[T]{}, constants.DocMapError
		}
		items = append(items, item)
	}
	// 构造返回结构
	result := pagination.PageResult[T]{Items: items, HasMore: hasMore}

	if len(docs) > 0 {
		lastDoc := docs[len(docs)-1]
		firstDoc := docs[0]

		// 取排序字段的值用于 cursor（取第一个 SortBy 字段）
		sortField := ""
		if len(q.SortBy) > 0 {
			sortField = q.SortBy[0].Field
		}
		getSortValue := func(doc *firestore.DocumentSnapshot) any {
			if sortField == "" {
				return nil
			}
			v, _ := doc.DataAt(sortField)
			return v
		}

		// 正向：最后一条文档作为下一页游标
		result.NextCursor, _ = pagination.EncodeCursor(pagination.CursorData{
			DocID:     lastDoc.Ref.ID,
			SortField: sortField,
			SortValue: getSortValue(lastDoc),
		})
		// 有入参游标说明不是第一页，生成上一页游标
		if q.Cursor != nil && q.Direction != pagination.CursorRefresh {
			result.PrevCursor, _ = pagination.EncodeCursor(pagination.CursorData{
				DocID:     firstDoc.Ref.ID,
				SortField: sortField,
				SortValue: getSortValue(firstDoc),
			})
		}
	}

	if q.IncludeTotalCount {
		count, err := r.aggregateCount(ctx, q)
		if err == nil {
			result.TotalCount = &count
		}
		// 总数查询失败不影响主查询，记录日志即可（此处省略 logger 依赖）
	}

	return result, nil
}

// Get 单文档查询。
func (r *FirestoreRepository[T]) Get(ctx context.Context, id string) (T, error) {
	var zero T
	snap, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return zero, err
	}
	return r.mapper(snap)
}

func (r *FirestoreRepository[T]) buildQuery(q pagination.PageQuery) (firestore.Query, error) {
	base := r.client.Collection(r.collection).Query

	if len(q.Filters) > 0 {
		entityFilter := buildEntityFilter(q.Filters)
		base = base.WhereEntity(entityFilter)
	}

	reverse := q.Direction == pagination.CursorBackward

	for _, s := range q.SortBy {
		desc := s.Descending
		if reverse {
			desc = !desc
		}
		dir := firestore.Asc
		if desc {
			dir = firestore.Desc
		}
		base = base.OrderBy(s.Field, dir)
	}
	// 追加 __name__ 保证游标唯一性（与 sort 字段相同方向）
	idDir := firestore.Asc
	if reverse && len(q.SortBy) > 0 && q.SortBy[0].Descending {
		idDir = firestore.Desc
	}
	base = base.OrderBy(firestore.DocumentID, idDir)

	// 应用游标
	if q.Cursor != nil && q.Direction != pagination.CursorRefresh {
		cursorData := q.Cursor
		if cursorData.SortValue != nil && cursorData.DocID != "" {
			switch q.Direction {
			case pagination.CursorForward:
				base = base.StartAfter(cursorData.SortValue, cursorData.DocID)
			case pagination.CursorBackward:
				base = base.EndBefore(cursorData.SortValue, cursorData.DocID)
			}
		}
	}

	return base, nil
}

// ─────────────────────────────────────────────────────────────
// buildEntityFilter — FilterCriteria 树 → firestore.EntityFilter
// 支持任意深度的 AND / OR 嵌套
// ─────────────────────────────────────────────────────────────
func buildEntityFilter(filters []pagination.FilterCriteria) firestore.EntityFilter {
	// 如果是单个的叶子节点 直接返回PropertyFilter
	if len(filters) == 1 && len(filters[0].Children) == 0 {
		return firestore.PropertyFilter{
			Path:     filters[0].Field,
			Operator: string(filters[0].Op),
			Value:    filters[0].Value,
		}
	}

	// 递归构造子节点
	children := make([]firestore.EntityFilter, 0, len(filters))
	for _, f := range filters {
		if len(f.Children) > 0 {
			children = append(children, buildEntityFilter(f.Children))
		} else {
			children = append(children, firestore.PropertyFilter{
				Path:     f.Field,
				Operator: string(f.Op),
				Value:    f.Value,
			})
		}
	}

	if filters[0].Logic == pagination.LogicOr {
		return firestore.OrFilter{Filters: children}
	}
	return firestore.AndFilter{Filters: children}
}

// ─────────────────────────────────────────────────────────────
// aggregateCount — AggregationQuery 总数（不下载文档）
// ─────────────────────────────────────────────────────────────
func (r *FirestoreRepository[T]) aggregateCount(ctx context.Context, q pagination.PageQuery) (int64, error) {
	base := r.client.Collection(r.collection).Query

	if len(q.Filters) > 0 {
		base = base.WhereEntity(buildEntityFilter(q.Filters))
	}

	result, err := base.NewAggregationQuery().
		WithCount("count").
		Get(ctx)
	if err != nil {
		return 0, err
	}

	v, ok := result["count"]
	if !ok {
		return 0, fmt.Errorf("aggregation result missing count")
	}

	n, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected count value type: %T", v)
	}

	return n, nil
}
