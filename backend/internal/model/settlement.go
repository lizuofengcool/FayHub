package model

import "time"

type SettlementRecord struct {
	TenantModel
	OrderNo        string     `gorm:"size:50;not null;index" json:"order_no"`
	PaymentOrderID uint       `gorm:"not null;index" json:"payment_order_id"`
	MarketOrderID  string     `gorm:"size:50;index" json:"market_order_id"`
	TotalAmount    int64      `gorm:"not null" json:"total_amount"`
	PlatformAmount int64      `gorm:"not null" json:"platform_amount"`
	TenantAmount   int64      `gorm:"not null" json:"tenant_amount"`
	PlatformRate   int        `gorm:"not null" json:"platform_rate"`
	Status         string     `gorm:"size:20;not null;index" json:"status"`
	SettledAt      *time.Time `json:"settled_at"`
	SettlementNo   string     `gorm:"size:50;unique" json:"settlement_no"`
	FailReason     string     `gorm:"size:500" json:"fail_reason"`
}

func (SettlementRecord) TableName() string {
	return "settlement_records"
}

const (
	SettlementStatusPending = "pending"
	SettlementStatusSettled = "settled"
	SettlementStatusFailed  = "failed"
)

type SettlementConfig struct {
	TenantModel
	PlatformRate int        `gorm:"not null" json:"platform_rate"`
	MinAmount    int64      `gorm:"default:0" json:"min_amount"`
	Status       int        `gorm:"default:1" json:"status"`
}

func (SettlementConfig) TableName() string {
	return "settlement_configs"
}
