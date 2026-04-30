package model

// Role 角色表
// 用于RBAC权限管理，定义平台和租户的角色
// 平台角色：super_admin（超级管理员）、platform_admin（平台管理员）
// 租户角色：tenant_admin（租户管理员）、tenant_user（租户用户）
type Role struct {
	TenantModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	Type        int    `gorm:"default:1" json:"type"`
	Status      int    `gorm:"default:1" json:"status"`
	DataScope   int    `gorm:"default:1" json:"data_scope"`
	DeptID      uint   `gorm:"default:0" json:"dept_id"`
}

const (
	DataScopeAll       = 1
	DataScopeDept      = 2
	DataScopeDeptAndSub = 3
	DataScopeSelf      = 4
	DataScopeCustom    = 5
)

// Menu 菜单表
// 定义系统菜单和权限点
// 平台菜单：系统设置、租户管理、用户管理等
// 租户菜单：业务功能菜单
type Menu struct {
	BaseModel
	ParentID   uint   `gorm:"index" json:"parent_id"`
	Title      string `gorm:"size:100;not null" json:"title"`
	Path       string `gorm:"size:200" json:"path"`
	Component  string `gorm:"size:200" json:"component"`
	Icon       string `gorm:"size:100" json:"icon"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Type       int    `gorm:"default:1" json:"type"`
	Status     int    `gorm:"default:1" json:"status"`
	Permission string `gorm:"size:200" json:"permission"`
	Layout     string `gorm:"size:20;default:embedded" json:"layout"`
	Children   []Menu `gorm:"-" json:"children,omitempty"`
}

// API 接口权限表
// 定义系统API接口的权限控制
type API struct {
	BaseModel
	Path        string `gorm:"size:500;not null" json:"path"`  // API路径
	Method      string `gorm:"size:10;not null" json:"method"` // HTTP方法
	Description string `gorm:"size:500" json:"description"`    // 接口描述
	Group       string `gorm:"size:100" json:"group"`          // 接口分组
	Status      int    `gorm:"default:1" json:"status"`        // 接口状态（1：正常，0：禁用）
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	TenantModel
	RoleID uint `gorm:"index;not null" json:"role_id"`
	MenuID uint `gorm:"index;not null" json:"menu_id"`
}

type RoleAPI struct {
	TenantModel
	RoleID uint `gorm:"index;not null" json:"role_id"`
	APIID  uint `gorm:"index;not null" json:"api_id"`
}

type UserRole struct {
	TenantModel
	UserID uint `gorm:"index;not null" json:"user_id"`
	RoleID uint `gorm:"index;not null" json:"role_id"`
}

// TenantRole 租户角色关联表
type TenantRole struct {
	TenantModel
	RoleID uint `gorm:"index;not null" json:"role_id"` // 角色ID
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
