package app

import (
	"context"
	"net/http"
	"time"

	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func startServerFx(
	lc fx.Lifecycle,
	server *http.Server,
	db *database.MySQL,
	rdb *cache.Redis,
	logger *zap.Logger,
	shutdowner fx.Shutdowner,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Server starting", zap.String("addr", server.Addr))
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Server error", zap.Error(err))
					_ = shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Server shutting down...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Error("Server shutdown error", zap.Error(err))
			}
			_ = logger.Sync()
			if err := db.Close(); err != nil {
				logger.Error("DB close error", zap.Error(err))
			}
			if err := rdb.Close(); err != nil {
				logger.Error("Redis close error", zap.Error(err))
			}
			return nil
		},
	})
}
