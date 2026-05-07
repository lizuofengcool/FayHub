package controller

import (
	"strconv"

	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type NotificationChannelController struct{}

func (ctrl *NotificationChannelController) CreateChannel(ctx *gin.Context) {
	var ch model.NotificationChannel
	if err := ctx.ShouldBindJSON(&ch); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.NotificationChannelService.CreateChannel(ctx.Request.Context(), &ch); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", ch)
}

func (ctrl *NotificationChannelController) UpdateChannel(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	var ch model.NotificationChannel
	if err := ctx.ShouldBindJSON(&ch); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	ch.ID = id

	if err := service.ServiceGroupApp.NotificationChannelService.UpdateChannel(ctx.Request.Context(), &ch); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", ch)
}

func (ctrl *NotificationChannelController) DeleteChannel(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.NotificationChannelService.DeleteChannel(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

func (ctrl *NotificationChannelController) GetChannel(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	ch, err := service.ServiceGroupApp.NotificationChannelService.GetChannel(ctx.Request.Context(), id)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, ch)
}

func (ctrl *NotificationChannelController) ListChannels(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	channels, total, err := service.ServiceGroupApp.NotificationChannelService.ListChannels(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      channels,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *NotificationChannelController) CreateTemplate(ctx *gin.Context) {
	var tpl model.NotificationTemplate
	if err := ctx.ShouldBindJSON(&tpl); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.NotificationChannelService.CreateTemplate(ctx.Request.Context(), &tpl); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", tpl)
}

func (ctrl *NotificationChannelController) UpdateTemplate(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	var tpl model.NotificationTemplate
	if err := ctx.ShouldBindJSON(&tpl); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	tpl.ID = id

	if err := service.ServiceGroupApp.NotificationChannelService.UpdateTemplate(ctx.Request.Context(), &tpl); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", tpl)
}

func (ctrl *NotificationChannelController) DeleteTemplate(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.NotificationChannelService.DeleteTemplate(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

func (ctrl *NotificationChannelController) ListTemplates(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	templates, total, err := service.ServiceGroupApp.NotificationChannelService.ListTemplates(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *NotificationChannelController) Send(ctx *gin.Context) {
	var record model.NotificationRecord
	if err := ctx.ShouldBindJSON(&record); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.NotificationChannelService.Send(ctx.Request.Context(), &record); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "发送成功", nil)
}

func (ctrl *NotificationChannelController) GetRecords(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	records, total, err := service.ServiceGroupApp.NotificationChannelService.GetRecords(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
