package utils

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// ContextKey 定义上下文键类型
type ContextKey string

const (
	// GinContextKey Gin上下文在标准上下文中的键名
	GinContextKey ContextKey = "gin_context"
)

// SetGinContext 将Gin上下文存储到标准上下文中
func SetGinContext(ctx context.Context, ginCtx *gin.Context) context.Context {
	return context.WithValue(ctx, GinContextKey, ginCtx)
}

// GetGinContext 从标准上下文中获取Gin上下文
func GetGinContext(ctx context.Context) (*gin.Context, error) {
	ginCtxValue := ctx.Value(GinContextKey)
	if ginCtxValue == nil {
		return nil, fmt.Errorf("Gin上下文不存在")
	}

	ginCtx, ok := ginCtxValue.(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("Gin上下文类型错误")
	}

	return ginCtx, nil
}

// GetUserIDFromContext 从标准上下文中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return 0, false
	}

	userID, exists := ginCtx.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return userIDUint, true
}

// GetTenantIDFromContext 从标准上下文中获取租户ID
func GetTenantIDFromContext(ctx context.Context) (uint, bool) {
	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return 0, false
	}

	tenantID, exists := ginCtx.Get("tenant_id")
	if !exists {
		return 0, false
	}

	tenantIDUint, ok := tenantID.(uint)
	if !ok {
		return 0, false
	}

	return tenantIDUint, true
}

// GetUsernameFromContext 从标准上下文中获取用户名
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return "", false
	}

	username, exists := ginCtx.Get("username")
	if !exists {
		return "", false
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", false
	}

	return usernameStr, true
}

// GetRoleFromContext 从标准上下文中获取角色
func GetRoleFromContext(ctx context.Context) (string, bool) {
	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return "", false
	}

	role, exists := ginCtx.Get("role")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", false
	}

	return roleStr, true
}
