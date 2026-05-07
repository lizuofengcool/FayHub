// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package plugin

import (
	"context"
	"log"
	"time"
)

// PluginStatus 插件状态
type PluginStatus struct {
	PluginID      string
	Version       string
	Status        string // active/disabled/error
	InstalledAt   *time.Time
	LastHeartbeat *time.Time
}

// InstalledPluginInfo 已安装插件信息
type InstalledPluginInfo struct {
	PluginID    string
	Name        string
	Version     string
	Status      string
	Icon        string
	Description string
	InstalledAt *time.Time
}

// PluginEngine 插件引擎抽象接口
//
// FayHub 底座插件通信引擎 100% 基于 WASM (WebAssembly) 技术，
// 具体实现库为 tetratelabs/wazero。
//
// 架构选择 WASM 的三大核心优势：
// 1. 极致轻量：2核4G服务器可流畅运行 50+ 插件（每个插件运行时仅占数MB内存）
// 2. 沙箱隔离：WASM 天然提供内存隔离和权限控制，插件无法越权访问底座资源
// 3. 热更新：无需重启底座，瞬间加载/卸载/替换插件模块
//
// 通信模型：
// - 底座 → 插件：通过 Call 方法调用插件导出函数
// - 插件 → 底座：通过 Host Functions（host_log, host_http_request, host_db_query 等）
// - 所有跨沙箱通信必须经过 wazero 内存拷贝，零共享内存
type PluginEngine interface {
	// Install 安装插件：下载 WASM 模块 → 校验 Manifest → 注册 Host Functions → 实例化模块
	Install(ctx context.Context, tenantID int64, pluginID string, version string, licenseKey string) error

	// Uninstall 卸载插件：关闭 WASM 模块 → 释放 wazero 运行时内存 → 注销路由/API/菜单
	Uninstall(ctx context.Context, tenantID int64, pluginID string) error

	// Enable 启用插件：重新实例化 WASM 模块 → 注册路由/API/菜单
	Enable(ctx context.Context, tenantID int64, pluginID string) error

	// Disable 禁用插件：关闭 WASM 模块实例 → 释放内存 → 注销路由/API/菜单（保留元数据，可再次 Enable）
	Disable(ctx context.Context, tenantID int64, pluginID string) error

	// Upgrade 升级插件：卸载旧版本 → 安装新版本（原子操作，失败则回滚）
	Upgrade(ctx context.Context, tenantID int64, pluginID string, oldVersion string, newVersion string, licenseKey string) error

	// Call 底座向插件发起调用的核心入口
	// 调用指定插件的导出函数，传入 payload（JSON 编码的请求参数），返回插件处理结果
	// 典型场景：基座收到 HTTP 请求 → 路由到对应插件 → Call 执行插件逻辑 → 返回响应
	// 安全约束：
	//   - functionName 必须在 Manifest.exports 中声明，否则拒绝调用
	//   - payload 大小受 SandboxConfig.MaxMemoryBytes 限制
	//   - 调用超时由 ctx 的 Deadline 控制，超时自动终止 WASM 执行
	Call(ctx context.Context, tenantID int64, pluginID string, functionName string, payload []byte) ([]byte, error)

	// GetStatus 获取插件运行状态（active/disabled/error）
	GetStatus(ctx context.Context, tenantID int64, pluginID string) (*PluginStatus, error)

	// ListPlugins 列出指定租户已安装的所有插件
	ListPlugins(ctx context.Context, tenantID int64) ([]*InstalledPluginInfo, error)

	// HealthCheck 健康检查：验证插件模块是否正常加载、状态是否为 active
	HealthCheck(ctx context.Context, tenantID int64, pluginID string) error

	// Start 引擎启动：初始化 wazero Runtime（仅执行一次，重复调用自动跳过）
	Start(ctx context.Context) error

	// Stop 引擎停止：关闭所有插件模块 → 关闭 wazero Runtime → 释放所有内存
	Stop(ctx context.Context) error
}

// NoopEngine 空实现（开发阶段使用）
type NoopEngine struct{}

func (e *NoopEngine) Install(ctx context.Context, tenantID int64, pluginID string, version string, licenseKey string) error {
	log.Printf("[NoopEngine] 安装插件: tenant=%d, plugin=%s, version=%s", tenantID, pluginID, version)
	return nil
}

func (e *NoopEngine) Uninstall(ctx context.Context, tenantID int64, pluginID string) error {
	log.Printf("[NoopEngine] 卸载插件: tenant=%d, plugin=%s", tenantID, pluginID)
	return nil
}

func (e *NoopEngine) Enable(ctx context.Context, tenantID int64, pluginID string) error {
	log.Printf("[NoopEngine] 启用插件: tenant=%d, plugin=%s", tenantID, pluginID)
	return nil
}

func (e *NoopEngine) Disable(ctx context.Context, tenantID int64, pluginID string) error {
	log.Printf("[NoopEngine] 禁用插件: tenant=%d, plugin=%s", tenantID, pluginID)
	return nil
}

func (e *NoopEngine) Upgrade(ctx context.Context, tenantID int64, pluginID string, oldVersion string, newVersion string, licenseKey string) error {
	log.Printf("[NoopEngine] 升级插件: tenant=%d, plugin=%s, %s→%s", tenantID, pluginID, oldVersion, newVersion)
	return nil
}

func (e *NoopEngine) GetStatus(ctx context.Context, tenantID int64, pluginID string) (*PluginStatus, error) {
	return &PluginStatus{
		PluginID: pluginID,
		Version:  "1.0.0",
		Status:   "active",
	}, nil
}

func (e *NoopEngine) ListPlugins(ctx context.Context, tenantID int64) ([]*InstalledPluginInfo, error) {
	return []*InstalledPluginInfo{}, nil
}

func (e *NoopEngine) Call(ctx context.Context, tenantID int64, pluginID string, functionName string, payload []byte) ([]byte, error) {
	log.Printf("[NoopEngine] 调用插件函数: tenant=%d, plugin=%s, function=%s, payloadSize=%d", tenantID, pluginID, functionName, len(payload))
	return []byte(`{"noop":true}`), nil
}

func (e *NoopEngine) HealthCheck(ctx context.Context, tenantID int64, pluginID string) error {
	log.Printf("[NoopEngine] 健康检查: tenant=%d, plugin=%s", tenantID, pluginID)
	return nil
}

func (e *NoopEngine) Start(ctx context.Context) error {
	log.Printf("[NoopEngine] 引擎启动")
	return nil
}

func (e *NoopEngine) Stop(ctx context.Context) error {
	log.Printf("[NoopEngine] 引擎停止")
	return nil
}

var DefaultEngine PluginEngine = &NoopEngine{}

func SetEngine(engine PluginEngine) {
	DefaultEngine = engine
}

func GetEngine() PluginEngine {
	return DefaultEngine
}
