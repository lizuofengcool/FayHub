package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

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
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		c.Next()

		duration := time.Since(start).Milliseconds()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		tenantID, _ := c.Get("tenant_id")

		userIDUint, _ := userID.(uint)
		usernameStr, _ := username.(string)
		tenantIDUint, _ := tenantID.(uint)

		action := mapMethodToAction(c.Request.Method)
		resource := extractResource(c.Request.URL.Path)
		resourceID := c.Param("id")

		var detail json.RawMessage
		if len(requestBody) > 0 {
			var sanitized map[string]interface{}
			if err := json.Unmarshal(requestBody, &sanitized); err == nil {
				delete(sanitized, "password")
				delete(sanitized, "secret")
				delete(sanitized, "license_key")
				delete(sanitized, "old_password")
				delete(sanitized, "new_password")
				if d, err := json.Marshal(sanitized); err == nil {
					detail = json.RawMessage(d)
				}
			}
		}

		success := c.Writer.Status() < 400

		auditLog := &model.AuditLog{
			UserID:     userIDUint,
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
		auditLog.TenantID = tenantIDUint

		go func() {
			ctx := context.Background()
			if err := service.AuditServiceApp.Record(ctx, auditLog); err != nil {
				_ = err
			}
		}()
	}
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
