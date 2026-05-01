package middleware

import (
	"context"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fayhub/pkg/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限校验中间件
func PermissionMiddleware(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			response.GinError(c, errors.ErrUnauthorized, "未获取到用户信息")
			c.Abort()
			return
		}

		role, _ := GetRoleFromContext(c)
		if role == "super_admin" || role == "platform_admin" {
			c.Next()
			return
		}

		hasPermission, err := checkPermission(c.Request.Context(), userID, permission)
		if err != nil {
			response.GinError(c, errors.ErrInternalServer, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.GinError(c, errors.ErrPermissionDenied, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkPermission 检查用户权限（带租户隔离）
func checkPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, fmt.Errorf("数据库未连接")
	}

	var userRoles []model.UserRole
	roleQuery := db.Where("user_id = ?", userID)
	if err := roleQuery.Find(&userRoles).Error; err != nil {
		return false, fmt.Errorf("查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return false, nil
	}

	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	var roleMenus []model.RoleMenu
	menuQuery := db.Where("role_id IN ?", roleIDs)
	if err := menuQuery.Find(&roleMenus).Error; err != nil {
		return false, fmt.Errorf("查询角色菜单失败")
	}

	var menuIDs []uint
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuID)
	}

	if len(menuIDs) == 0 {
		return false, nil
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var menus []model.Menu
	if err := platformDB.Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
		return false, fmt.Errorf("查询菜单权限失败")
	}

	for _, menu := range menus {
		if menu.Permission == permission {
			return true, nil
		}
	}

	return false, nil
}

// APIPermissionMiddleware API权限校验中间件
func APIPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			response.GinError(c, errors.ErrUnauthorized, "未获取到用户信息")
			c.Abort()
			return
		}

		role, _ := GetRoleFromContext(c)
		if role == "super_admin" || role == "platform_admin" {
			c.Next()
			return
		}

		path := c.FullPath()
		method := c.Request.Method

		if path == "" {
			c.Next()
			return
		}

		hasPermission, err := service.ServiceGroupApp.RBACService.CheckAPIPermission(c.Request.Context(), userID, path, method)
		if err != nil {
			response.GinError(c, errors.ErrInternalServer, "API权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.GinError(c, errors.ErrPermissionDenied, "无权限访问该API")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminMiddleware 超级管理员权限中间件
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if exists && role == "super_admin" {
			c.Next()
			return
		}

		userID, exists := GetUserIDFromContext(c)
		if !exists {
			response.GinError(c, errors.ErrUnauthorized, "未获取到用户信息")
			c.Abort()
			return
		}

		isSuperAdmin, err := checkSuperAdmin(c.Request.Context(), userID)
		if err != nil {
			response.GinError(c, errors.ErrInternalServer, "角色查询失败")
			c.Abort()
			return
		}

		if !isSuperAdmin {
			response.GinError(c, errors.ErrPermissionDenied, "需要超级管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkSuperAdmin 检查是否为超级管理员
func checkSuperAdmin(ctx context.Context, userID uint) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, fmt.Errorf("数据库未连接")
	}

	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, fmt.Errorf("查询用户角色失败")
	}

	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	var roles []model.Role
	if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return false, fmt.Errorf("查询角色失败")
	}

	for _, role := range roles {
		if role.Name == "super_admin" {
			return true, nil
		}
	}

	return false, nil
}

// TenantAdminMiddleware 租户管理员权限中间件
func TenantAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			response.GinError(c, errors.ErrUnauthorized, "未获取到用户信息")
			c.Abort()
			return
		}

		// 检查是否为租户管理员
		isTenantAdmin, err := checkTenantAdmin(c.Request.Context(), userID)
		if err != nil {
			response.GinError(c, errors.ErrInternalServer, "角色查询失败")
			c.Abort()
			return
		}

		if !isTenantAdmin {
			response.GinError(c, errors.ErrPermissionDenied, "需要租户管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkTenantAdmin 检查是否为租户管理员
func checkTenantAdmin(ctx context.Context, userID uint) (bool, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return false, fmt.Errorf("数据库未连接")
	}

	var userRoles []model.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return false, fmt.Errorf("查询用户角色失败")
	}

	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	var roles []model.Role
	if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return false, fmt.Errorf("查询角色失败")
	}

	for _, role := range roles {
		if role.Name == "tenant_admin" {
			return true, nil
		}
	}

	return false, nil
}
