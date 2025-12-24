package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimitMiddleware 限制请求体大小的中间件
// maxSize 单位为字节，例如 1MB = 1 * 1024 * 1024
func BodySizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求体最大大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// 继续处理请求
		c.Next()

		// 检查是否因为请求体过大而失败
		if c.Errors.Last() != nil {
			if c.Errors.Last().Err.Error() == "http: request body too large" {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"code":    http.StatusRequestEntityTooLarge,
					"message": "请求体过大",
				})
				c.Abort()
			}
		}
	}
}
