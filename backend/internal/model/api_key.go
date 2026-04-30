package model

import "time"

type APIKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	KeyHash     string    `gorm:"size:64;unique;not null" json:"-"` // SHA256哈希
	KeyPrefix   string    `gorm:"size:8;not null" json:"key_prefix"` // 密钥前缀用于显示
	Secret      string    `gorm:"-" json:"secret,omitempty"` // 仅在创建时返回明文密钥
	Permissions string    `gorm:"type:text" json:"permissions"` // JSON格式的权限列表
	RateLimit   int       `gorm:"default:1000" json:"rate_limit"` // 每小时请求限制
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	Status      int       `gorm:"default:1;index" json:"status"` // 1=启用, 0=禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
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
