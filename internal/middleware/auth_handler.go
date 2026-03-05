package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yuhang1130/gin-server/internal/model"
	jwtZ "github.com/yuhang1130/gin-server/pkg/jwt"
)

/**
 * AuthCache 实现了 TokenBlacklistChecker 接口（已有 IsInBlacklistByTokenID 方法）
 * Fx 自动将 *auth.AuthCache 注入为 middleware.TokenBlacklistChecker 接口
 * middleware 不再依赖 auth 模块，解除了循环依赖
 */
// TokenBlacklistChecker 检查 token 是否在黑名单中的接口
type TokenBlacklistChecker interface {
	IsInBlacklistByTokenID(ctx context.Context, tokenString string) (bool, error)
}

// SessionDataProvider 获取用户会话数据的接口
type SessionDataProvider interface {
	GetUserSessionData(ctx *gin.Context, userID, tenantID uint64, claimsID string) (*model.UserSessionData, error)
}

// validateAuthHeader 验证 Authorization header 并提取 token
func validateAuthHeader(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", jwt.ErrInvalidKey
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", jwt.ErrTokenMalformed
	}

	return tokenParts[1], nil
}

// checkTokenBlacklist 检查 token 是否在黑名单中
func checkTokenBlacklist(ctx *gin.Context, token string, checker TokenBlacklistChecker) error {
	isBlacklisted, err := checker.IsInBlacklistByTokenID(ctx, token)
	if err == nil && isBlacklisted {
		return jwt.ErrInvalidType
	}
	return nil
}

// validateTokenClaims 验证 token 并返回 claims
func validateTokenClaims(token string, j jwtZ.JWT) (*jwtZ.Claims, error) {
	claims, err := j.ParseAccessToken(token)
	if err != nil {
		return nil, jwt.ErrTokenSignatureInvalid
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Before(time.Now()) {
		return nil, jwt.ErrTokenExpired
	}

	if claims == nil || claims.UserID == 0 {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// setUserContext 设置用户信息到上下文
func setUserContext(ctx *gin.Context, claims *jwtZ.Claims, sessionData *model.UserSessionData) {
	ctx.Set("userID", claims.UserID)
	ctx.Set("userRole", claims.Role)
	ctx.Set("tenantID", claims.TenantID)
	ctx.Set("userSessionData", sessionData)
}

// APIAuthHandler API 鉴权中间件.
func APIAuthHandler(
	j jwtZ.JWT,
	blacklistChecker TokenBlacklistChecker,
	sessionProvider SessionDataProvider,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 验证并提取 token
		token, err := validateAuthHeader(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		// 检查黑名单
		if err := checkTokenBlacklist(ctx, token, blacklistChecker); err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		// 验证 token 并获取 claims
		claims, err := validateTokenClaims(token, j)
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		// 获取用户会话数据
		userSessionData, err := sessionProvider.GetUserSessionData(ctx, claims.UserID, claims.TenantID, claims.ID)
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, errors.New("Failed to get user session data"))
			return
		}

		// 设置用户信息到上下文
		setUserContext(ctx, claims, userSessionData)

		ctx.Next()
	}
}

func RoleAuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithError(http.StatusForbidden, errors.New("Role not found in context"))
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.AbortWithError(http.StatusForbidden, errors.New("Role not found in context"))
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
			c.AbortWithError(http.StatusForbidden, errors.New("Insufficient permissions"))
			return
		}

		c.Next()
	}
}
