package model

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type WebhookSubscription struct {
	TenantModel
	Name        string     `json:"name" gorm:"size:200;not null"`
	URL         string     `json:"url" gorm:"size:500;not null"`
	Secret      string     `json:"secret,omitempty" gorm:"size:200" fayhub:"encrypt"`
	Events      StringList `json:"events" gorm:"type:text"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	RetryCount  int        `json:"retry_count" gorm:"default:3"`
	TimeoutSec  int        `json:"timeout_sec" gorm:"default:10"`
	LastSuccess *time.Time `json:"last_success,omitempty"`
	LastFailure *time.Time `json:"last_failure,omitempty"`
	FailCount   int        `json:"fail_count" gorm:"default:0"`
}

func (WebhookSubscription) TableName() string {
	return "webhook_subscriptions"
}

type WebhookDelivery struct {
	TenantModel
	SubscriptionID uint            `json:"subscription_id" gorm:"index"`
	Event          string          `json:"event" gorm:"size:100;index"`
	Payload        json.RawMessage `json:"payload" gorm:"type:text"`
	Status         string          `json:"status" gorm:"size:20;default:pending"`
	Attempts       int             `json:"attempts" gorm:"default:0"`
	NextRetryAt    *time.Time      `json:"next_retry_at,omitempty"`
	LastAttemptAt  *time.Time      `json:"last_attempt_at,omitempty"`
	LastResponse   string          `json:"last_response,omitempty" gorm:"size:1000"`
	LastStatusCode int             `json:"last_status_code,omitempty"`
}

func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

type WebhookEvent struct {
	Event     string      `json:"event"`
	TenantID  uint        `json:"tenant_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type StringList []string

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var list []string
	if err := json.Unmarshal(value.([]byte), &list); err != nil {
		return err
	}
	*s = list
	return nil
}

func (s StringList) Value() (interface{}, error) {
	if s == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SignPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(payload []byte, secret string, signature string) bool {
	expected := SignPayload(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func BuildWebhookPayload(event *WebhookEvent) ([]byte, error) {
	wrapper := map[string]interface{}{
		"event":     event.Event,
		"tenant_id": event.TenantID,
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"data":      event.Data,
	}
	return json.Marshal(wrapper)
}

func GetNextRetryDelay(attempt int) time.Duration {
	delays := []time.Duration{
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
	}
	if attempt >= len(delays) {
		return 1 * time.Hour
	}
	return delays[attempt]
}

func BuildWebhookSignatureHeader(payload []byte, secret string) string {
	return fmt.Sprintf("sha256=%s", SignPayload(payload, secret))
}
