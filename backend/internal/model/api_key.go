package model

import "time"

type APIKey struct {
	SnowflakeTenantModel
	UserID      int64       `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	KeyHash     string     `gorm:"size:64;unique;not null" json:"-"`
	KeyPrefix   string     `gorm:"size:8;not null" json:"key_prefix"`
	Secret      string     `gorm:"-" json:"secret,omitempty"`
	Permissions string     `gorm:"type:text" json:"permissions"`
	RateLimit   int        `gorm:"default:1000" json:"rate_limit"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	Status      int        `gorm:"default:1;index" json:"status"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

type APIKeyPermission struct {
	Resource string `json:"resource"` // 资源类型：plugin, payment, user, etc.
	Action   string `json:"action"`   // 操作：read, write, delete
}

func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

func (k *APIKey) IsActive() bool {
	return k.Status == 1 && !k.IsExpired()
}
