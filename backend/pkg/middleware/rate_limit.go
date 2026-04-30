package middleware

import (
	"context"
	"fayhub/pkg/errors"
	"fayhub/pkg/redisclient"
	"fayhub/pkg/response"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count     int
	expiresAt time.Time
}

type fallbackLimiter struct {
	mu          sync.RWMutex
	visitors    map[string]*visitor
	maxRequests int
	window      time.Duration
}

func newFallbackLimiter(maxRequests int, window time.Duration) *fallbackLimiter {
	limiter := &fallbackLimiter{
		visitors:    make(map[string]*visitor),
		maxRequests: maxRequests,
		window:      window,
	}
	go limiter.cleanup()
	return limiter
}

func (fl *fallbackLimiter) cleanup() {
	ticker := time.NewTicker(fl.window)
	defer ticker.Stop()
	for range ticker.C {
		fl.mu.Lock()
		now := time.Now()
		for ip, v := range fl.visitors {
			if now.After(v.expiresAt) {
				delete(fl.visitors, ip)
			}
		}
		fl.mu.Unlock()
	}
}

func (fl *fallbackLimiter) Allow(ip string) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	now := time.Now()
	v, exists := fl.visitors[ip]
	if !exists || now.After(v.expiresAt) {
		fl.visitors[ip] = &visitor{
			count:     1,
			expiresAt: now.Add(fl.window),
		}
		return true
	}

	v.count++
	if v.count > fl.maxRequests {
		return false
	}
	return true
}

const rateLimitPrefix = "ratelimit:"

func redisRateLimitAllow(ctx context.Context, key string, maxRequests int, window time.Duration) bool {
	if !redisclient.IsEnabled() {
		return false
	}

	redisKey := rateLimitPrefix + key

	count, err := redisclient.Incr(ctx, redisKey)
	if err != nil {
		return true
	}

	if count == 1 {
		redisclient.Expire(ctx, redisKey, window)
	}

	return count <= int64(maxRequests)
}

func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	fallback := newFallbackLimiter(maxRequests, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())

		if redisclient.IsEnabled() {
			if !redisRateLimitAllow(c.Request.Context(), key+":"+ip, maxRequests, window) {
				response.GinError(c, errors.ErrOperationFailed, "请求过于频繁，请稍后再试")
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		} else {
			if !fallback.Allow(key + ":" + ip) {
				response.GinError(c, errors.ErrOperationFailed, "请求过于频繁，请稍后再试")
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}

		c.Next()
	}
}
