package model

type Department struct {
	TenantModel
	Name     string `gorm:"size:100;not null" json:"name"`
	ParentID uint   `gorm:"index;default:0" json:"parent_id"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"`
	LeaderID uint   `gorm:"default:0" json:"leader_id"`
	Children []Department `gorm:"-" json:"children,omitempty"`
}

type UserDepartment struct {
	TenantModel
	UserID uint `gorm:"index;not null" json:"user_id"`
	DeptID uint `gorm:"index;not null" json:"dept_id"`
}

type RoleDept struct {
	TenantModel
	RoleID uint `gorm:"index;not null" json:"role_id"`
	DeptID uint `gorm:"index;not null" json:"dept_id"`
}

func (Department) TableName() string {
	return "departments"
}

func (UserDepartment) TableName() string {
	return "user_departments"
}

func (RoleDept) TableName() string {
	return "role_depts"
}
