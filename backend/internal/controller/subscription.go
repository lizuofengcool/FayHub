package controller

import (
	"strconv"

	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type SubscriptionController struct{}

func (ctrl *SubscriptionController) Create(ctx *gin.Context) {
	var sub model.Subscription
	if err := ctx.ShouldBindJSON(&sub); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.SubscriptionService.Create(ctx.Request.Context(), &sub); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", sub)
}

func (ctrl *SubscriptionController) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	var sub model.Subscription
	if err := ctx.ShouldBindJSON(&sub); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	sub.ID = id

	if err := service.ServiceGroupApp.SubscriptionService.Update(ctx.Request.Context(), &sub); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", sub)
}

func (ctrl *SubscriptionController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.SubscriptionService.Delete(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

func (ctrl *SubscriptionController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	sub, err := service.ServiceGroupApp.SubscriptionService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, sub)
}

func (ctrl *SubscriptionController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	subs, total, err := service.ServiceGroupApp.SubscriptionService.List(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      subs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *SubscriptionController) Cancel(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.SubscriptionService.Cancel(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "已取消订阅", nil)
}

func (ctrl *SubscriptionController) Renew(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	var req struct {
		Months int `json:"months" binding:"required,min=1,max=36"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.SubscriptionService.Renew(ctx.Request.Context(), id, req.Months); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "续费成功", nil)
}

func (ctrl *SubscriptionController) GetInvoices(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	invoices, total, err := service.ServiceGroupApp.SubscriptionService.GetInvoices(ctx.Request.Context(), id, page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      invoices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
