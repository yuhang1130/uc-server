package router

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	timeout "github.com/vearne/gin-timeout"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewEngine(
	cfg *config.Config,
	logger logger.Logger,
) *gin.Engine {
	isProd := cfg.Server.Mode == "production"
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.New()

	if !isProd {
		pprof.Register(app)
	}

	// apply middlewares
	app.Use(requestid.New())
	app.Use(middleware.BodySizeLimit(cfg.Server.MaxBodySize)) // 10MB limit
	app.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-Request-ID", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:8090"
		},
		MaxAge: 1 * time.Hour, // 减少不必要的 OPTIONS 请求，降低网络开销，提升跨域请求的性能
	}))
	app.Use(ginzap.GinzapWithConfig(logger.Logger(), &ginzap.Config{
		UTC:        false,
		TimeFormat: time.DateTime,
		Context: func(ctx *gin.Context) []zapcore.Field {
			var fields []zapcore.Field
			// log request ID
			if rid := requestid.Get(ctx); rid != "" {
				fields = append(fields, zap.String("request_id", rid))
			}

			// log trace and span ID
			if trace.SpanFromContext(ctx.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(ctx.Request.Context()).SpanContext().SpanID().String()))
			}

			// log request body (skip if too large)
			const maxBodySize = 1024 * 10 // 10KB
			if ctx.Request.ContentLength > 0 && ctx.Request.ContentLength <= maxBodySize {
				var body []byte
				var buf bytes.Buffer
				tee := io.TeeReader(ctx.Request.Body, &buf)
				body, _ = io.ReadAll(tee)
				ctx.Request.Body = io.NopCloser(&buf)
				fields = append(fields, zap.String("body", string(body)))
			} else if ctx.Request.ContentLength > maxBodySize {
				fields = append(fields, zap.String("body", "[body too large to log]"))
			}

			return fields
		},
	}))
	app.Use(ginzap.RecoveryWithZap(logger.Logger(), true))
	app.Use(timeout.Timeout(timeout.WithTimeout(10 * time.Second)))
	app.Use(middleware.ErrorHandler(logger))

	// TODO Swagger

	return app
}
