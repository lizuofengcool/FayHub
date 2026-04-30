package redisclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fayhub/pkg/config"
	"fayhub/pkg/utils"

	"github.com/redis/go-redis/v9"
)

var (
	client     *redis.Client
	enabled    bool = false
	fallbackDB *FallbackStore
	logger     func(format string, args ...interface{})
)

type FallbackEntry struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

type FallbackStore struct {
	data map[string]*FallbackEntry
	mu   sync.RWMutex
}

func init() {
	fallbackDB = &FallbackStore{
		data: make(map[string]*FallbackEntry),
	}
	go cleanupExpired()
}

func Init(cfg *config.Config, logFn func(string, ...interface{})) error {
	logger = logFn
	if !cfg.Redis.Enabled {
		logger("⚠️  Redis未启用，使用降级存储模式")
		enabled = false
		return nil
	}

	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger("⚠️  Redis连接失败(%s:%d): %v，使用降级存储模式", cfg.Redis.Host, cfg.Redis.Port, err)
		enabled = false
		client = nil
		return nil
	}

	enabled = true
	logger("✅ Redis连接成功 (%s:%d)", cfg.Redis.Host, cfg.Redis.Port)
	return nil
}

func IsEnabled() bool {
	return enabled
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if enabled && client != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return client.Set(ctx, key, data, expiration).Err()
	}
	return fallbackSet(key, value, expiration)
}

func Get(ctx context.Context, key string, dest interface{}) error {
	if enabled && client != nil {
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), dest)
	}
	return fallbackGet(key, dest)
}

func Del(ctx context.Context, keys ...string) error {
	if enabled && client != nil {
		return client.Del(ctx, keys...).Err()
	}
	for _, k := range keys {
		fallbackDel(k)
	}
	return nil
}

func Exists(ctx context.Context, keys ...string) (int64, error) {
	if enabled && client != nil {
		return client.Exists(ctx, keys...).Result()
	}
	count := int64(0)
	for _, k := range keys {
		if fallbackExists(k) {
			count++
		}
	}
	return count, nil
}

func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if enabled && client != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return false, err
		}
		return client.SetNX(ctx, key, data, expiration).Result()
	}
	return fallbackSetNX(key, value, expiration)
}

func Incr(ctx context.Context, key string) (int64, error) {
	if enabled && client != nil {
		return client.Incr(ctx, key).Result()
	}
	return fallbackIncr(key)
}

func Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if enabled && client != nil {
		return client.Expire(ctx, key, expiration).Result()
	}
	return fallbackExpire(key, expiration)
}

func TTL(ctx context.Context, key string) (time.Duration, error) {
	if enabled && client != nil {
		return client.TTL(ctx, key).Result()
	}
	return fallbackTTL(key)
}

func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

func GetRawClient() *redis.Client {
	return client
}

func fallbackSet(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	entry := &FallbackEntry{
		Value:     string(data),
		ExpiresAt: time.Now().Add(expiration),
	}
	fallbackDB.mu.Lock()
	fallbackDB.data[key] = entry
	fallbackDB.mu.Unlock()
	return nil
}

func fallbackGet(key string, dest interface{}) error {
	fallbackDB.mu.RLock()
	entry, ok := fallbackDB.data[key]
	fallbackDB.mu.RUnlock()

	if !ok || entry == nil {
		return utils.ErrRedisKeyNotFound
	}
	if time.Now().After(entry.ExpiresAt) {
		fallbackDB.mu.Lock()
		delete(fallbackDB.data, key)
		fallbackDB.mu.Unlock()
		return utils.ErrRedisKeyNotFound
	}
	return json.Unmarshal([]byte(entry.Value), dest)
}

func fallbackDel(key string) {
	fallbackDB.mu.Lock()
	delete(fallbackDB.data, key)
	fallbackDB.mu.Unlock()
}

func fallbackExists(key string) bool {
	fallbackDB.mu.RLock()
	entry, ok := fallbackDB.data[key]
	fallbackDB.mu.RUnlock()

	if !ok || entry == nil {
		return false
	}
	if time.Now().After(entry.ExpiresAt) {
		fallbackDB.mu.Lock()
		delete(fallbackDB.data, key)
		fallbackDB.mu.Unlock()
		return false
	}
	return true
}

func fallbackSetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	fallbackDB.mu.Lock()
	defer fallbackDB.mu.Unlock()

	if _, ok := fallbackDB.data[key]; ok {
		return false, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	fallbackDB.data[key] = &FallbackEntry{
		Value:     string(data),
		ExpiresAt: time.Now().Add(expiration),
	}
	return true, nil
}

func fallbackIncr(key string) (int64, error) {
	fallbackDB.mu.Lock()
	defer fallbackDB.mu.Unlock()

	entry, ok := fallbackDB.data[key]
	if !ok {
		fallbackDB.data[key] = &FallbackEntry{
			Value:     "1",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		return 1, nil
	}

	var val int64
	if err := json.Unmarshal([]byte(entry.Value), &val); err != nil {
		return 0, err
	}
	val++
	data, _ := json.Marshal(val)
	entry.Value = string(data)
	return val, nil
}

func fallbackExpire(key string, expiration time.Duration) (bool, error) {
	fallbackDB.mu.Lock()
	defer fallbackDB.mu.Unlock()

	entry, ok := fallbackDB.data[key]
	if !ok {
		return false, nil
	}
	entry.ExpiresAt = time.Now().Add(expiration)
	return true, nil
}

func fallbackTTL(key string) (time.Duration, error) {
	fallbackDB.mu.RLock()
	entry, ok := fallbackDB.data[key]
	fallbackDB.mu.RUnlock()

	if !ok {
		return -2 * time.Second, nil
	}
	remaining := time.Until(entry.ExpiresAt)
	if remaining <= 0 {
		return -1 * time.Second, nil
	}
	return remaining, nil
}

func cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		fallbackDB.mu.Lock()
		for k, v := range fallbackDB.data {
			if now.After(v.ExpiresAt) {
				delete(fallbackDB.data, k)
			}
		}
		fallbackDB.mu.Unlock()
	}
}
