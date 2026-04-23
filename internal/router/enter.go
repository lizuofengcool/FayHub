package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RouterGroup 路由组管理（GVA 标准工程实践）
// 作用：统一管理所有路由分组，避免零散注册
type RouterGroup struct {
	// 系统核心路由组（阶段一先搭骨架）
	SystemRouter
	// 预留扩展：后续可添加 TenantRouter、UserRouter 等
	// TenantRouter
	// UserRouter
}

// 实例化全局路由组（对外暴露，供主程序调用）
var RouterGroupApp = new(RouterGroup)

// ==================== 系统路由子组（阶段一核心）====================
// SystemRouter 系统基础路由
type SystemRouter struct{}

// Init 初始化系统路由
// @Summary 注册系统相关路由
// @Description 注册健康检查等系统接口，强制挂载租户中间件
func (s *SystemRouter) Init(router *gin.Engine) {
	// 创建系统API分组
	systemGroup := router.Group("/api")
	
	// 强制挂载租户中间件（多租户隔离核心）
	systemGroup.Use(middleware.TenantMiddleware())
	
	// 注册健康检查接口
	systemGroup.GET("/health", controller.ControllerGroupApp.SystemController.HealthCheck)
	
	// 预留扩展：后续可添加其他系统接口
	// systemGroup.GET("/config", ...)
	// systemGroup.POST("/login", ...)
}

// InitAllRouters 初始化所有路由组
// @Summary 统一初始化所有路由
// @Description 集中管理所有路由组的初始化，便于主程序调用
func (r *RouterGroup) InitAllRouters(router *gin.Engine) {
	// 初始化系统路由
	r.SystemRouter.Init(router)
	
	// 预留扩展：后续初始化其他路由组
	// r.TenantRouter.Init(router)
	// r.UserRouter.Init(router)
}