package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/pkg/redis"
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

var (
	ErrCacheMiss = errors.New("cache: key not found")
)

// Cache 用户缓存操作
type UserCache struct {
	redis *redis.Redis
}

// NewUserCache 创建用户缓存实例
func NewUserCache(redis *redis.Redis) *UserCache {
	return &UserCache{redis: redis}
}

// SetUser 缓存用户信息（按 ID）
func (c *UserCache) SetUser(ctx context.Context, user *model.User) error {
	key := fmt.Sprintf("%s%d", userCachePrefix, user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.redis.Client.Set(ctx, key, data, defaultUserCacheTTL).Err()
}

// GetUser 获取用户信息（按 ID）
func (c *UserCache) GetUser(ctx context.Context, userID uint64) (*model.User, error) {
	key := fmt.Sprintf("%s%d", userCachePrefix, userID)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 获取用户信息（按 ID）- uint64 版本
func (c *UserCache) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	return c.GetUser(ctx, userID)
}

// SetUserByEmail 缓存用户信息（按 Email）
func (c *UserCache) SetUserByEmail(ctx context.Context, email string, user *model.User) error {
	key := fmt.Sprintf("%s%s", userEmailCachePrefix, email)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.redis.Client.Set(ctx, key, data, defaultUserCacheTTL).Err()
}

// GetUserByEmail 获取用户信息（按 Email）
func (c *UserCache) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	key := fmt.Sprintf("%s%s", userEmailCachePrefix, email)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser 删除用户缓存
func (c *UserCache) DeleteUser(ctx context.Context, userID uint64, email string) error {
	keys := []string{
		fmt.Sprintf("%s%d", userCachePrefix, userID),
		fmt.Sprintf("%s%s", userEmailCachePrefix, email),
	}
	return c.redis.Client.Del(ctx, keys...).Err()
}

// SetUserToken 缓存用户 Token（如验证码、重置密码 token 等）
func (c *UserCache) SetUserToken(ctx context.Context, tokenType string, identifier string, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	if ttl == 0 {
		ttl = defaultTokenTTL
	}
	return c.redis.Client.Set(ctx, key, token, ttl).Err()
}

// GetUserToken 获取用户 Token
func (c *UserCache) GetUserToken(ctx context.Context, tokenType string, identifier string) (string, error) {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", ErrCacheMiss
	}
	return val, err
}

// DeleteUserToken 删除用户 Token
func (c *UserCache) DeleteUserToken(ctx context.Context, tokenType string, identifier string) error {
	key := fmt.Sprintf("%s%s:%s", userTokenCachePrefix, tokenType, identifier)
	return c.redis.Client.Del(ctx, key).Err()
}

// SetUserSession 设置用户会话（用于记录登录状态）
func (c *UserCache) SetUserSession(ctx context.Context, userID uint64, sessionID string, data *model.UserSessionData) error {
	// 设置会话数据
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := c.redis.Client.Set(ctx, key, jsonData, defaultSessionTTL).Err(); err != nil {
		return err
	}

	// 将 sessionID 添加到用户的会话索引集合中
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	if err := c.redis.Client.SAdd(ctx, indexKey, sessionID).Err(); err != nil {
		return err
	}

	// 为索引 key 设置过期时间（比会话稍长，确保能清理干净）
	return c.redis.Client.Expire(ctx, indexKey, defaultSessionTTL+1*time.Hour).Err()
}

// GetUserSession 获取用户会话
func (c *UserCache) GetUserSession(ctx context.Context, userID uint64, sessionID string) (*model.UserSessionData, error) {
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	var session model.UserSessionData
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteUserSession 删除用户会话
func (c *UserCache) DeleteUserSession(ctx context.Context, userID uint64, sessionID string) error {
	// 删除会话数据
	key := fmt.Sprintf("%s%d:%s", userSessionPrefix, userID, sessionID)
	if err := c.redis.Client.Del(ctx, key).Err(); err != nil {
		return err
	}

	// 从用户的会话索引集合中移除 sessionID
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	return c.redis.Client.SRem(ctx, indexKey, sessionID).Err()
}

// DeleteAllUserSessions 删除用户的所有会话（用于修改密码、账号禁用等场景）
func (c *UserCache) DeleteAllUserSessions(ctx context.Context, userID uint64) error {
	// 获取用户的所有 sessionID
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	sessionIDs, err := c.redis.Client.SMembers(ctx, indexKey).Result()
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
	return c.redis.Client.Del(ctx, keys...).Err()
}

// GetAllUserSessionIDs 获取用户的所有会话ID（用于管理和监控）
func (c *UserCache) GetAllUserSessionIDs(ctx context.Context, userID uint64) ([]string, error) {
	indexKey := fmt.Sprintf("%s%d", userSessionIndexPrefix, userID)
	return c.redis.Client.SMembers(ctx, indexKey).Result()
}

// IncrementClientIPLoginAttempts 增加ClientIP登录失败次数
func (c *UserCache) IncrementClientIPLoginAttempts(ctx context.Context, clientIP string, duration time.Duration) (uint64, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	count, err := c.redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if duration == 0 {
		duration = defaultClientIPLoginTL
	}
	_ = c.redis.Client.Expire(ctx, key, duration)
	return uint64(count), nil
}

// GetClientIPLoginAttempts 获取ClientIP登录失败次数
func (c *UserCache) GetClientIPLoginAttempts(ctx context.Context, clientIP string) (uint64, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
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
func (c *UserCache) ResetClientIPLoginAttempts(ctx context.Context, clientIP string) error {
	key := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	return c.redis.Client.Del(ctx, key).Err()
}

// LockClientIP 锁定ClientIP（登录尝试次数过多时）
func (c *UserCache) LockClientIP(ctx context.Context, clientIP string, duration time.Duration) error {
	key := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	if duration == 0 {
		duration = defaultClientIPLoginLockedTL
	}
	return c.redis.Client.Set(ctx, key, "true", duration).Err()
}

// IsClientIPLocked 检查ClientIP是否被锁定
func (c *UserCache) IsClientIPLocked(ctx context.Context, clientIP string) (bool, time.Duration, error) {
	key := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	// 获取剩余锁定时间
	ttl, err := c.redis.Client.TTL(ctx, key).Result()
	if err != nil {
		return val == "true", 0, err
	}

	return val == "true", ttl, nil
}

// UnlockClientIP 解锁ClientIP
func (c *UserCache) UnlockClientIP(ctx context.Context, clientIP string) error {
	lockKey := fmt.Sprintf("%s%s", userClientIPLoginLockedPrefix, clientIP)
	attemptKey := fmt.Sprintf("%s%s", userClientIPLoginAttemptsPrefix, clientIP)
	return c.redis.Client.Del(ctx, lockKey, attemptKey).Err()
}

// IncrementLoginAttempts 增加登录失败次数（用于用户名限流）
func (c *UserCache) IncrementLoginAttempts(ctx context.Context, username string, duration time.Duration) (uint64, error) {
	key := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	count, err := c.redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// 设置过期时间
	_ = c.redis.Client.Expire(ctx, key, duration)
	return uint64(count), nil
}

// GetLoginAttempts 获取登录失败次数（用于用户名限流）
func (c *UserCache) GetLoginAttempts(ctx context.Context, username string) (uint64, error) {
	key := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
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
func (c *UserCache) LockAccount(ctx context.Context, username string, duration time.Duration) error {
	key := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	return c.redis.Client.Set(ctx, key, "true", duration).Err()
}

// IsAccountLocked 检查账户是否被锁定（用于用户名限流）
func (c *UserCache) IsAccountLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	key := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	val, err := c.redis.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	// 获取剩余锁定时间
	ttl, err := c.redis.Client.TTL(ctx, key).Result()
	if err != nil {
		return val == "true", 0, err
	}

	return val == "true", ttl, nil
}

// UnlockAccount 解锁账户（用于用户名限流）
func (c *UserCache) UnlockAccount(ctx context.Context, username string) error {
	lockKey := fmt.Sprintf("%s%s", userLoginLockedPrefix, username)
	attemptKey := fmt.Sprintf("%s%s", userLoginAttemptsPrefix, username)
	return c.redis.Client.Del(ctx, lockKey, attemptKey).Err()
}
