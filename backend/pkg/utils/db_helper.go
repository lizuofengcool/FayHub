package utils

import (
	"context"
	"strconv"

	"gorm.io/gorm"
)

// globalDB 全局数据库实例（模拟定义，实际项目中应从global包导入）
var globalDB *gorm.DB

// SetGlobalDB 设置全局数据库实例（用于测试和初始化）
// @Summary 设置全局数据库实例
// @Description 初始化全局数据库连接，供GetDB函数使用
// @Tags 数据库工具
func SetGlobalDB(db *gorm.DB) {
	globalDB = db
}

// GetGlobalDB 获取原始全局数据库实例（不包含租户隔离）
// @Summary 获取全局数据库
// @Description 获取原始数据库实例，用于初始化和迁移
// @Tags 数据库工具
func GetGlobalDB() *gorm.DB {
	return globalDB
}

// GetDB 智能数据库管家
// @Summary 获取带租户隔离的数据库实例
// @Description 根据上下文中的租户ID自动追加WHERE条件，实现多租户数据隔离
// @Tags 数据库工具
// @Param ctx context.Context true "请求上下文"
// @Return *gorm.DB "带租户隔离的数据库实例"
func GetDB(ctx context.Context) *gorm.DB {
	if globalDB == nil {
		// 阶段一：数据库未初始化，返回nil（健康检查接口不需要数据库）
		return nil
	}

	// 获取带上下文的基础DB实例
	db := globalDB.WithContext(ctx)

	// 尝试从上下文中提取租户ID
	tenantID := extractTenantIDFromContext(ctx)

	if tenantID > 0 {
		// 商家租户操作：自动追加租户隔离条件
		return db.Where("tenant_id = ?", tenantID)
	}

	// 全局超管操作：直接返回，不追加租户条件
	return db
}

// extractTenantIDFromContext 从上下文中提取租户ID
// @Summary 提取租户ID
// @Description 从context.Context中安全提取tenant_id
// @Tags 内部工具
func extractTenantIDFromContext(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}

	// 尝试从标准库context中获取
	tenantIDValue := ctx.Value("tenant_id")
	if tenantIDValue == nil {
		return 0
	}

	switch v := tenantIDValue.(type) {
	case uint:
		return v
	case int:
		if v > 0 {
			return uint(v)
		}
	case int64:
		if v > 0 {
			return uint(v)
		}
	case string:
		// 如果是字符串类型，尝试转换
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(id)
		}
	}

	return 0
}

// SkipTenantIsolation 跳过租户隔离
// @Summary 跳过租户隔离检查
// @Description 用于需要跨租户查询的特殊场景
// @Tags 数据库工具
// @Param ctx context.Context true "请求上下文"
// @Return context.Context "跳过租户隔离的上下文"
func SkipTenantIsolation(ctx context.Context) context.Context {
	return context.WithValue(ctx, "skip_tenant_isolation", true)
}

// IsTenantIsolationSkipped 检查是否跳过租户隔离
// @Summary 检查租户隔离状态
// @Description 判断当前上下文是否设置了跳过租户隔离
// @Tags 数据库工具
// @Param ctx context.Context true "请求上下文"
// @Return bool "是否跳过租户隔离"
func IsTenantIskipped(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	skipValue := ctx.Value("skip_tenant_isolation")
	if skipValue == nil {
		return false
	}

	skip, ok := skipValue.(bool)
	return ok && skip
}
