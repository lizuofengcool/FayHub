package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

// RBACController RBAC权限控制器
// @Summary RBAC权限控制器
// @Description 处理角色、权限相关的HTTP请求
// @Tags 权限管理
// @Security ApiKeyAuth
type RBACController struct{}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param createRoleRequest body service.CreateRoleRequest true "创建角色请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "角色已存在"
// @Router /api/rbac/roles [post]
func (c *RBACController) CreateRole(ctx *gin.Context) {
	var req service.CreateRoleRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	// 调用Service层创建角色逻辑
	role, err := service.ServiceGroupApp.RBACService.CreateRole(ctx.Request.Context(), req)
	if err != nil {
		if err.Error() == "角色名称已存在" {
			response.GinError(ctx, errors.ConflictError.Code, err.Error())
		} else {
			response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		}
		return
	}

	// 返回创建成功响应
	response.GinSuccessWithMessage(ctx, "角色创建成功", role)
}

// AssignRoleToUser 为用户分配角色
// @Summary 为用户分配角色
// @Description 为用户分配角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param assignRoleRequest body service.AssignRoleRequest true "分配角色请求"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户或角色不存在"
// @Failure 409 {object} map[string]interface{} "角色已分配"
// @Router /api/rbac/assign-role [post]
func (c *RBACController) AssignRoleToUser(ctx *gin.Context) {
	var req service.AssignRoleRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	// 调用Service层分配角色逻辑
	err := service.ServiceGroupApp.RBACService.AssignRoleToUser(ctx.Request.Context(), req.UserID, req.RoleID)
	if err != nil {
		if err.Error() == "用户不存在" || err.Error() == "角色不存在" {
			response.GinError(ctx, errors.ResourceNotFoundError.Code, err.Error())
		} else if err.Error() == "角色已分配给用户" {
			response.GinError(ctx, errors.ConflictError.Code, err.Error())
		} else {
			response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		}
		return
	}

	// 返回分配成功响应
	response.GinSuccessWithMessage(ctx, "角色分配成功", nil)
}

// GetUserRoles 获取用户角色列表
// @Summary 获取用户角色列表
// @Description 获取指定用户的角色列表
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userID path int true "用户ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Router /api/rbac/users/{userID}/roles [get]
func (c *RBACController) GetUserRoles(ctx *gin.Context) {
	userIDStr := ctx.Param("userID")
	if userIDStr == "" {
		response.GinError(ctx, errors.ParamMissingError.Code, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		response.GinError(ctx, errors.ParamFormatError.Code, "用户ID格式错误")
		return
	}

	// 调用Service层获取用户角色逻辑
	roles, err := service.ServiceGroupApp.RBACService.GetUserRoles(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	// 返回获取成功响应
	response.GinSuccessWithMessage(ctx, "获取成功", roles)
}

// GetUserPermissions 获取用户权限列表
// @Summary 获取用户权限列表
// @Description 获取指定用户的权限列表
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userID path int true "用户ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Router /api/rbac/users/{userID}/permissions [get]
func (c *RBACController) GetUserPermissions(ctx *gin.Context) {
	userIDStr := ctx.Param("userID")
	if userIDStr == "" {
		response.GinError(ctx, errors.ParamMissingError.Code, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		response.GinError(ctx, errors.ParamFormatError.Code, "用户ID格式错误")
		return
	}

	// 调用Service层获取用户权限逻辑
	permissions, err := service.ServiceGroupApp.RBACService.GetUserPermissions(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	// 返回获取成功响应
	response.GinSuccessWithMessage(ctx, "获取成功", permissions)
}

// RemoveRoleFromUser 移除用户的角色
// @Summary 移除用户的角色
// @Description 移除用户的角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param removeRoleRequest body service.RemoveRoleRequest true "移除角色请求"
// @Success 200 {object} map[string]interface{} "移除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户或角色不存在"
// @Router /api/rbac/remove-role [post]
func (c *RBACController) RemoveRoleFromUser(ctx *gin.Context) {
	var req service.RemoveRoleRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	// 调用Service层移除角色逻辑
	err := service.ServiceGroupApp.RBACService.RemoveRoleFromUser(ctx.Request.Context(), req.UserID, req.RoleID)
	if err != nil {
		if err.Error() == "用户未分配该角色" {
			response.GinError(ctx, errors.ResourceNotFoundError.Code, err.Error())
		} else {
			response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		}
		return
	}

	// 返回移除成功响应
	response.GinSuccessWithMessage(ctx, "角色移除成功", nil)
}
