package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantChannelController struct{}

func (tcc *TenantChannelController) ListConfigs(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	channelType := c.Query("channel_type")

	var status *int
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	configs, total, err := service.TenantChannelServiceApp.List(ctx, channelType, status, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询渠道配置失败")
		return
	}

	if configs == nil {
		configs = []*model.TenantChannelConfig{}
	}

	response.GinSuccess(c, gin.H{
		"list":      configs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (tcc *TenantChannelController) GetConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	config, err := service.TenantChannelServiceApp.GetByID(ctx, id)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询渠道配置失败")
		return
	}

	response.GinSuccess(c, config)
}

func (tcc *TenantChannelController) GetConfigByChannelType(c *gin.Context) {
	channelType := c.Param("channel_type")
	if channelType == "" {
		response.GinError(c, errs.ErrParamMissing, "无效的渠道类型")
		return
	}

	ctx := c.Request.Context()
	config, err := service.TenantChannelServiceApp.GetByChannelType(ctx, channelType)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询渠道配置失败")
		return
	}

	response.GinSuccess(c, config)
}

func (tcc *TenantChannelController) CreateConfig(c *gin.Context) {
	var req service.CreateChannelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	config, err := service.TenantChannelServiceApp.Create(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "创建渠道配置失败")
		return
	}

	response.GinSuccess(c, config)
}

func (tcc *TenantChannelController) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	var req service.UpdateChannelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	config, err := service.TenantChannelServiceApp.Update(ctx, id, req)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "更新渠道配置失败")
		return
	}

	response.GinSuccess(c, config)
}

func (tcc *TenantChannelController) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.TenantChannelServiceApp.Delete(ctx, id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除渠道配置失败")
		return
	}

	response.GinSuccess(c, nil)
}

func (tcc *TenantChannelController) GetUserBindings(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的用户ID")
		return
	}
	channelType := c.Query("channel_type")

	ctx := c.Request.Context()
	bindings, err := service.TenantChannelServiceApp.GetThirdPartyBindings(ctx, userID, channelType)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询用户绑定失败")
		return
	}

	if bindings == nil {
		bindings = []*model.UserThirdParty{}
	}

	response.GinSuccess(c, gin.H{
		"list": bindings,
	})
}

func (tcc *TenantChannelController) DeleteUserBinding(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.TenantChannelServiceApp.DeleteThirdPartyBinding(ctx, id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除用户绑定失败")
		return
	}

	response.GinSuccess(c, nil)
}
