package model

import (
	"fayhub/pkg/crypto"
	"time"
)

type PaymentConfig struct {
	TenantModel
	Channel     string `gorm:"type:varchar(20);not null;uniqueIndex:idx_tenant_channel" json:"channel"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
	MchID       string `gorm:"type:varchar(64)" json:"mch_id"`
	AppID       string `gorm:"type:varchar(64)" json:"app_id"`
	APIKey      string `gorm:"type:varchar(512)" json:"-"`
	PrivateKey  string `gorm:"type:text" json:"-"`
	PublicKey   string `gorm:"type:text" json:"-"`
	SerialNo    string `gorm:"type:varchar(64)" json:"serial_no"`
	NotifyURL   string `gorm:"type:varchar(256)" json:"notify_url"`
	Sandbox     bool   `gorm:"default:false" json:"sandbox"`
	ExtraConfig string `gorm:"type:text" json:"extra_config"`
}

func (c *PaymentConfig) AfterFind() error {
	if c.APIKey != "" {
		c.APIKey = crypto.DecryptField(c.APIKey)
	}
	if c.PrivateKey != "" {
		c.PrivateKey = crypto.DecryptField(c.PrivateKey)
	}
	if c.PublicKey != "" {
		c.PublicKey = crypto.DecryptField(c.PublicKey)
	}
	return nil
}

func (c *PaymentConfig) BeforeSave() error {
	if c.APIKey != "" && !crypto.IsEncrypted(c.APIKey) {
		c.APIKey = crypto.EncryptField(c.APIKey)
	}
	if c.PrivateKey != "" && !crypto.IsEncrypted(c.PrivateKey) {
		c.PrivateKey = crypto.EncryptField(c.PrivateKey)
	}
	if c.PublicKey != "" && !crypto.IsEncrypted(c.PublicKey) {
		c.PublicKey = crypto.EncryptField(c.PublicKey)
	}
	return nil
}

func (PaymentConfig) TableName() string { return "payment_configs" }

type PaymentOrder struct {
	TenantModel
	OrderNo      string     `gorm:"type:varchar(64);uniqueIndex" json:"order_no"`
	OutTradeNo   string     `gorm:"type:varchar(64);index" json:"out_trade_no"`
	Channel      string     `gorm:"type:varchar(20);not null;index" json:"channel"`
	Status       int        `gorm:"type:smallint;not null;default:0;index" json:"status"`
	Amount       int64      `gorm:"type:bigint;not null" json:"amount"`
	Currency     string     `gorm:"type:varchar(10);default:'CNY'" json:"currency"`
	Subject      string     `gorm:"type:varchar(256)" json:"subject"`
	Description  string     `gorm:"type:varchar(512)" json:"description"`
	UserID       uint       `gorm:"index" json:"user_id"`
	PluginID     string     `gorm:"type:varchar(128);index" json:"plugin_id"`
	PaidAt       *time.Time `json:"paid_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	NotifyData   string     `gorm:"type:text" json:"notify_data"`
	RefundStatus int        `gorm:"type:smallint;default:0" json:"refund_status"`
	RefundAmount int64      `gorm:"type:bigint;default:0" json:"refund_amount"`
}

func (PaymentOrder) TableName() string { return "payment_orders" }

const (
	PaymentStatusPending   = 0
	PaymentStatusPaid      = 1
	PaymentStatusFailed    = 2
	PaymentStatusClosed    = 3
	PaymentStatusRefunding = 4
	PaymentStatusRefunded  = 5

	PaymentChannelWechat = "wechat"
	PaymentChannelAlipay = "alipay"

	RefundStatusNone     = 0
	RefundStatusPartial  = 1
	RefundStatusFull     = 2
)
