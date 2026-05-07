// 示例插件 - 展示FayHub插件SDK的使用方法
package main

import (
	"encoding/json"
	"fmt"
	"fayhub/pkg/plugin"
)

func main() {
	fmt.Println("=== FayHub 插件开发示例 ===\n")

	// 1. 创建SDK实例
	sdk := plugin.NewSDK()
	fmt.Println("✓ SDK创建成功")

	// 2. 获取SDK信息
	info := sdk.GetSDKInfo()
	fmt.Printf("  SDK版本: %s\n", info.Version)
	fmt.Printf("  支持语言: %v\n", info.SupportedLangs)
	fmt.Printf("  可用权限: %v\n", info.Permissions)
	fmt.Printf("  Host函数: %v\n\n", info.HostFunctions)

	// 3. 创建插件清单
	fmt.Println("3. 创建插件清单")
	manifest := sdk.CreateManifest(
		"Hello World",
		"1.0.0",
		"一个简单的Hello World插件",
	)

	// 4. 添加权限
	fmt.Println("4. 添加权限")
	err := sdk.AddPermission(manifest, plugin.PermLogWrite)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Println("  ✓ log:write权限已添加")
	}

	err = sdk.AddPermission(manifest, plugin.PermDBRead)
	if err == nil {
		fmt.Println("  ✓ db:read权限已添加")
	}

	// 5. 添加路由
	fmt.Println("\n5. 添加路由")
	sdk.AddRoute(manifest, "GET", "/hello", "handleHello")
	sdk.AddRoute(manifest, "POST", "/data", "handlePostData")
	fmt.Printf("  ✓ 已添加 %d 个路由\n", len(manifest.Routes))

	// 6. 添加API端点
	fmt.Println("\n6. 添加API端点")
	sdk.AddAPI(manifest, "GET", "/api/info", "plugin")
	fmt.Printf("  ✓ 已添加 %d 个API\n", len(manifest.APIs))

	// 7. 添加菜单
	fmt.Println("\n7. 添加菜单")
	sdk.AddMenu(manifest, "Hello World", "/plugin/hello", "home", 10)
	fmt.Printf("  ✓ 已添加 %d 个菜单\n", len(manifest.Menus))

	// 8. 验证并生成清单
	fmt.Println("\n8. 验证并生成清单")
	manifestJSON, err := sdk.ValidateAndMarshal(manifest)
	if err != nil {
		fmt.Printf("  验证失败: %v\n", err)
		return
	}
	fmt.Println("  ✓ 清单验证通过")
	fmt.Println("\n=== 生成的 manifest.json ===")
	fmt.Println(manifestJSON)

	// 9. 使用辅助工具
	fmt.Println("\n9. 使用辅助工具")
	helper := plugin.NewHelper()

	pluginID := "hello-world"
	isValid := helper.ValidatePluginID(pluginID)
	fmt.Printf("  插件ID '%s' 验证: %v\n", pluginID, isValid)

	routePath := helper.BuildRoutePath(pluginID, "hello")
	fmt.Printf("  路由路径: %s\n", routePath)

	apiPath := helper.BuildAPIPath(pluginID, "v1/info")
	fmt.Printf("  API路径: %s\n", apiPath)

	// 10. 生成插件模板
	fmt.Println("\n10. 生成插件模板")
	template := &plugin.PluginTemplate{
		Name:        "Hello World",
		Description: "一个简单的Hello World插件",
	}

	basicTemplate := template.GenerateBasicTemplate()
	fmt.Println("  ✓ 基础模板已生成 (前200字符):")
	if len(basicTemplate) > 200 {
		fmt.Println(basicTemplate[:200] + "...")
	} else {
		fmt.Println(basicTemplate)
	}

	fmt.Println("\n=== 示例完成 ===")
}

// 插件响应结构示例
type PluginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 创建插件响应
func CreatePluginResponse(success bool, message string, data interface{}) []byte {
	resp := PluginResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	jsonData, _ := json.Marshal(resp)
	return jsonData
}
