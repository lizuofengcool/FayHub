package service

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/logger"
	"fayhub/pkg/utils"
	"time"

	"go.uber.org/zap"
)

type SecurityEventService struct{}

var SecurityEventServiceApp = &SecurityEventService{}

type SecurityEventType string

const (
	SecurityEventLoginFailed       SecurityEventType = "login_failed"
	SecurityEventCrossTenantAccess SecurityEventType = "cross_tenant_access"
	SecurityEventAPIKeyAbuse       SecurityEventType = "api_key_abuse"
	SecurityEventPaymentFraud      SecurityEventType = "payment_fraud"
	SecurityEventPluginViolation   SecurityEventType = "plugin_violation"
	SecurityEventRateLimitExceeded SecurityEventType = "rate_limit_exceeded"
	SecurityEventSuspiciousIP      SecurityEventType = "suspicious_ip"
)

type SecurityEvent struct {
	Type        SecurityEventType      `json:"type"`
	TenantID    int64                  `json:"tenant_id"`
	UserID      int64                  `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	IP          string                 `json:"ip"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Severity    string                 `json:"severity"` // low, medium, high, critical
}

func (s *SecurityEventService) RecordSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	logger.Info(ctx, "安全事件记录",
		zap.String("type", string(event.Type)),
		zap.Int64("tenant_id", event.TenantID),
		zap.Int64("user_id", event.UserID),
		zap.String("username", event.Username),
		zap.String("ip", event.IP),
		zap.String("description", event.Description),
		zap.String("severity", event.Severity),
		zap.Any("details", event.Details),
		zap.Time("timestamp", time.Now()))

	if event.Severity == "high" || event.Severity == "critical" {
		go s.sendSecurityAlert(ctx, event)
	}

	return nil
}

func (s *SecurityEventService) sendSecurityAlert(ctx context.Context, event *SecurityEvent) {
	notificationReq := &SendNotificationRequest{
		Title:    s.getAlertTitle(event),
		Content:  s.getAlertContent(event),
		Type:     model.NotifyTypeSecurity,
		Category: model.NotifyCategoryWarning,
		Priority: s.getAlertPriority(event),
		Data:     event.Details,
		ExpireIn: 24 * time.Hour,
	}

	if event.TenantID > 0 {
		admins, err := s.getTenantAdmins(ctx, event.TenantID)
		if err != nil {
			logger.Error(ctx, "获取租户管理员失败", zap.Error(err))
			return
		}

		var adminIDs []int64
		for _, admin := range admins {
			adminIDs = append(adminIDs, admin.ID)
		}

		notificationReq.UserIDs = adminIDs
		notificationReq.SenderName = "安全监控系统"

		if err := NotificationServiceApp.Send(ctx, notificationReq); err != nil {
			logger.Error(ctx, "发送安全告警失败", zap.Error(err))
		}
	}
}

func (s *SecurityEventService) getAlertTitle(event *SecurityEvent) string {
	switch event.Type {
	case SecurityEventLoginFailed:
		return "登录失败告警"
	case SecurityEventCrossTenantAccess:
		return "跨租户访问告警"
	case SecurityEventAPIKeyAbuse:
		return "API密钥滥用告警"
	case SecurityEventPaymentFraud:
		return "支付欺诈告警"
	case SecurityEventPluginViolation:
		return "插件违规告警"
	case SecurityEventRateLimitExceeded:
		return "限流超限告警"
	case SecurityEventSuspiciousIP:
		return "可疑IP告警"
	default:
		return "安全事件告警"
	}
}

func (s *SecurityEventService) getAlertContent(event *SecurityEvent) string {
	return event.Description
}

func (s *SecurityEventService) getAlertPriority(event *SecurityEvent) int {
	switch event.Severity {
	case "critical":
		return 10
	case "high":
		return 8
	case "medium":
		return 5
	default:
		return 3
	}
}

func (s *SecurityEventService) getTenantAdmins(ctx context.Context, tenantID int64) ([]model.User, error) {
	var admins []model.User
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, nil
	}

	err := db.Where("tenant_id = ? AND role IN ?", tenantID, []string{"admin", "super_admin"}).
		Where("status = 1").
		Find(&admins).Error
	if err != nil {
		return nil, err
	}

	return admins, nil
}
