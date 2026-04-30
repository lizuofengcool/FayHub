package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SettlementRouter struct{}

func (r *SettlementRouter) Init(router *gin.Engine) {
	settlementGroup := router.Group("/api/settlement")
	settlementGroup.Use(middleware.JwtAuthMiddleware())
	settlementGroup.Use(middleware.TenantMiddleware())
	{
		settlementGroup.POST("", controller.ControllerGroupApp.SettlementController.CreateSettlement)
		settlementGroup.GET("/config", controller.ControllerGroupApp.SettlementController.GetSettlementConfig)
		settlementGroup.PUT("/config", controller.ControllerGroupApp.SettlementController.UpdateSettlementConfig)
		settlementGroup.POST("/process/:settlement_no", controller.ControllerGroupApp.SettlementController.ProcessSettlement)
		settlementGroup.GET("/stats", controller.ControllerGroupApp.SettlementController.GetSettlementStats)
	}
}
