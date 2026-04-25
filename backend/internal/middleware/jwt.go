package middleware

import (
	"fayhub/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware JWT认证中间件
// @Summary JWT认证中间件
// @Description 验证JWT Token的有效性，并将用户信息存入上下文
// @Tags 系统中间件
// @Router /api/* [middleware]
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请提供认证Token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 检查Token格式（Bearer token）
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 解析Token
		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 继续执行后续中间件和处理器
		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
// @Summary 获取当前请求的用户ID
// @Description 从Gin Context中提取用户ID，用于业务层判断
// @Tags 工具函数
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return userIDUint, true
}

// GetUsernameFromContext 从上下文中获取用户名
// @Summary 获取当前请求的用户名
// @Description 从Gin Context中提取用户名
// @Tags 工具函数
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", false
	}

	return usernameStr, true
}

// GetRoleFromContext 从上下文中获取用户角色
// @Summary 获取当前请求的用户角色
// @Description 从Gin Context中提取用户角色
// @Tags 工具函数
func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", false
	}

	return roleStr, true
}