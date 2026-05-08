package service

import (
	"context"
	"fmt"
	"go-echo-demo/internal/domain"
	"go-echo-demo/internal/infra/firestore/dto"
	"go-echo-demo/internal/model"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Task TaskService 要求实现TaskRepository接口
type Task struct {
	client *firestore.Client
}

func NewTask(client *firestore.Client) *Task {
	return &Task{client: client}
}

func (s *Task) CreateTask(ctx context.Context, data *model.Task) error {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	td := dto.NewTask(data)
	docRef, _, err := s.client.Collection("tasks").Add(ctx, td)
	if err != nil {
		return err
	}
	data.ID = docRef.ID
	return nil
}

func (s *Task) GetTaskDetail(ctx context.Context, taskId string) (*model.Task, error) {
	//docRef, err := s.client.Collection("tasks").Doc(taskId).Get(ctx)
	docRef := s.client.Collection("tasks").Doc(taskId)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var rs dto.Task
	if err := docSnap.DataTo(&rs); err != nil {
		return nil, err
	}
	rs.ID = docSnap.Ref.ID
	return rs.ToEntity(), nil
}

func (s *Task) ModifyTask(ctx context.Context, data *model.Task) (*model.Task, error) {
	data.UpdatedAt = time.Now()
	td := dto.NewTask(data)
	_, err := s.client.Collection("tasks").Doc(data.ID).Set(ctx, td.ToMap(), firestore.MergeAll)
	if err != nil {
		return nil, fmt.Errorf("modify task error %v", err)
	}
	docRef, err := s.client.Collection("tasks").Doc(data.ID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var result model.Task
	if err := docRef.DataTo(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Task) DeleteTask(ctx context.Context, taskId string) error {
	_, err := s.client.Collection("tasks").Doc(taskId).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete task error %v", err)
	}
	return nil
}
