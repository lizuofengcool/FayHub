package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TenantContextKey 租户上下文键
type TenantContextKey struct{}

// SkipTenantIsolationKey 跳过租户隔离键
type SkipTenantIsolationKey struct{}

// TenantMiddleware 租户拦截中间件
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从HTTP Header的X-Tenant-ID中获取租户ID
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			// 如果没有租户ID，设置为0（表示总后台管理员）
			c.Set("tenant_id", uint(0))
		} else {
			// 转换为uint类型
			tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "invalid tenant ID",
				})
				return
			}
			c.Set("tenant_id", uint(tenantID))
		}

		// 将租户ID存入标准库的context.Context中
	ctx := context.WithValue(c.Request.Context(), "tenant_id", c.GetUint("tenant_id"))
	c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// SkipTenantIsolation 跳过租户隔离
func SkipTenantIsolation(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipTenantIsolationKey{}, true)
}