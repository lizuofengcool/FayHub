package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type StatsController struct{}

func (sc *StatsController) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.StatsService{}

	stats, err := svc.GetDashboardStats(ctx)
	if err != nil {
		response.GinError(c, 500, "获取统计数据失败")
		return
	}

	response.GinSuccess(c, stats)
}
