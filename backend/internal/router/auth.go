package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthRouter 认证路由
// @Summary 认证路由
// @Description 注册认证相关的路由
// @Tags 路由管理
type AuthRouter struct{}

// Init 初始化认证路由
// @Summary 初始化认证路由
// @Description 注册登录、登出、刷新Token等认证接口
// @Tags 路由管理
func (s *AuthRouter) Init(router *gin.Engine) {
	// 创建认证API分组
	authGroup := router.Group("/api/auth")
	
	// 登录接口不需要JWT认证
	authGroup.POST("/login", controller.ControllerGroupApp.AuthController.Login)
	
	// 以下接口需要JWT认证
	authGroup.Use(middleware.JwtAuthMiddleware())
	{
		authGroup.POST("/logout", controller.ControllerGroupApp.AuthController.Logout)
		authGroup.POST("/refresh", controller.ControllerGroupApp.AuthController.RefreshToken)
		authGroup.GET("/me", controller.ControllerGroupApp.AuthController.GetCurrentUser)
	}
}