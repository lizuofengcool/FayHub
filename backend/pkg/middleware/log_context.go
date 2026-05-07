package middleware

import (
	"context"
	"fayhub/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"fayhub/pkg/logger"
)

func LogContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Set("start_time", startTime)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "path", c.Request.URL.Path)
		ctx = context.WithValue(ctx, "method", c.Request.Method)
		ctx = context.WithValue(ctx, "ip", c.ClientIP())

		if tid, exists := c.Get("tenant_id"); exists {
			if tidInt64, ok := tid.(int64); ok {
				ctx = utils.WithTenantID(ctx, tidInt64)
			}
		}

		if uid, exists := c.Get("user_id"); exists {
			if uidInt64, ok := uid.(int64); ok {
				ctx = context.WithValue(ctx, "user_id", uidInt64)
			}
		}

		c.Request = c.Request.WithContext(ctx)

		logger.Info(ctx, "请求开始",
			zap.String("request_id", requestID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		endTime := time.Now()
		duration := endTime.Sub(startTime)
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

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

func generateRequestID() string {
	return "req_" + uuid.New().String()
}

func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func LoggingMiddleware() gin.HandlerFunc {
	return LogContextMiddleware()
}
