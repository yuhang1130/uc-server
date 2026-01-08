package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/yuhang1130/gin-server/internal/dto"
	"github.com/yuhang1130/gin-server/internal/model"
)

const (
	// 缓存键前缀
	userCachePrefix        = "user:"
	userEmailCachePrefix   = "user:email:"
	userTokenCachePrefix   = "user:token:"
	userSessionPrefix      = "user:session:"
	userSessionIndexPrefix = "user:session:index:" // 用户会话索引（维护用户的所有 claimsID）

	userClientIPLoginLockedPrefix   = "user:login:ip:locked:"          // IP 登录锁定
	userClientIPLoginAttemptsPrefix = "user:login:client_ip:attempts:" // IP 登录尝试次数

	userLoginLockedPrefix   = "user:login:locked:"   // 用户名登录锁定
	userLoginAttemptsPrefix = "user:login:attempts:" // 用户名登录尝试次数

	// 默认过期时间
	defaultUserCacheTTL          = 30 * time.Minute
	defaultSessionTTL            = 24 * time.Hour
	defaultTokenTTL              = 15 * time.Minute
	defaultClientIPLoginTL       = 10 * time.Minute // 默认时间窗户
	defaultClientIPLoginLockedTL = 15 * time.Minute // 默认锁定时间
)

// UserSessionData 用户会话数据结构，用于缓存用户登录会话
type UserSessionData struct {
	UserID        uint64            `json:"user_id"`
	User          *dto.UserResponse `json:"user"`
	TenantID      uint64            `json:"tenant_id"`
	TenantName    string            `json:"tenant_name"`
	Role          string            `json:"role"`
	Tenants       []dto.Tenant      `json:"tenants"`
	IsGlobalAdmin bool              `json:"is_global_admin"`
	IsProxy       bool              `json:"is_proxy"`      // 是否是代理token
	ProxyUserID   uint64            `json:"proxy_user_id"` // 被代理用户ID
}

// UserCache 用户缓存操作
type UserCache struct {
	redis *Redis
}

// NewUserCache 创建用户缓存实例
func NewUserCache(redis *Redis) *UserCache {
	return &UserCache{redis: redis}
}

// SetUser 缓存用户信息（按 ID）
func (uc *UserCache) SetUser(ctx context.Context, user *model.User) error {
	key := fmt.Sprintf("%s%d", userCachePrefix, user.ID)
	return uc.redis.SetJSON(ctx, key, user, defaultUserCacheTTL)
}

// GetUser 获取用户信息（按 ID）- uint64 版本
func (uc *UserCache) GetUser(ctx context.Context, userID uint64) (*model.User, error) {
	key := fmt.Sprintf("%s%d", userCachePrefix, userID)
	var user model.User
	err := uc.redis.GetJSON(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 获取用户信息（按 ID）- uint64 版本
func (uc *UserCache) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	key := fmt.Sprintf("%s%d", userCachePrefix, userID)
	var user model.User
	err := uc.redis.GetJSON(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUserByEmail 缓存用户信息（按 Email）
func (uc *UserCache) SetUserByEmail(ctx context.Context, email string, user *model.User) error {
	key := fmt.Sprintf("%s%s", userEmailCachePrefix, email)
	return uc.redis.SetJSON(ctx, key, user, defaultUserCacheTTL)
}

// GetUserByEmail 获取用户信息（按 Email）
func (uc *UserCache) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	key := fmt.Sprintf("%s%s", userEmailCachePrefix, email)
	var user model.User
	err := uc.redis.GetJSON(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser 删除用户缓存
func (uc *UserCache) DeleteUser(ctx context.Context, userID uint64, email string) error {
	keys := []string{
		fmt.Sprintf("%s%d", userCachePrefix, userID),
		fmt.Sprintf("%s%s", userEmailCachePrefix, email),
	}
	return uc.redis.Delete(ctx, keys...)
}

// SetUserToken 缓存用户 Token（如验证码、重置密码 token 等）
func (uc *UserCache) SetUserToken(ctx context.Context, tokenType string, identifier string, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	if ttl == 0 {
		ttl = defaultTokenTTL
	}
	return uc.redis.Set(ctx, key, token, ttl)
}

// GetUserToken 获取用户 Token
func (uc *UserCache) GetUserToken(ctx context.Context, tokenType string, identifier string) (string, error) {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	return uc.redis.Get(ctx, key)
}

// DeleteUserToken 删除用户 Token
func (uc *UserCache) DeleteUserToken(ctx context.Context, tokenType string, identifier string) error {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	return uc.redis.Delete(ctx, key)
}

// SetUserSession 设置用户会话（用于记录登录状态）
func (uc *UserCache) SetUserSession(ctx context.Context, userID uint64, sessionID string, data *UserSessionData) error {
	// 设置会话数据
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	if err := uc.redis.SetJSON(ctx, key, data, defaultSessionTTL); err != nil {
		return err
	}

	// 将 sessionID 添加到用户的会话索引集合中
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	if err := uc.redis.Client.SAdd(ctx, indexKey, sessionID).Err(); err != nil {
		return err
	}

	// 为索引 key 设置过期时间（比会话稍长，确保能清理干净）
	return uc.redis.Client.Expire(ctx, indexKey, defaultSessionTTL+1*time.Hour).Err()
}

// GetUserSession 获取用户会话
func (uc *UserCache) GetUserSession(ctx context.Context, userID uint64, sessionID string) (*UserSessionData, error) {
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	var session UserSessionData
	err := uc.redis.GetJSON(ctx, key, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteUserSession 删除用户会话
func (uc *UserCache) DeleteUserSession(ctx context.Context, userID uint64, sessionID string) error {
	// 删除会话数据
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	if err := uc.redis.Delete(ctx, key); err != nil {
		return err
	}

	// 从用户的会话索引集合中移除 sessionID
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	return uc.redis.Client.SRem(ctx, indexKey, sessionID).Err()
}

// DeleteAllUserSessions 删除用户的所有会话（用于修改密码、账号禁用等场景）
func (uc *UserCache) DeleteAllUserSessions(ctx context.Context, userID uint64) error {
	// 获取用户的所有 sessionID
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	sessionIDs, err := uc.redis.Client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return err
	}

	// 如果没有会话，直接返回
	if len(sessionIDs) == 0 {
		return nil
	}

	// 构建所有会话 key
	keys := make([]string, 0, len(sessionIDs)+1)
	for _, sessionID := range sessionIDs {
		key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
		keys = append(keys, key)
	}
	// 同时删除索引 key
	keys = append(keys, indexKey)

	// 批量删除所有会话
	return uc.redis.Delete(ctx, keys...)
}

// GetAllUserSessionIDs 获取用户的所有会话ID（用于管理和监控）
func (uc *UserCache) GetAllUserSessionIDs(ctx context.Context, userID uint64) ([]string, error) {
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	return uc.redis.Client.SMembers(ctx, indexKey).Result()
}

// IncrementClientIPLoginAttempts 增加ClientIP登录失败次数
func (uc *UserCache) IncrementClientIPLoginAttempts(ctx context.Context, clientIP string, duration time.Duration) (uint64, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	count, err := uc.redis.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if duration == 0 {
		duration = defaultClientIPLoginTL
	}
	_ = uc.redis.Expire(ctx, key, duration)
	return uint64(count), nil
}

// GetClientIPLoginAttempts 获取ClientIP登录失败次数
func (uc *UserCache) GetClientIPLoginAttempts(ctx context.Context, clientIP string) (uint64, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	val, err := uc.redis.Get(ctx, key)
	if err == ErrCacheMiss {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var count uint64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

// ResetClientIPLoginAttempts 重置ClientIP登录失败次数
func (uc *UserCache) ResetClientIPLoginAttempts(ctx context.Context, clientIP string) error {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	return uc.redis.Delete(ctx, key)
}

// LockClientIP 锁定ClientIP（登录尝试次数过多时）
func (uc *UserCache) LockClientIP(ctx context.Context, clientIP string, duration time.Duration) error {
	key := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	if duration == 0 {
		duration = defaultClientIPLoginLockedTL
	}
	return uc.redis.Set(ctx, key, "true", duration)
}

// IsClientIPLocked 检查ClientIP是否被锁定
func (uc *UserCache) IsClientIPLocked(ctx context.Context, clientIP string) (bool, time.Duration, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	val, err := uc.redis.Get(ctx, key)
	if err == ErrCacheMiss {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	// 获取剩余锁定时间
	ttl, err := uc.redis.Client.TTL(ctx, key).Result()
	if err != nil {
		return val == "true", 0, err
	}

	return val == "true", ttl, nil
}

// UnlockClientIP 解锁ClientIP
func (uc *UserCache) UnlockClientIP(ctx context.Context, clientIP string) error {
	lockKey := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	attemptKey := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	return uc.redis.Delete(ctx, lockKey, attemptKey)
}

// IncrementLoginAttempts 增加登录失败次数（用于用户名限流）
func (uc *UserCache) IncrementLoginAttempts(ctx context.Context, username string, duration time.Duration) (uint64, error) {
	key := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	count, err := uc.redis.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	// 设置过期时间为 15 分钟
	_ = uc.redis.Expire(ctx, key, duration)
	return uint64(count), nil
}

// GetLoginAttempts 获取登录失败次数（用于用户名限流）
func (uc *UserCache) GetLoginAttempts(ctx context.Context, username string) (uint64, error) {
	key := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	val, err := uc.redis.Get(ctx, key)
	if err == ErrCacheMiss {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var count uint64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

// LockAccount 锁定账户（用于用户名限流）
func (uc *UserCache) LockAccount(ctx context.Context, username string, duration time.Duration) error {
	key := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	return uc.redis.Set(ctx, key, "true", duration)
}

// IsAccountLocked 检查账户是否被锁定（用于用户名限流）
func (uc *UserCache) IsAccountLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	key := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	val, err := uc.redis.Get(ctx, key)
	if err == ErrCacheMiss {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	// 获取剩余锁定时间
	ttl, err := uc.redis.Client.TTL(ctx, key).Result()
	if err != nil {
		return val == "true", 0, err
	}

	return val == "true", ttl, nil
}

// UnlockAccount 解锁账户（用于用户名限流）
func (uc *UserCache) UnlockAccount(ctx context.Context, username string) error {
	lockKey := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	attemptKey := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	return uc.redis.Delete(ctx, lockKey, attemptKey)
}
