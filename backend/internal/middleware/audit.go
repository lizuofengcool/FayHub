package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/logger"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ctxKeyOperModule = "oper_module"
	ctxKeyOperAction = "oper_action"
)

func OperLog(module, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ctxKeyOperModule, module)
		c.Set(ctxKeyOperAction, action)
		c.Next()
	}
}

var auditedPaths = map[string]bool{
	"/api/auth/login":                     true,
	"/api/auth/logout":                    true,
	"/api/users":                          true,
	"/api/tenants":                        true,
	"/api/plugin-engine/install-callback": true,
}

var auditedMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

func shouldAudit(method string, path string) bool {
	if auditedPaths[path] {
		return true
	}

	if auditedMethods[method] {
		for p := range auditedPaths {
			if len(path) >= len(p) && path[:len(p)] == p {
				return true
			}
		}
	}

	return false
}

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !shouldAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil && c.Request.Method != "GET" {
			requestBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, 64*1024))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		c.Next()

		duration := time.Since(start).Milliseconds()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		tenantID, _ := c.Get("tenant_id")

		userIDInt64, _ := userID.(int64)
		usernameStr, _ := username.(string)
		tenantIDInt64, _ := tenantID.(int64)

		action := getOperAction(c)
		resource := getOperModule(c)
		if resource == "" {
			resource = extractResource(c.Request.URL.Path)
		}
		resourceID := c.Param("id")

		var detail model.JSONRawMessage
		if len(requestBody) > 0 {
			var sanitized map[string]interface{}
			if err := json.Unmarshal(requestBody, &sanitized); err == nil {
				delete(sanitized, "password")
				delete(sanitized, "secret")
				delete(sanitized, "license_key")
				delete(sanitized, "old_password")
				delete(sanitized, "new_password")
				if d, err := json.Marshal(sanitized); err == nil {
					detail = model.JSONRawMessage(d)
				}
			}
		}

		success := c.Writer.Status() < 400

		auditLog := &model.AuditLog{
			UserID:     userIDInt64,
			Username:   usernameStr,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Detail:     detail,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			RequestID:  c.GetString("request_id"),
			StatusCode: c.Writer.Status(),
			Success:    success,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Duration:   duration,
		}
		auditLog.TenantID = tenantIDInt64

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(context.Background(), "审计日志记录panic", zap.Any("error", r))
				}
			}()
			ctx := context.Background()
			if err := service.AuditServiceApp.Record(ctx, auditLog); err != nil {
				logger.Error(ctx, "审计日志记录失败", zap.Error(err))
			}
		}()
	}
}

func getOperModule(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyOperModule); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getOperAction(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyOperAction); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return mapMethodToAction(c.Request.Method)
}

func mapMethodToAction(method string) string {
	switch method {
	case "POST":
		return string(model.AuditActionCreate)
	case "PUT", "PATCH":
		return string(model.AuditActionUpdate)
	case "DELETE":
		return string(model.AuditActionDelete)
	default:
		return method
	}
}

func extractResource(path string) string {
	parts := splitPath(path)
	if len(parts) >= 2 {
		return parts[1]
	}
	return path
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
