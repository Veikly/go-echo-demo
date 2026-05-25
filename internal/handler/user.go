package handler

import (
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	UserUseCase usecase.UserUseCase
}

func NewUser(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		UserUseCase: userUseCase,
	}
}

func (h *UserHandler) GetMyDetail(c echo.Context) error {
	userId := c.Param("id")
	output, err := h.UserUseCase.GetMyDetail(c.Request().Context(), userId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, output)
}

func (h *UserHandler) CompleteUserInfo(c echo.Context) error {
	var req request.CompleteUserInfoInput
	if err := c.Bind(&req); err != nil {
		return nil
	}
	output, err := h.UserUseCase.CompleteUserInfo(c.Request().Context(), req.ToCompleteUserInfoInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, output)
}
