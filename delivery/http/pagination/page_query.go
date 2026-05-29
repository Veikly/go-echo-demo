package pagination

import (
	"go-echo-demo/internal/domain/pagination"
	"net/http"

	"github.com/labstack/echo/v4"
)

// BasePageQuery 通用分页基础参数，各资源 DTO 嵌入此结构。
type BasePageQuery struct {
	Scene     string `query:"scene"`
	Cursor    string `query:"cursor"`
	Direction string `query:"direction"`
	Limit     int    `query:"limit"`
}

// ValidateBaseParams 校验通用分页参数的合法性。
func ValidateBaseParams(base BasePageQuery, registry *pagination.Registry) error {
	if base.Scene == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "scene is required")
	}
	found := false
	for _, id := range registry.KnownScenes() {
		if string(id) == base.Scene {
			found = true
			break
		}
	}
	if !found {
		return echo.NewHTTPError(http.StatusBadRequest, "unknown scene: "+base.Scene)
	}

	if base.Limit < 0 || base.Limit > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be between 1 and 100")
	}

	if base.Direction != "" &&
		base.Direction != string(pagination.CursorForward) &&
		base.Direction != string(pagination.CursorBackward) &&
		base.Direction != string(pagination.CursorRefresh) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid direction")
	}

	if base.Cursor != "" {
		if _, err := pagination.DecodeCursor(base.Cursor); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
		}
	}

	return nil
}
