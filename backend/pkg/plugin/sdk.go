package plugin

import (
	"encoding/json"
	"fmt"
)

// SDK 插件开发工具包
// 提供便捷的工具函数和类型定义，简化插件开发
type SDK struct{}

// NewSDK 创建SDK实例
func NewSDK() *SDK {
	return &SDK{}
}

// CreateManifest 创建基础插件清单
func (s *SDK) CreateManifest(name, version, description string) *Manifest {
	return &Manifest{
		Name:         name,
		Version:      version,
		Description:  description,
		EntryPoint:   "_start",
		Layout:       "embedded",
		Permissions:  []string{},
		Routes:       []RouteRegistration{},
		APIs:         []APIRegistration{},
		Menus:        []MenuRegistration{},
	}
}

// AddPermission 添加权限声明
func (s *SDK) AddPermission(m *Manifest, permission string) error {
	if !validPermissions[permission] {
		return fmt.Errorf("无效的权限: %s", permission)
	}
	
	for _, p := range m.Permissions {
		if p == permission {
			return nil // 已存在
		}
	}
	
	m.Permissions = append(m.Permissions, permission)
	return nil
}

// AddRoute 添加路由注册
func (s *SDK) AddRoute(m *Manifest, method, path, handler string) {
	m.Routes = append(m.Routes, RouteRegistration{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

// AddAPI 添加API注册
func (s *SDK) AddAPI(m *Manifest, method, path, group string) {
	m.APIs = append(m.APIs, APIRegistration{
		Method: method,
		Path:   path,
		Group:  group,
	})
}

// AddMenu 添加菜单注册
func (s *SDK) AddMenu(m *Manifest, title, path, icon string, sort int) {
	m.Menus = append(m.Menus, MenuRegistration{
		Title: title,
		Path:  path,
		Icon:  icon,
		Sort:  sort,
	})
}

// MarshalManifest 序列化清单为JSON
func (s *SDK) MarshalManifest(m *Manifest) (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ValidateAndMarshal 验证并序列化清单
func (s *SDK) ValidateAndMarshal(m *Manifest) (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	return s.MarshalManifest(m)
}

// GetAllPermissions 获取所有可用权限列表
func (s *SDK) GetAllPermissions() []string {
	perms := make([]string, 0, len(validPermissions))
	for p := range validPermissions {
		perms = append(perms, p)
	}
	return perms
}

// Permission 权限常量
const (
	PermLogWrite   = "log:write"
	PermHTTPGet    = "http:get"
	PermHTTPPost   = "http:post"
	PermHTTPPut    = "http:put"
	PermHTTPDelete = "http:delete"
	PermDBRead     = "db:read"
	PermDBWrite    = "db:write"
	PermCacheRead  = "cache:read"
	PermCacheWrite = "cache:write"
)

// PluginTemplate 插件模板生成器
type PluginTemplate struct {
	Name        string
	Description string
}

// GenerateBasicTemplate 生成基础插件模板
func (pt *PluginTemplate) GenerateBasicTemplate() string {
	return fmt.Sprintf(`# %s 插件模板
## 说明
%s

## 快速开始
1. 编写WASM模块
2. 配置manifest.json
3. 测试插件
4. 打包发布

## 文件结构
plugin/
├── main.wasm
├── manifest.json
└── README.md

## manifest.json 示例
{
  "name": "%s",
  "version": "0.0.1",
  "entry_point": "_start",
  "description": "%s",
  "permissions": ["log:write"],
  "routes": [],
  "apis": [],
  "menus": []
}
`, pt.Name, pt.Description, pt.Name, pt.Description)
}

// GenerateAPITemplate 生成API插件模板
func (pt *PluginTemplate) GenerateAPITemplate() string {
	return fmt.Sprintf(`# %s API插件模板
## 说明
%s - 提供API接口的插件模板

## 功能特性
- 自定义API端点
- 数据库访问
- HTTP请求
- 缓存操作

## manifest.json 示例
{
  "name": "%s",
  "version": "0.0.1",
  "entry_point": "_start",
  "description": "%s",
  "permissions": ["log:write", "db:read", "db:write"],
  "apis": [
    {
      "method": "GET",
      "path": "/api/v1/data",
      "group": "plugin"
    }
  ],
  "routes": [
    {
      "method": "GET",
      "path": "/plugin/data",
      "handler": "handleGetData"
    }
  ]
}
`, pt.Name, pt.Description, pt.Name, pt.Description)
}

// Helper 辅助工具
type Helper struct{}

// NewHelper 创建辅助工具
func NewHelper() *Helper {
	return &Helper{}
}

// BuildRoutePath 构建插件路由路径
func (h *Helper) BuildRoutePath(pluginID, path string) string {
	if path == "" {
		return fmt.Sprintf("/plugin/%s", pluginID)
	}
	return fmt.Sprintf("/plugin/%s/%s", pluginID, path)
}

// BuildAPIPath 构建插件API路径
func (h *Helper) BuildAPIPath(pluginID, path string) string {
	if path == "" {
		return fmt.Sprintf("/api/plugin/%s", pluginID)
	}
	return fmt.Sprintf("/api/plugin/%s/%s", pluginID, path)
}

// ValidatePluginID 验证插件ID格式
func (h *Helper) ValidatePluginID(pluginID string) bool {
	if len(pluginID) < 3 || len(pluginID) > 50 {
		return false
	}
	for _, c := range pluginID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// SDKInfo SDK信息
type SDKInfo struct {
	Version          string
	SupportedLangs   []string
	MinAppVersion    string
	Permissions      []string
	HostFunctions    []string
}

// GetSDKInfo 获取SDK信息
func (s *SDK) GetSDKInfo() *SDKInfo {
	return &SDKInfo{
		Version:       "1.0.0",
		SupportedLangs: []string{"Go", "Rust", "AssemblyScript", "C/C++"},
		MinAppVersion: "1.0.0",
		Permissions:    s.GetAllPermissions(),
		HostFunctions: []string{
			"host_log",
			"host_http_request",
			"host_db_query",
			"host_db_exec",
			"host_cache_get",
			"host_cache_set",
		},
	}
}
