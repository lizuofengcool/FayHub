package middleware

import (
	"context"
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

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = utils.WithTenantID(ctx, claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
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

func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	return utils.GetUserIDFromContext(c.Request.Context())
}

func GetUsernameFromContext(c *gin.Context) (string, bool) {
	return utils.GetUsernameFromContext(c.Request.Context())
}

func GetRoleFromContext(c *gin.Context) (string, bool) {
	return utils.GetRoleFromContext(c.Request.Context())
}
