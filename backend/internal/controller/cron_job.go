package controller

import (
	"strconv"

	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type CronJobController struct{}

func (ctrl *CronJobController) Create(ctx *gin.Context) {
	var job model.CronJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.CronJobService.Create(ctx.Request.Context(), &job); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", job)
}

func (ctrl *CronJobController) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	var job model.CronJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	job.ID = id

	if err := service.ServiceGroupApp.CronJobService.Update(ctx.Request.Context(), &job); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", job)
}

func (ctrl *CronJobController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.CronJobService.Delete(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

func (ctrl *CronJobController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	job, err := service.ServiceGroupApp.CronJobService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, job)
}

func (ctrl *CronJobController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	jobs, total, err := service.ServiceGroupApp.CronJobService.List(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      jobs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *CronJobController) ToggleStatus(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.CronJobService.ToggleStatus(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "操作成功", nil)
}

func (ctrl *CronJobController) ExecuteOnce(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	if err := service.ServiceGroupApp.CronJobService.ExecuteOnce(ctx.Request.Context(), id); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "已触发执行", nil)
}

func (ctrl *CronJobController) GetLogs(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	logs, total, err := service.ServiceGroupApp.CronJobService.GetLogs(ctx.Request.Context(), id, page, pageSize)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(ctx, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
