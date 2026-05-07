package controller

import (
	"fayhub/pkg/metrics"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type MonitorController struct{}

func (mc *MonitorController) GetHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": metrics.GetMetrics()["uptime_seconds"],
		"services": map[string]interface{}{
			"api":    "healthy",
			"db":     checkDBHealth(),
			"cache":  checkCacheHealth(),
			"plugin": checkPluginEngineHealth(),
		},
	}

	response.GinSuccess(c, health)
}

func (mc *MonitorController) GetSystemMetrics(c *gin.Context) {
	metricsData := metrics.GetMetrics()
	response.GinSuccess(c, metricsData)
}

func (mc *MonitorController) GetPrometheusMetrics(c *gin.Context) {
	promMetrics := metrics.GetPrometheusFormat()
	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(200, promMetrics)
}

func checkDBHealth() string {
	db := metrics.GetMetrics()["db_pool_stats"]
	if db == nil {
		return "unhealthy"
	}
	return "healthy"
}

func checkCacheHealth() string {
	return "healthy"
}

func checkPluginEngineHealth() string {
	return "healthy"
}
