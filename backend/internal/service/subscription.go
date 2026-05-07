package service

import (
	"context"
	"fmt"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type SubscriptionService struct{}

func (s *SubscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var existing model.Subscription
	if err := db.Where("tenant_id = ?", sub.TenantID).First(&existing).Error; err == nil {
		return fmt.Errorf("该租户已有订阅")
	}

	return db.Create(sub).Error
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Save(sub).Error
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Delete(&model.Subscription{}, id).Error
}

func (s *SubscriptionService) GetByID(ctx context.Context, id int64) (*model.Subscription, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	var sub model.Subscription
	if err := db.First(&sub, id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SubscriptionService) GetByTenantID(ctx context.Context, tenantID int64) (*model.Subscription, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	var sub model.Subscription
	if err := db.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, page, pageSize int) ([]model.Subscription, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var subs []model.Subscription

	query := db.Model(&model.Subscription{})
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

func (s *SubscriptionService) Cancel(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Model(&model.Subscription{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "cancelled",
		"auto_renew": 0,
	}).Error
}

func (s *SubscriptionService) Renew(ctx context.Context, id int64, months int) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var sub model.Subscription
	if err := db.First(&sub, id).Error; err != nil {
		return err
	}

	newEnd := sub.EndDate.AddDate(0, months, 0)
	if time.Now().After(sub.EndDate) {
		newEnd = time.Now().AddDate(0, months, 0)
	}

	return db.Model(&sub).Updates(map[string]interface{}{
		"status":   "active",
		"end_date": newEnd,
	}).Error
}

func (s *SubscriptionService) CheckExpired(ctx context.Context) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Model(&model.Subscription{}).
		Where("status = ? AND end_date < ?", "active", time.Now()).
		Update("status", "expired").Error
}

func (s *SubscriptionService) CreateInvoice(ctx context.Context, invoice *model.SubscriptionInvoice) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Create(invoice).Error
}

func (s *SubscriptionService) GetInvoices(ctx context.Context, subscriptionID int64, page, pageSize int) ([]model.SubscriptionInvoice, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var invoices []model.SubscriptionInvoice

	query := db.Model(&model.SubscriptionInvoice{}).Where("subscription_id = ?", subscriptionID)
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (s *SubscriptionService) GetDB() *gorm.DB {
	return utils.GetDB(context.Background())
}
