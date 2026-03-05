package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/yuhang1130/gin-server/pkg/redis"
	"github.com/yuhang1130/gin-server/pkg/utils"
)

const (
	// 缓存键前缀
	authEmailCachePrefix = "auth:email:verifyCode:"
	blacklistPrefix      = "blacklist:"

	// 默认过期时间
	defaultVerifyCodeTTL = 5 * time.Minute
)

var (
	ErrCacheMiss = errors.New("cache: key not found")
)

// AuthCache 认证缓存操作
type AuthCache struct {
	redis *redis.Redis
}

// NewAuthCache 创建认证缓存实例
func NewAuthCache(redis *redis.Redis) *AuthCache {
	return &AuthCache{
		redis: redis,
	}
}

// GenerateVerifyCode 生成邮件验证码并缓存
// timeout 为 0 时使用默认过期时间（5分钟）
func (c *AuthCache) GenerateVerifyCode(ctx context.Context, email string, timeout time.Duration) (int, error) {
	key := fmt.Sprintf("%s%s", authEmailCachePrefix, email)

	// 检查是否已存在验证码
	existedCode, err := c.redis.Client.Get(ctx, key).Result()
	if err == nil && existedCode != "" {
		// 已存在验证码，转换并返回
		var code int
		fmt.Sscanf(existedCode, "%d", &code)
		return code, nil
	}

	// 生成新的验证码
	code := utils.Random(100000, 999999)
	ttl := timeout
	if ttl == 0 {
		ttl = defaultVerifyCodeTTL
	}

	// 缓存验证码
	err = c.redis.Client.Set(ctx, key, fmt.Sprintf("%d", code), ttl).Err()
	if err != nil {
		return 0, err
	}

	return code, nil
}

// ValidateVerifyCode 校验邮件验证码
func (c *AuthCache) ValidateVerifyCode(ctx context.Context, email string, code int, del bool) bool {
	key := fmt.Sprintf("%s%s", authEmailCachePrefix, email)
	cacheCode, err := c.redis.Client.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if cacheCode == fmt.Sprintf("%d", code) {
		if del {
			c.redis.Client.Del(ctx, key)
		}
		return true
	}
	return false
}

// AddToBlacklistByTokenString 将token加入黑名单（使用token字符串的hash）
func (c *AuthCache) AddToBlacklistByTokenString(ctx context.Context, tokenString string, ttl time.Duration) error {
	tokenKey := c.generateAccessTokenKeyByHash(tokenString)
	return c.redis.Client.Set(ctx, tokenKey, "true", ttl).Err()
}

// IsInBlacklistByTokenString 检查token是否在黑名单中（使用token字符串的hash）
func (c *AuthCache) IsInBlacklistByTokenString(ctx context.Context, tokenString string) (bool, error) {
	tokenKey := c.generateAccessTokenKeyByHash(tokenString)
	result, err := c.redis.Client.Get(ctx, tokenKey).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// AddToBlacklistByTokenID 将token加入黑名单（使用tokenID）
func (c *AuthCache) AddToBlacklistByTokenID(ctx context.Context, tokenString string, ttl time.Duration) error {
	tokenID := utils.GenerateTokenIDFromToken(tokenString)
	tokenKey := blacklistPrefix + tokenID
	return c.redis.Client.Set(ctx, tokenKey, "true", ttl).Err()
}

// IsInBlacklistByTokenID 检查token是否在黑名单中（使用tokenID）
func (c *AuthCache) IsInBlacklistByTokenID(ctx context.Context, tokenString string) (bool, error) {
	tokenID := utils.GenerateTokenIDFromToken(tokenString)
	tokenKey := blacklistPrefix + tokenID
	result, err := c.redis.Client.Get(ctx, tokenKey).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// generateAccessTokenKeyByHash 生成访问令牌的键名（使用token字符串的hash）
func (c *AuthCache) generateAccessTokenKeyByHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return blacklistPrefix + hex.EncodeToString(hash[:16])
}
