package middleware

import (
	"bytes"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestIDMiddleware 为每个请求生成唯一ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := generateRequestID()
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算耗时
		cost := time.Since(startTime)

		// 获取请求信息
		requestID, _ := c.Get("requestID")
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("cost", cost),
			zap.String("req_id", requestID.(string)),
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			logger.Info("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}

// ResponseBodyLogger 响应体日志中间件
func ResponseBodyLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建响应缓冲区
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 保存响应体
		responseBody := blw.body.String()
		c.Set("response_body", responseBody)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func generateRequestID() string {
	// 生成唯一请求ID，可以使用UUID或其他方法
	return "req-" + time.Now().Format("20060102150405") + "-" + string(rune(rand.Intn(90)+65))
}
