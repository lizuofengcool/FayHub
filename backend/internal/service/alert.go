package service

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/metrics"
	"fayhub/pkg/utils"
	"time"
)

type AlertService struct{}

type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

type AlertRule struct {
	ID          string
	Name        string
	Level       AlertLevel
	Description string
	Evaluate    func(ctx context.Context) (bool, string)
	Enabled     bool
}

var alertRules = []*AlertRule{}

func (s *AlertService) RegisterRule(rule *AlertRule) {
	alertRules = append(alertRules, rule)
}

func (s *AlertService) CheckAndAlert(ctx context.Context) {
	for _, rule := range alertRules {
		if !rule.Enabled {
			continue
		}

		triggered, message := rule.Evaluate(ctx)
		if triggered {
			s.SendAlert(ctx, rule, message)
		}
	}
}

func (s *AlertService) SendAlert(ctx context.Context, rule *AlertRule, message string) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	now := time.Now()
	alert := &model.Notification{
		Title:     rule.Name,
		Content:   message,
		Type:      model.NotifyTypeAlert,
		Category:  getCategoryByLevel(rule.Level),
		Priority:  getPriorityByLevel(rule.Level),
		IsRead:    false,
		CreatedAt: &now,
	}

	db.Create(alert)
}

func (s *AlertService) CheckSystemResources(ctx context.Context) {
	metricsData := metrics.GetMetrics()

	memoryAllocMb, _ := metricsData["memory_alloc_mb"].(float64)

	if memoryAllocMb > 2048 {
		s.SendAlert(ctx, &AlertRule{
			Name:        "内存使用率过高",
			Level:       AlertLevelWarning,
			Description: "系统内存使用超过阈值",
		}, "系统内存使用量过高")
	}
}

func getCategoryByLevel(level AlertLevel) string {
	switch level {
	case AlertLevelCritical:
		return model.NotifyCategoryCritical
	case AlertLevelError:
		return model.NotifyCategoryError
	case AlertLevelWarning:
		return model.NotifyCategoryWarning
	default:
		return model.NotifyCategoryInfo
	}
}

func getPriorityByLevel(level AlertLevel) int {
	switch level {
	case AlertLevelCritical:
		return 4
	case AlertLevelError:
		return 3
	case AlertLevelWarning:
		return 2
	default:
		return 1
	}
}

func (s *AlertService) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.CheckSystemResources(ctx)
			s.CheckAndAlert(ctx)
		}
	}
}
