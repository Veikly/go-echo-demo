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

func (r FirestoreRepository[T]) Query(ctx context.Context, q pagination.PageQuery) (pagination.PageResult[T], error) {
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

// buildEntityFilter — FilterCriteria 树 → firestore.EntityFilter
func buildEntityFilter(filters []pagination.FilterCriteria) firestore.EntityFilter {
	if len(filters) == 1 && len(filters[0].Children) == 0 {
		return firestore.PropertyFilter{
			Path:     filters[0].Field,
			Operator: string(filters[0].Op),
			Value:    filters[0].Value,
		}
	}

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
