package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"fayhub/internal/config"
	"fayhub/internal/initialize"
	"fayhub/internal/middleware"
	"fayhub/internal/router"
	"fayhub/pkg/middleware"
	"fayhub/pkg/utils"
)

func main() {
	log.Println("🚀 开始启动 FayHub 服务...")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("⚠️  加载配置文件失败，使用默认配置: %v", err)
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				Type:     "postgresql",
				Host:     "localhost",
				Port:     5432,
				Username: "fayhub",
				Password: "fayhub123",
				Database: "fayhub",
				Charset:  "utf8mb4",
			},
			Server: config.ServerConfig{
				Port: 8080,
				Mode: "debug",
			},
			JWT: config.JWTConfig{
				Secret: "fayhub-secret-key-2024",
				Expire: 168,
				Issuer: "fayhub",
			},
		}
		log.Println("✅ 使用默认配置启动（内存模式）")
	} else {
		log.Println("✅ 配置文件加载成功")
	}

	log.Println("🔑 初始化JWT配置...")
	utils.InitJWTConfig(cfg.JWT.Secret, cfg.JWT.Expire, cfg.JWT.Issuer)
	log.Println("✅ JWT配置初始化完成")

	// 初始化全局数据库连接
	if cfg.Database.Host != "" && cfg.Database.Port != 0 {
		log.Println("🗄️  尝试连接数据库...")
		log.Printf("📝 数据库配置: 类型=%s, 地址=%s:%d, 数据库=%s",
			cfg.Database.Type, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

		db, dbErr := initialize.InitDB(&cfg.Database)
		if dbErr != nil {
			log.Printf("⚠️  数据库连接失败: %v，使用内存模式", dbErr)
		} else {
			utils.SetGlobalDB(db)
			log.Println("✅ 数据库连接成功")

			if migrateErr := initialize.MigrateDatabase(db); migrateErr != nil {
				log.Printf("⚠️  数据库迁移失败: %v", migrateErr)
			}

			if initDataErr := initialize.InitTestData(db); initDataErr != nil {
				log.Printf("⚠️  初始化测试数据失败: %v", initDataErr)
			}
		}
	} else {
		log.Println("⚠️  数据库配置不完整，使用内存模式")
		utils.SetGlobalDB(nil)
	}

	log.Println("🌐 初始化Gin引擎...")
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()

	// 注册全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 注册自定义中间件
	r.Use(pkgMiddleware.GlobalExceptionMiddleware()) // 全局异常处理
	r.Use(pkgMiddleware.LoggingMiddleware())         // 日志上下文注入
	r.Use(middleware.JwtAuthMiddleware())            // JWT认证

	log.Println("🛣️  初始化路由...")
	router.RouterGroupApp.InitAllRouters(r)
	log.Println("✅ 路由初始化完成")

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🚀 FayHub 服务启动成功！")
	log.Printf("📍 服务地址: http://localhost%s", port)
	log.Printf("🔍 健康检查: http://localhost%s/api/health", port)
	log.Printf("🗄️  数据库版本: PostgreSQL 17 (首选) / MySQL 8.0")

	if err := r.Run(port); err != nil {
		log.Fatalf("❌ 启动服务失败: %v", err)
	}
}
