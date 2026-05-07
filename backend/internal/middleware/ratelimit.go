package middleware

import (
	"fayhub/pkg/ratelimit"
	"fayhub/pkg/response"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(category string) gin.HandlerFunc {
	cfg := ratelimit.GetRateLimitConfig(category)

	return func(c *gin.Context) {
		identifier := c.ClientIP()

		if tenantID, exists := c.Get("tenant_id"); exists {
			if tid, ok := tenantID.(uint); ok && tid > 0 {
				identifier = fmt.Sprintf("tenant:%d:%s", tid, c.ClientIP())
			}
		}

		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok && uid > 0 {
				identifier = fmt.Sprintf("user:%d", uid)
			}
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxTokens))
		c.Header("X-RateLimit-Window", cfg.Window.String())

		if !ratelimit.CheckRateLimit(category, identifier) {
			c.Header("Retry-After", strconv.Itoa(int(cfg.Window.Seconds())))
			response.GinError(c, 42900, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

func StrictRateLimitMiddleware(maxTokens int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()

		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok && uid > 0 {
				identifier = fmt.Sprintf("user:%d", uid)
			}
		}

		limiter := ratelimit.NewRateLimiter(
			fmt.Sprintf("strict:%s", identifier),
			maxTokens,
			window,
		)

		c.Header("X-RateLimit-Limit", strconv.Itoa(maxTokens))
		c.Header("X-RateLimit-Window", window.String())

		if !limiter.Allow() {
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			response.GinError(c, 42900, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
