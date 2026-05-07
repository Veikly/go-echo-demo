package repository

import (
	"context"
	"go-echo-demo/internal/model"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, data *model.Task) error
	GetTaskDetail(ctx context.Context, taskId string) (*model.Task, error)
	ModifyTask(ctx context.Context, data *model.Task) (*model.Task, error)
	DeleteTask(ctx context.Context, taskId string) error
}
