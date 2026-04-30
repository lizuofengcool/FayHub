package model

import "time"

type SSOAuthorizationCode struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	TenantID  uint      `gorm:"not null" json:"tenant_id"`
	Username  string    `gorm:"size:100;not null" json:"username"`
	Role      string    `gorm:"size:50;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

type SSOTokenData struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Token     string    `gorm:"size:64;uniqueIndex;not null" json:"token"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	TenantID  uint      `gorm:"not null" json:"tenant_id"`
	Username  string    `gorm:"size:100;not null" json:"username"`
	Role      string    `gorm:"size:50;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

func (SSOAuthorizationCode) TableName() string { return "sso_auth_codes" }
func (SSOTokenData) TableName() string         { return "sso_tokens" }
