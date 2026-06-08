package adapters

import (
	"context"
	domainpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/usecase/usecaseio"
)

// define all use cases

type TaskUseCase interface {
	CreateTask(ctx context.Context, input usecaseio.CreateTaskInput) (usecaseio.CreateTaskOutput, error)
	GetTaskDetail(ctx context.Context, taskId string) (usecaseio.TaskDetailOutput, error)
	ModifyTask(ctx context.Context, input usecaseio.ModifyTaskInput) (usecaseio.TaskDetailOutput, error)
	DeleteTask(ctx context.Context, taskId string) error
	// BatchArchieveTask 批量归档已完成的任务，并原子更新用户归档统计
	// 事务由 usecase 内部控制，handler 层无需感知
	BatchArchieveTask(ctx context.Context, ids []string) error
}

type UserUseCase interface {
	GetMyDetail(ctx context.Context, userId string) (usecaseio.UserDetailOutput, error)
	CompleteUserInfo(ctx context.Context, input usecaseio.CompleteUserInfoDetail) (usecaseio.CompleteUserInfoDetail, error)
}

type PageUseCase[DTO any] interface {
	Execute(ctx context.Context, input usecaseio.ExecuteInput) (domainpagination.PageResult[DTO], error)
}
