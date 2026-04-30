package controller

import (
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type SettlementController struct{}

func (sc *SettlementController) CreateSettlement(c *gin.Context) {
	var req service.CreateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	settlementService := &service.SettlementService{}
	result, err := settlementService.CreateSettlement(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "分账记录创建成功", result)
}

func (sc *SettlementController) GetSettlementConfig(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID, _ := c.Get("tenant_id")

	settlementService := &service.SettlementService{}
	config, err := settlementService.GetSettlementConfig(ctx, tenantID.(uint))
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, config)
}

func (sc *SettlementController) UpdateSettlementConfig(c *gin.Context) {
	var req struct {
		PlatformRate int   `json:"platform_rate" binding:"required"`
		MinAmount    int64 `json:"min_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	tenantID, _ := c.Get("tenant_id")

	settlementService := &service.SettlementService{}
	if err := settlementService.UpdateSettlementConfig(ctx, tenantID.(uint), req.PlatformRate, req.MinAmount); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "分账配置更新成功", nil)
}

func (sc *SettlementController) ProcessSettlement(c *gin.Context) {
	settlementNo := c.Param("settlement_no")
	if settlementNo == "" {
		response.GinError(c, errs.ErrParamValidation, "缺少结算单号")
		return
	}

	ctx := c.Request.Context()
	settlementService := &service.SettlementService{}
	if err := settlementService.ProcessSettlement(ctx, settlementNo); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "分账处理成功", nil)
}

func (sc *SettlementController) GetSettlementStats(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID, _ := c.Get("tenant_id")

	settlementService := &service.SettlementService{}
	stats, err := settlementService.GetSettlementStats(ctx, tenantID.(uint))
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, stats)
}
