package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	ErrRedisKeyNotFound = errors.New("redis key not found")
)

type ContextKey string

const (
	GinContextKey ContextKey = "gin_context"
)

func SetGinContext(ctx context.Context, ginCtx *gin.Context) context.Context {
	return context.WithValue(ctx, GinContextKey, ginCtx)
}

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

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	if ctx != nil {
		if id, ok := ctx.Value("user_id").(int64); ok {
			return id, true
		}
	}

	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return 0, false
	}

	userID, exists := ginCtx.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		return 0, false
	}

	return userIDInt64, true
}

func GetTenantIDFromContext(ctx context.Context) (int64, bool) {
	if id, ok := GetTenantIDFromCtx(ctx); ok {
		return id, true
	}

	ginCtx, err := GetGinContext(ctx)
	if err != nil {
		return 0, false
	}

	tenantID, exists := ginCtx.Get("tenant_id")
	if !exists {
		return 0, false
	}

	tenantIDInt64, ok := tenantID.(int64)
	if !ok {
		return 0, false
	}

	return tenantIDInt64, true
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	if ctx != nil {
		if username, ok := ctx.Value("username").(string); ok {
			return username, true
		}
	}

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

func GetRoleFromContext(ctx context.Context) (string, bool) {
	if ctx != nil {
		if role, ok := ctx.Value("role").(string); ok {
			return role, true
		}
	}

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
