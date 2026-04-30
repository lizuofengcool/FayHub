// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package model

import "time"

// InstalledPlugin 已安装插件表（租户级）
type InstalledPlugin struct {
	TenantModel
	PluginID              string     `json:"plugin_id" gorm:"size:100;uniqueIndex:idx_tenant_plugin;not null"`
	Name                  string     `json:"name" gorm:"size:100;not null"`
	Version               string     `json:"version" gorm:"size:20;not null"`
	Icon                  string     `json:"icon" gorm:"size:255"`
	Description           string     `json:"description" gorm:"size:500"`
	ConfigJSON            string     `json:"config_json" gorm:"type:text"`
	LicenseKey            string     `json:"license_key" gorm:"size:255;not null" fayhub:"encrypt"`
	RenderMode            string     `json:"render_mode" gorm:"size:20;default:schema"`
	Entry                 string     `json:"entry" gorm:"size:255"`
	Style                 string     `json:"style" gorm:"size:255"`
	Signature             string     `json:"signature" gorm:"size:512"`
	AllowedAPIPrefixes    string     `json:"allowed_api_prefixes" gorm:"size:500"`
	CompatibleBaseVersion string     `json:"compatible_base_version" gorm:"size:50"`
	UseShadowDOM          bool       `json:"use_shadow_dom" gorm:"default:false"`
	Status                string     `json:"status" gorm:"size:20;index;default:active"`
	InstalledAt           *time.Time `json:"installed_at"`
	UpdatedAt             *time.Time `json:"updated_at"`
}

func (InstalledPlugin) TableName() string {
	return "installed_plugins"
}

// PluginConfig 插件配置表（租户级）
type PluginConfig struct {
	TenantModel
	PluginID    string `json:"plugin_id" gorm:"size:100;not null"`
	ConfigKey   string `json:"config_key" gorm:"size:100;not null"`
	ConfigValue string `json:"config_value" gorm:"type:text"`
}

func (PluginConfig) TableName() string {
	return "plugin_configs"
}

// PluginEventLog 插件事件日志表（租户级）
type PluginEventLog struct {
	TenantModel
	PluginID  string     `json:"plugin_id" gorm:"size:100;not null"`
	EventType string     `json:"event_type" gorm:"size:50;not null"` // install/uninstall/enable/disable/upgrade/error
	EventData string     `json:"event_data" gorm:"type:text"`
	CreatedAt *time.Time `json:"created_at"`
}

func (PluginEventLog) TableName() string {
	return "plugin_event_logs"
}

type PluginVersionHistory struct {
	TenantModel
	PluginID     string     `json:"plugin_id" gorm:"size:100;index;not null"`
	Version      string     `json:"version" gorm:"size:20;not null"`
	PrevVersion  string     `json:"prev_version" gorm:"size:20"`
	Action       string     `json:"action" gorm:"size:20;not null"`
	WasmHash     string     `json:"wasm_hash" gorm:"size:64"`
	ManifestJSON string     `json:"manifest_json" gorm:"type:text"`
	RollbackAt   *time.Time `json:"rollback_at,omitempty"`
	CreatedAt    *time.Time `json:"created_at"`
}

func (PluginVersionHistory) TableName() string {
	return "plugin_version_histories"
}

type PluginDependency struct {
	TenantModel
	PluginID      string `json:"plugin_id" gorm:"size:100;index;not null"`
	DepPluginID   string `json:"dep_plugin_id" gorm:"size:100;not null"`
	DepVersionMin string `json:"dep_version_min" gorm:"size:20"`
	DepVersionMax string `json:"dep_version_max" gorm:"size:20"`
	IsRequired    bool   `json:"is_required" gorm:"default:true"`
}

func (PluginDependency) TableName() string {
	return "plugin_dependencies"
}
