package pagination

import (
	"go-echo-demo/internal/constants"
	"go-echo-demo/internal/domain/pagination"
)

// BasePageQuery 通用分页基础参数，各资源 DTO 嵌入此结构。
type BasePageQuery struct {
	Scene  string `query:"scene"`
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit"`
}

// ValidateBaseParams 校验通用分页参数的合法性，错误统一返回业务状态码。
func ValidateBaseParams(base BasePageQuery, registry *pagination.Registry) error {
	if base.Scene == "" {
		return constants.RequireAbsence
	}

	found := false
	for _, id := range registry.KnownScenes() {
		if string(id) == base.Scene {
			found = true
			break
		}
	}
	if !found {
		return constants.UnknownScene
	}

	if base.Limit < 0 || base.Limit > 100 {
		return constants.InvalidInputParam
	}

	if base.Cursor != "" {
		if _, err := pagination.DecodeCursor(base.Cursor); err != nil {
			return constants.InvalidCursor
		}
	}

	return nil
}
