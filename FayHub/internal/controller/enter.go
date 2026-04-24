package controller

import (
	"fayhub/internal/service"

	"github.com/gin-gonic/gin"
)

// ControllerGroup 控制器组管理（GVA 标准工程实践）
// 作用：统一管理所有API控制器实例，避免零散初始化
type ControllerGroup struct {
	SystemController
	AuthController
	TenantController
	UserController
}

// 实例化全局控制器组（对外暴露，供路由调用）
var ControllerGroupApp = new(ControllerGroup)

// ==================== 系统控制器子组（阶段一核心）====================
// SystemController 系统基础控制器
type SystemController struct{}

// HomePage 欢迎页面
// @Summary 系统欢迎页面
// @Description FayHub 系统首页，显示API信息
// @Tags 系统接口
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML页面"
// @Router / [get]
func (s *SystemController) HomePage(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FayHub - AI驱动的多租户SaaS平台</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 60px 40px;
            max-width: 800px;
            width: 100%;
            text-align: center;
        }
        h1 {
            color: #667eea;
            font-size: 3em;
            margin-bottom: 20px;
            font-weight: 800;
        }
        .subtitle {
            color: #666;
            font-size: 1.2em;
            margin-bottom: 40px;
        }
        .api-list {
            margin: 40px 0;
            text-align: left;
        }
        .api-item {
            background: #f8f9fa;
            padding: 20px;
            margin: 15px 0;
            border-radius: 10px;
            border-left: 4px solid #667eea;
            transition: transform 0.3s;
        }
        .api-item:hover {
            transform: translateX(10px);
        }
        .api-method {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 0.8em;
            margin-right: 10px;
        }
        .method-get {
            background: #61affe;
            color: white;
        }
        .method-post {
            background: #49cc90;
            color: white;
        }
        .api-path {
            font-family: monospace;
            color: #333;
            font-size: 1.1em;
        }
        .api-desc {
            color: #888;
            margin-top: 8px;
            font-size: 0.95em;
        }
        .tech-info {
            margin-top: 40px;
            padding-top: 30px;
            border-top: 2px solid #eee;
        }
        .tech-tag {
            display: inline-block;
            background: #e3f2fd;
            color: #1976d2;
            padding: 8px 16px;
            border-radius: 20px;
            margin: 5px;
            font-weight: 600;
        }
        .status {
            margin-top: 30px;
            padding: 15px;
            background: #e8f5e9;
            border-radius: 10px;
            color: #2e7d32;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 FayHub</h1>
        <p class="subtitle">AI驱动的多租户SaaS生态平台</p>
        
        <div class="status">
            ✅ 系统运行正常 - PostgreSQL 17 就绪
        </div>
        
        <div class="api-list">
            <h3 style="color: #333; margin-bottom: 20px;">📋 API接口列表</h3>
            
            <div class="api-item">
                <span class="api-method method-get">GET</span>
                <span class="api-path">/api/health</span>
                <p class="api-desc">系统健康检查 - 支持多租户隔离</p>
            </div>
            
            <div class="api-item">
                <span class="api-method method-post">POST</span>
                <span class="api-path">/api/auth/login</span>
                <p class="api-desc">用户登录 - 返回JWT Token</p>
            </div>
            
            <div class="api-item">
                <span class="api-method method-post">POST</span>
                <span class="api-path">/api/auth/logout</span>
                <p class="api-desc">用户登出</p>
            </div>
            
            <div class="api-item">
                <span class="api-method method-post">POST</span>
                <span class="api-path">/api/auth/refresh</span>
                <p class="api-desc">刷新JWT Token</p>
            </div>
            
            <div class="api-item">
                <span class="api-method method-get">GET</span>
                <span class="api-path">/api/auth/me</span>
                <p class="api-desc">获取当前用户信息（需要认证）</p>
            </div>
        </div>
        
        <div class="tech-info">
            <h3 style="color: #333; margin-bottom: 15px;">🛠️ 技术栈</h3>
            <span class="tech-tag">Go 1.26+</span>
            <span class="tech-tag">Gin Web</span>
            <span class="tech-tag">GORM v2</span>
            <span class="tech-tag">PostgreSQL 17</span>
            <span class="tech-tag">MySQL 8.0</span>
            <span class="tech-tag">JWT认证</span>
            <span class="tech-tag">多租户隔离</span>
        </div>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, html)
}

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
