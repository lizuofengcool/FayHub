package middleware

import (
	"context"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fayhub/pkg/tokenblacklist"
	"fayhub/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.GinError(c, errors.ErrUnauthorized, "未登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.GinError(c, errors.ErrTokenInvalid, "Token格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		if tokenblacklist.IsBlacklisted(c.Request.Context(), tokenString) {
			response.GinError(c, errors.ErrTokenRevoked, "Token已注销")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			go recordLoginFailedEvent(c.ClientIP(), authHeader)
			response.GinError(c, errors.ErrTokenInvalid, "Token无效或已过期")
			c.Abort()
			return
		}

		c.Set("token_string", tokenString)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("tenant_id", claims.TenantID)
		c.Set("is_impersonated", claims.IsImpersonated)
		c.Set("original_admin_id", claims.OriginalAdminID)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "is_impersonated", claims.IsImpersonated)
		ctx = context.WithValue(ctx, "original_admin_id", claims.OriginalAdminID)
		ctx = utils.WithTenantID(ctx, claims.TenantID)

		// 如果是超级管理员或平台管理员，跳过租户隔离和数据权限
		if claims.Role == "super_admin" || claims.Role == "platform_admin" {
			ctx = utils.SkipTenantIsolation(ctx)
			ctx = utils.SkipDataPermission(ctx)
		} else {
			if !utils.IsDataPermissionSkipped(ctx) {
				permSvc := &service.DataPermissionService{}
				filter, err := permSvc.GetDataScope(ctx, claims.UserID)
				if err == nil && filter != nil {
					ctx = utils.WithDataScopeFilter(ctx, filter)
				}
			}
		}

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		go recordOnlineActivity(c)
	}
}

func recordLoginFailedEvent(ip, authHeader string) {
	ctx := context.Background()
	securityService := &service.SecurityEventService{}
	securityService.RecordSecurityEvent(ctx, &service.SecurityEvent{
		Type:        service.SecurityEventLoginFailed,
		IP:          ip,
		Description: "JWT认证失败",
		Severity:    "medium",
		Details: map[string]interface{}{
			"auth_header_present": authHeader != "",
		},
	})
}

func GetTokenString(c *gin.Context) (string, bool) {
	tokenString, exists := c.Get("token_string")
	if !exists {
		return "", false
	}
	return tokenString.(string), true
}

func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	return utils.GetUserIDFromContext(c.Request.Context())
}

func GetUsernameFromContext(c *gin.Context) (string, bool) {
	return utils.GetUsernameFromContext(c.Request.Context())
}

func GetRoleFromContext(c *gin.Context) (string, bool) {
	return utils.GetRoleFromContext(c.Request.Context())
}

func recordOnlineActivity(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	tenantID, _ := c.Get("tenant_id")

	uid, _ := userID.(int64)
	tid, _ := tenantID.(int64)

	user := model.OnlineUser{
		UserID:    uid,
		Username:  toString(username),
		Role:      toString(role),
		TenantID:  tid,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	_ = service.ServiceGroupApp.OnlineUserService.RecordActivity(c.Request.Context(), user)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
