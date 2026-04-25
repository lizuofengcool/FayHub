package model

// Role 角色表
// 用于RBAC权限管理，定义平台和租户的角色
// 平台角色：super_admin（超级管理员）、platform_admin（平台管理员）
// 租户角色：tenant_admin（租户管理员）、tenant_user（租户用户）
type Role struct {
	BaseModel
	Name        string `gorm:"size:100;unique;not null" json:"name"`        // 角色名称
	Description string `gorm:"size:500" json:"description"`                 // 角色描述
	Type        int    `gorm:"default:1" json:"type"`                       // 角色类型（1：平台角色，2：租户角色）
	Status      int    `gorm:"default:1" json:"status"`                     // 角色状态（1：正常，0：禁用）
}

// Menu 菜单表
// 定义系统菜单和权限点
// 平台菜单：系统设置、租户管理、用户管理等
// 租户菜单：业务功能菜单
type Menu struct {
	BaseModel
	ParentID    uint   `gorm:"index" json:"parent_id"`                      // 父菜单ID
	Title       string `gorm:"size:100;not null" json:"title"`              // 菜单标题
	Path        string `gorm:"size:200" json:"path"`                        // 菜单路径
	Component   string `gorm:"size:200" json:"component"`                   // 组件路径
	Icon        string `gorm:"size:100" json:"icon"`                        // 菜单图标
	Sort        int    `gorm:"default:0" json:"sort"`                       // 排序
	Type        int    `gorm:"default:1" json:"type"`                       // 菜单类型（1：目录，2：菜单，3：按钮）
	Status      int    `gorm:"default:1" json:"status"`                     // 菜单状态（1：正常，0：禁用）
	Permission  string `gorm:"size:200" json:"permission"`                  // 权限标识
}

// API 接口权限表
// 定义系统API接口的权限控制
type API struct {
	BaseModel
	Path        string `gorm:"size:500;not null" json:"path"`               // API路径
	Method      string `gorm:"size:10;not null" json:"method"`              // HTTP方法
	Description string `gorm:"size:500" json:"description"`                 // 接口描述
	Group       string `gorm:"size:100" json:"group"`                       // 接口分组
	Status      int    `gorm:"default:1" json:"status"`                     // 接口状态（1：正常，0：禁用）
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	BaseModel
	RoleID uint `gorm:"index;not null" json:"role_id"`                      // 角色ID
	MenuID uint `gorm:"index;not null" json:"menu_id"`                      // 菜单ID
}

// RoleAPI 角色接口关联表
type RoleAPI struct {
	BaseModel
	RoleID uint `gorm:"index;not null" json:"role_id"`                      // 角色ID
	APIID  uint `gorm:"index;not null" json:"api_id"`                       // 接口ID
}

// UserRole 用户角色关联表
type UserRole struct {
	BaseModel
	UserID uint `gorm:"index;not null" json:"user_id"`                      // 用户ID
	RoleID uint `gorm:"index;not null" json:"role_id"`                      // 角色ID
}

// TenantRole 租户角色关联表
type TenantRole struct {
	TenantModel
	RoleID uint `gorm:"index;not null" json:"role_id"`                      // 角色ID
}

// TableName 设置表名
func (Role) TableName() string {
	return "roles"
}

func (Menu) TableName() string {
	return "menus"
}

func (API) TableName() string {
	return "apis"
}

func (RoleMenu) TableName() string {
	return "role_menus"
}

func (RoleAPI) TableName() string {
	return "role_apis"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (TenantRole) TableName() string {
	return "tenant_roles"
}