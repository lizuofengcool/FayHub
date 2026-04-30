package middleware

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取API密钥
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.GinError(c, errors.ErrUnauthorized, "缺少API密钥")
			c.Abort()
			return
		}

		// 验证Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.GinError(c, errors.ErrUnauthorized, "API密钥格式错误")
			c.Abort()
			return
		}

		apiKey := parts[1]
		if apiKey == "" {
			response.GinError(c, errors.ErrUnauthorized, "API密钥不能为空")
			c.Abort()
			return
		}

		// 验证API密钥
		keyService := &service.APIKeyService{}
		key, err := keyService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			response.GinError(c, errors.ErrUnauthorized, "无效的API密钥")
			c.Abort()
			return
		}

		// 检查限流
		allowed, _ := keyService.CheckRateLimit(c.Request.Context(), key.ID, key.RateLimit)
		if !allowed {
			response.GinError(c, 42900, "API请求频率超限")
			c.Abort()
			return
		}

		// 将API密钥信息存入上下文
		c.Set("api_key_id", key.ID)
		c.Set("api_key_tenant_id", key.TenantID)
		c.Set("api_key_user_id", key.UserID)
		c.Set("api_key", key)

		c.Next()
	}
}

func APIKeyPermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyInterface, exists := c.Get("api_key")
		if !exists {
			response.GinError(c, errors.ErrUnauthorized, "API密钥未验证")
			c.Abort()
			return
		}

		key, ok := keyInterface.(*model.APIKey)
		if !ok {
			response.GinError(c, errors.ErrInternalServer, "API密钥信息错误")
			c.Abort()
			return
		}

		keyService := &service.APIKeyService{}
		if !keyService.CheckPermission(key, resource, action) {
			response.GinError(c, errors.ErrForbidden, "无权限执行此操作")
			c.Abort()
			return
		}

		c.Next()
	}
}
