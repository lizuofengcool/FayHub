package service

import (
	"context"
	"fmt"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type NotificationChannelService struct{}

func (s *NotificationChannelService) CreateChannel(ctx context.Context, ch *model.NotificationChannel) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	if ch.IsDefault == 1 {
		db.Model(&model.NotificationChannel{}).Where("type = ?", ch.Type).Update("is_default", 0)
	}

	return db.Create(ch).Error
}

func (s *NotificationChannelService) UpdateChannel(ctx context.Context, ch *model.NotificationChannel) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	if ch.IsDefault == 1 {
		db.Model(&model.NotificationChannel{}).Where("type = ? AND id != ?", ch.Type, ch.ID).Update("is_default", 0)
	}

	return db.Save(ch).Error
}

func (s *NotificationChannelService) DeleteChannel(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Delete(&model.NotificationChannel{}, id).Error
}

func (s *NotificationChannelService) GetChannel(ctx context.Context, id int64) (*model.NotificationChannel, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	var ch model.NotificationChannel
	if err := db.First(&ch, id).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *NotificationChannelService) ListChannels(ctx context.Context, page, pageSize int) ([]model.NotificationChannel, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var channels []model.NotificationChannel

	query := db.Model(&model.NotificationChannel{})
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

func (s *NotificationChannelService) CreateTemplate(ctx context.Context, tpl *model.NotificationTemplate) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Create(tpl).Error
}

func (s *NotificationChannelService) UpdateTemplate(ctx context.Context, tpl *model.NotificationTemplate) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Save(tpl).Error
}

func (s *NotificationChannelService) DeleteTemplate(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return db.Delete(&model.NotificationTemplate{}, id).Error
}

func (s *NotificationChannelService) GetTemplate(ctx context.Context, id int64) (*model.NotificationTemplate, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	var tpl model.NotificationTemplate
	if err := db.First(&tpl, id).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (s *NotificationChannelService) ListTemplates(ctx context.Context, page, pageSize int) ([]model.NotificationTemplate, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var templates []model.NotificationTemplate

	query := db.Model(&model.NotificationTemplate{})
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (s *NotificationChannelService) Send(ctx context.Context, record *model.NotificationRecord) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	record.Status = "pending"
	if err := db.Create(record).Error; err != nil {
		return err
	}

	go s.processSend(record)

	return nil
}

func (s *NotificationChannelService) processSend(record *model.NotificationRecord) {
	ctx := utils.SkipTenantIsolation(context.Background())
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":  "success",
		"sent_at": now,
	}

	db.Model(&model.NotificationRecord{}).Where("id = ?", record.ID).Updates(updates)
}

func (s *NotificationChannelService) GetRecords(ctx context.Context, page, pageSize int) ([]model.NotificationRecord, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var records []model.NotificationRecord

	query := db.Model(&model.NotificationRecord{})
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (s *NotificationChannelService) GetDB() *gorm.DB {
	return utils.GetDB(context.Background())
}
