package provider

import (
	"context"
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain"
	dmpagination "go-echo-demo/internal/domain/pagination"
	"go-echo-demo/internal/handler"
	"go-echo-demo/internal/infra/firestore/mapper"
	fspagination "go-echo-demo/internal/infra/firestore/pagination"
	"go-echo-demo/internal/infra/firestore/scene"
	"go-echo-demo/internal/infra/firestore/service"
	"go-echo-demo/internal/model"
	"go-echo-demo/internal/response"
	"go-echo-demo/internal/usecase"
	ucpagination "go-echo-demo/internal/usecase/pagination"

	"cloud.google.com/go/firestore"
)

// NewTaskHandler 装配 task 相关的所有组件，返回完整的 TaskHandler。
func NewTaskHandler(client *firestore.Client) *handler.TaskHandler {
	// 基础 CRUD
	taskSvc := service.NewTask(client)
	taskUseCase := usecase.NewTask(taskSvc)

	// 分页查询
	registry := dmpagination.NewRegistry()
	scene.RegisterTaskScenes(registry)

	repo := fspagination.NewFirestoreRepository[model.Task](client, "tasks", mapper.TaskMapper)

	uc := ucpagination.NewQueryUseCase(ucpagination.QueryUseCaseConfig[model.Task, response.TaskItem]{
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
		InjectRules: func(ctx context.Context, q dmpagination.PageQuery) (dmpagination.PageQuery, error) {
			session, ok := domain.FromUserSession(ctx)
			if !ok {
				return q, constants.CredentialsAbsence
			}
			q.Filters = append(q.Filters, dmpagination.FilterCriteria{
				Field: "creator_id",
				Op:    dmpagination.FilterOpEq,
				Value: session.UID,
			})
			return q, nil
		},
	})

	listHandler := handler.NewTaskListHandler(uc, registry)
	return handler.NewTask(taskUseCase).WithListHandler(listHandler)
}
