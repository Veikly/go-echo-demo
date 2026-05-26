package appmiddleware

import (
	"go-echo-demo/internal/domain"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func NewAuthMiddleware(authenticator domain.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			zap.L().Info("请求认证中间件", zap.String("url", c.Request().URL.String()))
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			// 后续如果有需要对外开放的白名单接口 这里还需要进行判断
			if authHeader == "" {
				zap.L().Error("认证信息为空")
				return echo.NewHTTPError(http.StatusUnauthorized, "缺少鉴权凭证")
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				zap.L().Error("凭据格式错误")
				return echo.NewHTTPError(http.StatusUnauthorized, "凭据格式错误 不满足Bearer <token>")
			}
			idToken := parts[1]
			// 调用传入的认证
			userSession, err := authenticator.Authenticate(c.Request().Context(), idToken)
			if err != nil {
				zap.L().Error("凭据无效或已过期")
				return echo.NewHTTPError(http.StatusUnauthorized, "凭据无效或已过期!")
			}
			// 注入UserSession到Context中
			reqCtx := domain.WithUserSession(c.Request().Context(), *userSession)
			c.SetRequest(c.Request().WithContext(reqCtx))
			return next(c)
		}
	}
}
