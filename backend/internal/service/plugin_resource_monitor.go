package service

import (
	"context"
	"log"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/plugin"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PluginResourceMonitorService struct{}

func (s *PluginResourceMonitorService) Init() {
	monitor := plugin.GetResourceMonitor()
	monitor.SetPersister(s.persistStats)
	monitor.SetAlertHandler(s.handleAlert)
	log.Println("[PluginResourceMonitor] 持久化与告警回调已注册")
}

func (s *PluginResourceMonitorService) persistStats(stats *plugin.PluginRuntimeStats) {
	ctx := context.Background()
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	record := model.PluginResourceStats{
		PluginID:        stats.PluginID,
		CallCount:       stats.CallCount,
		ErrorCount:      stats.ErrorCount,
		TotalDurationMs: stats.TotalDurationMs,
		MaxDurationMs:   stats.MaxDurationMs,
		LastCallAt:      stats.LastCallAt,
		LastErrorAt:     stats.LastErrorAt,
		LastErrorMsg:    stats.LastErrorMsg,
		MemoryUsageKB:   stats.MemoryUsageKB,
		Status:          stats.Status,
	}

	if stats.TenantID > 0 {
		record.TenantID = stats.TenantID
	}

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plugin_id"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"call_count", "error_count", "total_duration_ms", "max_duration_ms",
			"last_call_at", "last_error_at", "last_error_msg", "memory_usage_kb",
			"status", "updated_at",
		}),
	}).Create(&record)

	if result.Error != nil {
		log.Printf("[PluginResourceMonitor] 持久化统计失败: plugin=%s, err=%v", stats.PluginID, result.Error)
	}
}

func (s *PluginResourceMonitorService) handleAlert(alert *plugin.ResourceAlert) {
	log.Printf("[PluginResourceMonitor] 资源告警: plugin=%s, type=%s, severity=%s, msg=%s, current=%d, threshold=%d",
		alert.PluginID, alert.AlertType, alert.Severity, alert.Message, alert.CurrentValue, alert.Threshold)

	s.saveAlertToDB(alert)
}

func (s *PluginResourceMonitorService) saveAlertToDB(alert *plugin.ResourceAlert) {
	ctx := context.Background()
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	alertLog := model.PluginEventLog{
		PluginID:  alert.PluginID,
		EventType: "resource_alert",
		EventData: alert.Message,
		CreatedAt: &alert.Timestamp,
	}

	if err := db.Create(&alertLog).Error; err != nil {
		log.Printf("[PluginResourceMonitor] 保存告警日志失败: plugin=%s, err=%v", alert.PluginID, err)
	}
}

func (s *PluginResourceMonitorService) GetPluginStats(ctx context.Context, pluginID string) (*model.PluginResourceStats, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var stats model.PluginResourceStats
	if err := db.Where("plugin_id = ?", pluginID).First(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *PluginResourceMonitorService) GetRuntimeStats(tenantID int64, pluginID string) *plugin.PluginRuntimeStats {
	monitor := plugin.GetResourceMonitor()
	return monitor.GetStats(tenantID, pluginID)
}

func (s *PluginResourceMonitorService) GetAllRuntimeStats(tenantID int64) []*plugin.PluginRuntimeStats {
	monitor := plugin.GetResourceMonitor()
	return monitor.GetAllStats(tenantID)
}

func (s *PluginResourceMonitorService) ResetStats(tenantID int64, pluginID string) {
	monitor := plugin.GetResourceMonitor()
	monitor.ResetStats(tenantID, pluginID)
}

func (s *PluginResourceMonitorService) GetRecentAlerts(ctx context.Context, pluginID string, limit int) ([]model.PluginEventLog, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var alerts []model.PluginEventLog
	if err := db.Where("plugin_id = ? AND event_type = ?", pluginID, "resource_alert").
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

func (s *PluginResourceMonitorService) GetStatsSummary(ctx context.Context) (map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var totalCalls int64
	var totalErrors int64
	var totalPlugins int64

	db.Model(&model.PluginResourceStats{}).Select("COALESCE(SUM(call_count), 0)").Scan(&totalCalls)
	db.Model(&model.PluginResourceStats{}).Select("COALESCE(SUM(error_count), 0)").Scan(&totalErrors)
	db.Model(&model.PluginResourceStats{}).Count(&totalPlugins)

	errorRate := float64(0)
	if totalCalls > 0 {
		errorRate = float64(totalErrors) / float64(totalCalls) * 100
	}

	return map[string]interface{}{
		"total_calls":   totalCalls,
		"total_errors":  totalErrors,
		"total_plugins": totalPlugins,
		"error_rate":    errorRate,
		"updated_at":    time.Now(),
	}, nil
}
