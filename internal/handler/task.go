package handler

import (
	"fmt"
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/response"
	"go-echo-demo/internal/usecase"
	"go-echo-demo/internal/usecase/usecaseio"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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
	err := c.Bind(&req)
	if err != nil {
		return fmt.Errorf("context bind error %w", err)
	}
	input := usecaseio.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		CompletedAt: req.CompletedAt,
	}
	output, err := h.TaskUseCase.CreateTask(c.Request().Context(), input)
	if err != nil {
		return fmt.Errorf("create task failed %v", err)
	}
	rsp := response.SaveTask{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		Status:      output.Status,
		CompletedAt: output.CompletedAt,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}
	return c.JSON(http.StatusCreated, rsp)
}

func (h *TaskHandler) GetTaskDetail(c echo.Context) error {
	taskId := c.Param("id")
	detail, err := h.TaskUseCase.GetTaskDetail(c.Request().Context(), taskId)
	if err != nil {
		log.Errorf("get task detail failed %v", err)
		return err
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *TaskHandler) ModifyTask(c echo.Context) error {
	taskId := c.Param("id")
	var req request.ModifyTaskRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("context bind error %v", err)
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
		log.Errorf("modify task failed %v", err)
		return err
	}
	return c.JSON(http.StatusOK, output)
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	taskId := c.Param("id")
	err := h.TaskUseCase.DeleteTask(c.Request().Context(), taskId)
	if err != nil {
		log.Errorf("delete task failed %v", err)
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
