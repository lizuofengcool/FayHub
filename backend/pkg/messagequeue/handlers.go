package messagequeue

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/logger"
	"fayhub/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterBusinessHandlers() {
	Subscribe("payment.paid", handlePaymentPaid)
	Subscribe("order.expired", handleOrderExpired)
	Subscribe("user.created", handleUserCreated)
	Subscribe("file.uploaded", handleFileUploaded)
}

func handlePaymentPaid(ctx context.Context, msg Message) error {
	orderNo, _ := msg.Payload["order_no"].(string)
	tenantID, _ := msg.Payload["tenant_id"].(float64)
	amount, _ := msg.Payload["amount"].(float64)

	logger.Info(ctx, "处理支付成功消息",
		zap.String("topic", "payment.paid"),
		zap.String("order_no", orderNo),
		zap.Float64("amount", amount))

	if tenantID > 0 {
		syncTenantStorageUsage(ctx, uint(tenantID))
	}

	return nil
}

func handleOrderExpired(ctx context.Context, msg Message) error {
	orderNo, _ := msg.Payload["order_no"].(string)

	logger.Info(ctx, "处理订单过期消息",
		zap.String("topic", "order.expired"),
		zap.String("order_no", orderNo))

	return nil
}

func handleUserCreated(ctx context.Context, msg Message) error {
	userID, _ := msg.Payload["user_id"].(float64)
	tenantID, _ := msg.Payload["tenant_id"].(float64)

	logger.Info(ctx, "处理用户创建消息",
		zap.String("topic", "user.created"),
		zap.Float64("user_id", userID),
		zap.Float64("tenant_id", tenantID))

	if tenantID > 0 {
		incrementTenantUserUsage(ctx, uint(tenantID))
	}

	return nil
}

func handleFileUploaded(ctx context.Context, msg Message) error {
	fileID, _ := msg.Payload["file_id"].(float64)
	tenantID, _ := msg.Payload["tenant_id"].(float64)
	fileSize, _ := msg.Payload["size"].(float64)

	logger.Info(ctx, "处理文件上传消息",
		zap.String("topic", "file.uploaded"),
		zap.Float64("file_id", fileID),
		zap.Float64("size", fileSize))

	if tenantID > 0 && fileSize > 0 {
		sizeMB := int(fileSize / 1024 / 1024)
		if sizeMB < 1 {
			sizeMB = 1
		}
		incrementTenantStorageUsage(ctx, uint(tenantID), sizeMB)
	}

	return nil
}

func incrementTenantUserUsage(ctx context.Context, tenantID uint) {
	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)
	if queryDB == nil {
		return
	}

	queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update("used_users", gorm.Expr("used_users + 1"))
}

func incrementTenantStorageUsage(ctx context.Context, tenantID uint, sizeMB int) {
	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)
	if queryDB == nil {
		return
	}

	queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update("used_storage_mb", gorm.Expr("used_storage_mb + ?", sizeMB))
}

func syncTenantStorageUsage(ctx context.Context, tenantID uint) {
	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)
	if queryDB == nil {
		return
	}

	var totalSize int64
	queryDB.Model(&model.FileRecord{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize)

	sizeMB := int(totalSize / 1024 / 1024)
	queryDB.Model(&model.TenantQuota{}).
		Where("tenant_id = ?", tenantID).
		Update("used_storage_mb", sizeMB)
}
