package scene

import (
	"testing"
	"time"

	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/constants/enums"
	dmpagination "go-echo-demo/internal/domain/pagination"
)

// ---------------------------------------------------------------------------
// 辅助：从 filters 中找到指定 field 的叶节点
// ---------------------------------------------------------------------------

func findFilter(filters []dmpagination.FilterCriteria, field string) (dmpagination.FilterCriteria, bool) {
	for _, f := range filters {
		if len(f.Children) > 0 {
			if found, ok := findFilter(f.Children, field); ok {
				return found, true
			}
		}
		if f.Field == field {
			return f, true
		}
	}
	return dmpagination.FilterCriteria{}, false
}

func findSortField(sortBy []dmpagination.SortField, field string) (dmpagination.SortField, bool) {
	for _, s := range sortBy {
		if s.Field == field {
			return s, true
		}
	}
	return dmpagination.SortField{}, false
}

// ===========================================================================
// buildTaskByStatusTitle
// ===========================================================================

func TestTaskByStatusTitle_StatusAsEnumType(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": enums.StatusTodo,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "status")
	if !ok {
		t.Fatal("status filter not found")
	}
	if f.Value != enums.StatusTodo {
		t.Errorf("status value: got %v, want %v", f.Value, enums.StatusTodo)
	}
}

func TestTaskByStatusTitle_StatusAsInt(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": int(1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "status")
	if !ok {
		t.Fatal("status filter not found")
	}
	if f.Value != enums.StatusInProgress {
		t.Errorf("status value: got %v, want %v", f.Value, enums.StatusInProgress)
	}
}

func TestTaskByStatusTitle_StatusAsFloat64(t *testing.T) {
	// JSON 解析默认产生 float64，这是实际请求中最常见的情况
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "status")
	if !ok {
		t.Fatal("status filter not found")
	}
	if f.Value != enums.StatusDone {
		t.Errorf("status value: got %v, want %v", f.Value, enums.StatusDone)
	}
}

func TestTaskByStatusTitle_StatusAsString(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": "1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "status")
	if !ok {
		t.Fatal("status filter not found")
	}
	if f.Value != enums.StatusInProgress {
		t.Errorf("status value: got %v, want %v", f.Value, enums.StatusInProgress)
	}
}

func TestTaskByStatusTitle_StatusInvalidString(t *testing.T) {
	_, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": "abc",
	})
	if err != constants.InvalidInputParam {
		t.Errorf("expected constants.InvalidInputParam, got %v", err)
	}
}

func TestTaskByStatusTitle_StatusMissing(t *testing.T) {
	_, err := buildTaskByStatusTitle(dmpagination.SceneParams{})
	if err != constants.RequireAbsence {
		t.Errorf("expected constants.RequireAbsence, got %v", err)
	}
}

func TestTaskByStatusTitle_TitlePresent(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": enums.StatusTodo,
		"title":  "my task",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "title")
	if !ok {
		t.Fatal("title filter not found")
	}
	if f.Value != "my task" {
		t.Errorf("title value: got %v, want my task", f.Value)
	}
}

func TestTaskByStatusTitle_TitleEmpty_NoFilter(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": enums.StatusTodo,
		"title":  "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := findFilter(q.Filters, "title"); ok {
		t.Error("title filter should not be added when title is empty string")
	}
}

func TestTaskByStatusTitle_DefaultDescending(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": enums.StatusTodo,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := findSortField(q.SortBy, "updated_at")
	if !ok {
		t.Fatal("updated_at sort field not found")
	}
	if !s.Descending {
		t.Error("default sort should be descending")
	}
}

func TestTaskByStatusTitle_DescFalse_Ascending(t *testing.T) {
	q, err := buildTaskByStatusTitle(dmpagination.SceneParams{
		"status": enums.StatusTodo,
		"desc":   "false",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := findSortField(q.SortBy, "updated_at")
	if !ok {
		t.Fatal("updated_at sort field not found")
	}
	if s.Descending {
		t.Error("desc=false should produce ascending sort")
	}
}

// ===========================================================================
// buildTaskByCreatedAt
// ===========================================================================

func TestTaskByCreatedAt_StatusMissing(t *testing.T) {
	_, err := buildTaskByCreatedAt(dmpagination.SceneParams{})
	if err != constants.RequireAbsence {
		t.Errorf("expected constants.RequireAbsence, got %v", err)
	}
}

func TestTaskByCreatedAt_CreatedAfterValid(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status":        enums.StatusTodo,
		"created_after": "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "created_at")
	if !ok {
		t.Fatal("created_at filter not found")
	}
	if f.Op != dmpagination.FilterOpGt {
		t.Errorf("created_after op: got %v, want >", f.Op)
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.Value.(time.Time).Equal(want) {
		t.Errorf("created_after value: got %v, want %v", f.Value, want)
	}
}

func TestTaskByCreatedAt_CreatedBeforeValid(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status":         enums.StatusTodo,
		"created_before": "2026-06-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := findFilter(q.Filters, "created_at")
	if !ok {
		t.Fatal("created_at filter not found")
	}
	if f.Op != dmpagination.FilterOpLt {
		t.Errorf("created_before op: got %v, want <", f.Op)
	}
}

func TestTaskByCreatedAt_BothTimeParams(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status":         enums.StatusTodo,
		"created_after":  "2026-01-01T00:00:00Z",
		"created_before": "2026-06-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该有两条 created_at 过滤条件
	count := 0
	for _, f := range q.Filters {
		if f.Field == "created_at" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 created_at filters, got %d", count)
	}
}

func TestTaskByCreatedAt_NoTimeParams_OnlyStatus(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status": enums.StatusTodo,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := findFilter(q.Filters, "created_at"); ok {
		t.Error("created_at filter should not exist when no time params provided")
	}
}

func TestTaskByCreatedAt_CreatedAfterInvalidFormat(t *testing.T) {
	_, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status":        enums.StatusTodo,
		"created_after": "not-a-date",
	})
	if err != constants.InvalidInputParam {
		t.Errorf("expected constants.InvalidInputParam, got %v", err)
	}
}

func TestTaskByCreatedAt_CreatedBeforeInvalidFormat(t *testing.T) {
	_, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status":         enums.StatusTodo,
		"created_before": "2026/06/01", // 不是 RFC3339
	})
	if err != constants.InvalidInputParam {
		t.Errorf("expected constants.InvalidInputParam, got %v", err)
	}
}

func TestTaskByCreatedAt_DefaultDescending(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status": enums.StatusTodo,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := findSortField(q.SortBy, "created_at")
	if !ok {
		t.Fatal("created_at sort field not found")
	}
	if !s.Descending {
		t.Error("default sort should be descending")
	}
}

func TestTaskByCreatedAt_DescFalse_Ascending(t *testing.T) {
	q, err := buildTaskByCreatedAt(dmpagination.SceneParams{
		"status": enums.StatusTodo,
		"desc":   "false",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := findSortField(q.SortBy, "created_at")
	if !ok {
		t.Fatal("created_at sort field not found")
	}
	if s.Descending {
		t.Error("desc=false should produce ascending sort")
	}
}

// ===========================================================================
// RegisterTaskScenes
// ===========================================================================

func TestRegisterTaskScenes_AllScenesRegistered(t *testing.T) {
	reg := dmpagination.NewRegistry()
	RegisterTaskScenes(reg)

	known := reg.KnownScenes()
	set := make(map[dmpagination.SceneID]bool)
	for _, id := range known {
		set[id] = true
	}
	if !set[TaskByStatusTitle] {
		t.Errorf("TaskByStatusTitle not registered")
	}
	if !set[TaskByCreatedAt] {
		t.Errorf("TaskByCreatedAt not registered")
	}
}
