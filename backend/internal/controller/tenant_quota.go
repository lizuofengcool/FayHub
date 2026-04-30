package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantQuotaController struct{}

func (qc *TenantQuotaController) GetQuota(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	svc := &service.TenantQuotaService{}
	quota, err := svc.GetQuota(ctx.Request.Context(), uint(tenantID))
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, quota)
}

func (qc *TenantQuotaController) UpdateQuota(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	var req service.UpdateQuotaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, 40000, "请求参数错误: "+err.Error())
		return
	}

	svc := &service.TenantQuotaService{}
	quota, err := svc.UpdateQuota(ctx.Request.Context(), uint(tenantID), req)
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, quota)
}

func (qc *TenantQuotaController) CheckUserQuota(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	svc := &service.TenantQuotaService{}
	result, err := svc.CheckUserQuota(ctx.Request.Context(), uint(tenantID))
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, result)
}

func (qc *TenantQuotaController) CheckStorageQuota(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	requiredMB, _ := strconv.Atoi(ctx.DefaultQuery("required_mb", "0"))

	svc := &service.TenantQuotaService{}
	result, err := svc.CheckStorageQuota(ctx.Request.Context(), uint(tenantID), requiredMB)
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, result)
}

func (qc *TenantQuotaController) CheckPluginQuota(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	svc := &service.TenantQuotaService{}
	result, err := svc.CheckPluginQuota(ctx.Request.Context(), uint(tenantID))
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, result)
}

func (qc *TenantQuotaController) SyncUsage(ctx *gin.Context) {
	tenantID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, 40000, "无效的租户ID")
		return
	}

	svc := &service.TenantQuotaService{}
	_ = svc.SyncUserCount(ctx.Request.Context(), uint(tenantID))
	_ = svc.SyncPluginCount(ctx.Request.Context(), uint(tenantID))

	quota, err := svc.GetQuota(ctx.Request.Context(), uint(tenantID))
	if err != nil {
		response.GinError(ctx, 50000, err.Error())
		return
	}

	response.GinSuccess(ctx, quota)
}
