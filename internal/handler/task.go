package handler

import (
	httppagination "go-echo-demo/delivery/http/pagination"
	"go-echo-demo/delivery/http/reponse"
	"go-echo-demo/internal/constants"
	dmpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/response"
	ucpagination "go-echo-demo/internal/usecase/pagination"
	"go-echo-demo/internal/usecase/usecaseio"

	"github.com/labstack/echo/v4"
)

type TaskHandler struct {
	TaskUseCase TaskUseCase
	listHandler echo.HandlerFunc // 可选，由 WithListHandler 挂载
}

func NewTask(taskUseCase TaskUseCase) *TaskHandler {
	return &TaskHandler{
		TaskUseCase: taskUseCase,
	}
}

func NewTaskListHandler(
	uc *ucpagination.QueryUseCase[model.Task, response.TaskItem],
	registry *dmpagination.Registry,
) echo.HandlerFunc {
	return PaginatedHandler[model.Task, response.TaskItem](
		uc,
		registry,
		bindTaskPageQuery,
	)
}

// WithListHandler 挂载分页查询能力，返回自身支持链式调用。
func (h *TaskHandler) WithListHandler(lh echo.HandlerFunc) *TaskHandler {
	h.listHandler = lh
	return h
}

// ListTasks 分页查询入口，委托给注入的 listHandler。
func (h *TaskHandler) ListTasks(c echo.Context) error {
	if h.listHandler == nil {
		return echo.NewHTTPError(404, "list not supported")
	}
	return h.listHandler(c)
}

func (h *TaskHandler) CreateTask(c echo.Context) error {
	var req request.CreateTask
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return reponse.Fail(c, err)
	}
	input := usecaseio.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		CompletedAt: req.CompletedAt,
	}
	output, err := h.TaskUseCase.CreateTask(c.Request().Context(), input)
	if err != nil {
		return reponse.Fail(c, err)
	}
	rsp := response.SaveTask{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		Status:      output.Status,
		CreatorID:   output.CreatorID,
		CompletedAt: output.CompletedAt,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}
	return reponse.Success(c, rsp)
}

func (h *TaskHandler) GetTaskDetail(c echo.Context) error {
	taskId := c.Param("id")
	detail, err := h.TaskUseCase.GetTaskDetail(c.Request().Context(), taskId)
	if err != nil {
		return reponse.Fail(c, err)
	}
	rsp := response.TaskDetail{
		ID:          detail.ID,
		Title:       detail.Title,
		Description: detail.Description,
		Status:      detail.Status,
		CreatorID:   detail.CreatorID,
		CompletedAt: detail.CompletedAt,
		CreatedAt:   detail.CreatedAt,
		UpdatedAt:   detail.UpdatedAt,
	}
	return reponse.Success(c, rsp)
}

func (h *TaskHandler) ModifyTask(c echo.Context) error {
	taskId := c.Param("id")
	var req request.ModifyTaskRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return reponse.Fail(c, err)
	}
	input := usecaseio.ModifyTaskInput{
		ID:          taskId,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		CompletedAt: req.CompletedAt,
	}
	output, err := h.TaskUseCase.ModifyTask(c.Request().Context(), input)
	if err != nil {
		return reponse.Fail(c, err)
	}
	rsp := response.TaskDetail{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		Status:      output.Status,
		CreatorID:   output.CreatorID,
		CompletedAt: output.CompletedAt,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}
	return reponse.Success(c, rsp)
}

func (h *TaskHandler) BatchArchieveTask(c echo.Context) error {
	var req request.BatchArchieveTask
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return reponse.Fail(c, err)
	}
	if err := h.TaskUseCase.BatchArchieveTask(c.Request().Context(), req.IDs); err != nil {
		return reponse.Fail(c, err)
	}
	return reponse.Success(c, nil)
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	taskId := c.Param("id")
	if err := h.TaskUseCase.DeleteTask(c.Request().Context(), taskId); err != nil {
		return reponse.Fail(c, err)
	}
	return reponse.Success(c, nil)
}

// bindTaskPageQuery 绑定并解析 task 分页请求参数。
func bindTaskPageQuery(c echo.Context) (httppagination.BasePageQuery, dmpagination.SceneParams, error) {
	var dto request.TaskPageQuery
	if err := c.Bind(&dto); err != nil {
		return httppagination.BasePageQuery{}, nil, constants.InvalidInputParam
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
