package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"fayhub/pkg/logger"
)

// LogContextMiddleware 日志上下文注入中间件
func LogContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		
		// 生成请求ID
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Set("start_time", startTime)
		
		// 注入日志上下文
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "path", c.Request.URL.Path)
		ctx = context.WithValue(ctx, "method", c.Request.Method)
		ctx = context.WithValue(ctx, "ip", c.ClientIP())
		
		// 如果有租户ID，也注入到上下文
		if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
			c.Set("tenant_id", tenantID)
			ctx = context.WithValue(ctx, "tenant_id", tenantID)
		}
		
		// 如果有用户ID，也注入到上下文
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
			ctx = context.WithValue(ctx, "user_id", userID)
		}
		
		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)
		
		// 记录请求开始日志
		logger.Info(ctx, "请求开始",
			zap.String("request_id", requestID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
		
		c.Next()
		
		// 记录请求结束日志
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		
		// 获取响应状态码
		statusCode := c.Writer.Status()
		
		logger.Info(ctx, "请求结束",
			zap.String("request_id", requestID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

// RequestIDMiddleware 请求ID中间件（简化版）
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID，如果没有则生成
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// TenantContextMiddleware 租户上下文中间件
func TenantContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取租户ID
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID != "" {
			c.Set("tenant_id", tenantID)
			
			// 注入到上下文
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, "tenant_id", tenantID)
			c.Request = c.Request.WithContext(ctx)
		}
		
		c.Next()
	}
}

// UserContextMiddleware 用户上下文中间件
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取用户ID
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("user_id", userID)
			
			// 注入到上下文
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, "user_id", userID)
			c.Request = c.Request.WithContext(ctx)
		}
		
		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return "req_" + uuid.New().String()
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTenantID 从上下文中获取租户ID
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// LoggingMiddleware 组合日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 按顺序执行各个中间件
		RequestIDMiddleware()(c)
		TenantContextMiddleware()(c)
		UserContextMiddleware()(c)
		LogContextMiddleware()(c)
	}
}