// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PluginEngineRouter struct{}

func (r *PluginEngineRouter) Init(router *gin.Engine) {
	pluginEngineGroup := router.Group("/api/plugin-engine")
	pluginEngineGroup.Use(middleware.JwtAuthMiddleware())
	pluginEngineGroup.Use(middleware.TenantMiddleware())
	{
		// 插件管理接口（需要租户权限）
		pluginEngineGroup.GET("/plugins", controller.ControllerGroupApp.PluginEngineController.ListPlugins)
		pluginEngineGroup.GET("/plugins/:id", controller.ControllerGroupApp.PluginEngineController.GetPlugin)
		pluginEngineGroup.DELETE("/plugins/:id", controller.ControllerGroupApp.PluginEngineController.UninstallPlugin)
		pluginEngineGroup.PUT("/plugins/:id/enable", controller.ControllerGroupApp.PluginEngineController.EnablePlugin)
		pluginEngineGroup.PUT("/plugins/:id/disable", controller.ControllerGroupApp.PluginEngineController.DisablePlugin)
		pluginEngineGroup.PUT("/plugins/:id/upgrade", controller.ControllerGroupApp.PluginEngineController.UpgradePlugin)
		pluginEngineGroup.GET("/plugins/:id/config", controller.ControllerGroupApp.PluginEngineController.GetPluginConfig)
		pluginEngineGroup.PUT("/plugins/:id/config", controller.ControllerGroupApp.PluginEngineController.UpdatePluginConfig)
		pluginEngineGroup.GET("/plugins/:id/page", controller.ControllerGroupApp.PluginEngineController.GetPluginPage)

		pluginEngineGroup.GET("/plugins/:id/versions", controller.ControllerGroupApp.PluginEngineController.GetVersionHistory)
		pluginEngineGroup.POST("/plugins/:id/rollback", controller.ControllerGroupApp.PluginEngineController.RollbackPlugin)
		pluginEngineGroup.GET("/plugins/:id/check-update", controller.ControllerGroupApp.PluginEngineController.CheckForUpdates)
		pluginEngineGroup.GET("/check-updates", controller.ControllerGroupApp.PluginEngineController.CheckAllUpdates)

		pluginEngineGroup.GET("/plugins/:id/dependencies", controller.ControllerGroupApp.PluginEngineController.GetDependencies)
		pluginEngineGroup.POST("/plugins/:id/dependencies", controller.ControllerGroupApp.PluginEngineController.SaveDependencies)
		pluginEngineGroup.GET("/plugins/:id/validate-deps", controller.ControllerGroupApp.PluginEngineController.ValidateDependencies)
	}

	pluginEngineGroup.POST("/install-callback", controller.ControllerGroupApp.PluginEngineController.InstallCallback)
	pluginEngineGroup.POST("/demo/install", controller.ControllerGroupApp.PluginEngineController.InstallDemoPlugin)

	marketGroup := router.Group("/api/plugin-engine/market")
	marketGroup.Use(middleware.JwtAuthMiddleware())
	marketGroup.Use(middleware.TenantMiddleware())
	{
		marketGroup.GET("/search", controller.ControllerGroupApp.PluginEngineController.SearchMarketPlugins)
		marketGroup.GET("/plugins/:id", controller.ControllerGroupApp.PluginEngineController.GetMarketPluginDetail)
		marketGroup.GET("/categories", controller.ControllerGroupApp.PluginEngineController.GetMarketCategories)
		marketGroup.POST("/install", controller.ControllerGroupApp.PluginEngineController.InstallFromMarket)
		marketGroup.GET("/public-key", controller.ControllerGroupApp.PluginEngineController.GetMarketPublicKey)
	}

	router.GET("/plugin-assets/:pluginId/*filepath", controller.ControllerGroupApp.PluginEngineController.ServePluginAsset)

	pluginDataGroup := router.Group("/api/plugin-data")
	pluginDataGroup.Use(middleware.JwtAuthMiddleware())
	pluginDataGroup.Use(middleware.TenantMiddleware())
	{
		pluginDataGroup.GET("/:table", controller.ControllerGroupApp.PluginEngineController.GetPluginData)
		pluginDataGroup.POST("/:table", controller.ControllerGroupApp.PluginEngineController.CreatePluginData)
		pluginDataGroup.PUT("/:table/:id", controller.ControllerGroupApp.PluginEngineController.UpdatePluginData)
		pluginDataGroup.DELETE("/:table/:id", controller.ControllerGroupApp.PluginEngineController.DeletePluginData)
	}
}
