package model

import (
	"time"
)

type Subscription struct {
	ID             int64       `gorm:"primaryKey" json:"id"`
	TenantID       int64       `gorm:"uniqueIndex;not null" json:"tenant_id"`
	PackageID      int64       `gorm:"not null" json:"package_id"`
	PackageName    string     `gorm:"size:100" json:"package_name"`
	Status         string     `gorm:"size:20;default:'active'" json:"status"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	TrialEndDate   *time.Time `json:"trial_end_date"`
	AutoRenew      int        `gorm:"default:0" json:"auto_renew"`
	MaxUsers       int        `gorm:"default:10" json:"max_users"`
	MaxStorage     int64      `gorm:"default:1073741824" json:"max_storage"`
	CurrentUsers   int        `gorm:"default:0" json:"current_users"`
	CurrentStorage int64      `gorm:"default:0" json:"current_storage"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type SubscriptionInvoice struct {
	ID             int64       `gorm:"primaryKey" json:"id"`
	SubscriptionID int64       `gorm:"index;not null" json:"subscription_id"`
	TenantID       int64       `gorm:"index;not null" json:"tenant_id"`
	Amount         float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency       string     `gorm:"size:10;default:'CNY'" json:"currency"`
	Status         string     `gorm:"size:20;default:'pending'" json:"status"`
	BillingPeriod  string     `gorm:"size:20" json:"billing_period"`
	PaidAt         *time.Time `json:"paid_at"`
	DueDate        time.Time  `json:"due_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (SubscriptionInvoice) TableName() string {
	return "subscription_invoices"
}
