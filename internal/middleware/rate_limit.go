package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	MaxAttempts    int           // 最大尝试次数
	LockDuration   time.Duration // 锁定时长
	WindowDuration time.Duration // 时间窗口
}

// DefaultIPRateLimit IP 限流配置（防止单个 IP 暴力破解）
var DefaultIPRateLimit = RateLimitConfig{
	MaxAttempts:    10,
	LockDuration:   30 * time.Minute,
	WindowDuration: 1 * time.Minute,
}

// IPRateLimiter 基于 IP 的登录速率限制中间件
// 防止同一 IP 地址的暴力破解攻击
// 这是第一层防护，在业务逻辑之前执行
func IPRateLimiter(userCache *cache.UserCache, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查 IP 是否被锁定
		locked, ttl, err := userCache.IsClientIPLocked(c, clientIP)
		if err != nil {
			// Redis 错误，允许请求通过（fail-open）
			c.Next()
			return
		}

		if locked {
			response.TooManyRequests(c, fmt.Sprintf("IP已被锁定，请在 %d 秒后重试", int(ttl.Seconds())))
			c.Abort()
			return
		}

		// 获取当前 IP 尝试次数
		attempts, err := userCache.GetClientIPLoginAttempts(c, clientIP)
		if err != nil {
			// Redis 错误，允许请求通过
			c.Next()
			return
		}

		// 检查是否超过最大尝试次数
		if attempts >= uint64(config.MaxAttempts) {
			// 锁定 IP
			if err := userCache.LockClientIP(c, clientIP, config.LockDuration); err == nil {
				response.TooManyRequests(c, fmt.Sprintf("IP登录尝试次数过多，已被锁定 %d 分钟", int(config.LockDuration.Minutes())))
			} else {
				response.TooManyRequests(c, "IP登录尝试次数过多")
			}
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()

		// 请求处理完成后，根据响应状态更新计数器
		// 401 表示认证失败（用户名或密码错误）
		if c.Writer.Status() == 401 {
			// 增加 IP 失败次数
			newAttempts, _ := userCache.IncrementClientIPLoginAttempts(c, clientIP, config.WindowDuration)

			// 在响应头中添加剩余尝试次数信息
			remaining := config.MaxAttempts - int(newAttempts)
			if remaining > 0 {
				c.Header("X-RateLimit-IP-Remaining", strconv.Itoa(remaining))
			}
		} else if c.Writer.Status() == 200 {
			// 登录成功，清除 IP 失败计数和锁定状态
			userCache.UnlockClientIP(c, clientIP)
		}
	}
}

// APIRateLimiter 通用 API 速率限制中间件
// 基于 IP 地址限制请求频率
func APIRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("api_rate:%s", clientIP)

		// 获取当前请求数
		count, err := redisClient.Get(c, key).Int()
		if err != nil && err != redis.Nil {
			// Redis 错误，允许请求通过（fail-open）
			c.Next()
			return
		}

		// 检查是否超过限制
		if count >= maxRequests {
			ttl, _ := redisClient.TTL(c, key).Result()
			response.TooManyRequests(c, fmt.Sprintf("请求频率过高，请在 %d 秒后重试", int(ttl.Seconds())))
			c.Abort()
			return
		}

		// 增加计数器
		pipe := redisClient.Pipeline()
		pipe.Incr(c, key)
		if count == 0 {
			// 首次请求，设置过期时间
			pipe.Expire(c, key, window)
		}
		_, err = pipe.Exec(c)
		if err != nil {
			// Redis 错误，允许请求通过
			c.Next()
			return
		}

		// 在响应头中添加速率限制信息
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-count-1))

		c.Next()
	}
}
