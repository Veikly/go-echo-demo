package usecase

import (
	"context"
	"fmt"
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/usecase/repository"
	"go-echo-demo/internal/usecase/usecaseio"
)

// Task 要求实现TaskUseCase接口
type Task struct {
	taskSvc repository.TaskRepository
}

func NewTask(s repository.TaskRepository) *Task {
	return &Task{
		taskSvc: s,
	}
}

func (u *Task) CreateTask(ctx context.Context, input usecaseio.CreateTaskInput) (usecaseio.CreateTaskOutput, error) {
	if input.Title == "" {
		return usecaseio.CreateTaskOutput{}, domain.ErrInvalidInput
	}
	task := model.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		CompletedAt: input.CompletedAt,
	}
	err := u.taskSvc.CreateTask(ctx, &task)
	if err != nil {
		return usecaseio.CreateTaskOutput{}, fmt.Errorf("create task failed %w", err)
	}
	output := usecaseio.CreateTaskOutput{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
	return output, nil
}

func (u *Task) GetTaskDetail(ctx context.Context, taskId string) (usecaseio.TaskDetailOutput, error) {
	taskDetail, err := u.taskSvc.GetTaskDetail(ctx, taskId)
	if err != nil {
		return usecaseio.TaskDetailOutput{}, err
	}
	return usecaseio.TaskDetailOutput{
		ID:          taskDetail.ID,
		Title:       taskDetail.Title,
		Description: taskDetail.Description,
		Status:      taskDetail.Status,
		CompletedAt: taskDetail.CompletedAt,
		CreatedAt:   taskDetail.CreatedAt,
		UpdatedAt:   taskDetail.UpdatedAt,
	}, nil
}

func (u *Task) ModifyTask(ctx context.Context, input usecaseio.ModifyTaskInput) (usecaseio.TaskDetailOutput, error) {
	if input.ID == "" {
		return usecaseio.TaskDetailOutput{}, domain.ErrInvalidInput
	}
	task := model.Task{
		ID:          input.ID,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		CompletedAt: input.CompletedAt,
	}
	result, err := u.taskSvc.ModifyTask(ctx, &task)
	if err != nil {
		return usecaseio.TaskDetailOutput{}, fmt.Errorf("modify task error %w", err)
	}
	return usecaseio.TaskDetailOutput{
		ID:          result.ID,
		Title:       result.Title,
		Description: result.Description,
		Status:      result.Status,
		CompletedAt: result.CompletedAt,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}, nil
}

func (u *Task) DeleteTask(ctx context.Context, taskId string) error {
	if taskId == "" {
		return domain.ErrInvalidInput
	}
	err := u.taskSvc.DeleteTask(ctx, taskId)
	if err != nil {
		return fmt.Errorf("delete task error %w", err)
	}
	return nil
}
