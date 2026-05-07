package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RBACController struct{}

// CreateRole godoc
// @Summary 创建角色
// @Description 在当前租户下创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateRoleRequest true "创建角色请求参数"
// @Success 200 {object} response.Response "角色创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "角色名称已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles [post]
func (c *RBACController) CreateRole(ctx *gin.Context) {
	var req service.CreateRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	role, err := service.ServiceGroupApp.RBACService.CreateRole(ctx.Request.Context(), req)
	if err != nil {
		if err.Error() == "角色名称已存在" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "角色创建成功", role)
}

// AssignRoleToUser godoc
// @Summary 分配角色给用户
// @Description 将指定角色分配给用户
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.AssignRoleRequest true "分配角色请求参数"
// @Success 200 {object} response.Response "角色分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户或角色不存在"
// @Failure 409 {object} response.Response "角色已分配给用户"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles/assign [post]
func (c *RBACController) AssignRoleToUser(ctx *gin.Context) {
	var req service.AssignRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err := service.ServiceGroupApp.RBACService.AssignRoleToUser(ctx.Request.Context(), req.UserID, req.RoleID)
	if err != nil {
		if err.Error() == "用户不存在" || err.Error() == "角色不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else if err.Error() == "角色已分配给用户" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "角色分配成功", nil)
}

// GetUserRoles godoc
// @Summary 获取用户角色列表
// @Description 获取指定用户的所有角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userID path int true "用户ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/users/{userID}/roles [get]
func (c *RBACController) GetUserRoles(ctx *gin.Context) {
	userIDStr := ctx.Param("userID")
	if userIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "用户ID不能为空")
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "用户ID格式错误")
		return
	}

	roles, err := service.ServiceGroupApp.RBACService.GetUserRoles(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", roles)
}

// GetUserPermissions godoc
// @Summary 获取用户权限列表
// @Description 获取指定用户的所有权限（菜单和API权限）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userID path int true "用户ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/users/{userID}/permissions [get]
func (c *RBACController) GetUserPermissions(ctx *gin.Context) {
	userIDStr := ctx.Param("userID")
	if userIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "用户ID不能为空")
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "用户ID格式错误")
		return
	}

	permissions, err := service.ServiceGroupApp.RBACService.GetUserPermissions(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", permissions)
}

// RemoveRoleFromUser godoc
// @Summary 移除用户角色
// @Description 移除指定用户的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.RemoveRoleRequest true "移除角色请求参数"
// @Success 200 {object} response.Response "角色移除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户未分配该角色"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles/remove [post]
func (c *RBACController) RemoveRoleFromUser(ctx *gin.Context) {
	var req service.RemoveRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err := service.ServiceGroupApp.RBACService.RemoveRoleFromUser(ctx.Request.Context(), req.UserID, req.RoleID)
	if err != nil {
		if err.Error() == "用户未分配该角色" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "角色移除成功", nil)
}

// GetRoleList godoc
// @Summary 获取角色列表
// @Description 分页查询当前租户下的角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles [get]
func (c *RBACController) GetRoleList(ctx *gin.Context) {
	var req service.GetRoleListRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	resp, err := service.ServiceGroupApp.RBACService.GetRoleList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", resp)
}

// GetRoleByID godoc
// @Summary 获取角色详情
// @Description 根据角色ID获取角色详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleID path int true "角色ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles/{roleID} [get]
func (c *RBACController) GetRoleByID(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleID")
	if roleIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "角色ID不能为空")
		return
	}

	var roleID int64
	if _, err := fmt.Sscanf(roleIDStr, "%d", &roleID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "角色ID格式错误")
		return
	}

	role, err := service.ServiceGroupApp.RBACService.GetRoleByID(ctx.Request.Context(), roleID)
	if err != nil {
		if err.Error() == "角色不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", role)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 根据角色ID更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleID path int true "角色ID"
// @Param body body service.UpdateRoleRequest true "更新角色请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "角色不存在或名称冲突"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles/{roleID} [put]
func (c *RBACController) UpdateRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleID")
	if roleIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "角色ID不能为空")
		return
	}

	var roleID int64
	if _, err := fmt.Sscanf(roleIDStr, "%d", &roleID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "角色ID格式错误")
		return
	}

	var req service.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	role, err := service.ServiceGroupApp.RBACService.UpdateRole(ctx.Request.Context(), roleID, req)
	if err != nil {
		if err.Error() == "角色不存在" || err.Error() == "角色名称已存在" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", role)
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 根据角色ID删除角色，超级管理员角色不可删除
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleID path int true "角色ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "超级管理员角色不可删除"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/rbac/roles/{roleID} [delete]
func (c *RBACController) DeleteRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleID")
	if roleIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "角色ID不能为空")
		return
	}

	var roleID int64
	if _, err := fmt.Sscanf(roleIDStr, "%d", &roleID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "角色ID格式错误")
		return
	}

	err := service.ServiceGroupApp.RBACService.DeleteRole(ctx.Request.Context(), roleID)
	if err != nil {
		if err.Error() == "角色不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else if err.Error() == "超级管理员角色不可删除" {
			response.GinError(ctx, errors.ErrPermissionDenied, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}
