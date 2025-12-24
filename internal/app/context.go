package app

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/handler"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/database"
	"github.com/yuhang1130/gin-server/internal/pkg/jwt"
	"github.com/yuhang1130/gin-server/internal/pkg/snowflake"
	"github.com/yuhang1130/gin-server/internal/repository"
	"github.com/yuhang1130/gin-server/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AppContext 应用上下文，管理所有应用级别的依赖和单例对象
type AppContext struct {
	// 基础设施
	Config *config.Config
	DB     *database.MySQL
	Redis  *cache.Redis

	// zap logger（全局使用）
	zapLogger *zap.Logger

	// JWT 工具单例
	jwtUtil     *jwt.JWTUtil
	jwtUtilOnce sync.Once

	// Snowflake ID 生成器单例
	snowflakeGenerator     *snowflake.Generator
	snowflakeGeneratorOnce sync.Once

	// Cache 单例
	authCache     *cache.AuthCache
	authCacheOnce sync.Once
	userCache     *cache.UserCache
	userCacheOnce sync.Once

	// Repository 单例
	userRepo           repository.UserRepository
	userRepoOnce       sync.Once
	tenantUserRepo     repository.TenantUserRepository
	tenantUserRepoOnce sync.Once

	// Service 单例
	authService     service.AuthService
	authServiceOnce sync.Once
	userService     service.UserService
	userServiceOnce sync.Once

	// Handler 单例
	authHandler     *handler.AuthHandler
	authHandlerOnce sync.Once
	userHandler     *handler.UserHandler
	userHandlerOnce sync.Once
}

var (
	appCtx *AppContext
)

// InitApp 初始化应用上下文
func InitApp() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// 初始化 zap logger
	var zapLogger *zap.Logger
	if cfg.Server.Mode == "production" {
		zapLogger, err = zap.NewProduction()
	} else {
		zapLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		return err
	}

	// 初始化数据库
	db, err := database.NewMySQL(cfg)
	if err != nil {
		zapLogger.Error("Failed to initialize database", zap.Error(err))
		return err
	}

	// 初始化Redis
	redis, err := cache.NewRedis(cfg)
	if err != nil {
		zapLogger.Error("Failed to initialize Redis", zap.Error(err))
		return err
	}

	appCtx = &AppContext{
		Config:    cfg,
		DB:        db,
		Redis:     redis,
		zapLogger: zapLogger,
	}

	return nil
}

// GetAppContext 获取全局 AppContext 实例
func GetAppContext() *AppContext {
	return appCtx
}

// GetConfig 获取配置
func GetConfig() *config.Config {
	return appCtx.Config
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return appCtx.DB.DB
}

// GetRedis 获取Redis客户端
func GetRedis() *cache.Redis {
	return appCtx.Redis
}

// GetLogger 获取 zap logger（兼容旧代码）
func GetLogger() *zap.Logger {
	return appCtx.GetZapLogger()
}

// GetZapLogger 获取 zap logger 单例
func (ctx *AppContext) GetZapLogger() *zap.Logger {
	return ctx.zapLogger
}

// GetJWTUtil 获取 JWT 工具单例
func (ctx *AppContext) GetJWTUtil() *jwt.JWTUtil {
	ctx.jwtUtilOnce.Do(func() {
		ctx.jwtUtil = jwt.NewJWTUtil(
			ctx.Config.JWT.SecretKey,
			ctx.Config.JWT.ExpiresIn,
		)
	})
	return ctx.jwtUtil
}

// GetSnowflakeGenerator 获取 Snowflake ID 生成器单例
func (ctx *AppContext) GetSnowflakeGenerator() *snowflake.Generator {
	ctx.snowflakeGeneratorOnce.Do(func() {
		generator, err := snowflake.NewGenerator(ctx.Config.Snowflake.NodeID)
		if err != nil {
			ctx.zapLogger.Error("Failed to create snowflake generator", zap.Error(err))
			// 使用默认节点 ID 1 作为 fallback
			generator, _ = snowflake.NewGenerator(1)
		}
		ctx.snowflakeGenerator = generator

		// 同时初始化全局默认生成器
		_ = snowflake.InitDefault(ctx.Config.Snowflake.NodeID)
	})
	return ctx.snowflakeGenerator
}

// GetAuthCache 获取认证缓存单例
func (ctx *AppContext) GetAuthCache() *cache.AuthCache {
	ctx.authCacheOnce.Do(func() {
		ctx.authCache = cache.NewAuthCache(ctx.Redis, ctx.GetZapLogger())
	})
	return ctx.authCache
}

// GetUserCache 获取用户缓存单例
func (ctx *AppContext) GetUserCache() *cache.UserCache {
	ctx.userCacheOnce.Do(func() {
		ctx.userCache = cache.NewUserCache(ctx.Redis)
	})
	return ctx.userCache
}

// GetUserRepository 获取用户仓库单例
func (ctx *AppContext) GetUserRepository() repository.UserRepository {
	ctx.userRepoOnce.Do(func() {
		ctx.userRepo = repository.NewUserRepository(ctx.DB.DB)
	})
	return ctx.userRepo
}

// GetTenantUserRepository 获取租户用户关联仓库单例
func (ctx *AppContext) GetTenantUserRepository() repository.TenantUserRepository {
	ctx.tenantUserRepoOnce.Do(func() {
		ctx.tenantUserRepo = repository.NewTenantUserRepository(ctx.DB.DB)
	})
	return ctx.tenantUserRepo
}

// GetAuthService 获取认证服务单例
func (ctx *AppContext) GetAuthService() service.AuthService {
	ctx.authServiceOnce.Do(func() {
		ctx.authService = service.NewAuthService(
			ctx.GetUserRepository(),
			ctx.GetTenantUserRepository(),
			ctx.GetJWTUtil(),
			ctx.GetAuthCache(),
			ctx.GetUserCache(),
			ctx.GetZapLogger(),
		)
	})
	return ctx.authService
}

// GetUserService 获取用户服务单例
func (ctx *AppContext) GetUserService() service.UserService {
	ctx.userServiceOnce.Do(func() {
		ctx.userService = service.NewUserService(
			ctx.GetUserRepository(),
			ctx.GetUserCache(),
			ctx.GetZapLogger(),
		)
	})
	return ctx.userService
}

// GetAuthHandler 获取认证处理器单例
func (ctx *AppContext) GetAuthHandler() *handler.AuthHandler {
	ctx.authHandlerOnce.Do(func() {
		ctx.authHandler = handler.NewAuthHandler(
			ctx.GetAuthService(),
			ctx.GetUserService(),
			ctx.Redis,
			ctx.Config,
			ctx.zapLogger,
		)
	})
	return ctx.authHandler
}

// GetUserHandler 获取用户处理器单例
func (ctx *AppContext) GetUserHandler() *handler.UserHandler {
	ctx.userHandlerOnce.Do(func() {
		ctx.userHandler = handler.NewUserHandler(ctx.GetUserService(), ctx.zapLogger)
	})
	return ctx.userHandler
}

// GetRedisClient 获取 Redis 客户端（用于中间件接口实现）
func (ctx *AppContext) GetRedisClient() *redis.Client {
	return ctx.Redis.Client
}

// GetContext 获取基础上下文
func GetContext() context.Context {
	return context.Background()
}

// Close 关闭资源
func Close() error {
	// 先同步 zap logger
	if appCtx.zapLogger != nil {
		_ = appCtx.zapLogger.Sync()
	}

	if appCtx.DB != nil {
		if err := appCtx.DB.Close(); err != nil {
			return err
		}
	}

	if appCtx.Redis != nil {
		if err := appCtx.Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}
