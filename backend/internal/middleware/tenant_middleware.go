package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TenantContextKey 租户上下文键
type TenantContextKey struct{}

// TenantMiddleware 租户拦截中间件
// @Summary 租户身份识别中间件
// @Description 从HTTP Header中提取X-Tenant-ID，存入Gin Context供后续使用
// @Tags 系统中间件
// @Router /api/* [middleware]
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从HTTP Header的X-Tenant-ID中获取租户ID
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		
		if tenantIDStr != "" {
			// 转换为uint类型
			tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
			if err == nil {
				// 将租户ID存入Gin Context
				c.Set("tenant_id", uint(tenantID))
				
				// 将租户ID存入标准库的context.Context中
				ctx := context.WithValue(c.Request.Context(), TenantContextKey{}, uint(tenantID))
				c.Request = c.Request.WithContext(ctx)
			}
			// 容错设计：即使转换失败也不报错，直接放行
		}
		
		// 继续执行后续中间件和处理器
		c.Next()
	}
}

// GetTenantIDFromContext 从上下文中获取租户ID
// @Summary 获取当前请求的租户ID
// @Description 从Gin Context中提取租户ID，用于业务层判断
// @Tags 工具函数
func GetTenantIDFromContext(c *gin.Context) (uint, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return 0, false
	}
	
	tenantIDUint, ok := tenantID.(uint)
	if !ok {
		return 0, false
	}
	
	return tenantIDUint, true
}