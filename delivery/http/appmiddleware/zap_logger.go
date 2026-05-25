package appmiddleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ZapLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		// 执行后续业务逻辑
		err := next(c)

		req := c.Request()
		res := c.Response()
		// 3. 业务执行完毕，用 Zap 打印标准化的 HTTP 请求报文信息
		zap.L().Info("HTTP Request Info",
			zap.String("来源", "Zap Logger"),
			zap.String("method", req.Method),
			zap.String("uri", req.RequestURI),
			zap.Int("status", res.Status),
			zap.Float64("latency_seconds", time.Since(start).Seconds()), // 计算耗时
			zap.String("ip", c.RealIP()),
		)
		return err
	}
}
