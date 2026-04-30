package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheManager struct {
	rdb *redis.Client
}

var GlobalCache *CacheManager

func InitCache(rdb *redis.Client) {
	GlobalCache = &CacheManager{rdb: rdb}
}

func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	if cm.rdb == nil {
		return fmt.Errorf("缓存未初始化")
	}

	data, err := cm.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("缓存未命中")
		}
		return fmt.Errorf("读取缓存失败: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("反序列化缓存失败: %w", err)
	}

	return nil
}

func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if cm.rdb == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	if err := cm.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("写入缓存失败: %w", err)
	}

	return nil
}

func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
	if cm.rdb == nil {
		return nil
	}

	return cm.rdb.Del(ctx, keys...).Err()
}

func (cm *CacheManager) DeleteByPattern(ctx context.Context, pattern string) error {
	if cm.rdb == nil {
		return nil
	}

	iter := cm.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("扫描缓存键失败: %w", err)
	}

	if len(keys) > 0 {
		return cm.rdb.Del(ctx, keys...).Err()
	}

	return nil
}

func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	err := cm.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	data, err := loader()
	if err != nil {
		return fmt.Errorf("加载数据失败: %w", err)
	}

	if err := cm.Set(ctx, key, data, ttl); err != nil {
		return fmt.Errorf("写入缓存失败: %w", err)
	}

	dataBytes, _ := json.Marshal(data)
	return json.Unmarshal(dataBytes, dest)
}

func BuildKey(parts ...string) string {
	key := "fayhub"
	for _, p := range parts {
		key += ":" + p
	}
	return key
}
