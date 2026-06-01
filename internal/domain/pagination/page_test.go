package pagination

import (
	"testing"
	"time"

	"go-echo-demo/internal/constants"
)

// ---------------------------------------------------------------------------
// ApplyPaging
// ---------------------------------------------------------------------------

func TestApplyPaging_LimitOverride(t *testing.T) {
	base := PageQuery{Limit: 20}
	q, err := ApplyPaging(base, "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Limit != 50 {
		t.Errorf("Limit: got %d, want 50", q.Limit)
	}
}

func TestApplyPaging_LimitZeroKeepsBase(t *testing.T) {
	base := PageQuery{Limit: 20}
	q, err := ApplyPaging(base, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Limit != 20 {
		t.Errorf("Limit: got %d, want 20 (base value)", q.Limit)
	}
}

func TestApplyPaging_LimitNegativeKeepsBase(t *testing.T) {
	base := PageQuery{Limit: 15}
	q, err := ApplyPaging(base, "", -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Limit != 15 {
		t.Errorf("Limit: got %d, want 15 (base value)", q.Limit)
	}
}

func TestApplyPaging_EmptyCursorNilCursor(t *testing.T) {
	base := PageQuery{Limit: 10}
	q, err := ApplyPaging(base, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Cursor != nil {
		t.Errorf("Cursor: got %+v, want nil", q.Cursor)
	}
}

func TestApplyPaging_ValidCursorDecoded(t *testing.T) {
	sortTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cursorStr, err := EncodeCursor(CursorData{
		DocID:     "doc-x",
		SortField: "updated_at",
		SortValue: sortTime,
	})
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}

	base := PageQuery{Limit: 10}
	q, err := ApplyPaging(base, cursorStr, 5)
	if err != nil {
		t.Fatalf("ApplyPaging error: %v", err)
	}
	if q.Cursor == nil {
		t.Fatal("Cursor: got nil, want non-nil")
	}
	if q.Cursor.DocID != "doc-x" {
		t.Errorf("Cursor.DocID: got %q, want doc-x", q.Cursor.DocID)
	}
	if !q.Cursor.SortValue.(time.Time).Equal(sortTime) {
		t.Errorf("Cursor.SortValue: got %v, want %v", q.Cursor.SortValue, sortTime)
	}
}

func TestApplyPaging_InvalidCursorReturnsError(t *testing.T) {
	_, err := ApplyPaging(PageQuery{}, "invalid-cursor!!!", 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != constants.InvalidCursor {
		t.Errorf("expected constants.InvalidCursor, got %v", err)
	}
}

func TestApplyPaging_BaseFiltersPreserved(t *testing.T) {
	base := PageQuery{
		Limit: 10,
		Filters: []FilterCriteria{
			{Field: "status", Op: FilterOpEq, Value: 1},
		},
		SortBy: []SortField{
			{Field: "updated_at", Descending: true},
		},
	}
	q, err := ApplyPaging(base, "", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Filters) != 1 || q.Filters[0].Field != "status" {
		t.Errorf("Filters not preserved: got %+v", q.Filters)
	}
	if len(q.SortBy) != 1 || q.SortBy[0].Field != "updated_at" {
		t.Errorf("SortBy not preserved: got %+v", q.SortBy)
	}
}
