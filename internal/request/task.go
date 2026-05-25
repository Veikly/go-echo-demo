package request

import (
	"go-echo-demo/internal/constants/enums"
	"time"
)

type CreateTask struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
}

type ModifyTaskRequest struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
}
