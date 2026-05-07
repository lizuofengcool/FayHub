package service

import (
	"bytes"
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/logger"
	"fayhub/pkg/utils"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookService struct{}

var WebhookServiceApp = new(WebhookService)

var (
	eventSubsCache sync.Map
)

func (s *WebhookService) CreateSubscription(ctx context.Context, sub *model.WebhookSubscription) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Create(sub).Error
}

func (s *WebhookService) UpdateSubscription(ctx context.Context, id int64, updates map[string]interface{}) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Model(&model.WebhookSubscription{}).Where("id = ?", id).Updates(updates).Error
}

func (s *WebhookService) DeleteSubscription(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Delete(&model.WebhookSubscription{}, id).Error
}

func (s *WebhookService) GetSubscription(ctx context.Context, id int64) (*model.WebhookSubscription, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	var sub model.WebhookSubscription
	if err := db.First(&sub, id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *WebhookService) ListSubscriptions(ctx context.Context, event string, page, pageSize int) ([]*model.WebhookSubscription, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.WebhookSubscription{})
	if event != "" {
		eventPattern := fmt.Sprintf(`%%"%s"%%`, event)
		query = query.Where("events LIKE ?", eventPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var subs []*model.WebhookSubscription
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

func (s *WebhookService) ListDeliveries(ctx context.Context, subscriptionID int64, status string, page, pageSize int) ([]*model.WebhookDelivery, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.WebhookDelivery{})
	if subscriptionID > 0 {
		query = query.Where("subscription_id = ?", subscriptionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var deliveries []*model.WebhookDelivery
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}

	return deliveries, total, nil
}

func (s *WebhookService) PublishEvent(ctx context.Context, event *model.WebhookEvent) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var subs []model.WebhookSubscription
	eventPattern := fmt.Sprintf(`%%"%s"%%`, event.Event)
	if err := db.Where("is_active = ? AND events LIKE ?", true, eventPattern).Find(&subs).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询订阅失败")
	}

	if len(subs) == 0 {
		return nil
	}

	payload, err := model.BuildWebhookPayload(event)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, "构建Webhook负载失败")
	}

	for i := range subs {
		delivery := &model.WebhookDelivery{
			SubscriptionID: subs[i].ID,
			Event:          event.Event,
			Payload:        payload,
			Status:         "pending",
			Attempts:       0,
		}

		delivery.TenantID = event.TenantID

		if err := db.Create(delivery).Error; err != nil {
			logger.Error(ctx, "创建Webhook投递记录失败",
				zap.Uint("subscription_id", subs[i].ID),
				zap.String("event", event.Event),
				zap.Error(err),
			)
			continue
		}

		go s.deliverWebhook(delivery, &subs[i], payload)
	}

	return nil
}

func (s *WebhookService) deliverWebhook(delivery *model.WebhookDelivery, sub *model.WebhookSubscription, payload []byte) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(context.Background(), "Webhook投递panic", zap.Any("error", r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db := utils.GetDB(ctx)

	maxRetries := sub.RetryCount
	if maxRetries <= 0 {
		maxRetries = 3
	}

	timeout := time.Duration(sub.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := model.GetNextRetryDelay(attempt - 1)
			nextRetry := time.Now().Add(delay)
			if db != nil {
				db.Model(delivery).Updates(map[string]interface{}{
					"status":        "retrying",
					"next_retry_at": nextRetry,
				})
			}
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}

		success, statusCode, responseBody := s.sendWebhookRequest(sub.URL, payload, sub.Secret, timeout)

		now := time.Now()
		delivery.Attempts = attempt + 1
		delivery.LastAttemptAt = &now
		delivery.LastStatusCode = statusCode
		if len(responseBody) > 1000 {
			responseBody = responseBody[:1000]
		}
		delivery.LastResponse = responseBody

		if success {
			delivery.Status = "delivered"
			if db != nil {
				db.Model(delivery).Updates(map[string]interface{}{
					"status":           "delivered",
					"attempts":         delivery.Attempts,
					"last_attempt_at":  now,
					"last_status_code": statusCode,
					"last_response":    responseBody,
				})
				db.Model(sub).Updates(map[string]interface{}{
					"last_success": now,
					"fail_count":   0,
				})
			}
			logger.Info(ctx, "Webhook投递成功",
				zap.Uint("subscription_id", sub.ID),
				zap.String("event", delivery.Event),
				zap.Int("attempts", delivery.Attempts),
			)
			return
		}

		logger.Warn(ctx, "Webhook投递失败",
			zap.Uint("subscription_id", sub.ID),
			zap.String("event", delivery.Event),
			zap.Int("attempt", attempt+1),
			zap.Int("status_code", statusCode),
		)
	}

	delivery.Status = "failed"
	now := time.Now()
	if db != nil {
		db.Model(delivery).Updates(map[string]interface{}{
			"status":           "failed",
			"attempts":         delivery.Attempts,
			"last_attempt_at":  now,
			"last_status_code": delivery.LastStatusCode,
			"last_response":    delivery.LastResponse,
		})
		db.Model(sub).Updates(map[string]interface{}{
			"last_failure": now,
			"fail_count":   gorm.Expr("fail_count + 1"),
		})
	}
}

func (s *WebhookService) sendWebhookRequest(url string, payload []byte, secret string, timeout time.Duration) (bool, int, string) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return false, 0, fmt.Sprintf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FayHub-Webhook/1.0")
	req.Header.Set("X-FayHub-Event", "webhook")

	if secret != "" {
		sig := model.BuildWebhookSignatureHeader(payload, secret)
		req.Header.Set("X-FayHub-Signature", sig)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, resp.StatusCode, responseBody
	}

	return false, resp.StatusCode, responseBody
}

func (s *WebhookService) Redeliver(ctx context.Context, deliveryID int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var delivery model.WebhookDelivery
	if err := db.First(&delivery, deliveryID).Error; err != nil {
		return errs.NewServiceError(errs.ErrResourceNotFound, "投递记录不存在")
	}

	var sub model.WebhookSubscription
	if err := db.First(&sub, delivery.SubscriptionID).Error; err != nil {
		return errs.NewServiceError(errs.ErrResourceNotFound, "订阅不存在")
	}

	delivery.Status = "pending"
	delivery.Attempts = 0
	db.Save(&delivery)

	go s.deliverWebhook(&delivery, &sub, delivery.Payload)

	return nil
}

func (s *WebhookService) GetDeliveryStats(ctx context.Context, subscriptionID int64) (map[string]int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	stats := make(map[string]int64)
	statuses := []string{"pending", "delivered", "retrying", "failed"}

	for _, status := range statuses {
		var count int64
		query := db.Model(&model.WebhookDelivery{}).Where("status = ?", status)
		if subscriptionID > 0 {
			query = query.Where("subscription_id = ?", subscriptionID)
		}
		query.Count(&count)
		stats[status] = count
	}

	return stats, nil
}

func (s *WebhookService) TestDelivery(ctx context.Context, subscriptionID int64) (*model.WebhookDelivery, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var sub model.WebhookSubscription
	if err := db.First(&sub, subscriptionID).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrResourceNotFound, "订阅不存在")
	}

	testPayload := []byte(`{"event":"test","timestamp":"` + time.Now().Format(time.RFC3339) + `","data":{"message":"This is a test delivery from FayHub"}}`)

	delivery := &model.WebhookDelivery{
		SubscriptionID: sub.ID,
		Event:          "test",
		Payload:        testPayload,
		Status:         "pending",
		Attempts:       0,
	}
	delivery.TenantID = sub.TenantID

	if err := db.Create(delivery).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建测试投递记录失败")
	}

	go s.deliverWebhook(delivery, &sub, testPayload)

	return delivery, nil
}
