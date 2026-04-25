package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type RBACService struct{}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Type        int    `json:"type" binding:"required,oneof=1 2"` // 1: 平台角色, 2: 租户角色
}

type AssignRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

type RemoveRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

// CheckPermission 检查用户权限（带租户隔离）
func (s *RBACService) CheckPermission(ctx context.Context, userID uint, tenantID uint, permission string) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, errors.New("数据库未连接")
	}

	// 获取用户的所有角色（考虑租户隔离）
	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, errors.New("查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return false, nil
	}

	// 获取角色ID列表
	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 检查菜单权限（带租户隔离）
	var menuCount int64
	if err := db.Model(&model.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN roles ON roles.id = role_menus.role_id").
		Where("role_menus.role_id IN ? AND menus.permission = ? AND menus.status = 1 AND (roles.tenant_id = 0 OR roles.tenant_id = ?)",
			roleIDs, permission, tenantID).
		Count(&menuCount).Error; err != nil {
		return false, errors.New("检查菜单权限失败")
	}

	if menuCount > 0 {
		return true, nil
	}

	// 检查API权限（带租户隔离）
	var apiCount int64
	if err := db.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Joins("JOIN roles ON roles.id = role_apis.role_id").
		Where("role_apis.role_id IN ? AND apis.path = ? AND apis.status = 1 AND (roles.tenant_id = 0 OR roles.tenant_id = ?)",
			roleIDs, permission, tenantID).
		Count(&apiCount).Error; err != nil {
		return false, errors.New("检查API权限失败")
	}

	return apiCount > 0, nil
}

// CheckAPIPermission 检查API权限
func (s *RBACService) CheckAPIPermission(ctx context.Context, userID uint, path, method string) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, errors.New("数据库未连接")
	}

	// 获取用户的所有角色
	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, errors.New("查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return false, nil
	}

	// 获取角色ID列表
	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 检查API权限
	var apiCount int64
	if err := db.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Where("role_apis.role_id IN ? AND apis.path = ? AND apis.method = ? AND apis.status = 1", roleIDs, path, method).
		Count(&apiCount).Error; err != nil {
		return false, errors.New("检查API权限失败")
	}

	return apiCount > 0, nil
}

// GetUserRoles 获取用户角色列表
func (s *RBACService) GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var roles []model.Role
	if err := db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Find(&roles).Error; err != nil {
		return nil, errors.New("查询用户角色失败")
	}

	return roles, nil
}

// GetUserPermissions 获取用户权限列表
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var permissions []string

	// 获取菜单权限
	var menuPermissions []string
	if err := db.Model(&model.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_menus.role_id").
		Where("user_roles.user_id = ? AND menus.status = 1 AND menus.permission != ''", userID).
		Pluck("DISTINCT menus.permission", &menuPermissions).Error; err != nil {
		return nil, errors.New("查询菜单权限失败")
	}

	// 获取API权限
	var apiPermissions []string
	if err := db.Model(&model.API{}).
		Joins("JOIN role_apis ON role_apis.api_id = apis.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_apis.role_id").
		Where("user_roles.user_id = ? AND apis.status = 1", userID).
		Pluck("DISTINCT CONCAT(apis.path, ':', apis.method)", &apiPermissions).Error; err != nil {
		return nil, errors.New("查询API权限失败")
	}

	permissions = append(permissions, menuPermissions...)
	permissions = append(permissions, apiPermissions...)

	return permissions, nil
}

// AssignRoleToUser 为用户分配角色
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errors.New("数据库未连接")
	}

	// 检查用户是否存在
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("查询用户失败")
	}

	// 检查角色是否存在
	var role model.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return errors.New("查询角色失败")
	}

	// 检查是否已分配
	var existingUserRole model.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingUserRole).Error; err == nil {
		return errors.New("角色已分配给用户")
	}

	// 分配角色
	userRole := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return errors.New("分配角色失败")
	}

	return nil
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(ctx context.Context, req CreateRoleRequest) (*model.Role, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	// 检查角色名称是否已存在
	var existingRole model.Role
	if err := db.Where("name = ?", req.Name).First(&existingRole).Error; err == nil {
		return nil, errors.New("角色名称已存在")
	}

	// 创建角色
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      1, // 正常状态
	}

	if err := db.Create(&role).Error; err != nil {
		return nil, errors.New("创建角色失败")
	}

	return &role, nil
}

// RemoveRoleFromUser 移除用户的角色
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errors.New("数据库未连接")
	}

	// 检查关联是否存在
	var userRole model.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户未分配该角色")
		}
		return errors.New("查询用户角色失败")
	}

	// 移除角色
	if err := db.Delete(&userRole).Error; err != nil {
		return errors.New("移除角色失败")
	}

	return nil
}
