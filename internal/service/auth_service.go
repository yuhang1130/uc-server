package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/dto"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/jwt"
	"github.com/yuhang1130/gin-server/internal/repository"
	"go.uber.org/zap"
)

const (
	// 用户名限流配置
	usernameRateLimitMaxAttempts  = 5                // 最大尝试次数
	usernameRateLimitLockDuration = 30 * time.Minute // 锁定时长
	usernameRateLimitWindowSize   = 1 * time.Minute  // 时间窗口
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx *gin.Context, username, password string) (*dto.LoginResponse, error)
	GetTokenExpiry(tokenString string) (time.Time, error)
	RevokeAllUserSessions(ctx *gin.Context, userID uint64) error
	GetUserSessionData(ctx *gin.Context, userID, tenantID uint64, claimsID string) (*cache.UserSessionData, error)
}

// authService 认证服务实现
type authService struct {
	userRepo       repository.UserRepository
	tenantUserRepo repository.TenantUserRepository
	jwtUtil        *jwt.JWTUtil
	authCache      *cache.AuthCache
	userCache      *cache.UserCache
	logger         *zap.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	userRepo repository.UserRepository,
	tenantUserRepo repository.TenantUserRepository,
	jwtUtil *jwt.JWTUtil,
	authCache *cache.AuthCache,
	userCache *cache.UserCache,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		tenantUserRepo: tenantUserRepo,
		jwtUtil:        jwtUtil,
		authCache:      authCache,
		userCache:      userCache,
		logger:         logger,
	}
}

// Login 用户登录
func (s *authService) Login(ctx *gin.Context, username, password string) (*dto.LoginResponse, error) {
	// 第一步：检查用户名限流
	if err := s.checkUsernameRateLimit(ctx, username); err != nil {
		return nil, err
	}

	// 第二步：查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// 尝试使用邮箱登录
		user, err = s.userRepo.FindByEmail(username)
		if err != nil {
			s.logger.Warn("Login failed for user. user not found", zap.String("username", username))
			// 记录失败次数
			s.recordLoginFailure(ctx, username)
			return nil, errors.New("invalid credentials")
		}
	}

	// 第三步：验证密码
	if !user.CheckPassword(password) {
		s.logger.Warn("Login failed for user. invalid password", zap.String("username", username))
		// 记录失败次数
		s.recordLoginFailure(ctx, username)
		return nil, errors.New("invalid credentials")
	}

	s.logger.Info("用户认证成功",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("client_ip", ctx.ClientIP()),
	)

	// 第四步：生成登录响应
	var loginResp *dto.LoginResponse
	if model.UserRoleAdminSystem == user.Role {
		loginResp, err = s.handleSystemAdminLogin(ctx, user)
	} else {
		loginResp, err = s.handleTenantUserLogin(ctx, user)
	}

	if err != nil {
		// 登录失败，记录失败次数
		s.recordLoginFailure(ctx, username)
		return nil, err
	}

	// 第五步：登录成功，清除限流记录
	s.clearLoginAttempts(ctx, username)

	return loginResp, nil
}

// handleSystemAdminLogin 处理系统管理员登录
func (s *authService) handleSystemAdminLogin(ctx *gin.Context, user *model.User) (*dto.LoginResponse, error) {
	claimsID, accessToken, expiresAt, err := s.jwtUtil.GenerateAccessToken(user.ID, 1, user.Role)
	if err != nil {
		s.logger.Error("Failed to generate access token: ", zap.Error(err))
		return nil, errors.New("failed to generate access token")
	}

	loginResp := &dto.LoginResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		User:                 s.buildUserResponse(user),
		CurrentTenant:        dto.Tenant{},
		Tenants:              []dto.Tenant{},
		IsGlobalAdmin:        true,
	}

	// 预热用户缓存
	s.WarmUpUserCache(ctx, claimsID, &cache.UserSessionData{
		UserID:        user.ID,
		User:          s.buildUserResponse(user),
		TenantID:      1,
		TenantName:    "",
		Role:          string(user.Role),
		Tenants:       []dto.Tenant{},
		IsGlobalAdmin: true,
		IsProxy:       false,
		ProxyUserID:   0,
	})

	return loginResp, nil
}

// handleTenantUserLogin 处理普通租户用户登录
func (s *authService) handleTenantUserLogin(ctx *gin.Context, user *model.User) (*dto.LoginResponse, error) {
	// 查询用户所属的所有租户
	tenantUsers, err := s.tenantUserRepo.FindUserTenantsWithTenant(user.ID)
	if err != nil {
		s.logger.Warn("User has no active tenants",
			zap.Uint64("user_id", user.ID),
			zap.Error(err))
		return nil, errors.New("user has no active tenants")
	}

	// 构建租户列表响应
	tenants := s.buildTenantList(tenantUsers)

	// 选择默认登录租户
	selectedTenant := s.selectDefaultTenant(tenantUsers)

	// 更新该租户的最后登录时间
	if err := s.tenantUserRepo.UpdateLastLoginAt(selectedTenant.TenantID, user.ID); err != nil {
		s.logger.Warn("Failed to update last login time",
			zap.Uint64("tenant_id", selectedTenant.TenantID),
			zap.Uint64("user_id", user.ID),
			zap.Error(err))
	}

	// 生成token
	claimsID, accessToken, expiresAt, err := s.jwtUtil.GenerateAccessToken(user.ID, selectedTenant.TenantID, user.Role)
	if err != nil {
		s.logger.Error("Failed to generate access token: ", zap.Error(err))
		return nil, errors.New("failed to generate access token")
	}

	loginResp := &dto.LoginResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		User:                 s.buildUserResponse(user),
		CurrentTenant: dto.Tenant{
			TenantID:   selectedTenant.TenantID,
			TenantName: selectedTenant.Tenant.TenantName,
			Role:       model.UserRoles(selectedTenant.Role),
		},
		Tenants:       tenants,
		IsGlobalAdmin: false,
	}

	// 预热用户缓存
	s.WarmUpUserCache(ctx, claimsID, &cache.UserSessionData{
		UserID: user.ID,
		User:   s.buildUserResponse(user),

		TenantID:      selectedTenant.TenantID,
		TenantName:    selectedTenant.Tenant.TenantName,
		Role:          selectedTenant.Role,
		Tenants:       tenants,
		IsGlobalAdmin: false,
		IsProxy:       false,
		ProxyUserID:   0,
	})

	return loginResp, nil
}

// buildUserResponse 构建用户响应数据
func (s *authService) buildUserResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// buildTenantList 构建租户列表
func (s *authService) buildTenantList(tenantUsers []*repository.TenantUserWithTenant) []dto.Tenant {
	tenants := make([]dto.Tenant, 0, len(tenantUsers))
	for _, tu := range tenantUsers {
		tenants = append(tenants, dto.Tenant{
			TenantID:   tu.TenantID,
			TenantName: tu.Tenant.TenantName,
			Role:       model.UserRoles(tu.Role),
		})
	}
	return tenants
}

// WarmUpUserCache 预热用户缓存
func (s *authService) WarmUpUserCache(
	ctx *gin.Context,
	claimsID string,
	sessionData *cache.UserSessionData,
) {
	if s.userCache == nil {
		return
	}

	if err := s.userCache.SetUserSession(ctx, sessionData.UserID, claimsID, sessionData); err != nil {
		s.logger.Warn("Failed to warm up user cache after login",
			zap.Error(err),
			zap.Uint64("userID", sessionData.UserID))
	}
}

// GetTokenExpiry 获取token过期时间
func (s *authService) GetTokenExpiry(tokenString string) (time.Time, error) {
	// 解析token
	claims, err := s.jwtUtil.ParseAccessToken(tokenString)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid token: %w", err)
	}

	// 返回过期时间
	if claims.ExpiresAt != nil {
		return claims.ExpiresAt.Time, nil
	}

	return time.Time{}, errors.New("token has no expiry time")
}

// RevokeAllUserSessions 撤销用户的所有登录会话
// 使用场景：
// 1. 用户修改密码后，强制所有设备重新登录
// 2. 账号被禁用或删除时
// 3. 检测到账号异常活动时
// 4. 用户主动选择"登出所有设备"时
func (s *authService) RevokeAllUserSessions(ctx *gin.Context, userID uint64) error {
	if s.userCache == nil {
		s.logger.Warn("User cache is not available, skip revoking sessions",
			zap.Uint64("userID", userID))
		return nil
	}

	err := s.userCache.DeleteAllUserSessions(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to revoke all user sessions",
			zap.Uint64("userID", userID),
			zap.Error(err))
		return err
	}

	s.logger.Info("Successfully revoked all user sessions",
		zap.Uint64("userID", userID),
		zap.String("client_ip", ctx.ClientIP()))

	return nil
}

// checkUsernameRateLimit 检查用户名是否被限流
// 返回 error 如果用户被锁定或超过限制
func (s *authService) checkUsernameRateLimit(ctx *gin.Context, username string) error {
	if s.userCache == nil || username == "" {
		return nil
	}

	// 检查用户是否被锁定
	locked, ttl, err := s.userCache.IsAccountLocked(ctx, username)
	if err != nil {
		// Redis 错误，允许通过（fail-open）
		s.logger.Warn("Failed to check username rate limit",
			zap.String("username", username),
			zap.Error(err))
		return nil
	}

	if locked {
		s.logger.Warn("Username is locked due to too many login attempts",
			zap.String("username", username),
			zap.Int("ttl_seconds", int(ttl.Seconds())))
		return fmt.Errorf("用户 '%s' 已被锁定，请在 %d 秒后重试", username, int(ttl.Seconds()))
	}

	// 获取尝试次数
	attempts, err := s.userCache.GetLoginAttempts(ctx, username)
	if err != nil {
		// Redis 错误，允许通过
		s.logger.Warn("Failed to get login attempts",
			zap.String("username", username),
			zap.Error(err))
		return nil
	}

	// 检查是否超过限制（默认5次）
	if attempts >= uint64(usernameRateLimitMaxAttempts) {
		// 锁定用户（默认15分钟）
		if err := s.userCache.LockAccount(ctx, username, usernameRateLimitLockDuration); err != nil {
			s.logger.Warn("Failed to lock account",
				zap.String("username", username),
				zap.Error(err))
		}

		s.logger.Warn("Username locked due to too many login attempts",
			zap.String("username", username),
			zap.Uint64("attempts", attempts))

		return fmt.Errorf("用户 '%s' 登录尝试次数过多，已被锁定 %d 分钟", username, int(usernameRateLimitLockDuration.Minutes()))
	}

	return nil
}

// recordLoginFailure 记录登录失败
func (s *authService) recordLoginFailure(ctx *gin.Context, username string) {
	if s.userCache == nil || username == "" {
		return
	}

	newAttempts, err := s.userCache.IncrementLoginAttempts(ctx, username, usernameRateLimitWindowSize)
	if err != nil {
		s.logger.Warn("Failed to increment login attempts",
			zap.String("username", username),
			zap.Error(err))
		return
	}

	s.logger.Info("Login failure recorded",
		zap.String("username", username),
		zap.Uint64("attempts", newAttempts))
}

// clearLoginAttempts 清除登录尝试记录（登录成功时调用）
func (s *authService) clearLoginAttempts(ctx *gin.Context, username string) {
	if s.userCache == nil || username == "" {
		return
	}

	if err := s.userCache.UnlockAccount(ctx, username); err != nil {
		s.logger.Warn("Failed to clear login attempts",
			zap.String("username", username),
			zap.Error(err))
	}
}

// selectDefaultTenant 选择用户默认登录的租户
// 选择逻辑：
// 1. 优先选择 tenant_admin 角色的租户
// 2. 如果只有普通用户租户，选择最近活跃的租户（按 last_login_at 降序）
// 3. 如果从未登录过任何租户，选择最早加入的租户（按 created_at 升序）
func (s *authService) selectDefaultTenant(tenantUsers []*repository.TenantUserWithTenant) *repository.TenantUserWithTenant {
	if len(tenantUsers) == 0 {
		return nil
	}

	// 如果只有一个租户，直接返回
	if len(tenantUsers) == 1 {
		return tenantUsers[0]
	}

	var (
		adminTenants   []*repository.TenantUserWithTenant
		regularTenants []*repository.TenantUserWithTenant
	)

	// 分类租户：管理员租户 vs 普通用户租户
	for _, tu := range tenantUsers {
		if tu.Role == string(model.UserRoleAdminTenant) {
			adminTenants = append(adminTenants, tu)
		} else {
			regularTenants = append(regularTenants, tu)
		}
	}

	// 1. 优先选择管理员租户
	if len(adminTenants) > 0 {
		// 如果有多个管理员租户，选择最近活跃的
		return s.selectMostRecentTenant(adminTenants)
	}

	// 2. 选择普通用户租户中最近活跃的
	return s.selectMostRecentTenant(regularTenants)
}

// selectMostRecentTenant 从租户列表中选择最近活跃的租户
// 如果有登录记录，按 last_login_at 降序选择
// 如果都没有登录记录，按 created_at 升序选择（最早加入的）
func (s *authService) selectMostRecentTenant(tenants []*repository.TenantUserWithTenant) *repository.TenantUserWithTenant {
	if len(tenants) == 0 {
		return nil
	}

	var (
		hasLoginRecord []*repository.TenantUserWithTenant
		noLoginRecord  []*repository.TenantUserWithTenant
		zeroTime       = int64(0)
	)

	// 分类：有登录记录 vs 无登录记录
	for _, tu := range tenants {
		if tu.LastLoginAt == zeroTime {
			noLoginRecord = append(noLoginRecord, tu)
		} else {
			hasLoginRecord = append(hasLoginRecord, tu)
		}
	}

	// 如果有登录记录的租户，选择最近登录的
	if len(hasLoginRecord) > 0 {
		mostRecent := hasLoginRecord[0]
		for _, tu := range hasLoginRecord[1:] {
			if tu.LastLoginAt > mostRecent.LastLoginAt {
				mostRecent = tu
			}
		}
		return mostRecent
	}

	// 如果都没有登录记录，选择最早加入的租户
	earliest := noLoginRecord[0]
	for _, tu := range noLoginRecord[1:] {
		if tu.CreatedAt < earliest.CreatedAt {
			earliest = tu
		}
	}
	return earliest
}

// GetUserSessionData 获取用户会话数据
func (s *authService) GetUserSessionData(ctx *gin.Context, userID, tenantID uint64, claimsID string) (*cache.UserSessionData, error) {
	// 首先尝试从缓存获取
	sessionData, err := s.getCachedUserSessionData(ctx, userID, claimsID)
	if err == nil && sessionData != nil {
		return sessionData, nil
	}

	// 缓存未命中，从数据库加载
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 获取用户租户信息
	tenantUsers, err := s.tenantUserRepo.FindUserTenantsWithTenant(userID)
	if err != nil {
		s.logger.Warn("Failed to get user tenants", zap.Uint64("user_id", userID), zap.Error(err))
		// 继续执行，可能用户是系统管理员
	}

	// 构建租户列表
	var tenants []dto.Tenant
	var tenantName string
	var role string
	var isGlobalAdmin = false

	if tenantUsers != nil {
		tenants = s.buildTenantList(tenantUsers)

		// 查找当前租户的信息
		for _, tu := range tenantUsers {
			if tu.TenantID == tenantID {
				tenantName = tu.Tenant.TenantName
				role = tu.Role
				break
			}
		}
	}

	// 检查是否为系统管理员
	if user.Role == model.UserRoleAdminSystem {
		isGlobalAdmin = true
		role = string(user.Role)
	}
	user.Role = model.UserRoles(role)

	sessionData = &cache.UserSessionData{
		UserID:        userID,
		User:          s.buildUserResponse(user),
		TenantID:      tenantID,
		TenantName:    tenantName,
		Role:          role,
		Tenants:       tenants,
		IsGlobalAdmin: isGlobalAdmin,
		IsProxy:       false,
		ProxyUserID:   0,
	}

	// 异步缓存用户数据
	go func() {
		if err := s.userCache.SetUserSession(context.Background(), userID, claimsID, sessionData); err != nil {
			s.logger.Warn("Failed to cache user session data",
				zap.Uint64("user_id", userID),
				zap.Error(err))
		}
	}()

	return sessionData, nil
}

// getCachedUserSessionData 从缓存获取用户会话数据
func (s *authService) getCachedUserSessionData(ctx *gin.Context, userID uint64, claimsID string) (*cache.UserSessionData, error) {
	if s.userCache == nil {
		return nil, errors.New("user cache not available")
	}

	// 直接获取 UserSessionData 类型，无需类型转换
	sessionData, err := s.userCache.GetUserSession(ctx, userID, claimsID)
	if err != nil {
		return nil, err
	}

	return sessionData, nil
}
