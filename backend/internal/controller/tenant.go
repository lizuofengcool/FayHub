package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantController struct{}

// CreateTenant godoc
// @Summary 创建租户
// @Description 创建新的租户/商家
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateTenantRequest true "创建租户请求参数"
// @Success 200 {object} response.Response "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/tenants [post]
func (tc *TenantController) CreateTenant(ctx *gin.Context) {
	var req service.CreateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.Create(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", tenant)
}

// UpdateTenant godoc
// @Summary 更新租户
// @Description 根据租户ID更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Param body body service.UpdateTenantRequest true "更新租户请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/tenants/{id} [put]
func (tc *TenantController) UpdateTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的租户ID")
		return
	}

	var req service.UpdateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.Update(ctx.Request.Context(), uint(id), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", tenant)
}

// DeleteTenant godoc
// @Summary 删除租户
// @Description 根据租户ID删除租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/tenants/{id} [delete]
func (tc *TenantController) DeleteTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的租户ID")
		return
	}

	if err := service.ServiceGroupApp.TenantService.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

// GetTenant godoc
// @Summary 获取租户详情
// @Description 根据租户ID获取租户详细信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/tenants/{id} [get]
func (tc *TenantController) GetTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的租户ID")
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", tenant)
}

// GetTenantList godoc
// @Summary 获取租户列表
// @Description 分页查询租户列表，支持关键词搜索和状态筛选
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param status query int false "状态筛选(0:禁用,1:正常)"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/tenants [get]
func (tc *TenantController) GetTenantList(ctx *gin.Context) {
	var req service.TenantListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	result, err := service.ServiceGroupApp.TenantService.GetList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", result)
}
