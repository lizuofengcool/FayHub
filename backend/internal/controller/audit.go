package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/export"
	"fayhub/pkg/response"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditController struct{}

func (ac *AuditController) ListAuditLogs(c *gin.Context) {
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

	filters := &service.AuditLogFilters{
		UserID:     parseInt64Query(c, "user_id"),
		Action:     c.Query("action"),
		Resource:   c.Query("resource"),
		ResourceID: c.Query("resource_id"),
		IP:         c.Query("ip"),
		Path:       c.Query("path"),
	}

	if s := c.Query("success"); s != "" {
		val := s == "true" || s == "1"
		filters.Success = &val
	}

	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			filters.StartTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			filters.EndTime = &t
		}
	}

	ctx := c.Request.Context()
	logs, total, err := service.AuditServiceApp.List(ctx, filters, page, pageSize)
	if err != nil {
		response.GinError(c, 50000, "查询审计日志失败: "+err.Error())
		return
	}

	if logs == nil {
		logs = []*model.AuditLog{}
	}

	response.GinSuccess(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ac *AuditController) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的日志ID")
		return
	}

	ctx := c.Request.Context()
	log, err := service.AuditServiceApp.GetByID(ctx, id)
	if err != nil {
		response.GinError(c, 40400, "审计日志不存在")
		return
	}

	response.GinSuccess(c, log)
}

func (ac *AuditController) GetAuditStats(c *gin.Context) {
	var startTime, endTime time.Time
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = t
		}
	}

	ctx := c.Request.Context()
	stats, err := service.AuditServiceApp.GetStats(ctx, startTime, endTime)
	if err != nil {
		response.GinError(c, 50000, "获取审计统计失败: "+err.Error())
		return
	}

	response.GinSuccess(c, stats)
}

func (ac *AuditController) CleanupAuditLogs(c *gin.Context) {
	var req struct {
		BeforeTime string `json:"before_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误: "+err.Error())
		return
	}

	before, err := time.Parse(time.RFC3339, req.BeforeTime)
	if err != nil {
		response.GinError(c, 40001, "时间格式错误，请使用RFC3339格式")
		return
	}

	ctx := c.Request.Context()
	affected, err := service.AuditServiceApp.Cleanup(ctx, before)
	if err != nil {
		response.GinError(c, 50000, "清理审计日志失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"affected_rows": affected,
	})
}

func parseInt64Query(c *gin.Context, key string) int64 {
	if v := c.Query(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func (ac *AuditController) ExportAuditLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "xlsx" {
		response.GinError(c, 40001, "不支持的导出格式，仅支持csv和xlsx")
		return
	}

	filters := &service.AuditLogFilters{
		UserID:     parseInt64Query(c, "user_id"),
		Action:     c.Query("action"),
		Resource:   c.Query("resource"),
		ResourceID: c.Query("resource_id"),
		IP:         c.Query("ip"),
		Path:       c.Query("path"),
	}

	if s := c.Query("success"); s != "" {
		val := s == "true" || s == "1"
		filters.Success = &val
	}

	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			filters.StartTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			filters.EndTime = &t
		}
	}

	ctx := c.Request.Context()
	logs, _, err := service.AuditServiceApp.List(ctx, filters, 1, 10000)
	if err != nil {
		response.GinError(c, 50000, "查询审计日志失败: "+err.Error())
		return
	}

	columns := []export.ExportColumn{
		{Header: "ID", Field: "ID"},
		{Header: "用户名", Field: "Username"},
		{Header: "操作", Field: "Action"},
		{Header: "资源", Field: "Resource"},
		{Header: "资源ID", Field: "ResourceID"},
		{Header: "IP", Field: "IP"},
		{Header: "方法", Field: "Method"},
		{Header: "路径", Field: "Path"},
		{Header: "状态码", Field: "StatusCode"},
		{Header: "成功", Field: "Success"},
		{Header: "错误信息", Field: "ErrorMsg"},
		{Header: "耗时(ms)", Field: "Duration"},
		{Header: "时间", Field: "CreatedAt"},
	}

	var data []byte
	var filename string

	if format == "xlsx" {
		data, err = export.ExportStructsExcel(logs, columns, "审计日志")
		filename = fmt.Sprintf("audit_logs_%s.xlsx", time.Now().Format("20060102150405"))
	} else {
		data, err = export.ExportStructsCSV(logs, columns)
		filename = fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102150405"))
	}

	if err != nil {
		response.GinError(c, 50000, "导出失败: "+err.Error())
		return
	}

	contentType := "text/csv; charset=utf-8"
	if format == "xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, data)
}
