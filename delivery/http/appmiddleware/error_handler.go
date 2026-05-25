package appmiddleware

import (
	"errors"
	"go-echo-demo/delivery/http/apperror"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	appError := apperror.MapError(err)
	var echoError *echo.HTTPError
	if errors.As(err, &echoError) {
		// 如果是echo框架自己抛出的错误 单独处理并返回
		appError.Status = echoError.Code

		if msg, ok := echoError.Message.(string); ok {
			appError.Code = msg
			appError.Message = msg
		}
	}

	if appError.Status >= http.StatusInternalServerError {
		zap.L().Error("request failed", zap.Error(err))
	}

	_ = c.JSON(appError.Status, ErrorResponse{
		Code:    appError.Code,
		Message: appError.Message,
	})
}
