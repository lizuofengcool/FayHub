package router

import (
	"fayhub/internal/controller"
	"fayhub/internal/middleware"
	"fayhub/pkg/metrics"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// RouterGroup 路由组管理（GVA 标准工程实践）
// 作用：统一管理所有路由分组，避免零散注册
type RouterGroup struct {
	SystemRouter
	SystemSettingRouter
	AuthRouter
	TenantRouter
	UserRouter
	RBACRouter
	MenuRouter
	APIRouter
	PluginEngineRouter
	EngineRouter
	SSORouter
	PaymentRouter
	LogRouter
	WebhookRouter
	AuditRouter
	NotificationRouter
	FileRouter
	DepartmentRouter
	APIKeyRouter
	SettlementRouter
	BackupRouter
	LoginLogRouter
	DictRouter
	ErrorCodeRouter
	TenantPackageRouter
	OnlineUserRouter
	CronJobRouter
	SubscriptionRouter
	NotificationChannelRouter
	MonitorRouter
	PluginResourceMonitorRouter
	SensitiveWordRouter
	ExcelRouter
	TenantChannelRouter
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
	distPath := "web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		distPath = "../web/dist"
	}
	if _, err := os.Stat(distPath); err == nil {
		router.GET("/", func(c *gin.Context) {
			c.File(filepath.Join(distPath, "index.html"))
		})
	} else {
		router.GET("/", controller.ControllerGroupApp.SystemController.HomePage)
	}

	// 创建系统API分组
	systemGroup := router.Group("/api")

	// 健康检查不需要租户隔离
	systemGroup.GET("/health", controller.ControllerGroupApp.SystemController.HealthCheck)

	protectedGroup := router.Group("/api")
	protectedGroup.Use(middleware.JwtAuthMiddleware())
	protectedGroup.GET("/metrics", gin.WrapF(metricsHandler()))
	protectedGroup.GET("/stats", controller.ControllerGroupApp.StatsController.GetDashboardStats)

	// 预留扩展：后续可添加其他系统接口
	// systemGroup.GET("/config", ...)
	// systemGroup.POST("/login", ...)
}

// InitAllRouters 初始化所有路由组
// @Summary 统一初始化所有路由
// @Description 集中管理所有路由组的初始化，便于主程序调用
func (r *RouterGroup) InitAllRouters(router *gin.Engine) {
	r.SystemRouter.Init(router)
	r.SystemSettingRouter.Init(router)
	r.AuthRouter.Init(router)
	r.TenantRouter.Init(router)
	r.UserRouter.Init(router)
	r.RBACRouter.Init(router)
	r.MenuRouter.Init(router)
	r.APIRouter.Init(router)
	r.PluginEngineRouter.Init(router)
	r.EngineRouter.Init(router)
	r.SSORouter.Init(router)
	r.PaymentRouter.Init(router)
	r.LogRouter.Init(router)
	r.WebhookRouter.Init(router)
	r.AuditRouter.Init(router)
	r.NotificationRouter.Init(router)
	r.FileRouter.Init(router)
	r.DepartmentRouter.Init(router)
	r.APIKeyRouter.Init(router)
	r.SettlementRouter.Init(router)
	r.BackupRouter.Init(router)
	r.LoginLogRouter.Init(router)
	r.DictRouter.Init(router)
	r.ErrorCodeRouter.Init(router)
	r.TenantPackageRouter.Init(router)
	r.OnlineUserRouter.Init(router)
	r.CronJobRouter.Init(router)
	r.SubscriptionRouter.Init(router)
	r.NotificationChannelRouter.Init(router)
	r.MonitorRouter.Init(router)
	r.PluginResourceMonitorRouter.Init(router)
	r.SensitiveWordRouter.Init(router)
	r.ExcelRouter.Init(router)
	r.TenantChannelRouter.Init(router)
}

func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(metrics.GetPrometheusFormat()))
	}
}
