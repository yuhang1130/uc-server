package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrCacheMiss = errors.New("cache: key not found")
	ErrCacheSet  = errors.New("cache: failed to set key")
)

// Set 设置缓存（字符串）
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get 获取缓存（字符串）
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	return val, err
}

// SetJSON 设置缓存（JSON 对象）
func (r *Redis) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, data, expiration).Err()
}

// GetJSON 获取缓存（JSON 对象）
func (r *Redis) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete 删除缓存
func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists 检查缓存是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	val, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// Expire 设置过期时间
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Incr 自增
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// IncrBy 增加指定值
func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

// Decr 自减
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}

// DecrBy 减少指定值
func (r *Redis) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希表字段
func (r *Redis) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.Client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希表字段
func (r *Redis) HGet(ctx context.Context, key string, field string) (string, error) {
	val, err := r.Client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	return val, err
}

// HGetAll 获取哈希表所有字段
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表字段
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// SAdd 添加集合成员
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// SIsMember 判断是否是集合成员
func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

// SRem 移除集合成员
func (r *Redis) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

// ZAdd 添加有序集合成员
func (r *Redis) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return r.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合指定范围成员（按分数从小到大）
func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange 获取有序集合指定范围成员（按分数从大到小）
func (r *Redis) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRem 移除有序集合成员
func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.ZRem(ctx, key, members...).Err()
}

// LPush 左侧推入列表
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

// RPush 右侧推入列表
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

// LPop 左侧弹出列表
func (r *Redis) LPop(ctx context.Context, key string) (string, error) {
	val, err := r.Client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	return val, err
}

// RPop 右侧弹出列表
func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	val, err := r.Client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	return val, err
}

// LRange 获取列表指定范围元素
func (r *Redis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (r *Redis) LLen(ctx context.Context, key string) (int64, error) {
	return r.Client.LLen(ctx, key).Result()
}
