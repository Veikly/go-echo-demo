package handler

import (
	"go-echo-demo/delivery/http/reponse"
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/response"
	"go-echo-demo/internal/usecase"
	"go-echo-demo/internal/usecase/usecaseio"

	"github.com/labstack/echo/v4"
)

type TaskHandler struct {
	TaskUseCase usecase.TaskUseCase
}

func NewTask(taskUseCase usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{
		TaskUseCase: taskUseCase,
	}
}

func (h *TaskHandler) CreateTask(c echo.Context) error {
	var req request.CreateTask
	if err := c.Bind(&req); err != nil {
		return err
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

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	taskId := c.Param("id")
	if err := h.TaskUseCase.DeleteTask(c.Request().Context(), taskId); err != nil {
		return reponse.Fail(c, err)
	}
	return reponse.Success(c, nil)
}
