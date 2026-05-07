package model

import "time"

type NotificationChannel struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TenantID  int64      `gorm:"index;default:0" json:"tenant_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	Provider  string    `gorm:"size:50;not null" json:"provider"`
	Config    string    `gorm:"type:text" json:"config"`
	Status    int       `gorm:"default:1" json:"status"`
	IsDefault int       `gorm:"default:0" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (NotificationChannel) TableName() string {
	return "notification_channels"
}

type NotificationTemplate struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TenantID  int64      `gorm:"index;default:0" json:"tenant_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Code      string    `gorm:"size:50;not null" json:"code"`
	ChannelID int64      `gorm:"not null" json:"channel_id"`
	Subject   string    `gorm:"size:200" json:"subject"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

type NotificationRecord struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	TenantID   int64      `gorm:"index;default:0" json:"tenant_id"`
	ChannelID  int64      `gorm:"not null" json:"channel_id"`
	TemplateID int64      `gorm:"default:0" json:"template_id"`
	Recipient  string    `gorm:"size:200;not null" json:"recipient"`
	Subject    string    `gorm:"size:200" json:"subject"`
	Content    string    `gorm:"type:text" json:"content"`
	Status     string    `gorm:"size:20;default:'pending'" json:"status"`
	Error      string    `gorm:"type:text" json:"error"`
	SentAt     *time.Time `json:"sent_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (NotificationRecord) TableName() string {
	return "notification_records"
}
