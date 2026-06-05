package usecase_test

import (
	"context"
	"errors"
	"testing"

	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/usecase"
	"go-echo-demo/internal/usecase/repository"
	"go-echo-demo/internal/usecase/usecaseio"
)

// ---------------------------------------------------------------------------
// fakeTaskRepo：repository.TaskRepository 的测试替身
// ---------------------------------------------------------------------------

type fakeTaskRepo struct {
	// BatchArchieveTask 的控制
	batchArchieveErr    error
	batchArchieveUserID string   // 记录收到的 userID
	batchArchieveIDs    []string // 记录收到的 ids
	batchArchieveTxSeen bool     // 记录调用时 ctx 里是否含有事务标志

	// GetTaskDetail 的控制（其他方法按需扩展）
	getTaskResult *model.Task
	getTaskErr    error
}

func (r *fakeTaskRepo) CreateTask(_ context.Context, _ *model.Task) error { return nil }

func (r *fakeTaskRepo) GetTaskDetail(_ context.Context, _ string) (*model.Task, error) {
	return r.getTaskResult, r.getTaskErr
}

func (r *fakeTaskRepo) ModifyTask(_ context.Context, _ *model.Task) (*model.Task, error) {
	return nil, nil
}

func (r *fakeTaskRepo) DeleteTask(_ context.Context, _ string) error { return nil }

func (r *fakeTaskRepo) BatchArchieveTask(ctx context.Context, ids []string, userID string) error {
	r.batchArchieveIDs = ids
	r.batchArchieveUserID = userID
	// 核心断言：ctx 里必须含有事务标志，否则说明 usecase 没有走事务路径
	_, r.batchArchieveTxSeen = repository.TxFromContext(ctx)
	return r.batchArchieveErr
}

// ---------------------------------------------------------------------------
// fakeTxManager：repository.TransactionManager 的测试替身
//
// 关键设计：调用 RunInTransaction 时，往 ctx 里注入一个 fakeTx 标记，
// 模拟真实 TransactionManager 注入 *FirestoreTx 的行为，
// 让 fakeTaskRepo 能感知到 ctx 里有事务。
// ---------------------------------------------------------------------------

// fakeTx 是一个空结构体，仅作为 TxContent 的占位标记
type fakeTx struct{}

type fakeTxManager struct {
	// 控制事务本身是否返回错误（模拟事务提交失败）
	txErr error
	// 记录 RunInTransaction 是否被调用
	called bool
}

func (m *fakeTxManager) RunInTransaction(ctx context.Context, fn repository.TxFunc) error {
	m.called = true
	if m.txErr != nil {
		return m.txErr
	}
	// 将 fakeTx 注入 ctx，模拟真实事务管理器的行为
	txCtx := repository.ContextWithTx(ctx, &fakeTx{})
	return fn(txCtx)
}

// ---------------------------------------------------------------------------
// 辅助：构造带 session 的 ctx
// ---------------------------------------------------------------------------

func ctxWithSession(uid string) context.Context {
	return domain.WithUserSession(context.Background(), domain.UserSession{
		UID:           uid,
		Email:         "test@example.com",
		EmailVerified: true,
	})
}

// ---------------------------------------------------------------------------
// BatchArchieveTask 测试用例
// ---------------------------------------------------------------------------

// TestBatchArchieve_EmptyIDs 验证空 ids 时在开启事务前就返回 InvalidInputParam
func TestBatchArchieve_EmptyIDs(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{}
	uc := usecase.NewTask(repo, txMgr)

	err := uc.BatchArchieveTask(ctxWithSession("user-1"), []string{})

	if !errors.Is(err, constants.InvalidInputParam) {
		t.Errorf("want InvalidInputParam, got %v", err)
	}
	// 空 ids 应该在事务开启前就拦截，txManager 不应被调用
	if txMgr.called {
		t.Error("txManager should NOT be called when ids is empty")
	}
}

// TestBatchArchieve_NoSession 验证 session 缺失时在开启事务前就返回 CredentialsAbsence
func TestBatchArchieve_NoSession(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{}
	uc := usecase.NewTask(repo, txMgr)

	err := uc.BatchArchieveTask(context.Background(), []string{"id-1", "id-2"})

	if !errors.Is(err, constants.CredentialsAbsence) {
		t.Errorf("want CredentialsAbsence, got %v", err)
	}
	if txMgr.called {
		t.Error("txManager should NOT be called when session is absent")
	}
}

// TestBatchArchieve_TxContextPropagation 验证核心设计：
// RunInTransaction 被调用，且 repo 收到的 ctx 里含有事务标志
func TestBatchArchieve_TxContextPropagation(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{}
	uc := usecase.NewTask(repo, txMgr)

	ids := []string{"task-1", "task-2"}
	err := uc.BatchArchieveTask(ctxWithSession("user-abc"), ids)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !txMgr.called {
		t.Error("txManager.RunInTransaction should be called")
	}
	// repo 收到的 ctx 里必须有事务标志
	if !repo.batchArchieveTxSeen {
		t.Error("repo.BatchArchieveTask should receive a ctx with transaction marker")
	}
}

// TestBatchArchieve_SessionUIDPassedToRepo 验证 session.UID 被正确传递给 repo
func TestBatchArchieve_SessionUIDPassedToRepo(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{}
	uc := usecase.NewTask(repo, txMgr)

	ids := []string{"task-1"}
	err := uc.BatchArchieveTask(ctxWithSession("user-xyz"), ids)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.batchArchieveUserID != "user-xyz" {
		t.Errorf("userID passed to repo: got %q, want user-xyz", repo.batchArchieveUserID)
	}
	if len(repo.batchArchieveIDs) != 1 || repo.batchArchieveIDs[0] != "task-1" {
		t.Errorf("ids passed to repo: got %v, want [task-1]", repo.batchArchieveIDs)
	}
}

// TestBatchArchieve_RepoErrorPropagated 验证 repo 返回的业务错误能正确透传出事务
func TestBatchArchieve_RepoErrorPropagated(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{batchArchieveErr: constants.PermissionDenied}
	uc := usecase.NewTask(repo, txMgr)

	err := uc.BatchArchieveTask(ctxWithSession("user-1"), []string{"task-1"})

	if !errors.Is(err, constants.PermissionDenied) {
		t.Errorf("want PermissionDenied, got %v", err)
	}
}

// TestBatchArchieve_TaskNotArchivableErrorPropagated 验证状态不满足时的业务错误透传
func TestBatchArchieve_TaskNotArchivableErrorPropagated(t *testing.T) {
	txMgr := &fakeTxManager{}
	repo := &fakeTaskRepo{batchArchieveErr: constants.TaskNotArchivable}
	uc := usecase.NewTask(repo, txMgr)

	err := uc.BatchArchieveTask(ctxWithSession("user-1"), []string{"task-1"})

	if !errors.Is(err, constants.TaskNotArchivable) {
		t.Errorf("want TaskNotArchivable, got %v", err)
	}
}

// TestBatchArchieve_TxCommitErrorPropagated 验证事务提交失败（txManager 本身报错）时错误透传
func TestBatchArchieve_TxCommitErrorPropagated(t *testing.T) {
	commitErr := errors.New("firestore: transaction aborted")
	txMgr := &fakeTxManager{txErr: commitErr}
	repo := &fakeTaskRepo{}
	uc := usecase.NewTask(repo, txMgr)

	err := uc.BatchArchieveTask(ctxWithSession("user-1"), []string{"task-1"})

	if !errors.Is(err, commitErr) {
		t.Errorf("want commitErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// 其他方法的基础测试（确保重构没有破坏已有逻辑）
// ---------------------------------------------------------------------------

func TestCreateTask_MissingTitle(t *testing.T) {
	uc := usecase.NewTask(&fakeTaskRepo{}, &fakeTxManager{})

	_, err := uc.CreateTask(ctxWithSession("user-1"), usecaseio.CreateTaskInput{Title: ""})

	if !errors.Is(err, constants.RequireAbsence) {
		t.Errorf("want RequireAbsence, got %v", err)
	}
}

func TestCreateTask_NoSession(t *testing.T) {
	uc := usecase.NewTask(&fakeTaskRepo{}, &fakeTxManager{})

	_, err := uc.CreateTask(context.Background(), usecaseio.CreateTaskInput{Title: "test"})

	if !errors.Is(err, constants.CredentialsAbsence) {
		t.Errorf("want CredentialsAbsence, got %v", err)
	}
}

func TestDeleteTask_EmptyID(t *testing.T) {
	uc := usecase.NewTask(&fakeTaskRepo{}, &fakeTxManager{})

	err := uc.DeleteTask(ctxWithSession("user-1"), "")

	if !errors.Is(err, constants.RequireAbsence) {
		t.Errorf("want RequireAbsence, got %v", err)
	}
}
