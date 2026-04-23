package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fayhub/internal/router"
	"fayhub/pkg/utils"
)

// main 主程序入口
func main() {
	// 阶段三：使用MySQL数据库，支持完整CRUD功能
	// 注意：当前使用内存模式，生产环境应配置真实MySQL连接
	log.Println("⚠️  阶段三使用内存模式，生产环境请配置MySQL数据库")
	var db *gorm.DB = nil

	// 初始化Gin引擎（生产环境建议设置为ReleaseMode）
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// 设置基础中间件
	r.Use(gin.Logger())   // 日志中间件
	r.Use(gin.Recovery()) // 恢复中间件

	// 设置全局数据库实例，供utils.GetDB使用
	utils.SetGlobalDB(db)

	// 初始化所有路由组
	router.RouterGroupApp.InitAllRouters(r)

	// 启动HTTP服务器
	port := ":8080"
	fmt.Printf("🚀 FayHub 服务启动成功！\n")
	fmt.Printf("📍 服务地址: http://localhost%s\n", port)
	fmt.Printf("🔍 健康检查: http://localhost%s/api/health\n", port)
	fmt.Printf("📋 测试命令: curl -H \"X-Tenant-ID: 1001\" http://localhost%s/api/health\n", port)

	log.Printf("FayHub 服务正在监听端口 %s", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
