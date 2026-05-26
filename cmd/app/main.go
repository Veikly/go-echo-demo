package main

import (
	"context"
	"errors"
	"go-echo-demo/delivery/http/appmiddleware"
	"go-echo-demo/internal/bootstrap"
	"go-echo-demo/internal/handler"
	"go-echo-demo/internal/infra/authenticator"
	"go-echo-demo/internal/infra/firestore/service"
	"go-echo-demo/internal/usecase"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// 项目启动入口 负责初始化组件 然后汇总 装配
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	ctx := context.Background()

	config, err := bootstrap.LoadConfig()
	if err != nil {
		zap.L().Error("loading the bootstrap config error", zap.Error(err))
		return
	}

	// init firestore
	fireStoreClient, err := bootstrap.InitFireStore(ctx, config.ProjectName)
	if err != nil {
		zap.L().Error("Exception occurred while initializing the database connection", zap.Error(err))
		return
	}
	bootstrap.InitFirebase()

	firebaseAuthenticator := authenticator.NewFirebaseAuthenticator()
	authMiddleware := appmiddleware.NewAuthMiddleware(firebaseAuthenticator)

	taskSvc := service.NewTask(fireStoreClient)
	taskUseCase := usecase.NewTask(taskSvc)
	taskHandler := handler.NewTask(taskUseCase)

	userSvc := service.NewUser(fireStoreClient)
	userUseCase := usecase.NewUser(userSvc)
	userHandler := handler.NewUser(userUseCase)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(authMiddleware)
	e.Use(appmiddleware.ZapLogger)
	e.HTTPErrorHandler = appmiddleware.CustomHTTPErrorHandler

	server := bootstrap.Server{Echo: e,
		TaskHandler: taskHandler,
		UserHandler: userHandler}
	// 为所有Handler绑定路由
	bootstrap.BindRoutes(&server)

	go func() {
		zap.L().Info("server starting", zap.String("port", ":8080"))

		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("server start error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	zap.L().Info("shutdown signal received", zap.String("signal", sig.String()))

	// 设置优雅停机超时时间
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("server shutdown error", zap.Error(err))
		return
	}

	zap.L().Info("server stopped gracefully")
}
