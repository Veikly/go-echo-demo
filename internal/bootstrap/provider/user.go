package provider

import (
	"go-echo-demo/internal/handler"
	"go-echo-demo/internal/infra/firestore/service"
	"go-echo-demo/internal/usecase"

	"cloud.google.com/go/firestore"
)

// NewUserHandler 装配 user 相关的所有组件，返回完整的 UserHandler。
func NewUserHandler(client *firestore.Client) *handler.UserHandler {
	userSvc := service.NewUser(client)
	userUseCase := usecase.NewUser(userSvc)
	return handler.NewUser(userUseCase)
}
