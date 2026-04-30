package middleware

import (
	"context"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/logger"
	"fayhub/pkg/response"
	"fayhub/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID uint
		var crossTenantAccess bool

		tokenTenantID, hasTokenTenant := c.Get("tenant_id")
		if hasTokenTenant {
			if tid, ok := tokenTenantID.(uint); ok {
				tenantID = tid
			}
		}

		role, hasRole := c.Get("role")
		roleStr := ""
		if hasRole {
			roleStr, _ = role.(string)
		}

		if roleStr == "super_admin" || roleStr == "platform_admin" {
			tenantIDStr := c.GetHeader("X-Tenant-ID")
			if tenantIDStr != "" {
				if tid, err := strconv.ParseUint(tenantIDStr, 10, 32); err == nil {
					targetTenantID := uint(tid)

					if targetTenantID != tenantID {
						crossTenantAccess = true

						if !validateTenantAccess(c, targetTenantID, roleStr) {
							response.GinError(c, errors.ErrForbidden, "无权访问该租户数据")
							c.Abort()
							return
						}

						logCrossTenantAccess(c, tenantID, targetTenantID, roleStr)
					}

					tenantID = targetTenantID
				}
			}
		} else {
			if !hasTokenTenant || tenantID == 0 {
				tenantID = resolveTenantByDomain(c)
				if tenantID == 0 {
					response.GinError(c, errors.ErrUnauthorized, "无法识别租户身份，请重新登录")
					c.Abort()
					return
				}
			}
		}

		c.Set("tenant_id", tenantID)
		c.Set("cross_tenant_access", crossTenantAccess)

		ctx := utils.WithTenantID(c.Request.Context(), tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func validateTenantAccess(c *gin.Context, targetTenantID uint, role string) bool {
	queryCtx := utils.SkipTenantIsolation(c.Request.Context())
	queryDB := utils.GetDB(queryCtx)
	if queryDB == nil {
		return false
	}

	var tenant model.Tenant
	if err := queryDB.Where("id = ? AND status = 1", targetTenantID).First(&tenant).Error; err != nil {
		logger.Warn(c.Request.Context(), "跨租户访问失败：目标租户不存在或已禁用",
			zap.Uint("target_tenant_id", targetTenantID),
			zap.String("role", role),
			zap.String("ip", c.ClientIP()),
			zap.Error(err))
		return false
	}

	return true
}

func logCrossTenantAccess(c *gin.Context, sourceTenantID, targetTenantID uint, role string) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	logger.Info(c.Request.Context(), "跨租户访问操作",
		zap.Any("user_id", userID),
		zap.Any("username", username),
		zap.Uint("source_tenant_id", sourceTenantID),
		zap.Uint("target_tenant_id", targetTenantID),
		zap.String("role", role),
		zap.String("ip", c.ClientIP()),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Time("timestamp", time.Now()))

	go func() {
		ctx := context.Background()
		securityService := &service.SecurityEventService{}
		securityService.RecordSecurityEvent(ctx, &service.SecurityEvent{
			Type:        service.SecurityEventCrossTenantAccess,
			TenantID:    sourceTenantID,
			UserID:      userID.(uint),
			Username:    username.(string),
			IP:          c.ClientIP(),
			Description: "跨租户访问操作",
			Severity:    "medium",
			Details: map[string]interface{}{
				"source_tenant_id": sourceTenantID,
				"target_tenant_id": targetTenantID,
				"role":             role,
				"method":           c.Request.Method,
				"path":             c.Request.URL.Path,
			},
		})
	}()
}

func resolveTenantByDomain(c *gin.Context) uint {
	host := c.Request.Host
	if host == "" {
		host = c.GetHeader("X-Forwarded-Host")
	}
	if host == "" {
		return 0
	}

	host = strings.Split(host, ":")[0]
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || host == "127.0.0.1" {
		return 0
	}

	queryCtx := utils.SkipTenantIsolation(c.Request.Context())
	queryDB := utils.GetDB(queryCtx)
	if queryDB == nil {
		return 0
	}

	var tenant model.Tenant
	if err := queryDB.Where("domain = ? AND status = 1", host).First(&tenant).Error; err != nil {
		return 0
	}

	return tenant.ID
}

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
