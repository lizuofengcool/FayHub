package utils

import (
	"context"
	"regexp"
	"strings"

	"fayhub/pkg/config"

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

func GetDBConfig() *config.DatabaseConfig {
	if config.GlobalConfig == nil {
		return nil
	}
	return &config.GlobalConfig.Database
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

var cuidPattern = regexp.MustCompile(`^[a-z0-9]{8,25}$`)
var uuidWithDashPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ValidateCUID(id string) bool {
	if id == "" {
		return false
	}
	return cuidPattern.MatchString(id) || uuidWithDashPattern.MatchString(id)
}

var tableNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

func ValidateTableName(name string) bool {
	if name == "" {
		return false
	}
	return tableNamePattern.MatchString(name)
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ValidateUUID(id string) bool {
	if id == "" {
		return false
	}
	return uuidPattern.MatchString(id)
}
