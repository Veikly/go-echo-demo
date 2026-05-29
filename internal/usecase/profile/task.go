package profile

import "go-echo-demo/internal/domain"

const TaskListProfile = "task.list"

// profile package被import时 函数自动执行
func init() {
	domain.RegisterProfile(&domain.QueryProfile{
		Key:         TaskListProfile,
		Resource:    "tasks",
		DefaultSort: "created_at_desc",
		MaxLimit:    50,
		Sorts: map[string]domain.SortSpec{
			"created_at_desc": {
				Key:   "created_at_desc",
				Field: "created_at",
				Dir:   domain.SortDesc,
				Type:  domain.CursorTime,
			},
			"created_at_asc": {
				Key:   "created_at_asc",
				Field: "created_at",
				Dir:   domain.SortAsc,
				Type:  domain.CursorTime,
			},
		},
		Filters: map[string]domain.FilterSpec{
			"status": {
				Name:  "status",
				Field: "status",
				Kind:  domain.FilterEqual,
			},
		},
	})
}
