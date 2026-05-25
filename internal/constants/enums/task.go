package enums

type TaskStatus int

const (
	StatusTodo       TaskStatus = 0 // 待办
	StatusInProgress TaskStatus = 1 // 进行中
	StatusDone       TaskStatus = 2 // 已完成
	StatusAbandoned  TaskStatus = 3 // 已废弃
	StatusArchived   TaskStatus = 4 // 已归档
)
