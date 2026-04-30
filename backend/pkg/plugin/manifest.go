// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package plugin

import (
	"encoding/json"
	"fmt"
)

type PageConfig struct {
	Type    string                   `json:"type"`
	Columns []map[string]interface{} `json:"columns,omitempty"`
	Actions []map[string]interface{} `json:"actions,omitempty"`
	Form    []map[string]interface{} `json:"form,omitempty"`
	Stats   []map[string]interface{} `json:"stats,omitempty"`
	Table   string                   `json:"table,omitempty"`
	API     string                   `json:"api,omitempty"`
}

type Manifest struct {
	Name                  string                 `json:"name"`
	Version               string                 `json:"version"`
	EntryPoint            string                 `json:"entry_point"`
	Description           string                 `json:"description"`
	MinAppVersion         string                 `json:"min_app_version"`
	CompatibleBaseVersion string                 `json:"compatible_base_version"`
	Layout                string                 `json:"layout"`
	RenderMode            string                 `json:"render_mode"`
	Entry                 string                 `json:"entry"`
	Style                 string                 `json:"style"`
	Signature             string                 `json:"signature"`
	UseShadowDOM          bool                   `json:"use_shadow_dom"`
	Permissions           []string               `json:"permissions"`
	AllowedAPIPrefixes    []string               `json:"allowed_api_prefixes"`
	ConfigSchema          map[string]interface{} `json:"config_schema"`
	Routes                []RouteRegistration    `json:"routes"`
	APIs                  []APIRegistration      `json:"apis"`
	Menus                 []MenuRegistration     `json:"menus"`
	Page                  *PageConfig            `json:"page,omitempty"`
}

type MenuRegistration struct {
	Title     string `json:"title"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	Component string `json:"component"`
	Sort      int    `json:"sort"`
	ParentID  uint   `json:"parent_id"`
}

type RouteRegistration struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

type APIRegistration struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Group  string `json:"group"`
}

var validPermissions = map[string]bool{
	"log:write":   true,
	"http:get":    true,
	"http:post":   true,
	"http:put":    true,
	"http:delete": true,
	"db:read":     true,
	"db:write":    true,
	"cache:read":  true,
	"cache:write": true,
}

func ParseManifest(jsonStr string) (*Manifest, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("清单JSON不能为空")
	}

	var m Manifest
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return nil, fmt.Errorf("清单JSON解析失败: %v", err)
	}

	if m.Name == "" {
		return nil, fmt.Errorf("插件名称不能为空")
	}
	if m.EntryPoint == "" {
		m.EntryPoint = "_start"
	}
	if m.Version == "" {
		m.Version = "0.0.1"
	}
	if m.Layout == "" {
		m.Layout = "embedded"
	}
	if m.Layout != "embedded" && m.Layout != "fullscreen" {
		return nil, fmt.Errorf("无效的layout值: %s，仅支持 embedded 或 fullscreen", m.Layout)
	}

	for _, perm := range m.Permissions {
		if !validPermissions[perm] {
			return nil, fmt.Errorf("无效的权限声明: %s", perm)
		}
	}

	return &m, nil
}

func (m *Manifest) HasPermission(perm string) bool {
	for _, p := range m.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	if m.EntryPoint == "" {
		return fmt.Errorf("入口函数不能为空")
	}
	for _, perm := range m.Permissions {
		if !validPermissions[perm] {
			return fmt.Errorf("无效的权限声明: %s", perm)
		}
	}
	return nil
}
