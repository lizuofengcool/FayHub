package service

import (
	"context"
	"encoding/json"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/market"
	"fayhub/pkg/plugin"
	"fayhub/pkg/pluginsign"
	"fayhub/pkg/utils"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PluginEngineService struct{}

type DBPluginDataSource struct{}

func (d *DBPluginDataSource) GetPluginResource(ctx context.Context, pluginID string, version string) (*plugin.PluginResource, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrPluginNotFound, "")
	}

	res := &plugin.PluginResource{
		Name:         p.Name,
		Slug:         p.PluginID,
		Version:      p.Version,
		Description:  p.Description,
		EntryPoint:   "_start",
		ManifestJSON: p.ConfigJSON,
	}

	return res, nil
}

type InstallPluginRequest struct {
	PluginID              string   `json:"plugin_id"`
	Version               string   `json:"version"`
	LicenseKey            string   `json:"license_key"`
	Name                  string   `json:"name"`
	Icon                  string   `json:"icon"`
	Description           string   `json:"description"`
	ConfigJSON            string   `json:"config_json"`
	RenderMode            string   `json:"render_mode"`
	Entry                 string   `json:"entry"`
	Style                 string   `json:"style"`
	Signature             string   `json:"signature"`
	AllowedAPIPrefixes    []string `json:"allowed_api_prefixes"`
	CompatibleBaseVersion string   `json:"compatible_base_version"`
	UseShadowDOM          bool     `json:"use_shadow_dom"`
}

type UpdatePluginConfigRequest struct {
	ConfigJSON string `json:"config_json"`
}

func (s *PluginEngineService) ListPlugins(ctx context.Context) ([]*model.InstalledPlugin, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var plugins []*model.InstalledPlugin
	if err := db.Find(&plugins).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询已安装插件失败")
	}

	return plugins, nil
}

func (s *PluginEngineService) GetPlugin(ctx context.Context, pluginID string) (*model.InstalledPlugin, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrPluginNotFound, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	return &p, nil
}

func (s *PluginEngineService) InstallPlugin(ctx context.Context, req *InstallPluginRequest) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	tenantID, _ := utils.GetTenantIDFromCtx(ctx)
	if tenantID == 0 {
		tenantID = 1
	}

	var existing model.InstalledPlugin
	err := db.Where("plugin_id = ?", req.PluginID).First(&existing).Error
	if err == nil {
		return errs.NewServiceError(errs.ErrPluginAlreadyInstalled, "")
	}
	if err != gorm.ErrRecordNotFound {
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	if req.LicenseKey != "" {
		licenseReq := &VerifyLicenseRequest{
			LicenseKey: req.LicenseKey,
			PluginID:   req.PluginID,
			Domain:     "",
		}
		licenseResp, err := ServiceGroupApp.LicenseService.VerifyLicense(ctx, licenseReq)
		if err != nil {
			return errs.NewServiceError(errs.ErrLicenseInvalid, fmt.Sprintf("License验证请求失败: %v", err))
		}
		if !licenseResp.Valid {
			return errs.NewServiceError(errs.ErrLicenseInvalid, fmt.Sprintf("License验证失败: %s", licenseResp.Message))
		}
	}

	engine := plugin.GetEngine()
	if err := engine.Install(ctx, tenantID, req.PluginID, req.Version, req.LicenseKey); err != nil {
		return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("安装插件失败: %v", err))
	}

	renderMode := req.RenderMode
	if renderMode == "" {
		renderMode = "schema"
	}
	entry := req.Entry
	if entry == "" {
		entry = "index.js"
	}
	allowedAPIPrefixesJSON, _ := json.Marshal(req.AllowedAPIPrefixes)

	now := time.Now()
	installedPlugin := &model.InstalledPlugin{
		TenantModel: model.TenantModel{
			TenantID: tenantID,
		},
		PluginID:              req.PluginID,
		Name:                  req.Name,
		Version:               req.Version,
		Icon:                  req.Icon,
		Description:           req.Description,
		ConfigJSON:            req.ConfigJSON,
		LicenseKey:            req.LicenseKey,
		RenderMode:            renderMode,
		Entry:                 entry,
		Style:                 req.Style,
		Signature:             req.Signature,
		AllowedAPIPrefixes:    string(allowedAPIPrefixesJSON),
		CompatibleBaseVersion: req.CompatibleBaseVersion,
		UseShadowDOM:          req.UseShadowDOM,
		Status:                "active",
		InstalledAt:           &now,
		UpdatedAt:             &now,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(installedPlugin).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "创建插件记录失败")
		}
		return nil
	})
	if err != nil {
		engine.Uninstall(ctx, tenantID, req.PluginID)
		return err
	}

	s.syncPluginMenusToDB(ctx, tenantID, req.PluginID)

	s.RecordVersionHistory(ctx, req.PluginID, req.Version, "", "install", "", req.ConfigJSON)

	s.logEvent(ctx, req.PluginID, "install", fmt.Sprintf("安装插件 %s 版本 %s", req.Name, req.Version))

	return nil
}

func (s *PluginEngineService) UninstallPlugin(ctx context.Context, pluginID string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	engine := plugin.GetEngine()
	if err := engine.Uninstall(ctx, p.TenantID, pluginID); err != nil {
		return errs.NewServiceError(errs.ErrPluginUninstallFailed, fmt.Sprintf("卸载插件失败: %v", err))
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&p).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "删除插件记录失败")
		}

		platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))
		if platformDB != nil {
			if p.ConfigJSON != "" {
				var manifest struct {
					Menus []struct {
						Path      string `json:"path"`
						Component string `json:"component"`
					} `json:"menus"`
				}
				if json.Unmarshal([]byte(p.ConfigJSON), &manifest) == nil {
					for _, menuReg := range manifest.Menus {
						platformDB.Where("path = ? AND component = ? AND tenant_id = ?", menuReg.Path, menuReg.Component, p.TenantID).Delete(&model.Menu{})
					}
				}
			}
			var pluginAppsMenu model.Menu
			if err := platformDB.Where("path = ? AND tenant_id = ?", "/plugin-apps", p.TenantID).First(&pluginAppsMenu).Error; err == nil {
				var childCount int64
				platformDB.Model(&model.Menu{}).Where("parent_id = ? AND tenant_id = ?", pluginAppsMenu.ID, p.TenantID).Count(&childCount)
				if childCount == 0 {
					platformDB.Where("tenant_id = ?", p.TenantID).Delete(&pluginAppsMenu)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.logEvent(ctx, pluginID, "uninstall", fmt.Sprintf("卸载插件 %s", p.Name))

	return nil
}

func (s *PluginEngineService) EnablePlugin(ctx context.Context, pluginID string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	engine := plugin.GetEngine()
	if err := engine.Enable(ctx, p.TenantID, pluginID); err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("启用插件失败: %v", err))
	}

	now := time.Now()
	p.Status = "active"
	p.UpdatedAt = &now
	if err := db.Save(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新插件状态失败")
	}

	s.syncPluginMenusToDB(ctx, p.TenantID, pluginID)

	s.logEvent(ctx, pluginID, "enable", fmt.Sprintf("启用插件 %s", p.Name))

	return nil
}

func (s *PluginEngineService) DisablePlugin(ctx context.Context, pluginID string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	engine := plugin.GetEngine()
	if err := engine.Disable(ctx, p.TenantID, pluginID); err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("禁用插件失败: %v", err))
	}

	now := time.Now()
	p.Status = "disabled"
	p.UpdatedAt = &now
	if err := db.Save(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新插件状态失败")
	}

	s.logEvent(ctx, pluginID, "disable", fmt.Sprintf("禁用插件 %s", p.Name))

	return nil
}

func (s *PluginEngineService) UpgradePlugin(ctx context.Context, pluginID string, newVersion string, newLicenseKey string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	engine := plugin.GetEngine()
	if err := engine.Upgrade(ctx, p.TenantID, pluginID, p.Version, newVersion, newLicenseKey); err != nil {
		return errs.NewServiceError(errs.ErrPluginVersionConflict, fmt.Sprintf("升级插件失败: %v", err))
	}

	prevVersion := p.Version

	now := time.Now()
	p.Version = newVersion
	p.LicenseKey = newLicenseKey
	p.UpdatedAt = &now
	if err := db.Save(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新插件版本失败")
	}

	s.RecordVersionHistory(ctx, pluginID, newVersion, prevVersion, "upgrade", "", p.ConfigJSON)

	s.logEvent(ctx, pluginID, "upgrade", fmt.Sprintf("升级插件 %s: %s→%s", p.Name, prevVersion, newVersion))

	return nil
}

func (s *PluginEngineService) GetPluginConfig(ctx context.Context, pluginID string) (map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	var config map[string]interface{}
	if p.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(p.ConfigJSON), &config); err != nil {
			return nil, errs.NewServiceError(errs.ErrInternalServer, "解析配置失败")
		}
	} else {
		config = make(map[string]interface{})
	}

	var configs []model.PluginConfig
	if err := db.Where("plugin_id = ?", pluginID).Find(&configs).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询插件配置失败")
	}

	for _, c := range configs {
		var val interface{}
		if err := json.Unmarshal([]byte(c.ConfigValue), &val); err != nil {
			config[c.ConfigKey] = c.ConfigValue
		} else {
			config[c.ConfigKey] = val
		}
	}

	return config, nil
}

func (s *PluginEngineService) UpdatePluginConfig(ctx context.Context, pluginID string, config map[string]interface{}) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, "序列化配置失败")
	}

	now := time.Now()
	p.ConfigJSON = string(configJSON)
	p.UpdatedAt = &now
	if err := db.Save(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新插件配置失败")
	}

	return nil
}

func (s *PluginEngineService) getDefaultIconURL(seed string) string {
	baseURL := "https://api.dicebear.com/7.x/identicon/svg"
	if config.GlobalConfig != nil && config.GlobalConfig.PluginEngine.DefaultIconURL != "" {
		baseURL = config.GlobalConfig.PluginEngine.DefaultIconURL
	}
	return baseURL + "?seed=" + seed
}

func (s *PluginEngineService) logEvent(ctx context.Context, pluginID string, eventType string, eventData string) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	now := time.Now()
	eventLog := &model.PluginEventLog{
		PluginID:  pluginID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: &now,
	}

	_ = db.Create(eventLog).Error
}

func (s *PluginEngineService) GetPluginPage(ctx context.Context, pluginID string) (map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrPluginNotFound, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	if p.Status != "active" {
		return nil, errs.NewServiceError(errs.ErrPluginDisabled, "")
	}

	manifest, err := plugin.ParseManifest(p.ConfigJSON)
	if err != nil {
		return map[string]interface{}{
			"type": "html",
			"html": fmt.Sprintf("<div class='text-center py-8'><h3 class='text-lg font-semibold text-slate-600'>%s</h3><p class='text-slate-400 mt-2'>插件配置解析失败</p></div>", p.Name),
		}, nil
	}

	if manifest.Page == nil {
		return map[string]interface{}{
			"type": "html",
			"html": fmt.Sprintf("<div class='text-center py-8'><h3 class='text-lg font-semibold text-slate-600'>%s</h3><p class='text-slate-400 mt-2'>插件页面数据加载中...</p></div>", p.Name),
		}, nil
	}

	result := map[string]interface{}{
		"type": manifest.Page.Type,
	}

	if manifest.Page.Columns != nil {
		result["columns"] = manifest.Page.Columns
	}
	if manifest.Page.Actions != nil {
		result["actions"] = manifest.Page.Actions
	}
	if manifest.Page.Form != nil {
		result["form"] = manifest.Page.Form
	}
	if manifest.Page.Stats != nil {
		result["stats"] = manifest.Page.Stats
	}

	if manifest.Page.Table != "" {
		items, err := s.loadPluginTableData(ctx, db, manifest.Page.Table)
		if err == nil {
			result["items"] = items
		} else {
			result["items"] = []interface{}{}
		}
	}

	if manifest.Page.API != "" {
		apiPath := manifest.Page.API
		if strings.HasPrefix(apiPath, "/api/plugin/") {
			apiPath = "/api/plugin-data/" + strings.TrimPrefix(apiPath, "/api/plugin/")
		}
		if strings.HasPrefix(apiPath, "/api/") {
			apiPath = strings.TrimPrefix(apiPath, "/api")
		}
		result["api"] = apiPath
	}

	return result, nil
}

func (s *PluginEngineService) loadPluginTableData(_ context.Context, db *gorm.DB, tableName string) ([]map[string]interface{}, error) {
	rows, err := db.Table(tableName).Order("created_at DESC").Limit(50).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	cols, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			byteVal, ok := val.([]byte)
			if ok {
				row[col] = string(byteVal)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}
	return results, nil
}

func (s *PluginEngineService) InstallDemoPlugin(ctx context.Context) error {
	tenantID, _ := utils.GetTenantIDFromCtx(ctx)

	req := &InstallPluginRequest{
		PluginID:    "com.fayhub.announcement",
		Version:     "1.0.0",
		LicenseKey:  "demo-license-key",
		Name:        "公告管理",
		Icon:        s.getDefaultIconURL("announcement"),
		Description: "企业公告发布与管理插件，支持公告的创建、编辑、发布和撤回",
	}

	if err := s.InstallPlugin(ctx, req); err != nil {
		if se, ok := err.(*errs.ServiceError); ok && se.Code == errs.ErrPluginAlreadyInstalled {
			s.ensurePluginTableExists(ctx, "plugin_announcements", map[string]string{
				"id":         "SERIAL PRIMARY KEY",
				"title":      "VARCHAR(255) NOT NULL DEFAULT ''",
				"content":    "TEXT DEFAULT ''",
				"status":     "VARCHAR(50) NOT NULL DEFAULT 'draft'",
				"is_top":     "BOOLEAN NOT NULL DEFAULT FALSE",
				"created_at": "TIMESTAMP NOT NULL DEFAULT NOW()",
				"updated_at": "TIMESTAMP NOT NULL DEFAULT NOW()",
			})
			return nil
		}
		return err
	}

	s.ensurePluginTableExists(ctx, "plugin_announcements", map[string]string{
		"id":         "SERIAL PRIMARY KEY",
		"title":      "VARCHAR(255) NOT NULL DEFAULT ''",
		"content":    "TEXT DEFAULT ''",
		"status":     "VARCHAR(50) NOT NULL DEFAULT 'draft'",
		"is_top":     "BOOLEAN NOT NULL DEFAULT FALSE",
		"created_at": "TIMESTAMP NOT NULL DEFAULT NOW()",
		"updated_at": "TIMESTAMP NOT NULL DEFAULT NOW()",
	})

	manifest := &plugin.Manifest{
		Name:       "公告管理",
		Version:    "1.0.0",
		EntryPoint: "_start",
		Routes: []plugin.RouteRegistration{
			{Method: "GET", Path: "/announcements", Handler: "list_announcements"},
			{Method: "POST", Path: "/announcements", Handler: "create_announcement"},
			{Method: "PUT", Path: "/announcements/:id", Handler: "update_announcement"},
			{Method: "DELETE", Path: "/announcements/:id", Handler: "delete_announcement"},
		},
		APIs: []plugin.APIRegistration{
			{Method: "GET", Path: "/api/plugin/announcements", Group: "公告管理"},
			{Method: "POST", Path: "/api/plugin/announcements", Group: "公告管理"},
		},
		Menus: []plugin.MenuRegistration{
			{
				Title:     "公告管理",
				Path:      "/plugin-apps/announcement",
				Icon:      "Tickets",
				Component: "com.fayhub.announcement",
				Sort:      1,
			},
		},
		Page: &plugin.PageConfig{
			Type:  "table",
			Table: "plugin_announcements",
			API:   "/api/plugin/announcements",
			Stats: []map[string]interface{}{
				{"key": "total", "label": "全部公告", "icon": "Document", "color": "blue"},
				{"key": "published", "label": "已发布", "icon": "CircleCheck", "color": "green", "filter": "status=published"},
				{"key": "draft", "label": "草稿", "icon": "EditPen", "color": "orange", "filter": "status=draft"},
			},
			Columns: []map[string]interface{}{
				{"key": "id", "label": "ID", "width": 80},
				{"key": "title", "label": "标题", "width": 220},
				{"key": "content", "label": "内容", "width": 300},
				{"key": "status", "label": "状态", "width": 100, "type": "tag", "options": []map[string]interface{}{{"value": "published", "label": "已发布", "type": "success"}, {"value": "draft", "label": "草稿", "type": "warning"}, {"value": "revoked", "label": "已撤回", "type": "info"}}},
				{"key": "created_at", "label": "创建时间", "width": 180},
			},
			Actions: []map[string]interface{}{
				{"key": "create", "label": "新建公告", "type": "primary", "icon": "Plus"},
				{"key": "edit", "label": "编辑", "type": "primary", "icon": ""},
				{"key": "delete", "label": "删除", "type": "danger", "icon": ""},
			},
			Form: []map[string]interface{}{
				{"key": "title", "label": "标题", "type": "input", "required": true, "maxlength": 100},
				{"key": "content", "label": "内容", "type": "textarea", "required": true, "maxlength": 2000, "rows": 6},
				{"key": "is_top", "label": "置顶", "type": "switch"},
				{"key": "status", "label": "状态", "type": "radio", "options": []map[string]interface{}{{"value": "draft", "label": "保存为草稿"}, {"value": "published", "label": "直接发布"}}, "default": "draft"},
			},
		},
	}

	manifestJSON, _ := json.Marshal(manifest)
	db := utils.GetDB(ctx)
	if db != nil {
		db.Model(&model.InstalledPlugin{}).Where("plugin_id = ?", req.PluginID).Update("config_json", string(manifestJSON))
	}

	engine := plugin.GetEngine()
	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		return nil
	}

	registry := wasmEngine.GetRegistry()
	registry.RegisterRoutes(tenantID, "com.fayhub.announcement", manifest.Routes)
	registry.RegisterAPIs(tenantID, "com.fayhub.announcement", manifest.APIs)
	registry.RegisterMenus(tenantID, "com.fayhub.announcement", manifest.Menus)

	s.syncPluginMenusToDB(ctx, tenantID, "com.fayhub.announcement")

	s.logEvent(ctx, "com.fayhub.announcement", "demo_install", "安装示例插件: 公告管理 v1.0.0")

	s.installDemoFrontendPlugin(ctx, tenantID)

	return nil
}

func (s *PluginEngineService) installDemoFrontendPlugin(ctx context.Context, tenantID uint) {
	demoReq := &InstallPluginRequest{
		PluginID:              "demo-plugin",
		Version:               "1.0.0",
		LicenseKey:            "demo-license-key",
		Name:                  "示例前端插件",
		Icon:                  s.getDefaultIconURL("demo-plugin"),
		Description:           "FayHub 前端自定义组件示例插件，演示沙箱加载与 Bridge 通信",
		RenderMode:            "custom",
		Entry:                 "index.js",
		Style:                 "style.css",
		AllowedAPIPrefixes:    []string{"/plugin-engine/plugins/demo-plugin/"},
		CompatibleBaseVersion: ">=1.0.0 <2.0.0",
	}

	if err := s.InstallPlugin(ctx, demoReq); err != nil {
		if se, ok := err.(*errs.ServiceError); ok && se.Code == errs.ErrPluginAlreadyInstalled {
			return
		}
		return
	}

	manifest := &plugin.Manifest{
		Name:       "示例前端插件",
		Version:    "1.0.0",
		EntryPoint: "_start",
		Menus: []plugin.MenuRegistration{
			{
				Title:     "示例前端插件",
				Path:      "/plugin-apps/demo-plugin",
				Icon:      "Box",
				Component: "demo-plugin",
				Sort:      2,
			},
		},
	}

	manifestJSON, _ := json.Marshal(manifest)
	db := utils.GetDB(ctx)
	if db != nil {
		db.Model(&model.InstalledPlugin{}).Where("plugin_id = ?", "demo-plugin").Update("config_json", string(manifestJSON))
	}

	engine := plugin.GetEngine()
	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		return
	}

	registry := wasmEngine.GetRegistry()
	registry.RegisterMenus(tenantID, "demo-plugin", manifest.Menus)

	s.syncPluginMenusToDB(ctx, tenantID, "demo-plugin")

	s.logEvent(ctx, "demo-plugin", "demo_install", "安装示例前端插件 v1.0.0")
}

func (s *PluginEngineService) syncPluginMenusToDB(ctx context.Context, tenantID uint, pluginID string) {
	engine := plugin.GetEngine()
	wasmEngine, ok := engine.(*plugin.WASMEngine)
	if !ok {
		return
	}

	manifest := wasmEngine.GetManifest(tenantID, pluginID)
	layout := "embedded"
	if manifest != nil && manifest.Layout != "" {
		layout = manifest.Layout
	}

	registry := wasmEngine.GetRegistry()
	menus := registry.GetMenus(tenantID, pluginID)
	if len(menus) == 0 {
		return
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if platformDB == nil {
		return
	}

	var pluginAppsMenu model.Menu
	if err := platformDB.Where("path = ?", "/plugin-apps").First(&pluginAppsMenu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pluginAppsMenu = model.Menu{
				ParentID:   0,
				Title:      "插件应用",
				Path:       "/plugin-apps",
				Icon:       "Grid",
				Sort:       4,
				Type:       1,
				Status:     1,
				Permission: "",
				Layout:     "embedded",
			}
			if createErr := platformDB.Create(&pluginAppsMenu).Error; createErr != nil {
				s.logEvent(ctx, "system", "menu_create_error", fmt.Sprintf("创建插件应用菜单失败: %v", createErr))
				return
			}
		} else {
			return
		}
	}

	for _, menuReg := range menus {
		var existingMenu model.Menu
		result := platformDB.Where("path = ? AND component = ?", menuReg.Path, menuReg.Component).First(&existingMenu)
		if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newMenu := model.Menu{
				ParentID:   pluginAppsMenu.ID,
				Title:      menuReg.Title,
				Path:       menuReg.Path,
				Component:  menuReg.Component,
				Icon:       menuReg.Icon,
				Sort:       menuReg.Sort,
				Type:       2,
				Status:     1,
				Permission: "",
				Layout:     layout,
			}
			platformDB.Create(&newMenu)
		}
	}
}

func (s *PluginEngineService) SearchMarketPlugins(ctx context.Context, keyword string, page, pageSize int, categoryID string) (*market.PluginListResponse, error) {
	client := market.GetClient()
	if client == nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "Market客户端未初始化")
	}

	result, err := client.SearchPlugins(ctx, keyword, page, pageSize, categoryID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("搜索Market插件失败: %v", err))
	}

	return result, nil
}

func (s *PluginEngineService) GetMarketPluginDetail(ctx context.Context, pluginID string) (*market.PluginDetail, error) {
	client := market.GetClient()
	if client == nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "Market客户端未初始化")
	}

	result, err := client.GetPluginDetail(ctx, pluginID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("获取Market插件详情失败: %v", err))
	}

	return result, nil
}

func (s *PluginEngineService) GetMarketCategories(ctx context.Context) ([]market.CategoryItem, error) {
	client := market.GetClient()
	if client == nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "Market客户端未初始化")
	}

	result, err := client.GetCategories(ctx)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("获取Market分类失败: %v", err))
	}

	return result, nil
}

func (s *PluginEngineService) InstallFromMarket(ctx context.Context, marketPluginID string, targetVersion string, licenseKey string, fayhubUserId string) error {
	client := market.GetClient()
	if client == nil {
		return errs.NewServiceError(errs.ErrInternalServer, "Market客户端未初始化")
	}

	if err := client.EnsureAuthenticated(ctx); err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("Market认证失败: %v", err))
	}

	detail, err := client.GetPluginDetail(ctx, marketPluginID)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("获取插件详情失败: %v", err))
	}

	installResp, err := client.InstallPlugin(ctx, marketPluginID, targetVersion, fayhubUserId)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("获取安装令牌失败: %v", err))
	}

	if !installResp.Success || installResp.InstallToken == "" {
		return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("安装令牌获取失败: %s", installResp.Message))
	}

	tokenResp, err := client.VerifyInstallToken(ctx, installResp.InstallToken)
	if err != nil {
		return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("验证安装令牌失败: %v", err))
	}

	if !tokenResp.Valid {
		return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("安装令牌无效: %s", tokenResp.Message))
	}

	effectiveLicenseKey := licenseKey
	if tokenResp.LicenseKey != "" {
		effectiveLicenseKey = tokenResp.LicenseKey
	}

	var wasmBytes []byte
	var manifestJSON string
	effectiveVersion := targetVersion

	if tokenResp.LatestVersion != nil {
		effectiveVersion = tokenResp.LatestVersion.Version

		if tokenResp.LatestVersion.DownloadURL != "" {
			wasmBytes, err = client.DownloadWASM(ctx, tokenResp.LatestVersion.DownloadURL)
			if err != nil {
				return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("下载WASM文件失败: %v", err))
			}
		}

		if tokenResp.LatestVersion.PackageHash != "" && len(wasmBytes) > 0 {
			if err := pluginsign.VerifyPluginHash(wasmBytes, tokenResp.LatestVersion.PackageHash); err != nil {
				return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("插件完整性校验失败: %v", err))
			}
		}
	} else if len(detail.Versions) > 0 {
		version := detail.Versions[0]
		if targetVersion != "" {
			for _, v := range detail.Versions {
				if v.Version == targetVersion {
					version = v
					break
				}
			}
		}
		effectiveVersion = version.Version

		if version.WasmURL != "" {
			wasmBytes, err = client.DownloadWASM(ctx, version.WasmURL)
			if err != nil {
				return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("下载WASM文件失败: %v", err))
			}
		}

		if version.Signature != "" && len(wasmBytes) > 0 {
			if err := pluginsign.VerifyPlugin(wasmBytes, version.Signature); err != nil {
				return errs.NewServiceError(errs.ErrPluginInstallFailed, fmt.Sprintf("插件签名校验失败: %v", err))
			}
		}
	}

	if manifestJSON == "" {
		manifestJSON = buildDefaultManifest(detail)
	}

	pluginName := tokenResp.PluginSlug
	if pluginName == "" {
		pluginName = detail.Slug
	}
	if pluginName == "" {
		pluginName = detail.Name
	}

	pluginIcon := tokenResp.PluginIcon
	if pluginIcon == "" {
		pluginIcon = detail.CoverImage
	}

	req := &InstallPluginRequest{
		PluginID:    marketPluginID,
		Version:     effectiveVersion,
		LicenseKey:  effectiveLicenseKey,
		Name:        pluginName,
		Icon:        pluginIcon,
		Description: tokenResp.PluginDescription,
		ConfigJSON:  manifestJSON,
	}

	if err := s.InstallPlugin(ctx, req); err != nil {
		return err
	}

	if len(wasmBytes) > 0 {
		s.logEvent(ctx, marketPluginID, "market_download", fmt.Sprintf("从Market下载WASM: %d bytes", len(wasmBytes)))
	}

	s.logEvent(ctx, marketPluginID, "market_install", fmt.Sprintf("从Market安装插件: %s v%s (installToken验证通过)", tokenResp.PluginName, effectiveVersion))

	return nil
}

func buildDefaultManifest(detail *market.PluginDetail) string {
	manifest := map[string]interface{}{
		"name":            detail.Slug,
		"version":         "1.0.0",
		"entry_point":     "_start",
		"description":     detail.Description,
		"min_app_version": "0.1.0",
		"layout":          "embedded",
		"permissions":     []string{"log:write"},
		"config_schema":   map[string]interface{}{},
		"routes":          []interface{}{},
		"apis":            []interface{}{},
	}

	jsonBytes, _ := json.Marshal(manifest)
	return string(jsonBytes)
}

func (s *PluginEngineService) RecordVersionHistory(ctx context.Context, pluginID string, version string, prevVersion string, action string, wasmHash string, manifestJSON string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	now := time.Now()
	history := &model.PluginVersionHistory{
		PluginID:     pluginID,
		Version:      version,
		PrevVersion:  prevVersion,
		Action:       action,
		WasmHash:     wasmHash,
		ManifestJSON: manifestJSON,
		CreatedAt:    &now,
	}

	return db.Create(history).Error
}

func (s *PluginEngineService) GetVersionHistory(ctx context.Context, pluginID string, page, pageSize int) ([]*model.PluginVersionHistory, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.PluginVersionHistory{}).Where("plugin_id = ?", pluginID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询版本历史失败")
	}

	var histories []*model.PluginVersionHistory
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&histories).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询版本历史失败")
	}

	return histories, total, nil
}

func (s *PluginEngineService) RollbackPlugin(ctx context.Context, pluginID string, targetVersion string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrPluginNotInstalled, "")
	}

	var history model.PluginVersionHistory
	if err := db.Where("plugin_id = ? AND version = ?", pluginID, targetVersion).First(&history).Error; err != nil {
		return errs.NewServiceError(errs.ErrPluginVersionConflict, "目标版本不存在于历史记录中")
	}

	currentVersion := p.Version

	engine := plugin.GetEngine()
	if err := engine.Upgrade(ctx, p.TenantID, pluginID, currentVersion, targetVersion, p.LicenseKey); err != nil {
		return errs.NewServiceError(errs.ErrPluginVersionConflict, fmt.Sprintf("回滚插件失败: %v", err))
	}

	now := time.Now()
	p.Version = targetVersion
	p.UpdatedAt = &now
	if history.ManifestJSON != "" {
		p.ConfigJSON = history.ManifestJSON
	}
	if err := db.Save(&p).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "回滚插件版本失败")
	}

	s.RecordVersionHistory(ctx, pluginID, targetVersion, currentVersion, "rollback", history.WasmHash, history.ManifestJSON)

	rollbackTime := now
	db.Model(&model.PluginVersionHistory{}).Where("id = ?", history.ID).Update("rollback_at", rollbackTime)

	s.logEvent(ctx, pluginID, "rollback", fmt.Sprintf("回滚插件 %s: %s→%s", p.Name, currentVersion, targetVersion))

	return nil
}

func (s *PluginEngineService) CheckForUpdates(ctx context.Context, pluginID string) (*market.PluginVersion, error) {
	client := market.GetClient()
	if client == nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "Market客户端未初始化")
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var p model.InstalledPlugin
	if err := db.Where("plugin_id = ?", pluginID).First(&p).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrPluginNotInstalled, "")
	}

	detail, err := client.GetPluginDetail(ctx, pluginID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, fmt.Sprintf("获取插件详情失败: %v", err))
	}

	if len(detail.Versions) == 0 {
		return nil, nil
	}

	latestVersion := detail.Versions[0]
	for _, v := range detail.Versions {
		if v.Version > latestVersion.Version {
			latestVersion = v
		}
	}

	if latestVersion.Version <= p.Version {
		return nil, nil
	}

	return &latestVersion, nil
}

func (s *PluginEngineService) CheckAllUpdates(ctx context.Context) ([]map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var plugins []model.InstalledPlugin
	if err := db.Where("status = ?", "active").Find(&plugins).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询已安装插件失败")
	}

	var updates []map[string]interface{}
	for _, p := range plugins {
		update, err := s.CheckForUpdates(ctx, p.PluginID)
		if err != nil {
			continue
		}
		if update != nil {
			updates = append(updates, map[string]interface{}{
				"plugin_id":       p.PluginID,
				"name":            p.Name,
				"current_version": p.Version,
				"latest_version":  update.Version,
				"changelog":       update.Changelog,
			})
		}
	}

	return updates, nil
}

func (s *PluginEngineService) SaveDependencies(ctx context.Context, pluginID string, deps []map[string]interface{}) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	db.Where("plugin_id = ?", pluginID).Delete(&model.PluginDependency{})

	for _, dep := range deps {
		depPluginID, _ := dep["dep_plugin_id"].(string)
		if depPluginID == "" {
			continue
		}

		versionMin, _ := dep["dep_version_min"].(string)
		versionMax, _ := dep["dep_version_max"].(string)
		isRequired := true
		if v, ok := dep["is_required"].(bool); ok {
			isRequired = v
		}

		dependency := &model.PluginDependency{
			PluginID:      pluginID,
			DepPluginID:   depPluginID,
			DepVersionMin: versionMin,
			DepVersionMax: versionMax,
			IsRequired:    isRequired,
		}

		if err := db.Create(dependency).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, fmt.Sprintf("保存依赖失败: %v", err))
		}
	}

	return nil
}

func (s *PluginEngineService) GetDependencies(ctx context.Context, pluginID string) ([]*model.PluginDependency, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var deps []*model.PluginDependency
	if err := db.Where("plugin_id = ?", pluginID).Find(&deps).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询依赖失败")
	}

	return deps, nil
}

func (s *PluginEngineService) ValidateDependencies(ctx context.Context, pluginID string) ([]map[string]interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var deps []*model.PluginDependency
	if err := db.Where("plugin_id = ? AND is_required = ?", pluginID, true).Find(&deps).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询依赖失败")
	}

	var issues []map[string]interface{}
	for _, dep := range deps {
		var installed model.InstalledPlugin
		err := db.Where("plugin_id = ? AND status = ?", dep.DepPluginID, "active").First(&installed).Error

		if err != nil {
			issues = append(issues, map[string]interface{}{
				"dep_plugin_id": dep.DepPluginID,
				"issue":         "missing",
				"message":       fmt.Sprintf("缺少必要依赖: %s", dep.DepPluginID),
			})
			continue
		}

		if dep.DepVersionMin != "" && installed.Version < dep.DepVersionMin {
			issues = append(issues, map[string]interface{}{
				"dep_plugin_id":   dep.DepPluginID,
				"issue":           "version_too_low",
				"current_version": installed.Version,
				"min_version":     dep.DepVersionMin,
				"message":         fmt.Sprintf("依赖 %s 版本过低: 当前 %s, 需要 >= %s", dep.DepPluginID, installed.Version, dep.DepVersionMin),
			})
		}

		if dep.DepVersionMax != "" && installed.Version > dep.DepVersionMax {
			issues = append(issues, map[string]interface{}{
				"dep_plugin_id":   dep.DepPluginID,
				"issue":           "version_too_high",
				"current_version": installed.Version,
				"max_version":     dep.DepVersionMax,
				"message":         fmt.Sprintf("依赖 %s 版本过高: 当前 %s, 需要 <= %s", dep.DepPluginID, installed.Version, dep.DepVersionMax),
			})
		}
	}

	return issues, nil
}

func (s *PluginEngineService) ensurePluginTableExists(ctx context.Context, tableName string, columns map[string]string) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	if db.Migrator().HasTable(tableName) {
		return
	}

	colDefs := make([]string, 0, len(columns))
	for name, def := range columns {
		colDefs = append(colDefs, fmt.Sprintf("%s %s", name, def))
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(colDefs, ", "))
	if err := db.Exec(sql).Error; err != nil {
		s.logEvent(ctx, "", "table_create_error", fmt.Sprintf("创建插件表 %s 失败: %v", tableName, err))
	} else {
		s.logEvent(ctx, "", "table_created", fmt.Sprintf("创建插件表 %s 成功", tableName))
	}
}

func (s *PluginEngineService) resolveTableNameFromAPI(ctx context.Context, apiPath string) (string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return "", errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	lookupPath := apiPath
	if strings.HasPrefix(lookupPath, "/api/plugin-data/") {
		lookupPath = "/api/plugin/" + strings.TrimPrefix(lookupPath, "/api/plugin-data/")
	}

	var plugins []model.InstalledPlugin
	if err := db.Where("status = ?", "active").Find(&plugins).Error; err != nil {
		return "", errs.NewServiceError(errs.ErrDatabase, "查询插件失败")
	}

	for _, p := range plugins {
		manifest, err := plugin.ParseManifest(p.ConfigJSON)
		if err != nil {
			continue
		}
		if manifest.Page != nil && manifest.Page.API == lookupPath && manifest.Page.Table != "" {
			return manifest.Page.Table, nil
		}
	}

	return "", errs.NewServiceError(errs.ErrResourceNotFound, "未找到对应的插件数据表")
}

func (s *PluginEngineService) GetPluginData(ctx context.Context, apiPath string) ([]map[string]interface{}, error) {
	tableName, err := s.resolveTableNameFromAPI(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	db := utils.GetDB(ctx)
	items, err := s.loadPluginTableData(ctx, db, tableName)
	if err != nil {
		return []map[string]interface{}{}, nil
	}
	return items, nil
}

func (s *PluginEngineService) CreatePluginData(ctx context.Context, apiPath string, data map[string]interface{}) (map[string]interface{}, error) {
	tableName, err := s.resolveTableNameFromAPI(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	data["created_at"] = time.Now()
	data["updated_at"] = time.Now()

	if err := db.Table(tableName).Create(data).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建数据失败: "+err.Error())
	}

	return data, nil
}

func (s *PluginEngineService) UpdatePluginData(ctx context.Context, apiPath string, recordID string, data map[string]interface{}) error {
	tableName, err := s.resolveTableNameFromAPI(ctx, apiPath)
	if err != nil {
		return err
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	data["updated_at"] = time.Now()

	result := db.Table(tableName).Where("id = ?", recordID).Updates(data)
	if result.Error != nil {
		return errs.NewServiceError(errs.ErrDatabase, "更新数据失败: "+result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errs.NewServiceError(errs.ErrResourceNotFound, "记录不存在")
	}

	return nil
}

func (s *PluginEngineService) DeletePluginData(ctx context.Context, apiPath string, recordID string) error {
	tableName, err := s.resolveTableNameFromAPI(ctx, apiPath)
	if err != nil {
		return err
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	result := db.Table(tableName).Where("id = ?", recordID).Delete(nil)
	if result.Error != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除数据失败: "+result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errs.NewServiceError(errs.ErrResourceNotFound, "记录不存在")
	}

	return nil
}
