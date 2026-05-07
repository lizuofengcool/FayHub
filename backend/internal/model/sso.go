package model

import "time"

type SSOAuthorizationCode struct {
	SnowflakeTenantModel
	Code     string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	UserID   int64     `gorm:"not null" json:"user_id"`
	Username string    `gorm:"size:100;not null" json:"username"`
	Role     string    `gorm:"size:50;not null" json:"role"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

type SSOTokenData struct {
	SnowflakeTenantModel
	Token    string    `gorm:"size:64;uniqueIndex;not null" json:"token"`
	UserID   int64     `gorm:"not null" json:"user_id"`
	Username string    `gorm:"size:100;not null" json:"username"`
	Role     string    `gorm:"size:50;not null" json:"role"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

func (SSOAuthorizationCode) TableName() string { return "sso_auth_codes" }
func (SSOTokenData) TableName() string         { return "sso_tokens" }
