package controller

import (
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type SystemSettingController struct{}

func (ssc *SystemSettingController) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()

	settings, err := service.ServiceGroupApp.SystemSettingService.GetSettings(ctx)
	if err != nil {
		response.GinError(c, errs.ErrConfigNotLoaded, err.Error())
		return
	}

	response.GinSuccess(c, settings)
}

func (ssc *SystemSettingController) UpdateSettings(c *gin.Context) {
	var req service.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.SystemSettingService.UpdateSettings(ctx, req); err != nil {
		response.GinError(c, errs.ErrConfigNotLoaded, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "系统设置更新成功", nil)
}
