package bootstrap

import (
	"go-echo-demo/internal/handler"

	"github.com/labstack/echo/v4"
)

// 定义服务结构体 描述echo实例包含的handler
type Server struct {
	Echo            *echo.Echo
	TaskHandler     *handler.TaskHandler
	UserHandler     *handler.UserHandler
	TaskPageHandler handler.TaskPageHandlerFunc
}

