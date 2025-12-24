package middleware

import (
	"context"
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
		tokenID := jwt.GenerateTokenIDFromToken(tokenString)
		redisClient := appCtx.GetRedisClient()
		isBlacklisted, err := redisClient.Get(c, "blacklist:"+tokenID).Result()
		if err == nil && isBlacklisted == "true" {
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

		// 使用 Redis 缓存优化：先尝试从缓存获取用户信息
		userCache := appCtx.GetUserCache()
		user, err := userCache.GetUserSession(c, claims.UserID, claims.ID)

		// 缓存未命中，从数据库加载并写入缓存
		if err != nil {
			userService := appCtx.GetUserService()
			user, err := userService.GetUserByID(c, claims.UserID)
			if err != nil {
				response.Unauthorized(c, "User not found")
				c.Abort()
				return
			}

			// 异步写入缓存，不阻塞请求
			go func() {
				ctx := context.Background()
				_ = userCache.SetUser(ctx, user)
			}()
		}

		c.Set("currentUser", user)

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
