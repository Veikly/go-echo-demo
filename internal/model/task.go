package model

import "time"

type Task struct {
	ID          string
	Title       string
	Description string
	Status      int // 0待办 1进行中 2已完成 3已放弃 4已归档
	CompletedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Creator     User
}
