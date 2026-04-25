package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantController struct{}

// CreateTenant
// @Summary 创建租户
// @Description 平台超管创建新租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body service.CreateTenantRequest true "创建租户请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Router /api/tenants [post]
func (tc *TenantController) CreateTenant(ctx *gin.Context) {
	var req service.CreateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.Create(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", tenant)
}

// UpdateTenant
// @Summary 更新租户
// @Description 平台超管更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "租户ID"
// @Param body body service.UpdateTenantRequest true "更新租户请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Router /api/tenants/{id} [put]
func (tc *TenantController) UpdateTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "无效的租户ID")
		return
	}

	var req service.UpdateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.Update(ctx.Request.Context(), uint(id), req)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", tenant)
}

// DeleteTenant
// @Summary 删除租户
// @Description 平台超管删除租户（软删除）
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Router /api/tenants/{id} [delete]
func (tc *TenantController) DeleteTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "无效的租户ID")
		return
	}

	if err := service.ServiceGroupApp.TenantService.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

// GetTenant
// @Summary 获取租户详情
// @Description 根据ID获取租户详细信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Router /api/tenants/{id} [get]
func (tc *TenantController) GetTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "无效的租户ID")
		return
	}

	tenant, err := service.ServiceGroupApp.TenantService.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", tenant)
}

// GetTenantList
// @Summary 获取租户列表
// @Description 分页获取租户列表，支持关键词搜索和状态筛选
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param status query int false "状态筛选"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Router /api/tenants [get]
func (tc *TenantController) GetTenantList(ctx *gin.Context) {
	var req service.TenantListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	result, err := service.ServiceGroupApp.TenantService.GetList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", result)
}
