package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/yuhang1130/gin-server/internal/pkg/utils"
	"go.uber.org/zap"
)

const (
	// 缓存键前缀
	authEmailCachePrefix = "auth:email:verifyCode:"

	// 默认过期时间
	defaultVerifyCodeTTL = 5 * time.Minute
)

// AuthCache 登录认证缓存操作
type AuthCache struct {
	redis  *Redis
	logger *zap.Logger
}

// NewUserCache 创建用户缓存实例
func NewAuthCache(redis *Redis, logger *zap.Logger) *AuthCache {
	return &AuthCache{
		redis:  redis,
		logger: logger,
	}
}

// GenerateVerifyCode 生成邮件验证码并缓存
// timeout 为 0 时使用默认过期时间（5分钟）
func (ac *AuthCache) GenerateVerifyCode(ctx context.Context, email string, timeout time.Duration) (int, error) {
	key := fmt.Sprintf("%s%s", authEmailCachePrefix, email)

	// 检查是否已存在验证码
	existedCode, err := ac.redis.Get(ctx, key)
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
	err = ac.redis.Set(ctx, key, fmt.Sprintf("%d", code), ttl)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// ValidateVerifyCode 校验邮件验证码
func (ac *AuthCache) ValidateVerifyCode(ctx context.Context, email string, code int, del bool) bool {
	key := fmt.Sprintf("%s%s", authEmailCachePrefix, email)
	cacheCode, err := ac.redis.Get(ctx, key)
	if err != nil {
		ac.logger.Info("获取邮件验证码失败", zap.Error(err))
		return false
	}
	if cacheCode == fmt.Sprintf("%d", code) {
		if del {
			ac.redis.Delete(ctx, key)
		}
		return true
	}
	return false
}
