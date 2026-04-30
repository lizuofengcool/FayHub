package controller

import (
	"fayhub/pkg/plugin"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type EngineController struct{}

func (ec *EngineController) GetEngineStatus(c *gin.Context) {
	engine := plugin.GetEngine()

	status := gin.H{
		"engine_type":  "noop",
		"is_running":   false,
		"plugin_count": 0,
	}

	if engine == nil {
		response.GinSuccess(c, status)
		return
	}

	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		response.GinSuccess(c, status)
		return
	}

	if !wasmEngine.IsRunning() {
		status = gin.H{
			"engine_type":  "wasm",
			"is_running":   false,
			"plugin_count": 0,
		}
		response.GinSuccess(c, status)
		return
	}

	infos := wasmEngine.GetPluginInfos()
	status = gin.H{
		"engine_type":  "wasm",
		"is_running":   true,
		"plugin_count": len(infos),
	}

	response.GinSuccess(c, status)
}

func (ec *EngineController) GetLoadedPlugins(c *gin.Context) {
	engine := plugin.GetEngine()

	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		response.GinSuccess(c, []interface{}{})
		return
	}

	infos := wasmEngine.GetPluginInfos()
	response.GinSuccess(c, infos)
}

func (ec *EngineController) GetPluginRoutes(c *gin.Context) {
	pluginID := c.Param("id")
	tenantID := getTenantIDFromContext(c)

	engine := plugin.GetEngine()
	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		response.GinSuccess(c, gin.H{
			"routes": []interface{}{},
			"apis":   []interface{}{},
		})
		return
	}

	registry := wasmEngine.GetRegistry()
	routes := registry.GetRoutes(tenantID, pluginID)
	apis := registry.GetAPIs(tenantID, pluginID)

	response.GinSuccess(c, gin.H{
		"routes": routes,
		"apis":   apis,
	})
}

func (ec *EngineController) HealthCheckPlugin(c *gin.Context) {
	pluginID := c.Param("id")
	ctx := c.Request.Context()
	tenantID := getTenantIDFromContext(c)

	engine := plugin.GetEngine()
	if healthErr := engine.HealthCheck(ctx, tenantID, pluginID); healthErr != nil {
		response.GinError(c, 500, "插件状态异常: "+healthErr.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件运行正常", gin.H{
		"plugin_id": pluginID,
		"status":    "healthy",
	})
}

func getTenantIDFromContext(c *gin.Context) uint {
	if tid, exists := c.Get("tenant_id"); exists {
		if id, ok := tid.(uint); ok {
			return id
		}
		if id, ok := tid.(float64); ok {
			return uint(id)
		}
	}
	return 0
}
