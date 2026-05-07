package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type MenuController struct{}

// CreateMenu godoc
// @Summary 创建菜单
// @Description 创建新的菜单项
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateMenuRequest true "创建菜单请求参数"
// @Success 200 {object} response.Response "菜单创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus [post]
func (c *MenuController) CreateMenu(ctx *gin.Context) {
	var req service.CreateMenuRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	menu, err := service.ServiceGroupApp.MenuService.CreateMenu(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "菜单创建成功", menu)
}

// GetMenuList godoc
// @Summary 获取菜单列表
// @Description 分页查询菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param parent_id query int false "父菜单ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus [get]
func (c *MenuController) GetMenuList(ctx *gin.Context) {
	var req service.GetMenuListRequest

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

	resp, err := service.ServiceGroupApp.MenuService.GetMenuList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", resp)
}

// GetMenuTree godoc
// @Summary 获取菜单树
// @Description 获取完整的菜单树形结构
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "获取成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/tree [get]
func (c *MenuController) GetMenuTree(ctx *gin.Context) {
	menus, err := service.ServiceGroupApp.MenuService.GetMenuTree(ctx.Request.Context())
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", menus)
}

// GetMenuByID godoc
// @Summary 获取菜单详情
// @Description 根据菜单ID获取菜单详细信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menuID path int true "菜单ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "菜单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/{menuID} [get]
func (c *MenuController) GetMenuByID(ctx *gin.Context) {
	menuIDStr := ctx.Param("menuID")
	if menuIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "菜单ID不能为空")
		return
	}

	var menuID int64
	if _, err := fmt.Sscanf(menuIDStr, "%d", &menuID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "菜单ID格式错误")
		return
	}

	menu, err := service.ServiceGroupApp.MenuService.GetMenuByID(ctx.Request.Context(), menuID)
	if err != nil {
		if err.Error() == "菜单不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", menu)
}

// UpdateMenu godoc
// @Summary 更新菜单
// @Description 根据菜单ID更新菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menuID path int true "菜单ID"
// @Param body body service.UpdateMenuRequest true "更新菜单请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "菜单不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/{menuID} [put]
func (c *MenuController) UpdateMenu(ctx *gin.Context) {
	menuIDStr := ctx.Param("menuID")
	if menuIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "菜单ID不能为空")
		return
	}

	var menuID int64
	if _, err := fmt.Sscanf(menuIDStr, "%d", &menuID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "菜单ID格式错误")
		return
	}

	var req service.UpdateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	menu, err := service.ServiceGroupApp.MenuService.UpdateMenu(ctx.Request.Context(), menuID, req)
	if err != nil {
		if err.Error() == "菜单不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", menu)
}

// DeleteMenu godoc
// @Summary 删除菜单
// @Description 根据菜单ID删除菜单，存在子菜单时不可删除
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menuID path int true "菜单ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "菜单不存在"
// @Failure 409 {object} response.Response "存在子菜单，不可删除"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/{menuID} [delete]
func (c *MenuController) DeleteMenu(ctx *gin.Context) {
	menuIDStr := ctx.Param("menuID")
	if menuIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "菜单ID不能为空")
		return
	}

	var menuID int64
	if _, err := fmt.Sscanf(menuIDStr, "%d", &menuID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "菜单ID格式错误")
		return
	}

	err := service.ServiceGroupApp.MenuService.DeleteMenu(ctx.Request.Context(), menuID)
	if err != nil {
		if err.Error() == "菜单不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else if err.Error() == "存在子菜单，不可删除" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

// AssignRoleMenus godoc
// @Summary 分配角色菜单权限
// @Description 为指定角色分配菜单权限
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.AssignRoleMenuRequest true "分配角色菜单请求参数"
// @Success 200 {object} response.Response "菜单权限分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/assign-roles [post]
func (c *MenuController) AssignRoleMenus(ctx *gin.Context) {
	var req service.AssignRoleMenuRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err := service.ServiceGroupApp.MenuService.AssignRoleMenus(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "菜单权限分配成功", nil)
}

// GetRoleMenus godoc
// @Summary 获取角色菜单列表
// @Description 获取指定角色的菜单权限列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleID path int true "角色ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/menus/roles/{roleID} [get]
func (c *MenuController) GetRoleMenus(ctx *gin.Context) {
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

	menus, err := service.ServiceGroupApp.MenuService.GetRoleMenus(ctx.Request.Context(), roleID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", menus)
}
