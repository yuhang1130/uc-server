package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/jwt"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
	"github.com/yuhang1130/gin-server/internal/service"
)

// AppContextProvider 定义应用上下文提供者接口，避免循环依赖
type AppContextProvider interface {
	GetJWTUtil() *jwt.JWTUtil
	GetUserService() service.UserService
	GetRedisClient() *redis.Client
	GetUserCache() *cache.UserCache
	GetAuthCache() *cache.AuthCache
	GetAuthService() service.AuthService
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(appCtx AppContextProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// 验证token格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid token format")
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 检查token是否在黑名单中
		authCache := appCtx.GetAuthCache()
		isBlacklisted, err := authCache.IsInBlacklistByTokenID(c, tokenString)
		if err == nil && isBlacklisted {
			response.Unauthorized(c, "Token has been revoked")
			c.Abort()
			return
		}

		// 使用 AppContext 中的单例 JWTUtil 解析token
		jwtUtil := appCtx.GetJWTUtil()
		claims, err := jwtUtil.ParseAccessToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 验证token是否过期
		if time.Unix(claims.ExpiresAt.Unix(), 0).Before(time.Now()) {
			response.Unauthorized(c, "Token has expired")
			c.Abort()
			return
		}

		// 将用户信息放入上下文
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("tenantID", claims.TenantID)

		// 使用 AppContext 中的单例 AuthService 获取用户完整信息
		authService := appCtx.GetAuthService()
		userSessionData, err := authService.GetUserSessionData(c, claims.UserID, claims.TenantID, claims.ID)
		if err != nil {
			response.Unauthorized(c, "Failed to get user session data")
			c.Abort()
			return
		}

		c.Set("userSessionData", userSessionData)

		c.Next()
	}
}

// RoleAuthMiddleware 基于角色的授权中间件
func RoleAuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			response.Forbidden(c, "Role not found in context")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Forbidden(c, "Invalid role format")
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的列表中
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
