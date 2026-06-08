package firestoreservice

import (
	"context"
	"fmt"
	"go-echo-demo/internal/adapters"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/constants/enums"
	"go-echo-demo/internal/infra/firestore/dto"
	"go-echo-demo/internal/infra/firestore/transaction"
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
			return nil, constants.TaskNotFound
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
	updates := map[string]interface{}{
		"title":       data.Title,
		"description": data.Description,
		"status":      data.Status,
		"updated_at":  time.Now(),
	}
	if !data.CompletedAt.IsZero() {
		updates["completed_at"] = data.CompletedAt
	}
	_, err := s.client.Collection("tasks").Doc(data.ID).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		return nil, fmt.Errorf("modify task error %v", err)
	}
	docRef, err := s.client.Collection("tasks").Doc(data.ID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var result dto.Task
	if err := docRef.DataTo(&result); err != nil {
		return nil, err
	}
	return result.ToEntity(), nil
}

func (s *Task) DeleteTask(ctx context.Context, taskId string) error {
	_, err := s.client.Collection("tasks").Doc(taskId).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete task error %v", err)
	}
	return nil
}

func (s *Task) BatchArchieveTask(ctx context.Context, ids []string, userID string) error {
	if tx, ok := adapters.TxFromContext(ctx); ok {
		ftx := tx.(*transaction.FirestoreTx)
		now := time.Now()
		// Phase 1: 在事务内批量读取所有目标任务
		refs := make([]*firestore.DocumentRef, len(ids))
		for i, id := range ids {
			refs[i] = s.client.Collection("tasks").Doc(id)
		}
		snaps, err := ftx.Tx.GetAll(refs)
		if err != nil {
			return constants.InternalError
		}

		// 读取用户归档统计文档（用于计数器的 read-modify-write）
		statsRef := s.client.Collection("user_stats").Doc(userID)
		statsSnap, err := ftx.Tx.Get(statsRef)
		// NotFound 是合法状态（首次归档），其他错误才需要返回
		if err != nil && status.Code(err) != codes.NotFound {
			return constants.InternalError
		}

		//  Phase 2: 业务校验（基于事务内读到的快照，不做额外 IO
		for _, snap := range snaps {
			if !snap.Exists() {
				return constants.TaskNotFound
			}
			var task dto.Task
			if err := snap.DataTo(&task); err != nil {
				return constants.DocMapError
			}
			// 权限校验：只能归档自己的任务
			if task.CreatorID != userID {
				return constants.PermissionDenied
			}
			// 业务规则：只有「已完成」状态的任务才能归档
			if task.Status != enums.StatusDone {
				return constants.TaskNotArchivable
			}
		}

		//  Phase 3: 在事务内执行所有写操作
		for _, snap := range snaps {
			if err := ftx.Tx.Update(snap.Ref, []firestore.Update{
				{Path: "status", Value: enums.StatusArchived},
				{Path: "updated_at", Value: now},
			}); err != nil {
				return constants.InternalError
			}
		}

		// 计数器 read-modify-write：基于事务读到的值做递增，保证并发安全
		// statsSnap 在 NotFound 时 Exists()==false，直接用 currentCount=0 即可
		currentCount := 0
		if statsSnap.Exists() {
			var stats dto.UserStats
			if err := statsSnap.DataTo(&stats); err != nil {
				return constants.DocMapError
			}
			currentCount = stats.ArchivedCount
		}
		if err := ftx.Tx.Set(statsRef, dto.UserStats{
			UserID:        userID,
			ArchivedCount: currentCount + len(ids),
			UpdatedAt:     now,
		}); err != nil {
			return constants.InternalError
		}
	} else {
		// 这个方法默认需要事务控制 如果没有收到事务信号 直接抛异常
		return constants.RequireTransaction
	}
	return nil
}
