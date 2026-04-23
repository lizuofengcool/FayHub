package controller

import (
	"fayhub/internal/service"
	"github.com/gin-gonic/gin"
)

// ControllerGroup 控制器组管理（GVA 标准工程实践）
// 作用：统一管理所有API控制器实例，避免零散初始化
type ControllerGroup struct {
	// 系统核心控制器（阶段一先搭骨架）
	SystemController
	// 预留扩展：后续可添加 TenantController、UserController 等
	// TenantController
	// UserController
}

// 实例化全局控制器组（对外暴露，供路由调用）
var ControllerGroupApp = new(ControllerGroup)

// ==================== 系统控制器子组（阶段一核心）====================
// SystemController 系统基础控制器
type SystemController struct{}

// HealthCheck 健康检查接口
// @Summary 系统健康检查
// @Description 验证系统运行状态并返回当前租户ID
// @Tags 系统接口
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "系统运行正常"
// @Router /api/health [get]
func (s *SystemController) HealthCheck(c *gin.Context) {
	// 从Gin Context转换为标准Context
	ctx := c.Request.Context()
	
	// 调用Service层健康检查方法
	tentantID, message, err := service.ServiceGroupApp.SystemService.HealthCheck(ctx)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "系统内部错误",
			"data":    nil,
		})
		return
	}
	
	// 返回统一格式的JSON响应
	c.JSON(200, gin.H{
		"code":    200,
		"message": message,
		"data": gin.H{
			"tenant_id": tentantID,
			"status":    "running",
		},
	})
}