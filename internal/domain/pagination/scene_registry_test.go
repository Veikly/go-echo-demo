package pagination

import (
	"testing"

	"go-echo-demo/internal/constants"
)

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

func TestRegistry_BuildRegisteredScene(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test.scene", func(params SceneParams) (PageQuery, error) {
		return NewQueryBuilder().Where("field", FilterOpEq, "value").Build(), nil
	})

	q, err := reg.Build("test.scene", SceneParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Filters) != 1 || q.Filters[0].Field != "field" {
		t.Errorf("unexpected query filters: %+v", q.Filters)
	}
}

func TestRegistry_BuildUnknownScene(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Build("not.exist", SceneParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != constants.UnknownScene {
		t.Errorf("expected constants.UnknownScene, got %v", err)
	}
}

func TestRegistry_KnownScenes(t *testing.T) {
	reg := NewRegistry()
	reg.Register("scene.a", func(SceneParams) (PageQuery, error) { return PageQuery{}, nil })
	reg.Register("scene.b", func(SceneParams) (PageQuery, error) { return PageQuery{}, nil })

	known := reg.KnownScenes()
	if len(known) != 2 {
		t.Fatalf("KnownScenes: got %d, want 2", len(known))
	}

	set := make(map[SceneID]bool)
	for _, id := range known {
		set[id] = true
	}
	if !set["scene.a"] || !set["scene.b"] {
		t.Errorf("KnownScenes missing expected IDs, got %v", known)
	}
}

func TestRegistry_RegisterOverwrite(t *testing.T) {
	// 注册同一 ID 两次，后者应覆盖前者（map 语义，当前设计预期行为）
	reg := NewRegistry()
	reg.Register("scene.x", func(SceneParams) (PageQuery, error) {
		return NewQueryBuilder().Where("v", FilterOpEq, "first").Build(), nil
	})
	reg.Register("scene.x", func(SceneParams) (PageQuery, error) {
		return NewQueryBuilder().Where("v", FilterOpEq, "second").Build(), nil
	})

	q, err := reg.Build("scene.x", SceneParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Filters[0].Value != "second" {
		t.Errorf("expected second registration to win, got value=%v", q.Filters[0].Value)
	}
}

func TestRegistry_BuildPassesParamsToBuilder(t *testing.T) {
	reg := NewRegistry()
	reg.Register("scene.params", func(params SceneParams) (PageQuery, error) {
		val := params["key"]
		return NewQueryBuilder().Where("key", FilterOpEq, val).Build(), nil
	})

	q, err := reg.Build("scene.params", SceneParams{"key": "my-value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Filters[0].Value != "my-value" {
		t.Errorf("params not passed to builder: got %v", q.Filters[0].Value)
	}
}
