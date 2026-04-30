// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type EngineRouter struct{}

func (r *EngineRouter) Init(router *gin.Engine) {
	engineGroup := router.Group("/api/engine")
	engineGroup.Use(middleware.JwtAuthMiddleware())
	engineGroup.Use(middleware.TenantMiddleware())
	{
		engineGroup.GET("/status", controller.ControllerGroupApp.EngineController.GetEngineStatus)
		engineGroup.GET("/plugins", controller.ControllerGroupApp.EngineController.GetLoadedPlugins)
		engineGroup.GET("/plugins/:id/routes", controller.ControllerGroupApp.EngineController.GetPluginRoutes)
		engineGroup.GET("/plugins/:id/health", controller.ControllerGroupApp.EngineController.HealthCheckPlugin)
	}
}
