package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/logger"
	"fayhub/pkg/utils"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SettlementService struct{}

var SettlementServiceApp = &SettlementService{}

type CreateSettlementRequest struct {
	OrderNo     string `json:"order_no" binding:"required"`
	TotalAmount int64  `json:"total_amount" binding:"required"`
}

type SettlementResponse struct {
	*model.SettlementRecord
}

func (s *SettlementService) CalculateSettlement(ctx context.Context, tenantID uint, totalAmount int64) (*model.SettlementRecord, error) {
	config, err := s.GetSettlementConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	platformAmount := totalAmount * int64(config.PlatformRate) / 10000
	tenantAmount := totalAmount - platformAmount

	settlement := &model.SettlementRecord{
		TenantID:       tenantID,
		TotalAmount:    totalAmount,
		PlatformAmount: platformAmount,
		TenantAmount:   tenantAmount,
		PlatformRate:   config.PlatformRate,
		Status:         model.SettlementStatusPending,
	}

	return settlement, nil
}

func (s *SettlementService) CreateSettlement(ctx context.Context, req CreateSettlementRequest) (*SettlementResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return nil, errs.NewServiceError(errs.ErrUnauthorized, "租户未识别")
	}

	var order model.PaymentOrder
	if err := db.Where("order_no = ?", req.OrderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrPaymentOrderNotFound, "订单不存在")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询订单失败")
	}

	if order.Status != model.PaymentStatusPaid {
		return nil, errs.NewServiceError(errs.ErrPaymentNotifyFailed, "订单状态不允许分账")
	}

	var existingSettlement model.SettlementRecord
	if err := db.Where("order_no = ?", req.OrderNo).First(&existingSettlement).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrConflict, "该订单已创建分账记录")
	}

	settlement, err := s.CalculateSettlement(ctx, tenantID, req.TotalAmount)
	if err != nil {
		return nil, err
	}

	settlement.OrderNo = req.OrderNo
	settlement.OrderID = order.ID
	settlement.SettlementNo = generateSettlementNo()

	if err := db.Create(settlement).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建分账记录失败")
	}

	logger.Info(ctx, "分账记录创建成功",
		zap.String("order_no", req.OrderNo),
		zap.String("settlement_no", settlement.SettlementNo),
		zap.Int64("total_amount", settlement.TotalAmount),
		zap.Int64("platform_amount", settlement.PlatformAmount),
		zap.Int64("tenant_amount", settlement.TenantAmount),
		zap.Int("platform_rate", settlement.PlatformRate))

	return &SettlementResponse{SettlementRecord: settlement}, nil
}

func (s *SettlementService) GetSettlementConfig(ctx context.Context, tenantID uint) (*model.SettlementConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var config model.SettlementConfig
	if err := db.Where("tenant_id = ? AND status = 1", tenantID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.getDefaultSettlementConfig(), nil
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询分账配置失败")
	}

	return &config, nil
}

func (s *SettlementService) getDefaultSettlementConfig() *model.SettlementConfig {
	return &model.SettlementConfig{
		PlatformRate: 1000, // 默认10%
		MinAmount:    100,  // 默认最小结算金额1元
		Status:       1,
	}
}

func (s *SettlementService) UpdateSettlementConfig(ctx context.Context, tenantID uint, platformRate int, minAmount int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if platformRate < 0 || platformRate > 10000 {
		return errs.NewServiceError(errs.ErrParamValidation, "分账比例必须在0-10000之间")
	}

	if minAmount < 0 {
		return errs.NewServiceError(errs.ErrParamValidation, "最小结算金额不能小于0")
	}

	var config model.SettlementConfig
	err := db.Where("tenant_id = ?", tenantID).First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = model.SettlementConfig{
			TenantID:     tenantID,
			PlatformRate: platformRate,
			MinAmount:    minAmount,
			Status:       1,
		}
		if err := db.Create(&config).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "创建分账配置失败")
		}
	} else if err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询分账配置失败")
	} else {
		updates := map[string]interface{}{
			"platform_rate": platformRate,
			"min_amount":    minAmount,
		}
		if err := db.Model(&config).Updates(updates).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "更新分账配置失败")
		}
	}

	logger.Info(ctx, "分账配置更新成功",
		zap.Uint("tenant_id", tenantID),
		zap.Int("platform_rate", platformRate),
		zap.Int64("min_amount", minAmount))

	return nil
}

func (s *SettlementService) ProcessSettlement(ctx context.Context, settlementNo string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var settlement model.SettlementRecord
	if err := db.Where("settlement_no = ?", settlementNo).First(&settlement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrResourceNotFound, "分账记录不存在")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询分账记录失败")
	}

	if settlement.Status != model.SettlementStatusPending {
		return errs.NewServiceError(errs.ErrOperationFailed, "分账记录状态不允许处理")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":     model.SettlementStatusSettled,
		"settled_at": now,
	}

	if err := db.Model(&settlement).Updates(updates).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新分账状态失败")
	}

	logger.Info(ctx, "分账处理成功",
		zap.String("settlement_no", settlementNo),
		zap.Int64("tenant_amount", settlement.TenantAmount),
		zap.Time("settled_at", now))

	return nil
}

func (s *SettlementService) GetSettlementStats(ctx context.Context, tenantID uint) (map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var totalAmount int64
	var platformAmount int64
	var tenantAmount int64
	var pendingCount int64
	var settledCount int64
	var failedCount int64

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalAmount)

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(platform_amount), 0)").Scan(&platformAmount)

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(tenant_amount), 0)").Scan(&tenantAmount)

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.SettlementStatusPending).
		Count(&pendingCount)

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.SettlementStatusSettled).
		Count(&settledCount)

	db.Model(&model.SettlementRecord{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.SettlementStatusFailed).
		Count(&failedCount)

	return map[string]interface{}{
		"total_amount":    totalAmount,
		"platform_amount": platformAmount,
		"tenant_amount":   tenantAmount,
		"pending_count":   pendingCount,
		"settled_count":   settledCount,
		"failed_count":    failedCount,
	}, nil
}

func (s *SettlementService) ListSettlements(ctx context.Context, tenantID uint, page, pageSize int, status string) ([]model.SettlementRecord, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.SettlementRecord{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var records []model.SettlementRecord
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询结算记录失败")
	}

	return records, total, nil
}

func generateSettlementNo() string {
	now := time.Now()
	randBytes := make([]byte, 3)
	if _, err := rand.Read(randBytes); err != nil {
		randBytes = []byte{byte(now.Nanosecond()), byte(now.Nanosecond() >> 8), byte(now.Nanosecond() >> 16)}
	}
	suffix := int(randBytes[0])<<16 | int(randBytes[1])<<8 | int(randBytes[2])
	return fmt.Sprintf("ST%s%06d", now.Format("20060102150405"), suffix%1000000)
}
