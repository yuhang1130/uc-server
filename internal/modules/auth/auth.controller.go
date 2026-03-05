package auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/internal/modules/user"
	"github.com/yuhang1130/gin-server/pkg/logger"
	"github.com/yuhang1130/gin-server/pkg/response"
	"github.com/yuhang1130/gin-server/pkg/validation"
)

// AuthController 认证处理器
type AuthController struct {
	authService AuthService
	userService user.UserService
	authCache   *AuthCache
	logger      logger.Logger
}

// NewAuthController
func NewAuthController(
	authService AuthService,
	userService user.UserService,
	logger logger.Logger,
	authCache *AuthCache,
) *AuthController {
	return &AuthController{
		authService: authService,
		userService: userService,
		authCache:   authCache,
		logger:      logger,
	}
}

// Login 用户登录
func (a *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		a.logger.Warnw("登录请求参数无效", "client_ip", c.ClientIP(), "error", errMessage)
		response.ValidationErrorFunc(c, "Invalid request: "+errMessage)
		return
	}

	a.logger.Infow("收到登录请求", "username", req.Username, "client_ip", c.ClientIP())

	// 执行登录
	loginResp, err := a.authService.Login(c, req.Username, req.Password)
	if err != nil {
		a.logger.Warnw("登录失败", "username", req.Username, "client_ip", c.ClientIP(), "error", err.Error())
		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "invalid credentials") {
			response.AuthInvalidCredentialsErrorFunc(c, "用户名或密码错误")
		} else if strings.Contains(err.Error(), "user has no active tenants") {
			response.UserTenantMismatchFunc(c, "用户未分配任何租户")
		} else {
			response.AuthLoginFailedFunc(c, err.Error())
		}
		return
	}

	response.SuccessFunc(c, loginResp)
}

// Logout 用户登出
func (a *AuthController) Logout(c *gin.Context) {
	// 从上下文中获取当前用户
	userSessionData, exists := c.Get("userSessionData")
	if !exists {
		a.logger.Warnw("未认证用户尝试登出", "client_ip", c.ClientIP())
		response.UnauthorizedFunc(c, "User not authenticated")
		return
	}

	currentUser, ok := userSessionData.(*model.UserSessionData)
	if !ok {
		a.logger.Errorw("获取当前用户失败", "userSessionData", userSessionData, "client_ip", c.ClientIP())
		response.InternalServerErrorFunc(c, "Failed to get current user")
		return
	}

	a.logger.Infow("收到登出请求", "user_id", currentUser.UserID, "client_ip", c.ClientIP())

	// 获取 Access Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) <= 7 {
		a.logger.Warnw("Authorization header 无效", "user_id", currentUser.UserID)
		response.ValidationErrorFunc(c, "Authorization header missing or invalid")
		return
	}

	tokenString := authHeader[7:] // 移除 "Bearer " 前缀

	// 获取 Token 过期时间
	exp, err := a.authService.GetTokenExpiry(tokenString)
	if err != nil {
		a.logger.Errorw("获取 Token 过期时间失败", "user_id", currentUser.UserID, "error", err.Error())
		response.InternalServerErrorFunc(c, "Failed to get token expiry")
		return
	}

	// 计算剩余有效时间
	now := time.Now()
	remainingTime := exp.Sub(now)

	// 将 Access Token 加入黑名单
	if remainingTime > 0 {
		if err := a.authCache.AddToBlacklistByTokenString(c, tokenString, remainingTime); err != nil {
			a.logger.Errorw("将 Token 加入黑名单失败", "user_id", currentUser.UserID, "error", err.Error())
			response.CacheOperationFailedFunc(c, "Failed to blacklist token")
			return
		}
	}

	a.logger.Infow("用户登出成功", "user_id", currentUser.UserID, "client_ip", c.ClientIP())

	response.SuccessFunc(c, gin.H{
		"message": "Logged out successfully",
		"user_id": currentUser.UserID,
	})
}

// ChangePassword 修改密码
func (a *AuthController) ChangePassword(c *gin.Context) {
	// 从上下文获取当前用户
	userSessionData, exists := c.Get("userSessionData")
	if !exists {
		a.logger.Warnw("未认证用户尝试修改密码", "client_ip", c.ClientIP())
		response.UnauthorizedFunc(c, "User not authenticated")
		return
	}

	currentUser, ok := userSessionData.(*model.UserSessionData)
	if !ok {
		a.logger.Errorw("获取当前用户失败", "client_ip", c.ClientIP())
		response.InternalServerErrorFunc(c, "Failed to get current user")
		return
	}

	a.logger.Infow("收到修改密码请求",
		"user_id", currentUser.UserID,
		"username", currentUser.User.Username,
		"client_ip", c.ClientIP(),
	)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		a.logger.Warnw("修改密码请求参数无效",
			"user_id", currentUser.UserID,
			"error", errMessage,
		)
		response.ValidationErrorFunc(c, "Invalid request: "+errMessage)
		return
	}

	err := a.userService.ChangePassword(c, currentUser.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "password mismatch" {
			response.UserPasswordIncorrectFunc(c, "原密码错误")
		} else {
			response.InternalServerErrorFunc(c, err.Error())
		}
		return
	}

	// 密码修改成功后，撤销用户的所有登录会话，强制重新登录
	a.authService.RevokeAllUserSessions(c, currentUser.UserID)

	a.logger.Infow("用户密码修改成功，所有会话已撤销",
		"user_id", currentUser.UserID,
		"username", currentUser.User.Username,
		"client_ip", c.ClientIP(),
	)

	response.SuccessFunc(c, gin.H{
		"message": "Password changed successfully. Please login again.",
	})
}
