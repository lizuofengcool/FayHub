package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type APIController struct{}

// CreateAPI godoc
// @Summary 创建API接口
// @Description 创建新的API接口权限
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateAPIRequest true "创建API接口请求参数"
// @Success 200 {object} response.Response "API接口创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "API接口已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis [post]
func (c *APIController) CreateAPI(ctx *gin.Context) {
	var req service.CreateAPIRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	api, err := service.ServiceGroupApp.APIService.CreateAPI(ctx.Request.Context(), req)
	if err != nil {
		if err.Error() == "API接口已存在" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "API接口创建成功", api)
}

// GetAPIList godoc
// @Summary 获取API接口列表
// @Description 分页查询API接口列表
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param method query string false "请求方法筛选"
// @Param group query string false "API分组筛选"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis [get]
func (c *APIController) GetAPIList(ctx *gin.Context) {
	var req service.GetAPIListRequest

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

	resp, err := service.ServiceGroupApp.APIService.GetAPIList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", resp)
}

// GetAPIByID godoc
// @Summary 获取API接口详情
// @Description 根据API接口ID获取详细信息
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param apiID path int true "API接口ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "API接口不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis/{apiID} [get]
func (c *APIController) GetAPIByID(ctx *gin.Context) {
	apiIDStr := ctx.Param("apiID")
	if apiIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "API接口ID不能为空")
		return
	}

	var apiID int64
	if _, err := fmt.Sscanf(apiIDStr, "%d", &apiID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "API接口ID格式错误")
		return
	}

	api, err := service.ServiceGroupApp.APIService.GetAPIByID(ctx.Request.Context(), apiID)
	if err != nil {
		if err.Error() == "API接口不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", api)
}

// UpdateAPI godoc
// @Summary 更新API接口
// @Description 根据API接口ID更新接口信息
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param apiID path int true "API接口ID"
// @Param body body service.UpdateAPIRequest true "更新API接口请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "API接口不存在或路径冲突"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis/{apiID} [put]
func (c *APIController) UpdateAPI(ctx *gin.Context) {
	apiIDStr := ctx.Param("apiID")
	if apiIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "API接口ID不能为空")
		return
	}

	var apiID int64
	if _, err := fmt.Sscanf(apiIDStr, "%d", &apiID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "API接口ID格式错误")
		return
	}

	var req service.UpdateAPIRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	api, err := service.ServiceGroupApp.APIService.UpdateAPI(ctx.Request.Context(), apiID, req)
	if err != nil {
		if err.Error() == "API接口不存在" || err.Error() == "API接口已存在" {
			response.GinError(ctx, errors.ErrConflict, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", api)
}

// DeleteAPI godoc
// @Summary 删除API接口
// @Description 根据API接口ID删除接口
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param apiID path int true "API接口ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "API接口不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis/{apiID} [delete]
func (c *APIController) DeleteAPI(ctx *gin.Context) {
	apiIDStr := ctx.Param("apiID")
	if apiIDStr == "" {
		response.GinError(ctx, errors.ErrParamMissing, "API接口ID不能为空")
		return
	}

	var apiID int64
	if _, err := fmt.Sscanf(apiIDStr, "%d", &apiID); err != nil {
		response.GinError(ctx, errors.ErrParamFormat, "API接口ID格式错误")
		return
	}

	err := service.ServiceGroupApp.APIService.DeleteAPI(ctx.Request.Context(), apiID)
	if err != nil {
		if err.Error() == "API接口不存在" {
			response.GinError(ctx, errors.ErrResourceNotFound, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

// AssignRoleAPIs godoc
// @Summary 分配角色API权限
// @Description 为指定角色分配API接口权限
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.AssignRoleAPIRequest true "分配角色API权限请求参数"
// @Success 200 {object} response.Response "API权限分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis/assign-roles [post]
func (c *APIController) AssignRoleAPIs(ctx *gin.Context) {
	var req service.AssignRoleAPIRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err := service.ServiceGroupApp.APIService.AssignRoleAPIs(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "API权限分配成功", nil)
}

// GetRoleAPIs godoc
// @Summary 获取角色API权限列表
// @Description 获取指定角色的API接口权限列表
// @Tags API管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleID path int true "角色ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/apis/roles/{roleID} [get]
func (c *APIController) GetRoleAPIs(ctx *gin.Context) {
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

	apis, err := service.ServiceGroupApp.APIService.GetRoleAPIs(ctx.Request.Context(), roleID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", apis)
}
