package handler

import (
	httppagination "go-echo-demo/delivery/http/pagination"
	"go-echo-demo/delivery/http/reponse"
	dmpagination "go-echo-demo/internal/domain/pagination"
	ucpagination "go-echo-demo/internal/usecase/pagination"
	"net/http"

	"github.com/labstack/echo/v4"
)

// PaginatedHandler 通用分页 Handler 工厂。
//
// DTO 是各资源自定义的请求结构体（嵌入 BasePageQuery）。
// Item 是响应列表中单条数据的类型。
//
// extractParams 负责从 DTO 中提取业务参数，由调用方实现，
// 使各资源的参数绑定逻辑与通用框架解耦。
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

		result, err := uc.Execute(c.Request().Context(), ucpagination.ExecuteInput{
			Scene:  dmpagination.SceneID(base.Scene),
			Params: params,
			Cursor: base.Cursor,
			Dir:    dmpagination.CursorDir(base.Direction),
			Limit:  base.Limit,
		})
		if err != nil {
			return err
		}

		return reponse.Success(c, httppagination.PageResponse[Item]{
			Items:      result.Items,
			NextCursor: result.NextCursor,
			PrevCursor: result.PrevCursor,
			HasMore:    result.HasMore,
			TotalCount: result.TotalCount,
		})
	}
}

// NewHTTPError 便捷封装，供 bindDTO 实现中使用。
func NewHTTPError(code int, msg string) error {
	return echo.NewHTTPError(code, msg)
}

// BadRequest 便捷封装。
func BadRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}
