package scene

import (
	"fmt"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/constants/enums"
	dmpagination "go-echo-demo/internal/domain/pagination"
	"time"
)

const (
	// TaskByStatusTitle status 等值 + title 等值，按 updated_at 排序
	TaskByStatusTitle = dmpagination.SceneID("task.by_status_title")

	// TaskByCreatedAt created_at 范围 + status 等值，按 created_at 排序
	TaskByCreatedAt = dmpagination.SceneID("task.by_created_at")
)

// RegisterTaskScenes 将 task 相关查询场景注册到 registry。
// 在 main.go 装配阶段调用。
func RegisterTaskScenes(r *dmpagination.Registry) {
	r.Register(TaskByStatusTitle, buildTaskByStatusTitle)
	r.Register(TaskByCreatedAt, buildTaskByCreatedAt)
}

// buildTaskByStatusTitle 场景一：status 等值 + title 等值，按 updated_at 排序
// 参数：
//
//	status  (必填) enums.TaskStatus 对应的整数值，如 "0"、"1"
//	title   (可选) 精确匹配 title 字段
//	desc    (可选) "true" 表示降序，默认降序
func buildTaskByStatusTitle(params dmpagination.SceneParams) (dmpagination.PageQuery, error) {
	b := dmpagination.NewQueryBuilder()

	// status 必填
	statusVal, ok := params["status"]
	if !ok {
		return dmpagination.PageQuery{}, constants.RequireAbsence
	}
	status, err := toTaskStatus(statusVal)
	if err != nil {
		return dmpagination.PageQuery{}, constants.InvalidInputParam
	}
	b.Where("status", dmpagination.FilterOpEq, status)

	// title 可选
	if titleVal, ok := params["title"]; ok && titleVal != "" {
		b.Where("title", dmpagination.FilterOpEq, titleVal)
	}

	// 排序：updated_at 默认降序
	desc := true
	if d, ok := params["desc"]; ok {
		desc = d != "false"
	}
	b.OrderBy("updated_at", desc)

	return b.Build(), nil
}

// buildTaskByCreatedAt 场景二：created_at 范围 + status 等值，按 created_at 排序
// 参数：
//
//	status      (必填) enums.TaskStatus 对应的整数值
//	created_after  (可选) RFC3339 时间字符串，筛选 created_at > 该值
//	created_before (可选) RFC3339 时间字符串，筛选 created_at < 该值
//	desc           (可选) "true" 表示降序，默认降序
func buildTaskByCreatedAt(params dmpagination.SceneParams) (dmpagination.PageQuery, error) {
	b := dmpagination.NewQueryBuilder()

	// status 必填
	statusVal, ok := params["status"]
	if !ok {
		return dmpagination.PageQuery{}, constants.RequireAbsence
	}
	status, err := toTaskStatus(statusVal)
	if err != nil {
		return dmpagination.PageQuery{}, constants.InvalidInputParam
	}
	b.Where("status", dmpagination.FilterOpEq, status)

	// created_after 可选
	if v, ok := params["created_after"]; ok && v != "" {
		t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", v))
		if err != nil {
			return dmpagination.PageQuery{}, constants.InvalidInputParam
		}
		b.Where("created_at", dmpagination.FilterOpGt, t)
	}

	// created_before 可选
	if v, ok := params["created_before"]; ok && v != "" {
		t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", v))
		if err != nil {
			return dmpagination.PageQuery{}, constants.InvalidInputParam
		}
		b.Where("created_at", dmpagination.FilterOpLt, t)
	}

	// 排序：created_at 默认降序
	desc := true
	if d, ok := params["desc"]; ok {
		desc = d != "false"
	}
	b.OrderBy("created_at", desc)

	return b.Build(), nil
}

// toTaskStatus 将 params 中的 status 值转换为 enums.TaskStatus
func toTaskStatus(val any) (enums.TaskStatus, error) {
	switch v := val.(type) {
	case enums.TaskStatus:
		return v, nil
	case int:
		return enums.TaskStatus(v), nil
	case float64:
		return enums.TaskStatus(int(v)), nil
	case string:
		var s int
		if _, err := fmt.Sscanf(v, "%d", &s); err != nil {
			return 0, constants.InvalidInputParam
		}
		return enums.TaskStatus(s), nil
	default:
		return 0, constants.InvalidInputParam
	}
}
