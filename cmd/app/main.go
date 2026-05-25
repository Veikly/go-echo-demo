package main

import (
	"context"
	"go-echo-demo/delivery/http/appmiddleware"
	"go-echo-demo/internal/bootstrap"
	"go-echo-demo/internal/handler"
	"go-echo-demo/internal/infra/firestore/service"
	"go-echo-demo/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// 项目启动入口 负责初始化组件 然后汇总 装配
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	ctx := context.Background()

	config, err := bootstrap.LoadConfig()
	if err != nil {
		zap.L().Error("loading the bootstrap config error", zap.Error(err))
		return
	}

	// init firestore
	fireStoreClient := bootstrap.InitFireStore(ctx, config.ProjectName)
	bootstrap.InitFirebase()

	taskSvc := service.NewTask(fireStoreClient)
	taskUseCase := usecase.NewTask(taskSvc)
	taskHandler := handler.NewTask(taskUseCase)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(appmiddleware.FirebaseAuthMiddleware)
	e.Use(appmiddleware.ZapLogger)
	e.HTTPErrorHandler = appmiddleware.CustomHTTPErrorHandler

	server := bootstrap.Server{Echo: e, TaskHandler: taskHandler}
	// 为所有Handler绑定路由
	bootstrap.BindRoutes(&server)

	zap.L().Fatal("server start error", zap.Error(e.Start(":8080")))
}
