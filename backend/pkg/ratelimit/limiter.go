package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fayhub/pkg/redisclient"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	key        string
	maxTokens  int
	window     time.Duration
	mu         sync.Mutex
	tokens     int
	lastRefill time.Time
}

func NewRateLimiter(key string, maxTokens int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		key:        key,
		maxTokens:  maxTokens,
		window:     window,
		tokens:     maxTokens,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rdb := redisclient.GetRawClient()
	if rdb != nil {
		return rl.allowRedis(rdb)
	}
	return rl.allowLocal()
}

func (rl *RateLimiter) allowRedis(rdb *redis.Client) bool {
	ctx := context.Background()
	key := fmt.Sprintf("ratelimit:%s", rl.key)

	now := time.Now().UnixMilli()
	windowMs := rl.window.Milliseconds()

	pipe := rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowMs))
	pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, rl.window+time.Second)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return rl.allowLocal()
	}

	count := results[1].(*redis.IntCmd).Val()

	if count >= int64(rl.maxTokens) {
		return false
	}

	return true
}

func (rl *RateLimiter) allowLocal() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	refillTokens := int(elapsed / rl.window) * rl.maxTokens
	if refillTokens > 0 {
		rl.tokens += refillTokens
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

type RateLimitConfig struct {
	MaxTokens int
	Window    time.Duration
}

var defaultLimits = map[string]RateLimitConfig{
	"login":    {MaxTokens: 5, Window: time.Minute},
	"api":      {MaxTokens: 100, Window: time.Minute},
	"webhook":  {MaxTokens: 30, Window: time.Minute},
	"upload":   {MaxTokens: 10, Window: time.Minute},
	"sso":      {MaxTokens: 20, Window: time.Minute},
	"backup":   {MaxTokens: 3, Window: time.Minute},
	"payment":  {MaxTokens: 10, Window: time.Minute},
	"plugin":   {MaxTokens: 30, Window: time.Minute},
}

func GetRateLimitConfig(category string) RateLimitConfig {
	if cfg, ok := defaultLimits[category]; ok {
		return cfg
	}
	return RateLimitConfig{MaxTokens: 60, Window: time.Minute}
}

func CheckRateLimit(category string, identifier string) bool {
	cfg := GetRateLimitConfig(category)
	key := fmt.Sprintf("%s:%s", category, identifier)
	limiter := NewRateLimiter(key, cfg.MaxTokens, cfg.Window)
	return limiter.Allow()
}
