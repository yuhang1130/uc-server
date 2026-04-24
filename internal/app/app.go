package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/modules"
	"github.com/yuhang1130/gin-server/internal/router"
	api_router "github.com/yuhang1130/gin-server/internal/router/api"
	http_server "github.com/yuhang1130/gin-server/pkg/http_server"
	"github.com/yuhang1130/gin-server/pkg/jwt"
	"github.com/yuhang1130/gin-server/pkg/logger"
	"github.com/yuhang1130/gin-server/pkg/mysql"
	"github.com/yuhang1130/gin-server/pkg/redis"
	"go.uber.org/fx"
)

func Run(cfg *config.Config) {
	fx.New(
		fx.Supply(cfg),
		fx.Provide(
			// logger
			provideLogger,
			// mysql
			fx.Annotate(
				provideMysql,
				fx.OnStop(func(logger logger.Logger, m *mysql.MySQL) error {
					logger.Info("app - Run - mysql closed")
					return m.Close()
				}),
			),
			// redis
			fx.Annotate(
				provideRedis,
				fx.OnStop(func(logger logger.Logger, rdb *redis.Redis) error {
					logger.Info("app - Run - redis closed")
					rdb.Close()
					return nil
				}),
			),
			// jwt
			provideJWT,
			// gin engine
			router.NewEngine,
			// api router
			api_router.NewAPIRouter,
			// http server
			provideHTTPServer,
		),
		// api modules
		modules.APIModule,
		// start http server
		fx.Invoke(startHTTPServer),
	).Run()
}

func provideLogger(cfg *config.Config) logger.Logger {
	return logger.New(cfg.Log.Dir, cfg.Log.Level, cfg.Server.Mode.IsProd())
}
func provideMysql(cfg *config.Config, logger logger.Logger) (*mysql.MySQL, error) {
	m, err := mysql.New(cfg.Database.MysqlURL, mysql.MaxPoolSize(cfg.Database.MysqlPoolMax))
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - mysql.New: %w", err))
		return nil, err
	}

	logger.Info("app - Run - mysql connected")
	return m, nil
}
func provideRedis(cfg *config.Config, logger logger.Logger) (*redis.Redis, error) {
	rdb, err := redis.New(cfg.Redis.URL, redis.MaxPoolSize(cfg.Redis.PoolMax))
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - redis.New: %w", err))
		return nil, err
	}

	logger.Info("app - Run - redis connected")
	return rdb, nil
}
func provideJWT(cfg *config.Config) jwt.JWT {
	return jwt.NewJWT(
		jwt.Expire(time.Duration(cfg.JWT.ExpiresIn)),
		jwt.Issuer(cfg.App.Name),
		jwt.Secret(cfg.JWT.SecretKey),
	)
}
func provideHTTPServer(cfg *config.Config, engine *gin.Engine) http_server.Server {
	return http_server.New(engine, http_server.Port(cfg.Server.Port))
}
func startHTTPServer(
	lc fx.Lifecycle,
	sd fx.Shutdowner,
	logger logger.Logger,
	httpServer http_server.Server,
) {
	lc.Append(
		fx.Hook{
			// start
			OnStart: func(ctx context.Context) error {
				logger.Info("http server - Starting HTTP Server...")

				go func() {
					if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
						logger.Errorf("http server - Start HTTP Server error: %v", err)
						sd.Shutdown()
					}
				}()

				logger.Info("http server - Listening on ", httpServer.GetAddress())
				return nil
			},
			// stop
			OnStop: func(ctx context.Context) error {
				logger.Info("http server - Stopping HTTP Server...")

				err := httpServer.Shutdown()
				if err != nil {
					return err
				}

				logger.Info("http server - Server - Shutting down")
				return nil
			},
		},
	)
}
