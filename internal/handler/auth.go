package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/config"
	"github.com/yuhang1130/gin-server/internal/dto"
	"github.com/yuhang1130/gin-server/internal/model"
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
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		redis:       redis,
		config:      config,
		logger:      logger,
	}
}

// generateRefreshTokenKey 生成 Refresh Token 的 Redis Key（使用 Token 的 hash）
func generateRefreshTokenKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return "refresh_token:" + hex.EncodeToString(hash[:16]) // 使用前 16 字节（32 个十六进制字符）
}

// generateAccessTokenKey 生成 Access Token 的黑名单 Redis Key
func generateAccessTokenKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return "blacklist:" + hex.EncodeToString(hash[:16])
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
		// middleware.RecordLoginFailure(c, h.redis.Client, req.Username, middleware.DefaultLoginRateLimit)
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	response.Success(c, loginResp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文中获取当前用户
	user, exists := c.Get("currentUser")
	if !exists {
		h.logger.Warn("未认证用户尝试登出", zap.String("client_ip", c.ClientIP()))
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		h.logger.Error("获取当前用户失败", zap.String("client_ip", c.ClientIP()))
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	h.logger.Info("收到登出请求",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 获取 Access Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) <= 7 {
		h.logger.Warn("Authorization header 无效",
			zap.Uint64("user_id", currentUser.ID),
		)
		response.BadRequest(c, "Authorization header missing or invalid")
		return
	}

	tokenString := authHeader[7:] // 移除 "Bearer " 前缀

	// 获取 Token 过期时间
	exp, err := h.authService.GetTokenExpiry(tokenString)
	if err != nil {
		h.logger.Error("获取 Token 过期时间失败",
			zap.Uint64("user_id", currentUser.ID),
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
		tokenKey := generateAccessTokenKey(tokenString)
		if err := h.redis.Client.Set(c, tokenKey, "true", remainingTime).Err(); err != nil {
			h.logger.Error("将 Token 加入黑名单失败",
				zap.Uint64("user_id", currentUser.ID),
				zap.Error(err),
			)
			response.InternalServerError(c, "Failed to blacklist token")
			return
		}
	}

	h.logger.Info("用户登出成功",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 可选：删除用户的所有 Refresh Token（如果需要全局登出）
	// 这里暂时只处理当前 Access Token

	response.Success(c, gin.H{
		"message": "Logged out successfully",
		"user_id": currentUser.ID,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// 从上下文获取当前用户
	user, exists := c.Get("currentUser")
	if !exists {
		h.logger.Warn("未认证用户尝试修改密码", zap.String("client_ip", c.ClientIP()))
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		h.logger.Error("获取当前用户失败", zap.String("client_ip", c.ClientIP()))
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	h.logger.Info("收到修改密码请求",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("修改密码请求参数无效",
			zap.Uint64("user_id", currentUser.ID),
			zap.Error(err),
		)
		response.BadRequest(c, "Invalid request: "+getBindingErrorMessage(err))
		return
	}

	// 验证旧密码
	if !currentUser.CheckPassword(req.OldPassword) {
		h.logger.Warn("旧密码验证失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Old password is incorrect")
		return
	}

	// 验证新密码强度
	if err := validation.ValidatePasswordWithCommonChecks(req.NewPassword); err != nil {
		h.logger.Warn("新密码强度验证失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	// 检查新密码是否与旧密码相同
	if req.OldPassword == req.NewPassword {
		h.logger.Warn("新密码与旧密码相同",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
		)
		response.BadRequest(c, "新密码不能与旧密码相同")
		return
	}

	// 设置新密码
	if err := currentUser.SetPassword(req.NewPassword); err != nil {
		h.logger.Error("密码加密失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to set new password")
		return
	}

	// 更新用户（通过 userService）
	if err := h.userService.UpdateUser(c, currentUser.ID, currentUser); err != nil {
		h.logger.Error("更新密码失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to update password")
		return
	}

	// 密码修改成功后，撤销用户的所有登录会话，强制重新登录
	if err := h.authService.RevokeAllUserSessions(c, currentUser.ID); err != nil {
		h.logger.Warn("撤销用户所有会话失败（密码已修改成功）",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		// 撤销会话失败不影响密码修改成功的结果，只记录警告
	}

	h.logger.Info("用户密码修改成功，所有会话已撤销",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	response.Success(c, gin.H{
		"message": "Password changed successfully. Please login again.",
	})
}
