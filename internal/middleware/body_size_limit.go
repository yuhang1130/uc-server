package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimit limits the maximum size of request body
func BodySizeLimit(maxSize int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.ContentLength > maxSize {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    http.StatusRequestEntityTooLarge,
				"message": "request body too large",
			})
			ctx.Abort()
			return
		}

		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxSize)
		ctx.Next()
	}
}
