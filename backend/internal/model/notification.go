package model

import (
	"encoding/json"
	"time"
)

type Notification struct {
	TenantModel
	UserID      int64            `json:"user_id" gorm:"index;not null"`
	Title       string          `json:"title" gorm:"size:200;not null"`
	Content     string          `json:"content" gorm:"type:text"`
	Type        string          `json:"type" gorm:"size:50;index;default:system"`
	Category    string          `json:"category" gorm:"size:50;index"`
	IsRead      bool            `json:"is_read" gorm:"default:false;index"`
	ReadAt      *time.Time      `json:"read_at"`
	Data        json.RawMessage `json:"data" gorm:"type:text"`
	SenderID    int64            `json:"sender_id"`
	SenderName  string          `json:"sender_name" gorm:"size:100"`
	Link        string          `json:"link" gorm:"size:500"`
	Priority    int             `json:"priority" gorm:"default:0"`
	ExpiredAt   *time.Time      `json:"expired_at"`
	CreatedAt   *time.Time      `json:"created_at" gorm:"index"`
}

const (
	NotifyTypeSystem    = "system"
	NotifyTypeSecurity  = "security"
	NotifyTypePlugin    = "plugin"
	NotifyTypePayment   = "payment"
	NotifyTypeWebhook   = "webhook"
	NotifyTypeAudit     = "audit"
	NotifyTypeAlert     = "alert"
)

const (
	NotifyCategoryInfo     = "info"
	NotifyCategoryWarning  = "warning"
	NotifyCategoryError    = "error"
	NotifyCategorySuccess  = "success"
	NotifyCategoryCritical = "critical"
)

func (Notification) TableName() string {
	return "notifications"
}
