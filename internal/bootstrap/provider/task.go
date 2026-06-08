package provider

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	domainpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/handler"
	"go-echo-demo/internal/infra/firestore/mapper"
	"go-echo-demo/internal/infra/firestore/page/scene"
	firestoreservice "go-echo-demo/internal/infra/firestore/service"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/response"
	usecase "go-echo-demo/internal/usecase/impl"
	usecasepagination "go-echo-demo/internal/usecase/pagination"

	"cloud.google.com/go/firestore"
)

// NewTaskHandler 装配 task 相关的所有组件，返回完整的 TaskHandler。
func NewTaskHandler(client *firestore.Client) *handler.TaskHandler {
	// 基础 CRUD
	taskSvc := firestoreservice.NewTask(client)
	taskUseCase := usecase.NewTask(taskSvc, GlobalTransationManger)
	// 分页查询
	registry := domainpagination.NewRegistry()
	scene.RegisterTaskScenes(registry)

	repo := firestoreservice.NewFirestoreRepository[model.Task](client, "tasks", mapper.TaskMapper)

	uc := usecasepagination.NewQueryUseCase(usecasepagination.QueryUseCaseConfig[model.Task, response.TaskItem]{
		Repo:     repo,
		Registry: registry,
		ToDTO: func(t model.Task) response.TaskItem {
			return response.TaskItem{
				ID:          t.ID,
				Title:       t.Title,
				Description: t.Description,
				Status:      t.Status,
				CreatorID:   t.Creator.ID,
				CompletedAt: t.CompletedAt,
				CreatedAt:   t.CreatedAt,
				UpdatedAt:   t.UpdatedAt,
			}
		},
		// 注入规则 只允许访问自己创建的任务 不允许访问其他人的
		InjectRules: func(ctx context.Context, q domainpagination.PageQuery) (domainpagination.PageQuery, error) {
			session, ok := domain.FromUserSession(ctx)
			if !ok {
				return q, constants.CredentialsAbsence
			}
			q.Filters = append(q.Filters, domainpagination.FilterCriteria{
				Field: "creator_id",
				Op:    domainpagination.FilterOpEq,
				Value: session.UID,
			})
			return q, nil
		},
	})

	listHandler := handler.NewTaskListHandler(uc, registry)
	return handler.NewTask(taskUseCase).WithListHandler(listHandler)
}
