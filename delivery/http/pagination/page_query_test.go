package pagination_test

import (
	"testing"

	httppagination "go-echo-demo/delivery/http/pagination"
	"go-echo-demo/internal/constants"
	dmpagination "go-echo-demo/internal/domain/pagination"
)

// ---------------------------------------------------------------------------
// 辅助：构造一个包含已知 scene 的 Registry
// ---------------------------------------------------------------------------

func newTestRegistry(scenes ...string) *dmpagination.Registry {
	reg := dmpagination.NewRegistry()
	for _, s := range scenes {
		id := dmpagination.SceneID(s)
		reg.Register(id, func(dmpagination.SceneParams) (dmpagination.PageQuery, error) {
			return dmpagination.PageQuery{}, nil
		})
	}
	return reg
}

// 合法 cursor，用于需要传入 cursor 的用例
func validCursor(t *testing.T) string {
	t.Helper()
	c, err := dmpagination.EncodeCursor(dmpagination.CursorData{
		DocID:     "doc1",
		SortField: "updated_at",
		SortValue: "some-value",
	})
	if err != nil {
		t.Fatalf("failed to encode cursor: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// ValidateBaseParams
// ---------------------------------------------------------------------------

func TestValidateBaseParams_SceneEmpty(t *testing.T) {
	reg := newTestRegistry("task.by_status_title")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "",
		Limit: 10,
	}, reg)
	if err != constants.RequireAbsence {
		t.Errorf("expected constants.RequireAbsence, got %v", err)
	}
}

func TestValidateBaseParams_SceneUnknown(t *testing.T) {
	reg := newTestRegistry("task.by_status_title")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "not.exist",
		Limit: 10,
	}, reg)
	if err != constants.UnknownScene {
		t.Errorf("expected constants.UnknownScene, got %v", err)
	}
}

func TestValidateBaseParams_SceneValid(t *testing.T) {
	reg := newTestRegistry("task.by_status_title")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "task.by_status_title",
		Limit: 10,
	}, reg)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateBaseParams_LimitNegative(t *testing.T) {
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "task.scene",
		Limit: -1,
	}, reg)
	if err != constants.InvalidInputParam {
		t.Errorf("expected constants.InvalidInputParam, got %v", err)
	}
}

func TestValidateBaseParams_LimitExceedsMax(t *testing.T) {
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "task.scene",
		Limit: 101,
	}, reg)
	if err != constants.InvalidInputParam {
		t.Errorf("expected constants.InvalidInputParam, got %v", err)
	}
}

func TestValidateBaseParams_LimitZero_Valid(t *testing.T) {
	// limit=0 表示使用默认值，应该通过校验
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "task.scene",
		Limit: 0,
	}, reg)
	if err != nil {
		t.Errorf("expected no error for limit=0, got %v", err)
	}
}

func TestValidateBaseParams_LimitMax_Valid(t *testing.T) {
	// limit=100 是上限本身，应该合法
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene: "task.scene",
		Limit: 100,
	}, reg)
	if err != nil {
		t.Errorf("expected no error for limit=100, got %v", err)
	}
}

func TestValidateBaseParams_CursorEmpty_Valid(t *testing.T) {
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene:  "task.scene",
		Limit:  10,
		Cursor: "",
	}, reg)
	if err != nil {
		t.Errorf("expected no error for empty cursor, got %v", err)
	}
}

func TestValidateBaseParams_CursorValid(t *testing.T) {
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene:  "task.scene",
		Limit:  10,
		Cursor: validCursor(t),
	}, reg)
	if err != nil {
		t.Errorf("expected no error for valid cursor, got %v", err)
	}
}

func TestValidateBaseParams_CursorInvalid(t *testing.T) {
	reg := newTestRegistry("task.scene")
	err := httppagination.ValidateBaseParams(httppagination.BasePageQuery{
		Scene:  "task.scene",
		Limit:  10,
		Cursor: "this-is-not-a-valid-cursor!!!",
	}, reg)
	if err != constants.InvalidCursor {
		t.Errorf("expected constants.InvalidCursor, got %v", err)
	}
}
