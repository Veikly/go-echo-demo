package handler

import (
	"go-echo-demo/delivery/http/reponse"
	"go-echo-demo/internal/request"
	"go-echo-demo/internal/response"
	"go-echo-demo/internal/usecase"

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
		return reponse.Fail(c, err)
	}
	rsp := response.UserDetail{
		Username: output.Username,
		Email:    output.Email,
		Age:      output.Age,
		Address:  response.Address(output.Address),
		Profile:  response.Profile(output.Profile),
	}
	return reponse.Success(c, rsp)
}

func (h *UserHandler) CompleteUserInfo(c echo.Context) error {
	var req request.CompleteUserInfoInput
	if err := c.Bind(&req); err != nil {
		return err
	}
	output, err := h.UserUseCase.CompleteUserInfo(c.Request().Context(), req.ToCompleteUserInfoInput())
	if err != nil {
		return reponse.Fail(c, err)
	}
	rsp := response.CompleteUserInfo{
		ID:       output.ID,
		Username: output.Username,
		Email:    output.Email,
		Age:      output.Age,
		Address:  response.Address(output.Address),
		Profile:  response.Profile(output.Profile),
	}
	return reponse.Success(c, rsp)
}
