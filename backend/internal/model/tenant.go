package model

// Tenant 租户表/商家表
type Tenant struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"` // 租户名称
	Domain      string `gorm:"size:200;unique" json:"domain"` // 租户域名
	Description string `gorm:"size:500" json:"description"`   // 租户描述
	Status      int    `gorm:"default:1" json:"status"`       // 租户状态（1：正常，0：禁用）
	ExpiredAt   int64  `gorm:"default:0" json:"expired_at"`   // 租户过期时间
}

// User 系统用户表
// 继承TenantModel，包含tenant_id字段，实现用户级多租户隔离
// super_admin的tenant_id为0（平台级用户），租户用户的tenant_id为所属租户ID
type User struct {
	TenantModel
	Username       string `gorm:"size:100;index:idx_tenant_username;not null" json:"username"`
	Password       string `gorm:"size:200;not null" json:"-"`
	Email          string `gorm:"size:200" json:"email"`
	Phone          string `gorm:"size:20" json:"phone"`
	Status         int    `gorm:"default:1" json:"status"`
	Role           string `gorm:"size:50;default:'tenant_user'" json:"role"`
	LastLoginAt    int64  `gorm:"default:0" json:"last_login_at"`
	LoginIP        string `gorm:"size:50" json:"login_ip"`
	Avatar         string `gorm:"size:500" json:"avatar"`
	RealName       string `gorm:"size:100" json:"real_name"`
	LoginFailCount int    `gorm:"default:0" json:"login_fail_count"`
	LockedUntil    int64  `gorm:"default:0" json:"locked_until"`
}

// TenantUser 商家员工表
type TenantUser struct {
	TenantModel
	UserID   uint   `gorm:"index;not null" json:"user_id"` // 用户ID
	Name     string `gorm:"size:100;not null" json:"name"` // 员工姓名
	Position string `gorm:"size:100" json:"position"`      // 员工职位
	Status   int    `gorm:"default:1" json:"status"`       // 员工状态（1：正常，0：禁用）
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
