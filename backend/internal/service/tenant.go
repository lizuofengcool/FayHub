package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type TenantService struct{}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

type UpdateTenantRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

type TenantListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   *int   `json:"status" form:"status"`
}

type TenantListResponse struct {
	List  []model.Tenant `json:"list"`
	Total int64          `json:"total"`
}

func (s *TenantService) Create(ctx context.Context, req CreateTenantRequest) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	tenant := &model.Tenant{
		Name:        req.Name,
		Domain:      req.Domain,
		Description: req.Description,
		Status:      1,
	}

	if err := db.Create(tenant).Error; err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *TenantService) Update(ctx context.Context, id uint, req UpdateTenantRequest) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("租户不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&tenant).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	db.First(&tenant, id)
	return &tenant, nil
}

func (s *TenantService) Delete(ctx context.Context, id uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errors.New("数据库未连接")
	}

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("租户不存在")
		}
		return err
	}

	return db.Delete(&tenant).Error
}

func (s *TenantService) GetByID(ctx context.Context, id uint) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("租户不存在")
		}
		return nil, err
	}

	return &tenant, nil
}

func (s *TenantService) GetList(ctx context.Context, req TenantListRequest) (*TenantListResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := db.Model(&model.Tenant{})

	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR domain LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var tenants []model.Tenant
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&tenants).Error; err != nil {
		return nil, err
	}

	return &TenantListResponse{
		List:  tenants,
		Total: total,
	}, nil
}
