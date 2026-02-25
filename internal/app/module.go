package app

import (
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/handler"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/database"
	"github.com/yuhang1130/gin-server/internal/repository"
	"github.com/yuhang1130/gin-server/internal/service"
	"go.uber.org/fx"
)

// Module 汇总所有 fx provider 和 invoke
var Module = fx.Options(
	database.Module,
	cache.Module,
	repository.Module,
	service.Module,
	handler.Module,
	fx.Provide(
		config.LoadConfig,
		provideLogger,
		provideJWTUtil,
		provideSnowflakeGenerator,
		provideAppContextAdapter,
		provideGinEngine,
		provideHTTPServer,
	),
	fx.Invoke(registerRoutesFx, startServerFx),
)
