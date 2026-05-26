package reponse

import (
	"errors"
	"go-echo-demo/internal/constants"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// 定义全局响应格式
type ApiResponse struct {
	Code    constants.BizCode `json:"code"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
}

func Success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, ApiResponse{
		Code:    constants.Success,
		Message: "",
		Data:    data,
	})
}

func Fail(c echo.Context, err error) error {
	// 默认的系统级兜底参数
	httpStatus := http.StatusInternalServerError
	bizCode := constants.InternalError
	msg := constants.InternalError.Error() // "系统内部错误！"

	// 关于AsType的用法 需要特别注意 如果自定义的错误是值接收者 理论上来讲AsType支持传入指针和值 但是具体怎么写 需要根据外部传入的类型
	// 如果是指针接收者 那么这里必须传入指针 若违反使用规则 这里的类型判断无法生效
	if c, ok := errors.AsType[constants.BizCode](err); ok {
		zap.L().Error("匹配到业务异常")
		httpStatus = c.HTTPStatus()
		bizCode = c
		msg = c.Error()
	} else {
		// 2. 如果是原生错误（如 DB 报错），记录日志，但不把具体堆栈扔给前端
		zap.L().Error("Uncaught system error", zap.Any("error", err))
	}

	return c.JSON(httpStatus, ApiResponse{
		Code:    bizCode,
		Message: msg,
		Data:    nil,
	})
}
