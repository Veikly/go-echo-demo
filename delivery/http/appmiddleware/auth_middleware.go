package appmiddleware

import (
	"go-echo-demo/internal/bootstrap"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func FirebaseAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		zap.L().Debug("请求认证中间件", zap.String("url", c.Request().URL.String()))
		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		// 后续如果有需要对外开放的白名单接口 这里还需要进行判断
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "缺少鉴权凭证")
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "凭据格式错误 不满足Bearer <token>")
		}
		idToken := parts[1]
		token, err := bootstrap.AuthClient.VerifyIDToken(c.Request().Context(), idToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "凭据无效或已过期：%v", token)
		}
		c.Set("firebase_token", token)
		return next(c)
	}
}
