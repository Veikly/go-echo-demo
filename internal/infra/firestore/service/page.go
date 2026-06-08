package firestoreservice

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain/pagination"

	"cloud.google.com/go/firestore"
	pb "cloud.google.com/go/firestore/apiv1/firestorepb"
)

// Mapper 将 Firestore DocumentSnapshot 映射到领域模型
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

func (r *FirestoreRepository[T]) Query(ctx context.Context, q pagination.PageQuery) (pagination.PageResult[T], error) {
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	fsQuery, err := r.buildQuery(q)
	if err != nil {
		return pagination.PageResult[T]{}, err
	}

	docs, err := fsQuery.Limit(q.Limit + 1).Documents(ctx).GetAll()
	if err != nil {
		return pagination.PageResult[T]{}, err
	}

	hasMore := len(docs) > q.Limit
	if hasMore {
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

	result := pagination.PageResult[T]{Items: items, HasMore: hasMore}

	if len(docs) > 0 {
		lastDoc := docs[len(docs)-1]
		sortField := ""
		if len(q.SortBy) > 0 {
			sortField = q.SortBy[0].Field
		}
		var sortValue any
		if sortField != "" {
			sortValue, _ = lastDoc.DataAt(sortField)
		}
		result.NextCursor, _ = pagination.EncodeCursor(pagination.CursorData{
			DocID:     lastDoc.Ref.ID,
			SortField: sortField,
			SortValue: sortValue,
		})
	}

	if q.IncludeTotalCount {
		count, err := r.aggregateCount(ctx, q)
		if err == nil {
			result.TotalCount = &count
		}
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
		base = base.WhereEntity(buildEntityFilter(q.Filters))
	}

	for _, s := range q.SortBy {
		dir := firestore.Asc // 默认情况下 是升序
		if s.Descending {
			dir = firestore.Desc
		}
		base = base.OrderBy(s.Field, dir)
	}

	// 追加 __name__ 保证游标唯一性，方向与第一个排序字段一致
	idDir := firestore.Asc
	if len(q.SortBy) > 0 && q.SortBy[0].Descending {
		idDir = firestore.Desc
	}
	base = base.OrderBy(firestore.DocumentID, idDir)

	// 应用游标
	if q.Cursor != nil {
		cursorData := q.Cursor
		if cursorData.SortValue != nil && cursorData.DocID != "" {
			base = base.StartAfter(cursorData.SortValue, cursorData.DocID)
		}
	}

	return base, nil
}

// buildEntityFilter 将 []FilterCriteria 列表转换为 firestore.EntityFilter。
//
// 约定：平级列表中的多个条目之间永远是 AND 关系。
// 若需要 OR，必须通过 QueryBuilder.WhereOr() 显式构造一个
// Logic=OR 的组合节点放入列表，由 buildSingleFilter 递归处理。
//
// 这样无论列表中混入多少组合节点，AND/OR 语义都由节点自身决定，
// 不再依赖 filters[0].Logic 的值，消除了原有的歧义。
func buildEntityFilter(filters []pagination.FilterCriteria) firestore.EntityFilter {
	if len(filters) == 1 {
		return buildSingleFilter(filters[0])
	}
	// 平级多条目 → AND
	children := make([]firestore.EntityFilter, len(filters))
	for i, f := range filters {
		children[i] = buildSingleFilter(f)
	}
	return firestore.AndFilter{Filters: children}
}

// buildSingleFilter 将单个 FilterCriteria 节点转换为 firestore.EntityFilter。
//
// 叶节点（Children 为空）：直接映射为 PropertyFilter。
// 组合节点（Children 非空）：递归处理子节点，Logic 字段决定使用 AndFilter 还是 OrFilter。
func buildSingleFilter(f pagination.FilterCriteria) firestore.EntityFilter {
	if len(f.Children) == 0 {
		// 叶节点
		return firestore.PropertyFilter{
			Path:     f.Field,
			Operator: string(f.Op),
			Value:    f.Value,
		}
	}
	// 组合节点：递归
	children := make([]firestore.EntityFilter, len(f.Children))
	for i, child := range f.Children {
		children[i] = buildSingleFilter(child)
	}
	if f.Logic == pagination.LogicOr {
		return firestore.OrFilter{Filters: children}
	}
	return firestore.AndFilter{Filters: children}
}

// aggregateCount — AggregationQuery 总数
func (r *FirestoreRepository[T]) aggregateCount(ctx context.Context, q pagination.PageQuery) (int64, error) {
	base := r.client.Collection(r.collection).Query
	if len(q.Filters) > 0 {
		base = base.WhereEntity(buildEntityFilter(q.Filters))
	}

	result, err := base.NewAggregationQuery().WithCount("count").Get(ctx)
	if err != nil {
		return 0, err
	}

	countValue, ok := result["count"]
	if !ok {
		return 0, constants.InternalError
	}

	pbVal, ok := countValue.(*pb.Value)
	if !ok {
		return 0, constants.InternalError
	}
	return pbVal.GetIntegerValue(), nil
}
