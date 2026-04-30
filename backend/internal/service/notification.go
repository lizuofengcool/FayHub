package service

import (
	"context"
	"encoding/json"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"fmt"
	"time"
)

type NotificationService struct{}

var NotificationServiceApp = &NotificationService{}

type NotificationFilters struct {
	Type     string
	Category string
	IsRead   *bool
}

type SendNotificationRequest struct {
	UserIDs    []uint        `json:"user_ids"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Type       string        `json:"type"`
	Category   string        `json:"category"`
	Data       interface{}   `json:"data"`
	SenderID   uint          `json:"sender_id"`
	SenderName string        `json:"sender_name"`
	Link       string        `json:"link"`
	Priority   int           `json:"priority"`
	ExpireIn   time.Duration `json:"expire_in"`
}

func (s *NotificationService) Send(ctx context.Context, req *SendNotificationRequest) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	if len(req.UserIDs) == 0 {
		return fmt.Errorf("接收用户不能为空")
	}

	if req.Type == "" {
		req.Type = model.NotifyTypeSystem
	}
	if req.Category == "" {
		req.Category = model.NotifyCategoryInfo
	}

	var dataJSON json.RawMessage
	if req.Data != nil {
		data, err := json.Marshal(req.Data)
		if err != nil {
			dataJSON = json.RawMessage(`{}`)
		} else {
			dataJSON = json.RawMessage(data)
		}
	}

	notifications := make([]*model.Notification, 0, len(req.UserIDs))
	now := time.Now()

	for _, userID := range req.UserIDs {
		n := &model.Notification{
			UserID:     userID,
			Title:      req.Title,
			Content:    req.Content,
			Type:       req.Type,
			Category:   req.Category,
			Data:       dataJSON,
			SenderID:   req.SenderID,
			SenderName: req.SenderName,
			Link:       req.Link,
			Priority:   req.Priority,
			IsRead:     false,
			CreatedAt:  &now,
		}

		if req.ExpireIn > 0 {
			expiredAt := now.Add(req.ExpireIn)
			n.ExpiredAt = &expiredAt
		}

		notifications = append(notifications, n)
	}

	if err := db.CreateInBatches(notifications, 100).Error; err != nil {
		return fmt.Errorf("创建通知失败: %w", err)
	}

	return nil
}

func (s *NotificationService) SendToTenant(ctx context.Context, tenantID uint, req *SendNotificationRequest) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var users []model.User
	if err := db.Where("tenant_id = ?", tenantID).Find(&users).Error; err != nil {
		return fmt.Errorf("查询租户用户失败: %w", err)
	}

	userIDs := make([]uint, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	req.UserIDs = userIDs
	return s.Send(ctx, req)
}

func (s *NotificationService) ListByUser(ctx context.Context, userID uint, filters *NotificationFilters, page, pageSize int) ([]*model.Notification, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	query := db.Model(&model.Notification{}).Where("user_id = ?", userID)

	if filters != nil {
		if filters.Type != "" {
			query = query.Where("type = ?", filters.Type)
		}
		if filters.Category != "" {
			query = query.Where("category = ?", filters.Category)
		}
		if filters.IsRead != nil {
			query = query.Where("is_read = ?", *filters.IsRead)
		}
	}

	query = query.Where("expired_at IS NULL OR expired_at > ?", time.Now())

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询通知总数失败: %w", err)
	}

	var notifications []*model.Notification
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("priority DESC, id DESC").Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("查询通知列表失败: %w", err)
	}

	return notifications, total, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID uint, notificationIDs []uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	now := time.Now()
	result := db.Model(&model.Notification{}).
		Where("user_id = ? AND id IN ?", userID, notificationIDs).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("标记已读失败: %w", result.Error)
	}

	return nil
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	now := time.Now()
	result := db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("标记全部已读失败: %w", result.Error)
	}

	return nil
}

func (s *NotificationService) Delete(ctx context.Context, userID uint, notificationIDs []uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	result := db.Where("user_id = ? AND id IN ?", userID, notificationIDs).Delete(&model.Notification{})
	if result.Error != nil {
		return fmt.Errorf("删除通知失败: %w", result.Error)
	}

	return nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	var count int64
	err := db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ? AND (expired_at IS NULL OR expired_at > ?)", userID, false, time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("查询未读数失败: %w", err)
	}

	return count, nil
}

func (s *NotificationService) CleanupExpired(ctx context.Context) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result := db.Where("expired_at IS NOT NULL AND expired_at < ?", time.Now()).Delete(&model.Notification{})
	if result.Error != nil {
		return 0, fmt.Errorf("清理过期通知失败: %w", result.Error)
	}

	return result.RowsAffected, nil
}
