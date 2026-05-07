package usecaseio

import "time"

// usecase不能感知到handler层(http)的存在 因此定义usecase层专用的数据模型 由外部传入 usecase只专注自身的业务逻辑
type CreateTaskInput struct {
	Title       string
	Description string
	Status      int // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time
}

type CreateTaskOutput struct {
	ID          string
	Title       string
	Description string
	Status      int // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskDetailOutput struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      int       `json:"status"` // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ModifyTaskInput struct {
	ID          string
	Title       string
	Description string
	Status      int // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time
}
