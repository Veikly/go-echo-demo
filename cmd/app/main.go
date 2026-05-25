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
	"github.com/labstack/gommon/log"
)

// 项目启动入口 负责初始化组件 然后汇总 装配
func main() {
	ctx := context.Background()

	config, err := bootstrap.LoadConfig()
	if err != nil {
		log.Errorf("loading the bootstrap config error %v", err)
		return
	}

	// init firestore
	fireStoreClient := bootstrap.InitFireStore(ctx, config.ProjectName)
	bootstrap.InitFirebase()

	taskSvc := service.NewTask(fireStoreClient)
	taskUseCase := usecase.NewTask(taskSvc)
	taskHandler := handler.NewTask(taskUseCase)

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(appmiddleware.FirebaseAuthMiddleware)
	e.HTTPErrorHandler = appmiddleware.CustomHTTPErrorHandler

	server := bootstrap.Server{Echo: e, TaskHandler: taskHandler}
	// 为所有Handler绑定路由
	bootstrap.BindRoutes(&server)

	e.Logger.Fatal(e.Start(":8080"))
}
