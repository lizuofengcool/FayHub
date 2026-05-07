package initialize

import (
	"context"
	"encoding/json"
	"fayhub/internal/service"
	"fayhub/pkg/market"
	"fayhub/pkg/plugin"
	"fayhub/pkg/utils"
	"log"

	"gorm.io/gorm"
)

func InitPluginEngine() error {
	ctx := context.Background()

	engine := plugin.NewWASMEngine()
	if err := engine.Start(ctx); err != nil {
		log.Printf("WASM引擎启动失败: %v，回退到NoopEngine", err)
		return err
	}

	dataSource := &service.DBPluginDataSource{}
	engine.SetDataSource(dataSource)

	plugin.SetEngine(engine)
	log.Println("WASM插件引擎初始化成功(含数据源注入)")

	service.ServiceGroupApp.PluginResourceMonitorService.Init()
	log.Println("插件资源监控服务初始化成功")

	market.InitClient()
	log.Println("Market API客户端初始化成功")

	return nil
}

func LoadInstalledPlugins() error {
	ctx := context.Background()
	platformCtx := utils.SkipTenantIsolation(ctx)
	db := utils.GetDB(platformCtx)
	if db == nil {
		log.Println("数据库未连接，跳过已安装插件加载")
		return nil
	}

	engine := plugin.GetEngine()

	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		log.Println("当前引擎非WASMEngine，跳过已安装插件加载")
		return nil
	}

	type installedInfo struct {
		TenantID   int64
		PluginID   string
		Version    string
		Status     string
		ConfigJSON string
	}

	var installed []installedInfo

	if err := db.Table("installed_plugins").
		Select("tenant_id, plugin_id, version, status, config_json").
		Where("status = ?", "active").
		Find(&installed).Error; err != nil {
		log.Printf("查询已安装插件失败: %v", err)
		return err
	}

	loadedCount := 0
	for _, inst := range installed {
		if err := wasmEngine.Install(ctx, inst.TenantID, inst.PluginID, inst.Version, ""); err != nil {
			log.Printf("加载已安装插件失败: tenant=%d, plugin=%s, err=%v",
				inst.TenantID, inst.PluginID, err)
			continue
		}

		// 解析并注册 manifest
		if inst.ConfigJSON != "" {
			var manifest plugin.Manifest
			if err := json.Unmarshal([]byte(inst.ConfigJSON), &manifest); err == nil {
				registry := wasmEngine.GetRegistry()
				if len(manifest.Routes) > 0 {
					registry.RegisterRoutes(inst.TenantID, inst.PluginID, manifest.Routes)
				}
				if len(manifest.APIs) > 0 {
					registry.RegisterAPIs(inst.TenantID, inst.PluginID, manifest.APIs)
				}
				if len(manifest.Menus) > 0 {
					registry.RegisterMenus(inst.TenantID, inst.PluginID, manifest.Menus)
				}
			}
		}

		loadedCount++
	}

	// 同步所有已加载插件的菜单到数据库
	for _, inst := range installed {
		service.ServiceGroupApp.PluginEngineService.SyncPluginMenus(ctx, inst.TenantID, inst.PluginID)
	}

	log.Printf("已加载 %d 个已安装插件", loadedCount)
	return nil
}

func ShutdownPluginEngine() {
	ctx := context.Background()
	engine := plugin.GetEngine()
	if err := engine.Stop(ctx); err != nil {
		log.Printf("WASM引擎关闭失败: %v", err)
	} else {
		log.Println("WASM插件引擎已安全关闭")
	}
}

func RestoreActivePlugins(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	ctx := context.Background()
	engine := plugin.GetEngine()

	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		return nil
	}

	type installedInfo struct {
		TenantID int64
		PluginID string
		Version  string
	}

	var installed []installedInfo
	platformCtx := utils.SkipTenantIsolation(ctx)
	platformDB := utils.GetDB(platformCtx)
	if platformDB == nil {
		return nil
	}

	if err := platformDB.Table("installed_plugins").
		Select("tenant_id, plugin_id, version").
		Where("status = ?", "active").
		Find(&installed).Error; err != nil {
		return err
	}

	for _, inst := range installed {
		if err := wasmEngine.Install(ctx, inst.TenantID, inst.PluginID, inst.Version, ""); err != nil {
			log.Printf("恢复已安装插件失败: tenant=%d, plugin=%s, err=%v",
				inst.TenantID, inst.PluginID, err)
		}
	}

	return nil
}
