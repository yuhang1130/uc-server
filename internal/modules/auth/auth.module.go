package auth

import (
	"github.com/yuhang1130/gin-server/internal/middleware"
	"github.com/yuhang1130/gin-server/internal/modules/user"
	api_router "github.com/yuhang1130/gin-server/internal/router/api"
	"github.com/yuhang1130/gin-server/pkg/jwt"
	"github.com/yuhang1130/gin-server/pkg/logger"
	"github.com/yuhang1130/gin-server/pkg/redis"
	"go.uber.org/fx"
)

// AuthCacheResult 用于同时提供 AuthCache 的具体类型和接口类型
type AuthCacheResult struct {
	fx.Out

	AuthCache        *AuthCache
	BlacklistChecker middleware.TokenBlacklistChecker
}

// ProvideAuthCache 提供 AuthCache，同时作为具体类型和接口类型
func ProvideAuthCache(redis *redis.Redis) AuthCacheResult {
	cache := NewAuthCache(redis)
	return AuthCacheResult{
		AuthCache:        cache,
		BlacklistChecker: cache,
	}
}

// AuthServiceResult 用于同时提供 AuthService 的具体类型和接口类型
type AuthServiceResult struct {
	fx.Out

	AuthService     AuthService
	SessionProvider middleware.SessionDataProvider
}

// ProvideAuthService 提供 AuthService，同时作为具体类型和接口类型
func ProvideAuthService(
	userRepo user.UserRepository,
	tenantUserRepo user.TenantUserRepository,
	jwtUtil jwt.JWT,
	authCache *AuthCache,
	userCache *user.UserCache,
	logger logger.Logger,
) AuthServiceResult {
	service := NewAuthService(userRepo, tenantUserRepo, jwtUtil, authCache, userCache, logger)
	return AuthServiceResult{
		AuthService:     service,
		SessionProvider: service,
	}
}

var Module = fx.Module(
	"auth",

	fx.Provide(
		// auth cache - 同时提供具体类型和接口类型
		ProvideAuthCache,
		// auth service - 同时提供具体类型和接口类型
		ProvideAuthService,
		// auth controller
		NewAuthController,
	),

	// register router
	fx.Invoke(func(
		router api_router.APIRouterParams,
		authController *AuthController,
	) {
		// public routers
		{
			authGroup := router.Public_API_V1.Group("/auth")
			authGroup.POST("/login", authController.Login)
		}

		// protected routers
		{
			authGroup := router.Private_API_V1.Group("/auth")
			authGroup.POST("/logout", authController.Logout)
			authGroup.POST("/change-password", authController.ChangePassword)
		}
	}),
)
