package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"fayhub/pkg/utils"
)

type TenantQuotaService struct{}

type UpdateQuotaRequest struct {
	MaxUsers     *int `json:"max_users"`
	MaxStorageMB *int `json:"max_storage_mb"`
	MaxPlugins   *int `json:"max_plugins"`
	MaxAPIPerDay *int `json:"max_api_per_day"`
}

type QuotaCheckResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	Used    int    `json:"used"`
	Max     int    `json:"max"`
}

func (s *TenantQuotaService) GetQuota(ctx context.Context, tenantID uint) (*model.TenantQuota, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var quota model.TenantQuota
	err := queryDB.Where("tenant_id = ?", tenantID).First(&quota).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.initDefaultQuota(queryDB, tenantID)
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询配额失败")
	}

	s.resetAPICounterIfNeeded(queryDB, &quota)

	return &quota, nil
}

func (s *TenantQuotaService) initDefaultQuota(db *gorm.DB, tenantID uint) (*model.TenantQuota, error) {
	quota := model.TenantQuota{
		TenantID:     tenantID,
		MaxUsers:     10,
		MaxStorageMB: 1024,
		MaxPlugins:   5,
		MaxAPIPerDay: 10000,
		APIResetDate: time.Now().Format("2006-01-02"),
	}

	if err := db.Create(&quota).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "初始化配额失败")
	}

	return &quota, nil
}

func (s *TenantQuotaService) UpdateQuota(ctx context.Context, tenantID uint, req UpdateQuotaRequest) (*model.TenantQuota, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var quota model.TenantQuota
	if err := queryDB.Where("tenant_id = ?", tenantID).First(&quota).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrTenantNotExist, "租户配额不存在")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询配额失败")
	}

	updates := map[string]interface{}{}
	if req.MaxUsers != nil && *req.MaxUsers >= 0 {
		updates["max_users"] = *req.MaxUsers
	}
	if req.MaxStorageMB != nil && *req.MaxStorageMB >= 0 {
		updates["max_storage_mb"] = *req.MaxStorageMB
	}
	if req.MaxPlugins != nil && *req.MaxPlugins >= 0 {
		updates["max_plugins"] = *req.MaxPlugins
	}
	if req.MaxAPIPerDay != nil && *req.MaxAPIPerDay >= 0 {
		updates["max_api_per_day"] = *req.MaxAPIPerDay
	}

	if len(updates) > 0 {
		if err := queryDB.Model(&quota).Updates(updates).Error; err != nil {
			return nil, errs.NewServiceError(errs.ErrDatabase, "更新配额失败")
		}
	}

	queryDB.Where("tenant_id = ?", tenantID).First(&quota)
	return &quota, nil
}

func (s *TenantQuotaService) CheckUserQuota(ctx context.Context, tenantID uint) (*QuotaCheckResult, error) {
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if quota.MaxUsers > 0 && quota.UsedUsers >= quota.MaxUsers {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  "用户数已达上限",
			Used:    quota.UsedUsers,
			Max:     quota.MaxUsers,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Used:    quota.UsedUsers,
		Max:     quota.MaxUsers,
	}, nil
}

func (s *TenantQuotaService) CheckStorageQuota(ctx context.Context, tenantID uint, requiredMB int) (*QuotaCheckResult, error) {
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if quota.MaxStorageMB > 0 && (quota.UsedStorageMB+requiredMB) > quota.MaxStorageMB {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  "存储空间不足",
			Used:    quota.UsedStorageMB,
			Max:     quota.MaxStorageMB,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Used:    quota.UsedStorageMB,
		Max:     quota.MaxStorageMB,
	}, nil
}

func (s *TenantQuotaService) CheckPluginQuota(ctx context.Context, tenantID uint) (*QuotaCheckResult, error) {
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if quota.MaxPlugins > 0 && quota.UsedPlugins >= quota.MaxPlugins {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  "插件数已达上限",
			Used:    quota.UsedPlugins,
			Max:     quota.MaxPlugins,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Used:    quota.UsedPlugins,
		Max:     quota.MaxPlugins,
	}, nil
}

func (s *TenantQuotaService) CheckAPIQuota(ctx context.Context, tenantID uint) (*QuotaCheckResult, error) {
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	s.resetAPICounterIfNeeded(utils.GetDB(utils.SkipTenantIsolation(ctx)), quota)

	if quota.MaxAPIPerDay > 0 && quota.UsedAPIPerDay >= quota.MaxAPIPerDay {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  "今日API调用次数已达上限",
			Used:    quota.UsedAPIPerDay,
			Max:     quota.MaxAPIPerDay,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Used:    quota.UsedAPIPerDay,
		Max:     quota.MaxAPIPerDay,
	}, nil
}

func (s *TenantQuotaService) IncrementUsage(ctx context.Context, tenantID uint, field string, delta int) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	result := queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update(field, gorm.Expr(field+" + ?", delta))

	if result.Error != nil {
		logger.Error(ctx, "更新配额使用量失败",
			zap.Uint("tenant_id", tenantID),
			zap.String("field", field),
			zap.Error(result.Error))
		return errs.NewServiceError(errs.ErrDatabase, "更新配额使用量失败")
	}

	if result.RowsAffected == 0 {
		_, _ = s.initDefaultQuota(queryDB, tenantID)
		queryDB.Model(&model.TenantQuota{}).
			Where("tenant_id = ?", tenantID).
			Update(field, gorm.Expr(field+" + ?", delta))
	}

	return nil
}

func (s *TenantQuotaService) DecrementUsage(ctx context.Context, tenantID uint, field string, delta int) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	result := queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ? AND "+field+" >= ?", tenantID, delta).
		Update(field, gorm.Expr(field+" - ?", delta))

	if result.Error != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新配额使用量失败")
	}

	return nil
}

func (s *TenantQuotaService) SyncUserCount(ctx context.Context, tenantID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var count int64
	queryDB.Model(&model.User{}).Where("tenant_id = ?", tenantID).Count(&count)

	return queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update("used_users", count).Error
}

func (s *TenantQuotaService) SyncPluginCount(ctx context.Context, tenantID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var count int64
	queryDB.Model(&model.InstalledPlugin{}).Where("tenant_id = ? AND status = ?", tenantID, 1).Count(&count)

	return queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update("used_plugins", count).Error
}

func (s *TenantQuotaService) resetAPICounterIfNeeded(db *gorm.DB, quota *model.TenantQuota) {
	today := time.Now().Format("2006-01-02")
	if quota.APIResetDate != today {
		db.Model(&model.TenantQuota{}).
			Where("tenant_id = ?", quota.TenantID).
			Updates(map[string]interface{}{
				"used_api_per_day": 0,
				"api_reset_date":   today,
			})
		quota.UsedAPIPerDay = 0
		quota.APIResetDate = today
	}
}
