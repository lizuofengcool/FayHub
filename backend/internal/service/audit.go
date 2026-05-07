package service

import (
	"context"
	"encoding/json"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"time"
)

type AuditService struct{}

var AuditServiceApp = new(AuditService)

func (s *AuditService) Record(ctx context.Context, log *model.AuditLog) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	now := time.Now()
	log.CreatedAt = &now

	return db.Create(log).Error
}

func (s *AuditService) RecordAction(ctx context.Context, tenantID int64, userID int64, username string, action model.AuditAction, resource string, resourceID string, detail interface{}, success bool, errMsg string) error {
	var detailJSON model.JSONRawMessage
	if detail != nil {
		data, err := json.Marshal(detail)
		if err != nil {
			detailJSON = model.JSONRawMessage(`{"error":"序列化失败"}`)
		} else {
			detailJSON = model.JSONRawMessage(data)
		}
	}

	auditLog := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     string(action),
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detailJSON,
		Success:    success,
		ErrorMsg:   errMsg,
	}
	auditLog.TenantID = tenantID

	return s.Record(ctx, auditLog)
}

func (s *AuditService) List(ctx context.Context, filters *AuditLogFilters, page, pageSize int) ([]*model.AuditLog, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.AuditLog{})

	if filters != nil {
		if filters.UserID > 0 {
			query = query.Where("user_id = ?", filters.UserID)
		}
		if filters.Action != "" {
			query = query.Where("action = ?", filters.Action)
		}
		if filters.Resource != "" {
			query = query.Where("resource = ?", filters.Resource)
		}
		if filters.ResourceID != "" {
			query = query.Where("resource_id = ?", filters.ResourceID)
		}
		if filters.Success != nil {
			query = query.Where("success = ?", *filters.Success)
		}
		if filters.StartTime != nil {
			query = query.Where("created_at >= ?", filters.StartTime)
		}
		if filters.EndTime != nil {
			query = query.Where("created_at <= ?", filters.EndTime)
		}
		if filters.IP != "" {
			query = query.Where("ip = ?", filters.IP)
		}
		if filters.Path != "" {
			query = query.Where("path LIKE ?", "%"+filters.Path+"%")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询审计日志总数失败")
	}

	var logs []*model.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询审计日志失败")
	}

	return logs, total, nil
}

func (s *AuditService) GetByID(ctx context.Context, id int64) (*model.AuditLog, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var log model.AuditLog
	if err := db.First(&log, id).Error; err != nil {
		return nil, err
	}

	return &log, nil
}

func (s *AuditService) GetStats(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.AuditLog{})
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	var totalCount int64
	query.Count(&totalCount)

	var successCount int64
	db.Model(&model.AuditLog{}).Where("success = ?", true).Count(&successCount)

	var failCount int64
	db.Model(&model.AuditLog{}).Where("success = ?", false).Count(&failCount)

	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	db.Model(&model.AuditLog{}).Where("created_at >= ?", today).Count(&todayCount)

	successRate := float64(0)
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount)
	}

	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actionCounts []ActionCount
	db.Model(&model.AuditLog{}).Select("action, count(*) as count").Group("action").Scan(&actionCounts)

	type ResourceCount struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	}
	var resourceCounts []ResourceCount
	db.Model(&model.AuditLog{}).Select("resource, count(*) as count").Group("resource").Scan(&resourceCounts)

	return map[string]interface{}{
		"total":        totalCount,
		"today":        todayCount,
		"success_rate": successRate,
		"success":      successCount,
		"failed":       failCount,
		"by_action":    actionCounts,
		"by_resource":  resourceCounts,
	}, nil
}

func (s *AuditService) Cleanup(ctx context.Context, before time.Time) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	result := db.Where("created_at < ?", before).Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}

type AuditLogFilters struct {
	UserID     int64      `form:"user_id"`
	Action     string     `form:"action"`
	Resource   string     `form:"resource"`
	ResourceID string     `form:"resource_id"`
	Success    *bool      `form:"success"`
	StartTime  *time.Time `form:"start_time"`
	EndTime    *time.Time `form:"end_time"`
	IP         string     `form:"ip"`
	Path       string     `form:"path"`
}
