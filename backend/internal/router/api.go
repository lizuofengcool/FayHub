package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type APIRouter struct{}

func (s *APIRouter) Init(router *gin.Engine) {
	apiGroup := router.Group("/api/apis")
	apiGroup.Use(middleware.JwtAuthMiddleware())
	apiGroup.Use(middleware.TenantMiddleware())

	apiGroup.POST("", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.APIController.CreateAPI)
	apiGroup.GET("", controller.ControllerGroupApp.APIController.GetAPIList)
	apiGroup.GET("/:apiID", controller.ControllerGroupApp.APIController.GetAPIByID)
	apiGroup.PUT("/:apiID", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.APIController.UpdateAPI)
	apiGroup.DELETE("/:apiID", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.APIController.DeleteAPI)

	apiGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controller.ControllerGroupApp.APIController.AssignRoleAPIs)
	apiGroup.GET("/role/:roleID", controller.ControllerGroupApp.APIController.GetRoleAPIs)
}
