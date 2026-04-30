package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

type DepartmentRouter struct{}

func (r *DepartmentRouter) Init(router *gin.Engine) {
	deptGroup := router.Group("/api/departments")
	deptGroup.Use(middleware.JwtAuthMiddleware())
	deptGroup.Use(middleware.TenantMiddleware())
	{
		deptGroup.GET("/tree", controller.ControllerGroupApp.DepartmentController.GetTree)
		deptGroup.POST("", controller.ControllerGroupApp.DepartmentController.Create)
		deptGroup.PUT("/:id", controller.ControllerGroupApp.DepartmentController.Update)
		deptGroup.DELETE("/:id", controller.ControllerGroupApp.DepartmentController.Delete)
		deptGroup.POST("/:id/users/:userId", controller.ControllerGroupApp.DepartmentController.AssignUser)
		deptGroup.DELETE("/:id/users/:userId", controller.ControllerGroupApp.DepartmentController.RemoveUser)
	}
}
