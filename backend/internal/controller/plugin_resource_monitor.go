package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/plugin"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type PluginResourceMonitorController struct{}

func (ctrl *PluginResourceMonitorController) GetRuntimeStats(c *gin.Context) {
	pluginID := c.Query("plugin_id")
	tenantID := c.GetInt64("tenant_id")

	svc := &service.PluginResourceMonitorService{}

	if pluginID != "" {
		stats := svc.GetRuntimeStats(tenantID, pluginID)
		if stats == nil {
			response.GinSuccess(c, gin.H{"plugin_id": pluginID, "status": "no_data"})
			return
		}
		response.GinSuccess(c, stats)
		return
	}

	allStats := svc.GetAllRuntimeStats(tenantID)
	if allStats == nil {
		allStats = make([]*plugin.PluginRuntimeStats, 0)
	}
	response.GinSuccess(c, allStats)
}

func (ctrl *PluginResourceMonitorController) GetDBStats(c *gin.Context) {
	pluginID := c.Query("plugin_id")
	ctx := c.Request.Context()

	svc := &service.PluginResourceMonitorService{}

	if pluginID != "" {
		stats, err := svc.GetPluginStats(ctx, pluginID)
		if err != nil {
			response.GinSuccess(c, gin.H{"plugin_id": pluginID, "status": "no_data"})
			return
		}
		response.GinSuccess(c, stats)
		return
	}

	summary, err := svc.GetStatsSummary(ctx)
	if err != nil {
		response.GinSuccess(c, gin.H{"status": "no_data"})
		return
	}
	response.GinSuccess(c, summary)
}

func (ctrl *PluginResourceMonitorController) GetAlerts(c *gin.Context) {
	pluginID := c.Query("plugin_id")
	ctx := c.Request.Context()

	svc := &service.PluginResourceMonitorService{}

	limit := 20
	alerts, err := svc.GetRecentAlerts(ctx, pluginID, limit)
	if err != nil {
		response.GinSuccess(c, make([]model.PluginEventLog, 0))
		return
	}

	if alerts == nil {
		alerts = make([]model.PluginEventLog, 0)
	}
	response.GinSuccess(c, alerts)
}

func (ctrl *PluginResourceMonitorController) ResetStats(c *gin.Context) {
	pluginID := c.Query("plugin_id")
	if pluginID == "" {
		response.GinError(c, 40000, "缺少plugin_id参数")
		return
	}

	tenantID := c.GetInt64("tenant_id")
	svc := &service.PluginResourceMonitorService{}
	svc.ResetStats(tenantID, pluginID)

	response.GinSuccessWithMessage(c, "统计已重置", nil)
}
