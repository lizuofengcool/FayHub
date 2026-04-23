package model

// Tenant 租户表/商家表
type Tenant struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`         // 租户名称
	Domain      string `gorm:"size:200;unique" json:"domain"`        // 租户域名
	Description string `gorm:"size:500" json:"description"`          // 租户描述
	Status      int    `gorm:"default:1" json:"status"`               // 租户状态（1：正常，0：禁用）
	ExpiredAt   int64  `gorm:"default:0" json:"expired_at"`          // 租户过期时间
}

// User 系统总管理员表
type User struct {
	BaseModel
	Username string `gorm:"size:100;unique;not null" json:"username"` // 用户名
	Password string `gorm:"size:200;not null" json:"password"`        // 密码
	Email    string `gorm:"size:200;unique" json:"email"`             // 邮箱
	Phone    string `gorm:"size:20;unique" json:"phone"`              // 手机号
	Status   int    `gorm:"default:1" json:"status"`                  // 用户状态（1：正常，0：禁用）
	Role     string `gorm:"size:50;default:'admin'" json:"role"`      // 用户角色
}

// TenantUser 商家员工表
type TenantUser struct {
	TenantModel
	UserID   uint   `gorm:"index;not null" json:"user_id"`         // 用户ID
	Name     string `gorm:"size:100;not null" json:"name"`         // 员工姓名
	Position string `gorm:"size:100" json:"position"`              // 员工职位
	Status   int    `gorm:"default:1" json:"status"`               // 员工状态（1：正常，0：禁用）
}

// TableName 设置表名
func (Tenant) TableName() string {
	return "tenants"
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

// TableName 设置表名
func (TenantUser) TableName() string {
	return "tenant_users"
}