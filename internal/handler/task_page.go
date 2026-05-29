package handler

import (
	httppagination "go-echo-demo/delivery/http/pagination"
	dmpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/response"
	ucpagination "go-echo-demo/internal/usecase/pagination"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TaskPageHandlerFunc 是 task 分页查询的 Echo handler 类型别名。
type TaskPageHandlerFunc = echo.HandlerFunc

// NewTaskPageHandler 构造 task 分页查询 handler。
func NewTaskPageHandler(
	uc *ucpagination.QueryUseCase[model.Task, response.TaskItem],
	registry *dmpagination.Registry,
) echo.HandlerFunc {
	return PaginatedHandler[model.Task, response.TaskItem](
		uc,
		registry,
		bindTaskPageQuery,
	)
}

// bindTaskPageQuery 绑定并解析 task 分页请求参数，返回通用基础参数和业务 SceneParams。
func bindTaskPageQuery(c echo.Context) (httppagination.BasePageQuery, dmpagination.SceneParams, error) {
	var dto request.TaskPageQuery
	if err := c.Bind(&dto); err != nil {
		return httppagination.BasePageQuery{}, nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	params := dmpagination.SceneParams{}

	if dto.Status != nil {
		params["status"] = *dto.Status
	}
	if dto.Title != "" {
		params["title"] = dto.Title
	}
	if dto.CreatedAfter != "" {
		params["created_after"] = dto.CreatedAfter
	}
	if dto.CreatedBefore != "" {
		params["created_before"] = dto.CreatedBefore
	}
	if dto.Desc != nil {
		if *dto.Desc {
			params["desc"] = "true"
		} else {
			params["desc"] = "false"
		}
	}

	return dto.BasePageQuery, params, nil
}
