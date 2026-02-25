package app

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/internal/pkg/jwt"
	"github.com/yuhang1130/gin-server/internal/pkg/snowflake"
	"go.uber.org/zap"
)

func provideLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Server.Mode == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func provideJWTUtil(cfg *config.Config) *jwt.JWTUtil {
	return jwt.NewJWTUtil(cfg.JWT.SecretKey, cfg.JWT.ExpiresIn)
}

func provideSnowflakeGenerator(cfg *config.Config) (*snowflake.Generator, error) {
	nodeID := cfg.Snowflake.NodeID
	gen, err := snowflake.NewGenerator(nodeID)
	if err != nil {
		// fallback to node 1
		nodeID = 1
		gen, err = snowflake.NewGenerator(nodeID)
	}
	if err != nil {
		return nil, err
	}
	_ = snowflake.InitDefault(nodeID)
	return gen, nil
}

func provideGinEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	isProd := cfg.Server.Mode == "production"
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	if !isProd {
		pprof.Register(engine)
	}

	engine.Use(
		middleware.RequestIDMiddleware(),
		middleware.BodySizeLimitMiddleware(cfg.Server.MaxBodySize),
		middleware.LoggingMiddleware(logger),
		middleware.RecoveryMiddleware(logger),
		middleware.CORSMiddleware(cfg.Server.CorsAllowed),
	)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			field := fld.Tag.Get("json")
			if field == "" {
				field = fld.Tag.Get("form")
			}
			name := strings.SplitN(field, ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		v.RegisterValidation("is_mobile", validateMobile)
	}

	return engine
}

func provideHTTPServer(cfg *config.Config, engine *gin.Engine) *http.Server {
	timeout := time.Duration(cfg.Server.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}
}

func validateMobile(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, fl.Field().String())
	return matched
}
