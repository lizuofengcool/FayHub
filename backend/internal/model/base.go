package model

import (
	"gorm.io/gorm"
	"time"
)

// BaseModel 所有模型的基类
// 包含ID、CreatedAt、UpdatedAt、DeletedAt（用于软删除）
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt time.Time      `json:"created_at"`           // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 删除时间（软删除）
}

// TenantModel 租户基类
// 继承自BaseModel，并强制包含tenant_id字段
type TenantModel struct {
	BaseModel
	TenantID uint `gorm:"index;not null" json:"tenant_id"` // 租户ID
}

// BeforeCreate GORM钩子函数，在创建记录前自动设置tenant_id
func (t *TenantModel) BeforeCreate(tx *gorm.DB) error {
	// 从上下文中获取当前租户ID
	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 检查是否跳过租户隔离
	if skip, ok := ctx.Value("skip_tenant_isolation").(bool); ok && skip {
		return nil
	}

	// 从上下文中获取租户ID
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return nil
	}

	// 设置租户ID
	t.TenantID = tenantID

	return nil
}

// BeforeFind GORM钩子函数，在查询记录前自动带上tenant_id
func (t *TenantModel) BeforeFind(tx *gorm.DB) error {
	// 从上下文中获取当前租户ID
	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 检查是否跳过租户隔离
	if skip, ok := ctx.Value("skip_tenant_isolation").(bool); ok && skip {
		return nil
	}

	// 从上下文中获取租户ID
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok || tenantID == 0 {
		return nil
	}

	// 追加租户ID查询条件
	tx.Where("tenant_id = ?", tenantID)

	return nil
}

// BeforeUpdate GORM钩子函数，在更新记录前自动带上tenant_id
func (t *TenantModel) BeforeUpdate(tx *gorm.DB) error {
	// 从上下文中获取当前租户ID
	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 检查是否跳过租户隔离
	if skip, ok := ctx.Value("skip_tenant_isolation").(bool); ok && skip {
		return nil
	}

	// 从上下文中获取租户ID
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok || tenantID == 0 {
		return nil
	}

	// 追加租户ID查询条件
	tx.Where("tenant_id = ?", tenantID)

	return nil
}

// BeforeDelete GORM钩子函数，在删除记录前自动带上tenant_id
func (t *TenantModel) BeforeDelete(tx *gorm.DB) error {
	// 从上下文中获取当前租户ID
	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	// 检查是否跳过租户隔离
	if skip, ok := ctx.Value("skip_tenant_isolation").(bool); ok && skip {
		return nil
	}

	// 从上下文中获取租户ID
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok || tenantID == 0 {
		return nil
	}

	// 追加租户ID查询条件
	tx.Where("tenant_id = ?", tenantID)

	return nil
}