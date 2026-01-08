package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/dto"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
	"github.com/yuhang1130/gin-server/internal/pkg/validation"
	"github.com/yuhang1130/gin-server/internal/service"
	"go.uber.org/zap"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
	redis       *cache.Redis
	authCache   *cache.AuthCache
	config      *config.Config
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(
	authService service.AuthService,
	userService service.UserService,
	redis *cache.Redis,
	config *config.Config,
	logger *zap.Logger,
	authCache *cache.AuthCache,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		redis:       redis,
		authCache:   authCache,
		config:      config,
		logger:      logger,
	}
}

// getBindingErrorMessage 获取友好的绑定错误信息
func getBindingErrorMessage(err error) string {
	return validation.TranslateValidationError(err)
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("登录请求参数无效",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+getBindingErrorMessage(err))
		return
	}

	h.logger.Info("收到登录请求",
		zap.String("username", req.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 执行登录
	loginResp, err := h.authService.Login(c, req.Username, req.Password)
	if err != nil {
		h.logger.Warn("登录失败",
			zap.String("username", req.Username),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		// 记录登录失败
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, loginResp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文中获取当前用户
	userSessionData, exists := c.Get("userSessionData")
	if !exists {
		h.logger.Warn("未认证用户尝试登出", zap.String("client_ip", c.ClientIP()))
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := userSessionData.(*cache.UserSessionData)
	if !ok {
		h.logger.Error("获取当前用户失败", zap.String("client_ip", c.ClientIP()))
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	h.logger.Info("收到登出请求",
		zap.Uint64("user_id", currentUser.UserID),
		zap.String("username", currentUser.User.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 获取 Access Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) <= 7 {
		h.logger.Warn("Authorization header 无效",
			zap.Uint64("user_id", currentUser.UserID),
		)
		response.BadRequest(c, "Authorization header missing or invalid")
		return
	}

	tokenString := authHeader[7:] // 移除 "Bearer " 前缀

	// 获取 Token 过期时间
	exp, err := h.authService.GetTokenExpiry(tokenString)
	if err != nil {
		h.logger.Error("获取 Token 过期时间失败",
			zap.Uint64("user_id", currentUser.UserID),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to get token expiry")
		return
	}

	// 计算剩余有效时间
	now := time.Now()
	remainingTime := exp.Sub(now)

	// 将 Access Token 加入黑名单
	if remainingTime > 0 {
		if err := h.authCache.AddToBlacklistByTokenString(c, tokenString, remainingTime); err != nil {
			h.logger.Error("将 Token 加入黑名单失败",
				zap.Uint64("user_id", currentUser.UserID),
				zap.Error(err),
			)
			response.InternalServerError(c, "Failed to blacklist token")
			return
		}
	}

	h.logger.Info("用户登出成功",
		zap.Uint64("user_id", currentUser.UserID),
		zap.String("username", currentUser.User.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 可选：删除用户的所有 Refresh Token（如果需要全局登出）
	// 这里暂时只处理当前 Access Token

	response.Success(c, gin.H{
		"message": "Logged out successfully",
		"user_id": currentUser.UserID,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// 从上下文获取当前用户
	userSessionData, exists := c.Get("userSessionData")
	if !exists {
		h.logger.Warn("未认证用户尝试修改密码", zap.String("client_ip", c.ClientIP()))
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := userSessionData.(*cache.UserSessionData)
	if !ok {
		h.logger.Error("获取当前用户失败", zap.String("client_ip", c.ClientIP()))
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	h.logger.Info("收到修改密码请求",
		zap.Uint64("user_id", currentUser.UserID),
		zap.String("username", currentUser.User.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("修改密码请求参数无效",
			zap.Uint64("user_id", currentUser.UserID),
			zap.Error(err),
		)
		response.BadRequest(c, "Invalid request: "+getBindingErrorMessage(err))
		return
	}

	err := h.userService.ChangePassword(c, currentUser.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	// 密码修改成功后，撤销用户的所有登录会话，强制重新登录
	if err := h.authService.RevokeAllUserSessions(c, currentUser.UserID); err != nil {
		h.logger.Warn("撤销用户所有会话失败（密码已修改成功）",
			zap.Uint64("user_id", currentUser.UserID),
			zap.String("username", currentUser.User.Username),
			zap.Error(err),
		)
		// 撤销会话失败不影响密码修改成功的结果，只记录警告
	}

	h.logger.Info("用户密码修改成功，所有会话已撤销",
		zap.Uint64("user_id", currentUser.UserID),
		zap.String("username", currentUser.User.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	response.Success(c, gin.H{
		"message": "Password changed successfully. Please login again.",
	})
}
