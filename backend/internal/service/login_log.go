package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"time"
)

type LoginLogService struct{}

var LoginLogServiceApp = new(LoginLogService)

type LoginLogFilters struct {
	Username    string
	LoginStatus string
	LoginIP     string
	StartTime   *time.Time
	EndTime     *time.Time
}

func (s *LoginLogService) Record(ctx context.Context, log *model.LoginLog) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Create(log).Error
}

func (s *LoginLogService) RecordLogin(ctx context.Context, userID int64, username string, tenantID int64, ip string, userAgent string, success bool, msg string) {
	status := "success"
	if !success {
		status = "failed"
	}

	browser, os := parseUserAgent(userAgent)

	log := &model.LoginLog{
		UserID:      userID,
		Username:    username,
		TenantID:    tenantID,
		LoginStatus: status,
		LoginIP:     ip,
		Browser:     browser,
		OS:          os,
		LoginTime:   time.Now(),
		Msg:         msg,
	}

	go func() {
		_ = s.Record(ctx, log)
	}()
}

func (s *LoginLogService) List(ctx context.Context, filters *LoginLogFilters, page, pageSize int) ([]*model.LoginLog, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.LoginLog{})

	if filters != nil {
		if filters.Username != "" {
			query = query.Where("username LIKE ?", "%"+filters.Username+"%")
		}
		if filters.LoginStatus != "" {
			query = query.Where("login_status = ?", filters.LoginStatus)
		}
		if filters.LoginIP != "" {
			query = query.Where("login_ip = ?", filters.LoginIP)
		}
		if filters.StartTime != nil {
			query = query.Where("login_time >= ?", filters.StartTime)
		}
		if filters.EndTime != nil {
			query = query.Where("login_time <= ?", filters.EndTime)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询登录日志总数失败")
	}

	var logs []*model.LoginLog
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询登录日志失败")
	}

	return logs, total, nil
}

func (s *LoginLogService) Cleanup(ctx context.Context, before time.Time) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	result := db.Where("login_time < ?", before).Delete(&model.LoginLog{})
	return result.RowsAffected, result.Error
}

func parseUserAgent(ua string) (browser, os string) {
	if ua == "" {
		return "-", "-"
	}

	switch {
	case contains(ua, "Edg/"):
		browser = "Edge"
	case contains(ua, "Chrome/") && !contains(ua, "Edg/"):
		browser = "Chrome"
	case contains(ua, "Firefox/"):
		browser = "Firefox"
	case contains(ua, "Safari/") && !contains(ua, "Chrome/"):
		browser = "Safari"
	default:
		browser = "Other"
	}

	switch {
	case contains(ua, "Windows"):
		os = "Windows"
	case contains(ua, "Mac OS"):
		os = "macOS"
	case contains(ua, "Linux"):
		os = "Linux"
	case contains(ua, "Android"):
		os = "Android"
	case contains(ua, "iPhone") || contains(ua, "iPad"):
		os = "iOS"
	default:
		os = "Other"
	}

	return browser, os
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
