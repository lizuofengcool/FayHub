package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "fayhub/docs/swagger"
	"fayhub/internal/initialize"
	intMiddleware "fayhub/internal/middleware"
	"fayhub/internal/model"
	"fayhub/internal/router"
	"fayhub/internal/service"
	"fayhub/pkg/cache"
	"fayhub/pkg/config"
	pkgCrypto "fayhub/pkg/crypto"
	"fayhub/pkg/eventbus"
	"fayhub/pkg/messagequeue"
	pkgMiddleware "fayhub/pkg/middleware"
	"fayhub/pkg/plugin"
	"fayhub/pkg/pluginsign"
	"fayhub/pkg/redisclient"
	"fayhub/pkg/storage"
	"fayhub/pkg/utils"
)

// @title FayHub SaaS平台API文档
// @version 1.0
// @description FayHub多租户SaaS平台后端API接口文档
// @termsOfService http://swagger.io/terms/

// @contact.name FayHub Team
// @contact.url http://www.fayhub.com
// @contact.email dev@fayhub.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	log.Println("🚀 开始启动 FayHub 服务...")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}
	log.Println("✅ 配置文件加载成功")

	if cfg.JWT.Secret == "" {
		log.Fatal("❌ 安全错误：JWT密钥未配置！请设置环境变量 FAYHUB_JWT_SECRET")
	}

	if os.Getenv("FAYHUB_JWT_SECRET") != "" {
		log.Println("✅ JWT密钥从环境变量加载")
	} else {
		log.Println("⚠️  JWT密钥来自配置文件，生产环境建议使用环境变量 FAYHUB_JWT_SECRET")
	}

	log.Println("🔑 初始化JWT配置...")
	if cfg.JWT.Algorithm == "RS256" {
		if rs256Err := utils.InitJWTConfigRS256(cfg.JWT.Secret, cfg.JWT.Expire, cfg.JWT.Issuer, cfg.JWT.Algorithm, cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath); rs256Err != nil {
			log.Fatalf("❌ RS256初始化失败: %v", rs256Err)
		}
		log.Println("✅ JWT RS256非对称加密配置初始化完成")
	} else {
		utils.InitJWTConfig(cfg.JWT.Secret, cfg.JWT.Expire, cfg.JWT.Issuer)
		log.Println("✅ JWT HS256对称加密配置初始化完成（建议升级RS256）")
	}

	if cfg.Database.Host != "" && cfg.Database.Port != 0 {
		if cfg.Database.Password == "" && os.Getenv("FAYHUB_DB_PASSWORD") == "" {
			log.Println("⚠️  数据库密码未配置，请设置环境变量 FAYHUB_DB_PASSWORD")
		} else if os.Getenv("FAYHUB_DB_PASSWORD") != "" {
			log.Println("✅ 数据库密码从环境变量加载")
		} else {
			log.Println("⚠️  数据库密码来自配置文件，生产环境建议使用环境变量 FAYHUB_DB_PASSWORD")
		}

		log.Println("🗄️  尝试连接数据库...")
		log.Printf("📝 数据库配置: 类型=%s, 地址=%s:%d, 数据库=%s",
			cfg.Database.Type, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

		db, dbErr := initialize.InitDB(&cfg.Database)
		if dbErr != nil {
			log.Printf("⚠️  数据库连接失败: %v，使用内存模式", dbErr)
		} else {
			utils.SetGlobalDB(db)
			log.Println("✅ 数据库连接成功")

			if callbackErr := model.RegisterTenantIsolationCallbacks(db); callbackErr != nil {
				log.Printf("⚠️  注册租户隔离回调失败: %v", callbackErr)
			} else {
				log.Println("✅ 租户隔离回调注册成功")
			}

			if encKey := os.Getenv("FAYHUB_ENCRYPTION_KEY"); encKey != "" {
				pkgCrypto.InitEncryptionKey(encKey)
				pkgCrypto.RegisterEncryptionCallbacks(db)
				log.Println("✅ 字段加密回调注册成功")
			} else {
				log.Println("⚠️  未配置FAYHUB_ENCRYPTION_KEY，敏感字段将明文存储")
			}

			if indexErr := model.CreateCompositeIndexes(db); indexErr != nil {
				log.Printf("⚠️  创建复合索引失败: %v", indexErr)
			} else {
				log.Println("✅ 复合索引创建成功")
			}

			if migrateErr := initialize.MigrateDatabase(db); migrateErr != nil {
				log.Printf("⚠️  数据库迁移失败: %v", migrateErr)
			}

			if initDataErr := initialize.InitTestData(db); initDataErr != nil {
				log.Printf("⚠️  初始化测试数据失败: %v", initDataErr)
			}

			if adminErr := initialize.InitDefaultAdmin(db); adminErr != nil {
				log.Printf("⚠️  初始化默认管理员失败: %v", adminErr)
			}

			if rolesErr := initialize.InitDefaultRoles(db); rolesErr != nil {
				log.Printf("⚠️  初始化默认角色失败: %v", rolesErr)
			}

			if menusErr := initialize.InitDefaultMenus(db); menusErr != nil {
				log.Printf("⚠️  初始化默认菜单失败: %v", menusErr)
			}

			initialize.FixMissingMenus(db)

			if apisErr := initialize.InitDefaultAPIs(db); apisErr != nil {
				log.Printf("⚠️  初始化默认API权限失败: %v", apisErr)
			}
		}
	} else {
		log.Println("⚠️  数据库配置不完整，使用内存模式")
		utils.SetGlobalDB(nil)
	}

	log.Println("🔌 初始化Redis...")
	if redisErr := redisclient.Init(cfg, log.Printf); redisErr != nil {
		log.Printf("⚠️  Redis初始化失败: %v，使用降级存储模式", redisErr)
	} else {
		cache.InitCache(redisclient.GetRawClient())
		log.Println("✅ 缓存管理器初始化成功")
	}

	log.Println("💾 初始化存储驱动...")
	if storageErr := storage.Init(cfg); storageErr != nil {
		log.Printf("⚠️  存储驱动初始化失败: %v", storageErr)
	} else {
		log.Printf("✅ 存储驱动初始化成功（%s）", cfg.Storage.Driver)
	}

	log.Println("📊 创建数据库复合索引...")
	if indexErr := initialize.MigrateCompositeIndexes(); indexErr != nil {
		log.Printf("⚠️  复合索引创建失败: %v", indexErr)
	} else {
		log.Println("✅ 数据库复合索引创建成功")
	}

	log.Println("🔌 初始化WASM插件引擎...")
	if engineErr := initialize.InitPluginEngine(); engineErr != nil {
		log.Printf("⚠️  WASM引擎初始化失败: %v，使用NoopEngine", engineErr)
	} else {
		log.Println("✅ WASM插件引擎初始化成功")

		if signKeyPath := os.Getenv("FAYHUB_PLUGIN_PUBLIC_KEY"); signKeyPath != "" {
			if signErr := pluginsign.InitPublicKey(signKeyPath); signErr != nil {
				log.Printf("⚠️  插件签名公钥加载失败: %v", signErr)
			} else {
				log.Println("✅ 插件签名校验已启用")
			}
		} else {
			log.Println("⚠️  未配置插件签名公钥，跳过签名校验")
		}

		plugin.RegisterDBFuncs(
			func(ctx context.Context, tenantKey string, query string, args ...interface{}) ([]map[string]interface{}, error) {
				db := utils.GetDB(ctx)
				if db == nil {
					return nil, fmt.Errorf("数据库未连接")
				}
				if err := plugin.ValidatePluginSQL(query); err != nil {
					return nil, fmt.Errorf("SQL校验失败: %v", err)
				}
				var results []map[string]interface{}
				if err := db.Raw(query, args...).Scan(&results).Error; err != nil {
					return nil, err
				}
				return results, nil
			},
			func(ctx context.Context, tenantKey string, query string, args ...interface{}) (int64, error) {
				db := utils.GetDB(ctx)
				if db == nil {
					return 0, fmt.Errorf("数据库未连接")
				}
				if err := plugin.ValidatePluginSQL(query); err != nil {
					return 0, fmt.Errorf("SQL校验失败: %v", err)
				}
				result := db.Exec(query, args...)
				return result.RowsAffected, result.Error
			},
		)
		log.Println("✅ 插件Host DB函数注册成功")

		if loadErr := initialize.LoadInstalledPlugins(); loadErr != nil {
			log.Printf("⚠️  加载已安装插件失败: %v", loadErr)
		}
	}

	log.Println("🌐 初始化Gin引擎...")
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(pkgMiddleware.CORSMiddleware())

	// 注册自定义中间件
	r.Use(pkgMiddleware.GlobalExceptionMiddleware())   // 全局异常处理
	r.Use(pkgMiddleware.LoggingMiddleware())           // 日志上下文注入
	r.Use(intMiddleware.AuditMiddleware())             // 审计日志
	r.Use(intMiddleware.MetricsMiddleware())           // 性能监控
	r.Use(intMiddleware.InputSanitizationMiddleware()) // 输入验证与过滤
	// 注意：租户中间件不在全局注册，由各路由组按需挂载
	// 登录/注册等无需租户隔离的接口不应经过 TenantMiddleware

	log.Println("🛣️  初始化路由...")
	router.RouterGroupApp.InitAllRouters(r)

	log.Println("🧹 启动SSO过期数据定时清理...")
	service.ServiceGroupApp.SSOService.StartCleanup()

	log.Println("🧹 启动过期订单定时关闭...")
	service.ServiceGroupApp.PaymentService.StartCloseExpiredOrders()

	log.Println("📡 初始化事件总线...")
	eventbus.InitAll()
	log.Println("✅ 事件总线已启动")

	log.Println("📨 初始化消息队列...")
	messagequeue.InitAndStart([]string{
		"payment.paid",
		"order.expired",
		"user.created",
		"file.uploaded",
	})
	messagequeue.RegisterBusinessHandlers()
	log.Println("✅ 消息队列已启动")

	if wasmEng, ok := plugin.GetEngine().(*plugin.WASMEngine); ok {
		wasmEng.SetGinEngine(r)
		wasmEng.SetupPluginProxyRoutes()
		log.Println("✅ 插件代理路由已注册")
	}

	log.Println("✅ 路由初始化完成")

	// 全局 404 / 405 处理（API 路径返回 JSON，而非纯文本）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":      40400,
				"msg":       "接口不存在",
				"data":      nil,
				"path":      path,
				"method":    c.Request.Method,
				"timestamp": time.Now().Unix(),
			})
		} else {
			c.String(http.StatusNotFound, "404 page not found")
		}
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":      40500,
			"msg":       "请求方法不允许",
			"data":      nil,
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"timestamp": time.Now().Unix(),
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("✅ Swagger文档路由注册完成")

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🚀 FayHub 服务启动成功！")
	log.Printf("📍 服务地址: http://0.0.0.0%s"+" (also available via api.fayhub.com)", port)
	log.Printf("🔍 健康检查: http://localhost%s/api/health", port)
	log.Printf("🗄️  数据库版本: PostgreSQL 17 (首选) / MySQL 8.0")

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 启动服务失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("🛑 收到信号 %v，开始优雅关机...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ 服务强制关机: %v", err)
	}

	if closeErr := redisclient.Close(); closeErr != nil {
		log.Printf("⚠️ Redis关闭失败: %v", closeErr)
	}

	eventbus.StopAll()

	log.Println("✅ FayHub 服务已安全退出")
}
