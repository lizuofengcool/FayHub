package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type APIService struct{}

type CreateAPIRequest struct {
	Path        string `json:"path" binding:"required,max=500"`
	Method      string `json:"method" binding:"required,max=10"`
	Description string `json:"description" binding:"max=500"`
	Group       string `json:"group" binding:"max=100"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

type UpdateAPIRequest struct {
	Path        string `json:"path" binding:"max=500"`
	Method      string `json:"method" binding:"max=10"`
	Description string `json:"description" binding:"max=500"`
	Group       string `json:"group" binding:"max=100"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

type GetAPIListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Path     string `form:"path"`
	Method   string `form:"method"`
	Group    string `form:"group"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}

type GetAPIListResponse struct {
	List  []model.API `json:"list"`
	Total int64       `json:"total"`
}

type AssignRoleAPIRequest struct {
	RoleID int64   `json:"role_id" binding:"required"`
	APIIDs []int64 `json:"api_ids" binding:"required,min=1"`
}

func (s *APIService) CreateAPI(ctx context.Context, req CreateAPIRequest) (*model.API, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var existing model.API
	if err := db.Where("path = ? AND method = ?", req.Path, req.Method).First(&existing).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrAPIAlreadyExist, "")
	}

	api := model.API{
		Path:        req.Path,
		Method:      req.Method,
		Description: req.Description,
		Group:       req.Group,
		Status:      req.Status,
	}

	if err := db.Create(&api).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建API接口失败")
	}

	return &api, nil
}

func (s *APIService) GetAPIList(ctx context.Context, req GetAPIListRequest) (*GetAPIListResponse, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.API{})

	if req.Path != "" {
		query = query.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Method != "" {
		query = query.Where("method = ?", req.Method)
	}
	if req.Group != "" {
		query = query.Where("\"group\" = ?", req.Group)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API总数失败")
	}

	var apis []model.API
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id ASC").Find(&apis).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API列表失败")
	}

	return &GetAPIListResponse{
		List:  apis,
		Total: total,
	}, nil
}

func (s *APIService) GetAPIByID(ctx context.Context, apiID int64) (*model.API, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var api model.API
	if err := db.First(&api, apiID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrAPINotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API接口失败")
	}

	return &api, nil
}

func (s *APIService) UpdateAPI(ctx context.Context, apiID int64, req UpdateAPIRequest) (*model.API, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var api model.API
	if err := db.First(&api, apiID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrAPINotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API接口失败")
	}

	if req.Path != "" && req.Method != "" && (req.Path != api.Path || req.Method != api.Method) {
		var existing model.API
		if err := db.Where("path = ? AND method = ? AND id != ?", req.Path, req.Method, apiID).First(&existing).Error; err == nil {
			return nil, errs.NewServiceError(errs.ErrAPIAlreadyExist, "")
		}
	}

	if req.Path != "" {
		api.Path = req.Path
	}
	if req.Method != "" {
		api.Method = req.Method
	}
	api.Description = req.Description
	api.Group = req.Group
	api.Status = req.Status

	if err := db.Save(&api).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "更新API接口失败")
	}

	return &api, nil
}

func (s *APIService) DeleteAPI(ctx context.Context, apiID int64) error {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var api model.API
	if err := db.First(&api, apiID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrAPINotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询API接口失败")
	}

	if err := db.Where("api_id = ?", apiID).Delete(&model.RoleAPI{}).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除API角色关联失败")
	}

	if err := db.Delete(&api).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除API接口失败")
	}

	return nil
}

func (s *APIService) AssignRoleAPIs(ctx context.Context, req AssignRoleAPIRequest) error {
	tenantDB := utils.GetDB(ctx)
	if tenantDB == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var role model.Role
	if err := tenantDB.First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var existingAPIs []model.API
	if err := platformDB.Where("id IN ?", req.APIIDs).Find(&existingAPIs).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询API接口失败")
	}
	if len(existingAPIs) != len(req.APIIDs) {
		return errs.NewServiceError(errs.ErrAPINotExist, "部分API接口不存在")
	}

	if err := tenantDB.Where("role_id = ?", req.RoleID).Delete(&model.RoleAPI{}).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "清除角色API关联失败")
	}

	for _, apiID := range req.APIIDs {
		roleAPI := model.RoleAPI{
			RoleID: req.RoleID,
			APIID:  apiID,
		}
		if err := tenantDB.Create(&roleAPI).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "分配API权限失败")
		}
	}

	return nil
}

func (s *APIService) GetRoleAPIs(ctx context.Context, roleID int64) ([]model.API, error) {
	tenantDB := utils.GetDB(ctx)
	if tenantDB == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var roleAPIs []model.RoleAPI
	if err := tenantDB.Where("role_id = ?", roleID).Find(&roleAPIs).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色API关联失败")
	}

	if len(roleAPIs) == 0 {
		return []model.API{}, nil
	}

	var apiIDs []int64
	for _, ra := range roleAPIs {
		apiIDs = append(apiIDs, ra.APIID)
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var apis []model.API
	if err := platformDB.Where("id IN ? AND status = 1", apiIDs).Order("id ASC").Find(&apis).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API接口失败")
	}

	return apis, nil
}
