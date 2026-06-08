package firestoreservice

import (
	"testing"

	"go-echo-demo/internal/domain/pagination"

	"cloud.google.com/go/firestore"
)

// ---------------------------------------------------------------------------
// 辅助断言
// ---------------------------------------------------------------------------

// assertPropertyFilter 断言 ef 是 PropertyFilter 并校验其字段。
func assertPropertyFilter(t *testing.T, ef firestore.EntityFilter, path, operator string, value any) {
	t.Helper()
	pf, ok := ef.(firestore.PropertyFilter)
	if !ok {
		t.Fatalf("expected PropertyFilter, got %T", ef)
	}
	if pf.Path != path {
		t.Errorf("PropertyFilter.Path: got %q, want %q", pf.Path, path)
	}
	if pf.Operator != operator {
		t.Errorf("PropertyFilter.Operator: got %q, want %q", pf.Operator, operator)
	}
	if pf.Value != value {
		t.Errorf("PropertyFilter.Value: got %v, want %v", pf.Value, value)
	}
}

// assertAndFilter 断言 ef 是 AndFilter 并返回其子节点列表。
func assertAndFilter(t *testing.T, ef firestore.EntityFilter, wantLen int) []firestore.EntityFilter {
	t.Helper()
	af, ok := ef.(firestore.AndFilter)
	if !ok {
		t.Fatalf("expected AndFilter, got %T", ef)
	}
	if len(af.Filters) != wantLen {
		t.Fatalf("AndFilter.Filters: got %d children, want %d", len(af.Filters), wantLen)
	}
	return af.Filters
}

// assertOrFilter 断言 ef 是 OrFilter 并返回其子节点列表。
func assertOrFilter(t *testing.T, ef firestore.EntityFilter, wantLen int) []firestore.EntityFilter {
	t.Helper()
	of, ok := ef.(firestore.OrFilter)
	if !ok {
		t.Fatalf("expected OrFilter, got %T", ef)
	}
	if len(of.Filters) != wantLen {
		t.Fatalf("OrFilter.Filters: got %d children, want %d", len(of.Filters), wantLen)
	}
	return of.Filters
}

// ===========================================================================
// buildSingleFilter
// ===========================================================================

// TestBuildSingleFilter_Leaf 叶节点应映射为 PropertyFilter。
func TestBuildSingleFilter_Leaf(t *testing.T) {
	f := pagination.FilterCriteria{
		Field: "status",
		Op:    pagination.FilterOpEq,
		Value: 1,
	}
	ef := buildSingleFilter(f)
	assertPropertyFilter(t, ef, "status", "==", 1)
}

// TestBuildSingleFilter_AndGroup 组合节点 Logic=AND 应映射为 AndFilter。
func TestBuildSingleFilter_AndGroup(t *testing.T) {
	f := pagination.FilterCriteria{
		Logic: pagination.LogicAnd,
		Children: []pagination.FilterCriteria{
			{Field: "a", Op: pagination.FilterOpEq, Value: 1},
			{Field: "b", Op: pagination.FilterOpEq, Value: 2},
		},
	}
	ef := buildSingleFilter(f)
	children := assertAndFilter(t, ef, 2)
	assertPropertyFilter(t, children[0], "a", "==", 1)
	assertPropertyFilter(t, children[1], "b", "==", 2)
}

// TestBuildSingleFilter_OrGroup 组合节点 Logic=OR 应映射为 OrFilter。
func TestBuildSingleFilter_OrGroup(t *testing.T) {
	f := pagination.FilterCriteria{
		Logic: pagination.LogicOr,
		Children: []pagination.FilterCriteria{
			{Field: "x", Op: pagination.FilterOpEq, Value: "foo"},
			{Field: "y", Op: pagination.FilterOpGt, Value: 100},
		},
	}
	ef := buildSingleFilter(f)
	children := assertOrFilter(t, ef, 2)
	assertPropertyFilter(t, children[0], "x", "==", "foo")
	assertPropertyFilter(t, children[1], "y", ">", 100)
}

// TestBuildSingleFilter_NestedAndInOr AND 嵌套在 OR 内：OR( AND(a,b), c )
func TestBuildSingleFilter_NestedAndInOr(t *testing.T) {
	f := pagination.FilterCriteria{
		Logic: pagination.LogicOr,
		Children: []pagination.FilterCriteria{
			{
				Logic: pagination.LogicAnd,
				Children: []pagination.FilterCriteria{
					{Field: "a", Op: pagination.FilterOpEq, Value: 1},
					{Field: "b", Op: pagination.FilterOpEq, Value: 2},
				},
			},
			{Field: "c", Op: pagination.FilterOpEq, Value: 3},
		},
	}
	ef := buildSingleFilter(f)
	orChildren := assertOrFilter(t, ef, 2)

	// 第一个子节点应是 AndFilter(a, b)
	andChildren := assertAndFilter(t, orChildren[0], 2)
	assertPropertyFilter(t, andChildren[0], "a", "==", 1)
	assertPropertyFilter(t, andChildren[1], "b", "==", 2)

	// 第二个子节点应是 PropertyFilter(c)
	assertPropertyFilter(t, orChildren[1], "c", "==", 3)
}

// ===========================================================================
// buildEntityFilter
// ===========================================================================

// TestBuildEntityFilter_SingleLeaf 只有一个叶节点时直接返回 PropertyFilter（不额外包 AndFilter）。
func TestBuildEntityFilter_SingleLeaf(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{Field: "status", Op: pagination.FilterOpEq, Value: 1},
	}
	ef := buildEntityFilter(filters)
	assertPropertyFilter(t, ef, "status", "==", 1)
}

// TestBuildEntityFilter_MultipleLeaves_AlwaysAnd 多个叶节点平级时永远是 AND。
func TestBuildEntityFilter_MultipleLeaves_AlwaysAnd(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{Field: "status", Op: pagination.FilterOpEq, Value: 1},
		{Field: "title", Op: pagination.FilterOpEq, Value: "test"},
	}
	ef := buildEntityFilter(filters)
	children := assertAndFilter(t, ef, 2)
	assertPropertyFilter(t, children[0], "status", "==", 1)
	assertPropertyFilter(t, children[1], "title", "==", "test")
}

// TestBuildEntityFilter_LeafAndOrGroup 平级列表中混入 OR 组合节点：
// [status==1, OR(created_at>t, title=="x")] → AND(status==1, OR(...))
// 这是修复 bug 的核心用例：原代码会因 filters[0].Logic=="" 而把整个列表包成 AND，
// 但无法正确处理 OR 子树；新代码通过 buildSingleFilter 递归，语义完全正确。
func TestBuildEntityFilter_LeafAndOrGroup(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{Field: "status", Op: pagination.FilterOpEq, Value: 1},
		{
			Logic: pagination.LogicOr,
			Children: []pagination.FilterCriteria{
				{Field: "created_at", Op: pagination.FilterOpGt, Value: "2026-01-01"},
				{Field: "title", Op: pagination.FilterOpEq, Value: "x"},
			},
		},
	}
	ef := buildEntityFilter(filters)

	// 顶层是 AND
	andChildren := assertAndFilter(t, ef, 2)

	// 第一个子节点：status==1 叶节点
	assertPropertyFilter(t, andChildren[0], "status", "==", 1)

	// 第二个子节点：OR(created_at>..., title==...)
	orChildren := assertOrFilter(t, andChildren[1], 2)
	assertPropertyFilter(t, orChildren[0], "created_at", ">", "2026-01-01")
	assertPropertyFilter(t, orChildren[1], "title", "==", "x")
}

// TestBuildEntityFilter_OrGroupFirst OR 组合节点排在列表第一位时，
// 整体仍然是 AND（平级列表之间永远 AND），不受节点顺序影响。
// 这直接验证了原 bug 的修复：原代码取 filters[0].Logic 来决定顶层逻辑，
// 若第一个节点恰好是 OR 组合节点，整个列表就会被错误地包成 OrFilter。
func TestBuildEntityFilter_OrGroupFirst(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{
			Logic: pagination.LogicOr,
			Children: []pagination.FilterCriteria{
				{Field: "a", Op: pagination.FilterOpEq, Value: 1},
				{Field: "b", Op: pagination.FilterOpEq, Value: 2},
			},
		},
		{Field: "creator_id", Op: pagination.FilterOpEq, Value: "uid-123"},
	}
	ef := buildEntityFilter(filters)

	// 顶层必须是 AND，不能是 OR
	andChildren := assertAndFilter(t, ef, 2)

	// 第一个子节点是 OR 子树
	assertOrFilter(t, andChildren[0], 2)

	// 第二个子节点是 creator_id 叶节点
	assertPropertyFilter(t, andChildren[1], "creator_id", "==", "uid-123")
}

// TestBuildEntityFilter_SingleOrGroup 列表只有一个 OR 组合节点时，
// 直接返回 OrFilter（不额外包 AndFilter）。
func TestBuildEntityFilter_SingleOrGroup(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{
			Logic: pagination.LogicOr,
			Children: []pagination.FilterCriteria{
				{Field: "x", Op: pagination.FilterOpEq, Value: "foo"},
				{Field: "y", Op: pagination.FilterOpEq, Value: "bar"},
			},
		},
	}
	ef := buildEntityFilter(filters)
	// 只有一个节点，走 buildSingleFilter，顶层是 OrFilter
	orChildren := assertOrFilter(t, ef, 2)
	assertPropertyFilter(t, orChildren[0], "x", "==", "foo")
	assertPropertyFilter(t, orChildren[1], "y", "==", "bar")
}

// TestBuildEntityFilter_ThreeLeaves 三个平级叶节点，全部 AND。
func TestBuildEntityFilter_ThreeLeaves(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{Field: "status", Op: pagination.FilterOpEq, Value: 1},
		{Field: "creator_id", Op: pagination.FilterOpEq, Value: "uid"},
		{Field: "title", Op: pagination.FilterOpEq, Value: "hello"},
	}
	ef := buildEntityFilter(filters)
	children := assertAndFilter(t, ef, 3)
	assertPropertyFilter(t, children[0], "status", "==", 1)
	assertPropertyFilter(t, children[1], "creator_id", "==", "uid")
	assertPropertyFilter(t, children[2], "title", "==", "hello")
}

// TestBuildEntityFilter_DeepNesting 深度嵌套：AND( OR(a, AND(b, c)), d )
func TestBuildEntityFilter_DeepNesting(t *testing.T) {
	filters := []pagination.FilterCriteria{
		{
			Logic: pagination.LogicOr,
			Children: []pagination.FilterCriteria{
				{Field: "a", Op: pagination.FilterOpEq, Value: 1},
				{
					Logic: pagination.LogicAnd,
					Children: []pagination.FilterCriteria{
						{Field: "b", Op: pagination.FilterOpEq, Value: 2},
						{Field: "c", Op: pagination.FilterOpEq, Value: 3},
					},
				},
			},
		},
		{Field: "d", Op: pagination.FilterOpEq, Value: 4},
	}
	ef := buildEntityFilter(filters)

	// 顶层 AND
	topAndChildren := assertAndFilter(t, ef, 2)

	// 第一个子节点：OR(a, AND(b,c))
	orChildren := assertOrFilter(t, topAndChildren[0], 2)
	assertPropertyFilter(t, orChildren[0], "a", "==", 1)
	innerAndChildren := assertAndFilter(t, orChildren[1], 2)
	assertPropertyFilter(t, innerAndChildren[0], "b", "==", 2)
	assertPropertyFilter(t, innerAndChildren[1], "c", "==", 3)

	// 第二个子节点：d 叶节点
	assertPropertyFilter(t, topAndChildren[1], "d", "==", 4)
}
