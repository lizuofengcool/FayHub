package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginLogController struct{}

func (lc *LoginLogController) ListLoginLogs(c *gin.Context) {
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

	filters := &service.LoginLogFilters{
		Username:    c.Query("username"),
		LoginStatus: c.Query("login_status"),
		LoginIP:     c.Query("login_ip"),
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
	logs, total, err := service.LoginLogServiceApp.List(ctx, filters, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询登录日志失败")
		return
	}

	if logs == nil {
		logs = []*model.LoginLog{}
	}

	response.GinSuccess(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (lc *LoginLogController) CleanupLoginLogs(c *gin.Context) {
	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	before := time.Now().AddDate(0, 0, -days)
	ctx := c.Request.Context()
	affected, err := service.LoginLogServiceApp.Cleanup(ctx, before)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "清理登录日志失败")
		return
	}

	response.GinSuccess(c, gin.H{
		"affected": affected,
	})
}
