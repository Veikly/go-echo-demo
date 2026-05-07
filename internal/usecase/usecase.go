package usecase

import (
	"context"
	"go-echo-demo/internal/usecase/usecaseio"
)

// define all use cases

type TaskUseCase interface {
	CreateTask(ctx context.Context, input usecaseio.CreateTaskInput) (usecaseio.CreateTaskOutput, error)
	GetTaskDetail(ctx context.Context, taskId string) (usecaseio.TaskDetailOutput, error)
	ModifyTask(ctx context.Context, input usecaseio.ModifyTaskInput) (usecaseio.TaskDetailOutput, error)
	DeleteTask(ctx context.Context, taskId string) error
}
