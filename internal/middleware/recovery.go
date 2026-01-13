package middleware

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
	"go.uber.org/zap"
)

// RecoveryMiddleware 带自定义日志的恢复中间件
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪信息
				stack := make([]byte, 4096)
				stackLen := runtime.Stack(stack, false)

				// 获取请求上下文信息
				requestID, _ := c.Get("requestID")

				// 记录详细的 panic 日志
				logger.Error("Panic recovered",
					zap.Any("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("panic", err),
					zap.String("stack", string(stack[:stackLen])),
				)

				// 检查是否已经发送过响应,避免重复发送
				if !c.Writer.Written() {
					response.InternalServerErrorFunc(c, "Internal server error")
				}

				// 中断后续处理
				c.Abort()
			}
		}()

		c.Next()
	}
}
