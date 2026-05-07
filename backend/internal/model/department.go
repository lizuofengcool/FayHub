package model

type Department struct {
	SnowflakeTenantModel
	Name     string `gorm:"size:100;not null" json:"name"`
	ParentID int64  `gorm:"index;default:0" json:"parent_id"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"`
	LeaderID int64  `gorm:"default:0" json:"leader_id"`
	Children []Department `gorm:"-" json:"children,omitempty"`
}

type UserDepartment struct {
	SnowflakeTenantModel
	UserID int64 `gorm:"index;not null" json:"user_id"`
	DeptID int64 `gorm:"index;not null" json:"dept_id"`
}

type RoleDept struct {
	SnowflakeTenantModel
	RoleID int64 `gorm:"index;not null" json:"role_id"`
	DeptID int64 `gorm:"index;not null" json:"dept_id"`
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
