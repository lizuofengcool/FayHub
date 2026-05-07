package model

import (
	"database/sql/driver"
	"fayhub/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type DeletedAt gorm.DeletedAt

func (d DeletedAt) MarshalJSON() ([]byte, error) {
	return gorm.DeletedAt(d).MarshalJSON()
}

func (d *DeletedAt) UnmarshalJSON(b []byte) error {
	return (*gorm.DeletedAt)(d).UnmarshalJSON(b)
}

func (d *DeletedAt) Scan(value interface{}) error {
	return (*gorm.DeletedAt)(d).Scan(value)
}

func (d DeletedAt) Value() (driver.Value, error) {
	return gorm.DeletedAt(d).Value()
}

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at" swaggertype:"string"`
}

type TenantModel struct {
	BaseModel
	TenantID int64 `gorm:"index;not null" json:"tenant_id"`
}

type SnowflakeModel struct {
	ID        int64     `gorm:"primarykey" json:"id,string"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at" swaggertype:"string"`
}

func (s *SnowflakeModel) BeforeCreate(tx *gorm.DB) error {
	if s.ID == 0 {
		s.ID = utils.GenerateSnowflakeID()
	}
	return nil
}

type SnowflakeTenantModel struct {
	SnowflakeModel
	TenantID int64 `gorm:"index;not null" json:"tenant_id"`
}

func (t *SnowflakeTenantModel) BeforeCreate(tx *gorm.DB) error {
	if err := t.SnowflakeModel.BeforeCreate(tx); err != nil {
		return err
	}

	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	if utils.IsTenantIsolationSkipped(ctx) {
		return nil
	}

	tenantID, ok := utils.GetTenantIDFromCtx(ctx)
	if !ok || tenantID == 0 {
		return nil
	}

	if t.TenantID == 0 {
		t.TenantID = tenantID
	}

	return nil
}

func (t *TenantModel) BeforeCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	if ctx == nil {
		return nil
	}

	if utils.IsTenantIsolationSkipped(ctx) {
		return nil
	}

	tenantID, ok := utils.GetTenantIDFromCtx(ctx)
	if !ok || tenantID == 0 {
		return nil
	}

	if t.TenantID == 0 {
		t.TenantID = tenantID
	}

	return nil
}

func RegisterTenantIsolationCallbacks(db *gorm.DB) error {
	if err := db.Callback().Query().Before("gorm:query").Register("tenant_isolation:query", tenantIsolationQueryCallback); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register("tenant_isolation:update", tenantIsolationUpdateCallback); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register("tenant_isolation:delete", tenantIsolationDeleteCallback); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register("tenant_isolation:row", tenantIsolationQueryCallback); err != nil {
		return err
	}
	if err := db.Callback().Query().Before("gorm:query").Register("data_permission:query", dataPermissionQueryCallback); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register("data_permission:row", dataPermissionQueryCallback); err != nil {
		return err
	}
	return nil
}

func CreateCompositeIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_status ON users(tenant_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_roles_tenant_name ON roles(tenant_id, name)",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_user_role ON user_roles(tenant_id, user_id, role_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_menus_tenant_role_menu ON role_menus(tenant_id, role_id, menu_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_apis_tenant_role_api ON role_apis(tenant_id, role_id, api_id)",
		"CREATE INDEX IF NOT EXISTS idx_apis_path_method ON apis(path, method)",
		"CREATE INDEX IF NOT EXISTS idx_menus_parent_status ON menus(parent_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_plugin_configs_tenant_plugin_key ON plugin_configs(tenant_id, plugin_id, config_key)",
		"CREATE INDEX IF NOT EXISTS idx_plugin_event_logs_tenant_plugin ON plugin_event_logs(tenant_id, plugin_id)",
		"CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_user ON tenant_users(tenant_id, user_id)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return err
		}
	}
	return nil
}

func tenantIsolationQueryCallback(db *gorm.DB) {
	applyTenantIsolation(db)
}

func tenantIsolationUpdateCallback(db *gorm.DB) {
	applyTenantIsolation(db)
}

func tenantIsolationDeleteCallback(db *gorm.DB) {
	applyTenantIsolation(db)
}

func applyTenantIsolation(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	if utils.IsTenantIsolationSkipped(ctx) {
		return
	}

	tenantID, ok := utils.GetTenantIDFromCtx(ctx)
	if !ok || tenantID == 0 {
		return
	}

	if !hasTenantIDColumn(db) {
		return
	}

	if db.Statement.Schema != nil {
		tableName := db.Statement.Schema.Table
		db.Where(tableName+".tenant_id = ?", tenantID)
	} else {
		db.Where("tenant_id = ?", tenantID)
	}
}

func hasTenantIDColumn(db *gorm.DB) bool {
	if db.Statement.Schema == nil {
		return true
	}

	for _, field := range db.Statement.Schema.Fields {
		if field.DBName == "tenant_id" {
			return true
		}
	}
	return false
}

func dataPermissionQueryCallback(db *gorm.DB) {
	applyDataPermission(db)
}

func applyDataPermission(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	if utils.IsDataPermissionSkipped(ctx) {
		return
	}

	filter, ok := utils.GetDataScopeFilterFromCtx(ctx)
	if !ok || filter == nil || filter.IsAdmin {
		return
	}

	if db.Statement.Schema == nil {
		return
	}

	hasDeptID := false
	hasUserID := false
	for _, field := range db.Statement.Schema.Fields {
		if field.DBName == "dept_id" {
			hasDeptID = true
		}
		if field.DBName == "user_id" {
			hasUserID = true
		}
	}

	if !hasDeptID && !hasUserID {
		return
	}

	tableName := db.Statement.Schema.Table

	switch filter.Scope {
	case DataScopeAll:
		return

	case DataScopeDept:
		if hasDeptID && len(filter.DeptIDs) > 0 {
			db.Where(tableName+".dept_id IN ?", filter.DeptIDs)
		} else if hasUserID {
			db.Where(tableName+".user_id = ?", filter.UserID)
		}

	case DataScopeDeptAndSub:
		if hasDeptID && len(filter.DeptIDs) > 0 {
			db.Where(tableName+".dept_id IN ?", filter.DeptIDs)
		} else if hasUserID {
			db.Where(tableName+".user_id = ?", filter.UserID)
		}

	case DataScopeSelf:
		if hasUserID {
			db.Where(tableName+".user_id = ?", filter.UserID)
		}

	case DataScopeCustom:
		if hasDeptID && len(filter.DeptIDs) > 0 {
			db.Where(tableName+".dept_id IN ? OR "+tableName+".user_id = ?", filter.DeptIDs, filter.UserID)
		} else if hasUserID {
			db.Where(tableName+".user_id = ?", filter.UserID)
		}

	default:
		if hasUserID {
			db.Where(tableName+".user_id = ?", filter.UserID)
		}
	}
}
