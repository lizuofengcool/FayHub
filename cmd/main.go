package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"fayhub/internal/controller"
	"fayhub/internal/initialize"
	"fayhub/internal/middleware"
	"fayhub/internal/model"
	"fayhub/internal/router"
	"fayhub/pkg/utils"
)

// main 主程序入口
func main() {
	// 初始化数据库
	dbConfig := &initialize.DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "fayhub",
		Charset:  "utf8mb4",
	}
	db, err := initialize.InitDB(dbConfig)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移数据库表
	err = initialize.AutoMigrate(db, &model.Tenant{}, &model.User{}, &model.TenantUser{})
	if err != nil {
		log.Fatalf("自动迁移数据库表失败: %v", err)
	}

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