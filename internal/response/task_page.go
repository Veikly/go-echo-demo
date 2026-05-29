package response

import (
	"go-echo-demo/internal/constants/enums"
	"time"
)

// TaskItem 分页列表中单条 task 的响应结构
type TaskItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      enums.TaskStatus `json:"status"`
	CreatorID   string           `json:"creator_id"`
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
