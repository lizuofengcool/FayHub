package utils

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type contextKey string

const (
	TenantIDKey      contextKey = "tenant_id"
	SkipIsolationKey contextKey = "skip_tenant_isolation"
)

var globalDB *gorm.DB

func SetGlobalDB(db *gorm.DB) {
	globalDB = db
}

func GetGlobalDB() *gorm.DB {
	return globalDB
}

func GetDB(ctx context.Context) *gorm.DB {
	if globalDB == nil {
		return nil
	}

	return globalDB.WithContext(ctx)
}

func SkipTenantIsolation(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipIsolationKey, true)
}

func IsTenantIsolationSkipped(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	skip, ok := ctx.Value(SkipIsolationKey).(bool)
	return ok && skip
}

func GetTenantIDFromCtx(ctx context.Context) (uint, bool) {
	if ctx == nil {
		return 0, false
	}
	id, ok := ctx.Value(TenantIDKey).(uint)
	return id, ok
}

func WithTenantID(ctx context.Context, tenantID uint) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
