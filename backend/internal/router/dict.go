package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type DictRouter struct{}

func (r *DictRouter) Init(router *gin.Engine) {
	dictGroup := router.Group("/api/dict")
	dictGroup.Use(middleware.JwtAuthMiddleware())
	dictGroup.Use(middleware.TenantMiddleware())
	{
		dictGroup.GET("/types", controller.ControllerGroupApp.DictController.ListDictTypes)
		dictGroup.GET("/types/:id", controller.ControllerGroupApp.DictController.GetDictType)
		dictGroup.POST("/types", controller.ControllerGroupApp.DictController.CreateDictType)
		dictGroup.PUT("/types/:id", controller.ControllerGroupApp.DictController.UpdateDictType)
		dictGroup.DELETE("/types/:id", controller.ControllerGroupApp.DictController.DeleteDictType)

		dictGroup.GET("/data", controller.ControllerGroupApp.DictController.ListDictData)
		dictGroup.GET("/data/:dict_type", controller.ControllerGroupApp.DictController.GetDictDataByType)
		dictGroup.POST("/data", controller.ControllerGroupApp.DictController.CreateDictData)
		dictGroup.PUT("/data/:id", controller.ControllerGroupApp.DictController.UpdateDictData)
		dictGroup.DELETE("/data/:id", controller.ControllerGroupApp.DictController.DeleteDictData)
	}
}
