package pagination_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/constants/enums"
	dmpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/infra/firestore/page/scene"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/response"
	ucpagination "go-echo-demo/internal/usecase/pagination"
	"go-echo-demo/internal/usecase/usecaseio"
)

// ---------------------------------------------------------------------------
// fakeTaskRepo：pagination.Repository[model.Task] 的测试替身
// ---------------------------------------------------------------------------

type fakeTaskRepo struct {
	result    dmpagination.PageResult[model.Task]
	err       error
	lastQuery dmpagination.PageQuery // 记录最近一次收到的 query，供断言使用
}

func (r *fakeTaskRepo) Query(_ context.Context, q dmpagination.PageQuery) (dmpagination.PageResult[model.Task], error) {
	r.lastQuery = q
	return r.result, r.err
}

func (r *fakeTaskRepo) Get(_ context.Context, _ string) (model.Task, error) {
	return model.Task{}, nil
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

// toDTO 将 model.Task 转换为 response.TaskItem，测试中复用。
func toDTO(t model.Task) response.TaskItem {
	return response.TaskItem{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// newUC 构造一个带真实 Registry 的 QueryUseCase，injectRules 可为 nil。
func newUC(
	repo *fakeTaskRepo,
	injectRules func(context.Context, dmpagination.PageQuery) (dmpagination.PageQuery, error),
) *ucpagination.QueryUseCase[model.Task, response.TaskItem] {
	reg := dmpagination.NewRegistry()
	scene.RegisterTaskScenes(reg)

	return ucpagination.NewQueryUseCase(ucpagination.QueryUseCaseConfig[model.Task, response.TaskItem]{
		Repo:        repo,
		Registry:    reg,
		ToDTO:       toDTO,
		InjectRules: injectRules,
	})
}

// ---------------------------------------------------------------------------
// 测试用例
// ---------------------------------------------------------------------------

// TestExecute_FirstPage 验证首页查询（无 cursor）的完整流程：
// registry 构建查询 → repo 被调用 → DTO 转换正确。
func TestExecute_FirstPage(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeTaskRepo{
		result: dmpagination.PageResult[model.Task]{
			Items: []model.Task{
				{ID: "t1", Title: "task one", Status: enums.StatusTodo, UpdatedAt: now},
				{ID: "t2", Title: "task two", Status: enums.StatusTodo, UpdatedAt: now.Add(-time.Minute)},
			},
			HasMore:    true,
			NextCursor: "some-cursor-string",
		},
	}

	uc := newUC(repo, nil)

	result, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("items count: got %d, want 2", len(result.Items))
	}
	if result.Items[0].ID != "t1" {
		t.Errorf("items[0].ID: got %q, want t1", result.Items[0].ID)
	}
	if result.Items[1].ID != "t2" {
		t.Errorf("items[1].ID: got %q, want t2", result.Items[1].ID)
	}
	if !result.HasMore {
		t.Error("HasMore: got false, want true")
	}
	if result.NextCursor != "some-cursor-string" {
		t.Errorf("NextCursor: got %q, want some-cursor-string", result.NextCursor)
	}
	// 首页时 repo 收到的 query 不应携带 cursor
	if repo.lastQuery.Cursor != nil {
		t.Error("first page query should have nil cursor")
	}
	// limit 应被正确传递
	if repo.lastQuery.Limit != 10 {
		t.Errorf("query.Limit: got %d, want 10", repo.lastQuery.Limit)
	}
}

// TestExecute_WithCursor 验证带 cursor 的翻页：
// cursor 被正确解码并传入 repo.Query。
func TestExecute_WithCursor(t *testing.T) {
	sortTime := time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC)
	cursorStr, err := dmpagination.EncodeCursor(dmpagination.CursorData{
		DocID:     "doc99",
		SortField: "updated_at",
		SortValue: sortTime,
	})
	if err != nil {
		t.Fatalf("failed to encode cursor: %v", err)
	}

	repo := &fakeTaskRepo{
		result: dmpagination.PageResult[model.Task]{Items: []model.Task{}},
	}
	uc := newUC(repo, nil)

	_, err = uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
		Cursor: cursorStr,
		Limit:  5,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastQuery.Cursor == nil {
		t.Fatal("expected cursor to be set in query, got nil")
	}
	if repo.lastQuery.Cursor.DocID != "doc99" {
		t.Errorf("cursor.DocID: got %q, want doc99", repo.lastQuery.Cursor.DocID)
	}
	if repo.lastQuery.Cursor.SortField != "updated_at" {
		t.Errorf("cursor.SortField: got %q, want updated_at", repo.lastQuery.Cursor.SortField)
	}
	decoded := repo.lastQuery.Cursor.SortValue.(time.Time)
	if !decoded.Equal(sortTime) {
		t.Errorf("cursor.SortValue: got %v, want %v", decoded, sortTime)
	}
	if repo.lastQuery.Limit != 5 {
		t.Errorf("query.Limit: got %d, want 5", repo.lastQuery.Limit)
	}
}

// TestExecute_InvalidCursor 验证非法 cursor 时返回 constants.InvalidCursor。
func TestExecute_InvalidCursor(t *testing.T) {
	repo := &fakeTaskRepo{}
	uc := newUC(repo, nil)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
		Cursor: "this-is-not-a-valid-cursor!!",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, constants.InvalidCursor) {
		t.Errorf("expected constants.InvalidCursor, got %v", err)
	}
}

// TestExecute_UnknownScene 验证未注册的 scene 返回 constants.UnknownScene。
func TestExecute_UnknownScene(t *testing.T) {
	repo := &fakeTaskRepo{}
	uc := newUC(repo, nil)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  dmpagination.SceneID("not.exist"),
		Params: dmpagination.SceneParams{},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, constants.UnknownScene) {
		t.Errorf("expected constants.UnknownScene, got %v", err)
	}
}

// TestExecute_MissingRequiredParam 验证 scene 缺少必填参数时返回 constants.RequireAbsence。
// TaskByStatusTitle 要求 status 必填。
func TestExecute_MissingRequiredParam(t *testing.T) {
	repo := &fakeTaskRepo{}
	uc := newUC(repo, nil)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{}, // 故意不传 status
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, constants.RequireAbsence) {
		t.Errorf("expected constants.RequireAbsence, got %v", err)
	}
}

// TestExecute_RepoError 验证 repo.Query 返回错误时，Execute 将错误向上透传。
func TestExecute_RepoError(t *testing.T) {
	repoErr := errors.New("firestore: deadline exceeded")
	repo := &fakeTaskRepo{err: repoErr}
	uc := newUC(repo, nil)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
	})

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repoErr, got %v", err)
	}
}

// TestExecute_InjectRules 验证 injectRules 注入的过滤条件出现在 repo 收到的 query 里。
func TestExecute_InjectRules(t *testing.T) {
	repo := &fakeTaskRepo{
		result: dmpagination.PageResult[model.Task]{Items: []model.Task{}},
	}

	injectRules := func(_ context.Context, q dmpagination.PageQuery) (dmpagination.PageQuery, error) {
		q.Filters = append(q.Filters, dmpagination.FilterCriteria{
			Field: "creator_id",
			Op:    dmpagination.FilterOpEq,
			Value: "user-abc",
		})
		return q, nil
	}

	uc := newUC(repo, injectRules)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range repo.lastQuery.Filters {
		if f.Field == "creator_id" && f.Value == "user-abc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("injected filter creator_id=user-abc not found in query.Filters")
	}
}

// TestExecute_InjectRulesError 验证 injectRules 返回错误时，Execute 将错误向上透传。
func TestExecute_InjectRulesError(t *testing.T) {
	repo := &fakeTaskRepo{}
	injectErr := errors.New("permission denied by inject rules")

	injectRules := func(_ context.Context, q dmpagination.PageQuery) (dmpagination.PageQuery, error) {
		return dmpagination.PageQuery{}, injectErr
	}

	uc := newUC(repo, injectRules)

	_, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
	})

	if err == nil {
		t.Fatal("expected error from injectRules, got nil")
	}
	if !errors.Is(err, injectErr) {
		t.Errorf("expected injectErr, got %v", err)
	}
}

// TestExecute_WithTotalCount 验证 WithTotalCount=true 时，IncludeTotalCount 被正确传入 repo。
func TestExecute_WithTotalCount(t *testing.T) {
	total := int64(42)
	repo := &fakeTaskRepo{
		result: dmpagination.PageResult[model.Task]{
			Items:      []model.Task{},
			TotalCount: &total,
		},
	}
	uc := newUC(repo, nil)

	result, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:          scene.TaskByStatusTitle,
		Params:         dmpagination.SceneParams{"status": enums.StatusTodo},
		WithTotalCount: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.lastQuery.IncludeTotalCount {
		t.Error("query.IncludeTotalCount: got false, want true")
	}
	if result.TotalCount == nil {
		t.Fatal("result.TotalCount: got nil, want non-nil")
	}
	if *result.TotalCount != 42 {
		t.Errorf("result.TotalCount: got %d, want 42", *result.TotalCount)
	}
}

// TestExecute_EmptyResult 验证 repo 返回空列表时，结果正常（不 panic，Items 为空切片）。
func TestExecute_EmptyResult(t *testing.T) {
	repo := &fakeTaskRepo{
		result: dmpagination.PageResult[model.Task]{
			Items:   []model.Task{},
			HasMore: false,
		},
	}
	uc := newUC(repo, nil)

	result, err := uc.Execute(context.Background(), usecaseio.ExecuteInput{
		Scene:  scene.TaskByStatusTitle,
		Params: dmpagination.SceneParams{"status": enums.StatusTodo},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("items count: got %d, want 0", len(result.Items))
	}
	if result.HasMore {
		t.Error("HasMore: got true, want false")
	}
}
