package dto

import (
	"go-echo-demo/internal/constants/enums"
	"go-echo-demo/internal/model"
	"time"
)

type Task struct {
	ID          string           `firestore:"-"`
	Title       string           `firestore:"title"`
	Description string           `firestore:"description"`
	Status      enums.TaskStatus `firestore:"status"` //0待办 1进行中 2已完成 3已放弃 4已归档
	CreatorID   string           `firestore:"creator_id"`
	CompletedAt time.Time        `firestore:"completed_at"`
	CreatedAt   time.Time        `firestore:"created_at"`
	UpdatedAt   time.Time        `firestore:"updated_at"`
}

func (t *Task) ToEntity() *model.Task {
	return &model.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Creator:     model.User{ID: t.CreatorID},
		CompletedAt: t.CompletedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (t *Task) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"title":        t.Title,
		"description":  t.Description,
		"status":       t.Status,
		"creator_id":   t.CreatorID,
		"completed_at": t.CompletedAt,
		"created_at":   t.CreatedAt,
		"updated_at":   t.UpdatedAt,
	}
}

func NewTask(t *model.Task) *Task {
	return &Task{
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatorID:   t.Creator.ID,
		CompletedAt: t.CompletedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
