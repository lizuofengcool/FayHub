package middleware

import (
	"fayhub/pkg/ratelimit"
	"fayhub/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(category string) gin.HandlerFunc {
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

		if !ratelimit.CheckRateLimit(category, identifier) {
			response.GinError(c, 42900, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
