package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type RBACService struct{}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Type        int    `json:"type" binding:"required,oneof=1 2"`
	DataScope   int    `json:"data_scope" binding:"omitempty,oneof=1 2 3 4 5"`
}

type AssignRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

type RemoveRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

func (s *RBACService) CheckPermission(ctx context.Context, userID uint, tenantID uint, permission string) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, errs.NewServiceError(errs.ErrDatabase, "查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return false, nil
	}

	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var menuCount int64
	if err := platformDB.Model(&model.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN roles ON roles.id = role_menus.role_id").
		Where("role_menus.role_id IN ? AND menus.permission = ? AND menus.status = 1 AND (roles.tenant_id = 0 OR roles.tenant_id = ?)",
			roleIDs, permission, tenantID).
		Count(&menuCount).Error; err != nil {
		return false, errs.NewServiceError(errs.ErrDatabase, "检查菜单权限失败")
	}

	if menuCount > 0 {
		return true, nil
	}

	var apiCount int64
	if err := platformDB.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Joins("JOIN roles ON roles.id = role_apis.role_id").
		Where("role_apis.role_id IN ? AND apis.path = ? AND apis.status = 1 AND (roles.tenant_id = 0 OR roles.tenant_id = ?)",
			roleIDs, permission, tenantID).
		Count(&apiCount).Error; err != nil {
		return false, errs.NewServiceError(errs.ErrDatabase, "检查API权限失败")
	}

	return apiCount > 0, nil
}

func (s *RBACService) CheckAPIPermission(ctx context.Context, userID uint, path, method string) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, errs.NewServiceError(errs.ErrDatabase, "查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return false, nil
	}

	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var apiCount int64
	if err := platformDB.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Where("role_apis.role_id IN ? AND apis.path = ? AND apis.method = ? AND apis.status = 1", roleIDs, path, method).
		Count(&apiCount).Error; err != nil {
		return false, errs.NewServiceError(errs.ErrDatabase, "检查API权限失败")
	}

	return apiCount > 0, nil
}

func (s *RBACService) GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var roles []model.Role
	if err := db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Find(&roles).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户角色失败")
	}

	return roles, nil
}

func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var permissions []string

	var menuPermissions []string
	if err := platformDB.Model(&model.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_menus.role_id").
		Where("user_roles.user_id = ? AND menus.status = 1 AND menus.permission != ''", userID).
		Pluck("DISTINCT menus.permission", &menuPermissions).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单权限失败")
	}

	var apiPermissions []string
	if err := platformDB.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_apis.role_id").
		Where("user_roles.user_id = ? AND apis.status = 1", userID).
		Pluck("DISTINCT apis.path || ':' || apis.method", &apiPermissions).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API权限失败")
	}

	permissions = append(permissions, menuPermissions...)
	permissions = append(permissions, apiPermissions...)

	return permissions, nil
}

func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	var role model.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	var existingUserRole model.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingUserRole).Error; err == nil {
		return errs.NewServiceError(errs.ErrRoleAlreadyBound, "")
	}

	userRole := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "分配角色失败")
	}

	return nil
}

func (s *RBACService) CreateRole(ctx context.Context, req CreateRoleRequest) (*model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var existingRole model.Role
	if err := db.Where("name = ?", req.Name).First(&existingRole).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrRoleAlreadyExist, "")
	}

	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      1,
		DataScope:   req.DataScope,
	}
	if role.DataScope == 0 {
		role.DataScope = model.DataScopeAll
	}

	if err := db.Create(&role).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建角色失败")
	}

	return &role, nil
}

func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var userRole model.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrRoleNotBound, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询用户角色失败")
	}

	if err := db.Delete(&userRole).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "移除角色失败")
	}

	return nil
}

type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Status      int    `json:"status" binding:"oneof=0 1"`
	DataScope   int    `json:"data_scope" binding:"omitempty,oneof=1 2 3 4 5"`
}

type GetRoleListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name"`
	Type     int    `form:"type" binding:"omitempty,oneof=0 1 2"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}

type GetRoleListResponse struct {
	List  []model.Role `json:"list"`
	Total int64        `json:"total"`
}

func (s *RBACService) GetRoleList(ctx context.Context, req GetRoleListRequest) (*GetRoleListResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.Role{})

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Type > 0 {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色总数失败")
	}

	var roles []model.Role
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id ASC").Find(&roles).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色列表失败")
	}

	return &GetRoleListResponse{
		List:  roles,
		Total: total,
	}, nil
}

func (s *RBACService) GetRoleByID(ctx context.Context, roleID uint) (*model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var role model.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	return &role, nil
}

func (s *RBACService) UpdateRole(ctx context.Context, roleID uint, req UpdateRoleRequest) (*model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var role model.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	if req.Name != "" && req.Name != role.Name {
		var existingRole model.Role
		if err := db.Where("name = ? AND id != ?", req.Name, roleID).First(&existingRole).Error; err == nil {
			return nil, errs.NewServiceError(errs.ErrRoleAlreadyExist, "")
		}
		role.Name = req.Name
	}

	if req.Description != "" {
		role.Description = req.Description
	}
	role.Status = req.Status
	if req.DataScope >= 1 && req.DataScope <= 5 {
		role.DataScope = req.DataScope
	}

	if err := db.Save(&role).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "更新角色失败")
	}

	return &role, nil
}

func (s *RBACService) DeleteRole(ctx context.Context, roleID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var role model.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	if role.Name == "super_admin" {
		return errs.NewServiceError(errs.ErrForbidden, "超级管理员角色不可删除")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.UserRole{}).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "删除角色用户关联失败")
		}

		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "删除角色菜单关联失败")
		}

		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleAPI{}).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "删除角色API关联失败")
		}

		if err := tx.Delete(&role).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "删除角色失败")
		}

		return nil
	})
}
