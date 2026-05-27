package usecase

import (
	"context"
	"fmt"
	"go-echo-demo/internal/constants"
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
		return usecaseio.CreateTaskOutput{}, constants.RequireAbsence
	}
	session, ok := domain.FromUserSession(ctx)
	if !ok {
		return usecaseio.CreateTaskOutput{}, constants.CredentialsAbsence
	}
	task := model.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		Creator:     model.User{ID: session.UID},
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
		CreatorID:   task.Creator.ID,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
	return output, nil
}

// getOwnedTask 获取任务并校验所有权，防止越权访问
func (u *Task) getOwnedTask(ctx context.Context, taskId string) (*model.Task, error) {
	session, ok := domain.FromUserSession(ctx)
	if !ok {
		return nil, constants.CredentialsAbsence
	}
	task, err := u.taskSvc.GetTaskDetail(ctx, taskId)
	if err != nil {
		return nil, err
	}
	if task.Creator.ID != session.UID {
		return nil, constants.PermissionDenied
	}
	return task, nil
}

func (u *Task) GetTaskDetail(ctx context.Context, taskId string) (usecaseio.TaskDetailOutput, error) {
	taskDetail, err := u.getOwnedTask(ctx, taskId)
	if err != nil {
		return usecaseio.TaskDetailOutput{}, err
	}
	return usecaseio.TaskDetailOutput{
		ID:          taskDetail.ID,
		Title:       taskDetail.Title,
		Description: taskDetail.Description,
		Status:      taskDetail.Status,
		CreatorID:   taskDetail.Creator.ID,
		CompletedAt: taskDetail.CompletedAt,
		CreatedAt:   taskDetail.CreatedAt,
		UpdatedAt:   taskDetail.UpdatedAt,
	}, nil
}

func (u *Task) ModifyTask(ctx context.Context, input usecaseio.ModifyTaskInput) (usecaseio.TaskDetailOutput, error) {
	if input.ID == "" {
		return usecaseio.TaskDetailOutput{}, constants.RequireAbsence
	}
	if _, err := u.getOwnedTask(ctx, input.ID); err != nil {
		return usecaseio.TaskDetailOutput{}, err
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
		CreatorID:   result.Creator.ID,
		CompletedAt: result.CompletedAt,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}, nil
}

func (u *Task) DeleteTask(ctx context.Context, taskId string) error {
	if taskId == "" {
		return constants.RequireAbsence
	}
	if _, err := u.getOwnedTask(ctx, taskId); err != nil {
		return err
	}
	if err := u.taskSvc.DeleteTask(ctx, taskId); err != nil {
		return fmt.Errorf("delete task error %w", err)
	}
	return nil
}
