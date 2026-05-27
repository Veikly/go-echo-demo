package request

import (
	"go-echo-demo/internal/constants/enums"
	"time"
)

type CreateTask struct {
	Title       string           `json:"title"       validate:"required,min=1,max=200"`
	Description string           `json:"description"  validate:"max=200000000"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
}

type ModifyTaskRequest struct {
	Title       string           `json:"title"       validate:"max=200"`
	Description string           `json:"description"  validate:"max=200000000"`
	Status      enums.TaskStatus `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time        `json:"completed_at"`
}
