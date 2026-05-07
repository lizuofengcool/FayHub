package model

import "time"

type LoginLog struct {
	SnowflakeModel
	UserID        int64      `json:"user_id" gorm:"index"`
	Username      string     `json:"username" gorm:"size:100;index"`
	TenantID      int64      `json:"tenant_id" gorm:"index"`
	LoginStatus   string     `json:"login_status" gorm:"size:20;index;not null"`
	LoginIP       string     `json:"login_ip" gorm:"size:128"`
	LoginLocation string     `json:"login_location" gorm:"size:255"`
	Browser       string     `json:"browser" gorm:"size:200"`
	OS            string     `json:"os" gorm:"size:200"`
	LoginTime     time.Time  `json:"login_time" gorm:"index"`
	LogoutTime    *time.Time `json:"logout_time"`
	Msg           string     `json:"msg" gorm:"size:500"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
