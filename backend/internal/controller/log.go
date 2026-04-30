package controller

import (
	"fayhub/pkg/logger"
	"fayhub/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LogController struct{}

func (lc *LogController) QueryTenantLogs(c *gin.Context) {
	tenantIDStr := c.Param("tenantId")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的租户ID")
		return
	}

	level := c.Query("level")
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	var startTime, endTime time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	entries, err := logger.QueryTenantLogs(uint(tenantID), level, limit, startTime, endTime)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	if entries == nil {
		entries = []*logger.TenantLogEntry{}
	}

	response.GinSuccess(c, gin.H{
		"tenant_id": tenantID,
		"count":     len(entries),
		"logs":      entries,
	})
}

func (lc *LogController) GetTenantLogCount(c *gin.Context) {
	tenantIDStr := c.Param("tenantId")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的租户ID")
		return
	}

	count, err := logger.GetTenantLogCount(uint(tenantID))
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"tenant_id": tenantID,
		"count":     count,
	})
}
