package handler

import (
	httppagination "go-echo-demo/delivery/http/pagination"
	"go-echo-demo/delivery/http/reponse"
	dmpagination "go-echo-demo/internal/domain/pagination"
	ucpagination "go-echo-demo/internal/usecase/pagination"
	"go-echo-demo/internal/usecase/usecaseio"

	"github.com/labstack/echo/v4"
)

// PaginatedHandler 通用分页 Handler 工厂。
// bindDTO 负责绑定请求参数并提取业务 SceneParams，由各资源自己实现。
func PaginatedHandler[T, Item any](
	uc *ucpagination.QueryUseCase[T, Item],
	registry *dmpagination.Registry,
	bindDTO func(c echo.Context) (httppagination.BasePageQuery, dmpagination.SceneParams, error),
) echo.HandlerFunc {
	return func(c echo.Context) error {
		base, params, err := bindDTO(c)
		if err != nil {
			return err
		}

		if err := httppagination.ValidateBaseParams(base, registry); err != nil {
			return err
		}

		result, err := uc.Execute(c.Request().Context(), usecaseio.ExecuteInput{
			Scene:          dmpagination.SceneID(base.Scene),
			Params:         params,
			Cursor:         base.Cursor,
			Limit:          base.Limit,
			WithTotalCount: base.WithTotalCount,
		})
		if err != nil {
			return err
		}

		return reponse.Success(c, httppagination.PageResponse[Item]{
			Items:      result.Items,
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
			TotalCount: result.TotalCount,
		})
	}
}
