package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
)

type TenantChannelService struct{}

var TenantChannelServiceApp = new(TenantChannelService)

type CreateChannelConfigRequest struct {
	ChannelType    string `json:"channel_type" binding:"required"`
	ChannelName    string `json:"channel_name"`
	AppID          string `json:"app_id"`
	AppSecret      string `json:"app_secret"`
	MerchantID     string `json:"merchant_id"`
	PayPublicKey   string `json:"pay_public_key"`
	PayPrivateKey  string `json:"pay_private_key"`
	CertSerialNo   string `json:"cert_serial_no"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	Extra          string `json:"extra"`
	Status         int    `json:"status"`
}

type UpdateChannelConfigRequest struct {
	ChannelName    string `json:"channel_name"`
	AppID          string `json:"app_id"`
	AppSecret      string `json:"app_secret"`
	MerchantID     string `json:"merchant_id"`
	PayPublicKey   string `json:"pay_public_key"`
	PayPrivateKey  string `json:"pay_private_key"`
	CertSerialNo   string `json:"cert_serial_no"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	Extra          string `json:"extra"`
	Status         *int   `json:"status"`
}

func (s *TenantChannelService) List(ctx context.Context, channelType string, status *int, page, pageSize int) ([]*model.TenantChannelConfig, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.TenantChannelConfig{})

	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询渠道配置总数失败")
	}

	var configs []*model.TenantChannelConfig
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&configs).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询渠道配置失败")
	}

	return configs, total, nil
}

func (s *TenantChannelService) GetByID(ctx context.Context, id int64) (*model.TenantChannelConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var config model.TenantChannelConfig
	if err := db.First(&config, id).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *TenantChannelService) GetByChannelType(ctx context.Context, channelType string) (*model.TenantChannelConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var config model.TenantChannelConfig
	if err := db.Where("channel_type = ?", channelType).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *TenantChannelService) Create(ctx context.Context, req CreateChannelConfigRequest) (*model.TenantChannelConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	config := &model.TenantChannelConfig{
		ChannelType:    req.ChannelType,
		ChannelName:    req.ChannelName,
		AppID:          req.AppID,
		AppSecret:      req.AppSecret,
		MerchantID:     req.MerchantID,
		PayPublicKey:   req.PayPublicKey,
		PayPrivateKey:  req.PayPrivateKey,
		CertSerialNo:   req.CertSerialNo,
		Token:          req.Token,
		EncodingAESKey: req.EncodingAESKey,
		Extra:          req.Extra,
		Status:         req.Status,
	}
	if config.Status == 0 {
		config.Status = 1
	}

	if err := db.Create(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (s *TenantChannelService) Update(ctx context.Context, id int64, req UpdateChannelConfigRequest) (*model.TenantChannelConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var config model.TenantChannelConfig
	if err := db.First(&config, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.ChannelName != "" {
		updates["channel_name"] = req.ChannelName
	}
	if req.AppID != "" {
		updates["app_id"] = req.AppID
	}
	if req.AppSecret != "" {
		updates["app_secret"] = req.AppSecret
	}
	if req.MerchantID != "" {
		updates["merchant_id"] = req.MerchantID
	}
	if req.PayPublicKey != "" {
		updates["pay_public_key"] = req.PayPublicKey
	}
	if req.PayPrivateKey != "" {
		updates["pay_private_key"] = req.PayPrivateKey
	}
	if req.CertSerialNo != "" {
		updates["cert_serial_no"] = req.CertSerialNo
	}
	if req.Token != "" {
		updates["token"] = req.Token
	}
	if req.EncodingAESKey != "" {
		updates["encoding_aes_key"] = req.EncodingAESKey
	}
	if req.Extra != "" {
		updates["extra"] = req.Extra
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&config).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	db.First(&config, id)
	return &config, nil
}

func (s *TenantChannelService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Delete(&model.TenantChannelConfig{}, id).Error
}

func (s *TenantChannelService) GetThirdPartyBindings(ctx context.Context, userID int64, channelType string) ([]*model.UserThirdParty, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var bindings []*model.UserThirdParty
	query := db.Where("user_id = ?", userID)
	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	if err := query.Order("id DESC").Find(&bindings).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询第三方绑定失败")
	}

	return bindings, nil
}

func (s *TenantChannelService) GetThirdPartyByOpenID(ctx context.Context, tenantID int64, channelType string, openID string) (*model.UserThirdParty, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var binding model.UserThirdParty
	if err := db.Where("tenant_id = ? AND channel_type = ? AND open_id = ?", tenantID, channelType, openID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func (s *TenantChannelService) DeleteThirdPartyBinding(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Delete(&model.UserThirdParty{}, id).Error
}
