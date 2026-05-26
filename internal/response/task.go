package response

import (
	"go-echo-demo/internal/constants/enums"
	"time"
)

type SaveTask struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type TaskDetail struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
