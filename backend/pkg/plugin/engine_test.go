package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type mockDataSource struct {
	resources map[string]*PluginResource
}

func newMockDataSource() *mockDataSource {
	return &mockDataSource{
		resources: make(map[string]*PluginResource),
	}
}

func (m *mockDataSource) GetPluginResource(ctx context.Context, pluginID string, version string) (*PluginResource, error) {
	if res, ok := m.resources[pluginID]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("插件资源不存在: pluginID=%s, version=%s", pluginID, version)
}

func (m *mockDataSource) addResource(pluginID string, res *PluginResource) {
	m.resources[pluginID] = res
}

func TestManifestValidation(t *testing.T) {
	validManifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"entry_point": "_start",
		"description": "测试插件",
		"permissions": ["log:write", "http:get"],
		"routes": [],
		"apis": []
	}`

	manifest := &Manifest{}
	if err := json.Unmarshal([]byte(validManifest), manifest); err != nil {
		t.Fatalf("解析manifest失败: %v", err)
	}

	if err := manifest.Validate(); err != nil {
		t.Fatalf("manifest校验失败: %v", err)
	}

	if manifest.Name != "test-plugin" {
		t.Errorf("期望name=test-plugin, 实际=%s", manifest.Name)
	}

	if !manifest.HasPermission("log:write") {
		t.Error("期望有log:write权限")
	}

	if manifest.HasPermission("db:write") {
		t.Error("不期望有db:write权限")
	}
}

func TestManifestInvalidPermission(t *testing.T) {
	invalidManifest := `{
		"name": "bad-plugin",
		"version": "1.0.0",
		"entry_point": "_start",
		"permissions": ["log:write", "evil:access"],
		"routes": [],
		"apis": []
	}`

	manifest := &Manifest{}
	if err := json.Unmarshal([]byte(invalidManifest), manifest); err != nil {
		t.Fatalf("解析manifest失败: %v", err)
	}

	if err := manifest.Validate(); err == nil {
		t.Error("期望校验失败，但通过了")
	}
}

func TestSandboxConfigDefaults(t *testing.T) {
	cfg := DefaultSandboxConfig()

	if cfg.MemoryLimitPages != 256 {
		t.Errorf("期望MemoryLimitPages=256, 实际=%d", cfg.MemoryLimitPages)
	}

	if cfg.ExecutionTimeout.Seconds() != 30 {
		t.Errorf("期望ExecutionTimeout=30s, 实际=%v", cfg.ExecutionTimeout)
	}

	if cfg.AllowNetwork {
		t.Error("默认不允许网络访问")
	}

	if cfg.AllowFileAccess {
		t.Error("默认不允许文件访问")
	}
}

func TestSandboxConfigValidation(t *testing.T) {
	cfg := DefaultSandboxConfig()
	cfg.MemoryLimitPages = 600

	if err := cfg.Validate(); err == nil {
		t.Error("期望校验失败(超过最大内存限制)，但通过了")
	}

	cfg = DefaultSandboxConfig()
	cfg.ExecutionTimeout = 600 * 1e9

	if err := cfg.Validate(); err == nil {
		t.Error("期望校验失败(超过最大超时)，但通过了")
	}

	cfg = DefaultSandboxConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("默认配置校验失败: %v", err)
	}
}

func TestWASMEngineStartStop(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	if engine.IsRunning() {
		t.Error("新引擎不应处于运行状态")
	}

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}

	if !engine.IsRunning() {
		t.Error("引擎应处于运行状态")
	}

	if err := engine.Stop(ctx); err != nil {
		t.Fatalf("引擎停止失败: %v", err)
	}

	if engine.IsRunning() {
		t.Error("引擎不应处于运行状态")
	}
}

func TestWASMEngineDoubleStart(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("重复启动不应报错: %v", err)
	}

	engine.Stop(ctx)
}

func TestWASMEngineInstallWithoutDataSource(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}
	defer engine.Stop(ctx)

	err := engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", "")
	if err != nil {
		t.Fatalf("无数据源安装插件失败: %v", err)
	}

	if engine.PluginCount() != 1 {
		t.Errorf("期望1个插件, 实际=%d", engine.PluginCount())
	}

	infos := engine.GetPluginInfos()
	if len(infos) != 1 {
		t.Errorf("期望1个插件信息, 实际=%d", len(infos))
	}

	if infos[0].Name != "plugin-com.fayhub.test" {
		t.Errorf("期望name=plugin-com.fayhub.test, 实际=%s", infos[0].Name)
	}

	if infos[0].HasModule {
		t.Error("无数据源时不应有WASM模块")
	}
}

func TestWASMEngineInstallWithDataSource(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}
	defer engine.Stop(ctx)

	ds := newMockDataSource()
	manifestJSON := `{
		"name": "data-driven-plugin",
		"version": "2.0.0",
		"entry_point": "_start",
		"description": "数据源驱动插件",
		"permissions": ["log:write"],
		"routes": [{"method": "GET", "path": "/api/test", "handler": "testHandler"}],
		"apis": [{"method": "GET", "path": "/api/test", "group": "test"}]
	}`

	ds.addResource("com.fayhub.dataplugin", &PluginResource{
		Name:         "data-driven-plugin",
		Version:      "2.0.0",
		Description:  "数据源驱动插件",
		EntryPoint:   "_start",
		ManifestJSON: manifestJSON,
	})

	engine.SetDataSource(ds)

	err := engine.Install(ctx, 1, "com.fayhub.dataplugin", "2.0.0", "")
	if err != nil {
		t.Fatalf("数据源安装插件失败: %v", err)
	}

	infos := engine.GetPluginInfos()
	if len(infos) != 1 {
		t.Fatalf("期望1个插件信息, 实际=%d", len(infos))
	}

	if infos[0].Name != "data-driven-plugin" {
		t.Errorf("期望name=data-driven-plugin, 实际=%s", infos[0].Name)
	}

	if len(infos[0].Permissions) != 1 || infos[0].Permissions[0] != "log:write" {
		t.Errorf("期望permissions=[log:write], 实际=%v", infos[0].Permissions)
	}

	registry := engine.GetRegistry()
	routes := registry.GetRoutes(1, "com.fayhub.dataplugin")
	if len(routes) != 1 {
		t.Errorf("期望1个路由, 实际=%d", len(routes))
	}

	apis := registry.GetAPIs(1, "com.fayhub.dataplugin")
	if len(apis) != 1 {
		t.Errorf("期望1个API, 实际=%d", len(apis))
	}
}

func TestWASMEngineUninstall(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	engine.Start(ctx)
	defer engine.Stop(ctx)

	engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", "")

	if engine.PluginCount() != 1 {
		t.Errorf("期望1个插件, 实际=%d", engine.PluginCount())
	}

	if err := engine.Uninstall(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("卸载插件失败: %v", err)
	}

	if engine.PluginCount() != 0 {
		t.Errorf("期望0个插件, 实际=%d", engine.PluginCount())
	}
}

func TestWASMEngineEnableDisable(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	engine.Start(ctx)
	defer engine.Stop(ctx)

	engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", "")

	if err := engine.Disable(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("禁用插件失败: %v", err)
	}

	infos := engine.GetPluginInfos()
	if len(infos) != 1 && infos[0].Status != "disabled" {
		t.Errorf("期望status=disabled, 实际=%s", infos[0].Status)
	}

	if err := engine.Enable(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("启用插件失败: %v", err)
	}

	infos = engine.GetPluginInfos()
	if len(infos) != 1 && infos[0].Status != "active" {
		t.Errorf("期望status=active, 实际=%s", infos[0].Status)
	}
}

func TestWASMEngineHealthCheck(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	engine.Start(ctx)
	defer engine.Stop(ctx)

	engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", "")

	if err := engine.HealthCheck(ctx, 1, "com.fayhub.test"); err != nil {
		t.Errorf("健康检查失败: %v", err)
	}

	if err := engine.HealthCheck(ctx, 1, "com.fayhub.nonexistent"); err == nil {
		t.Error("期望不存在的插件健康检查失败")
	}
}

func TestWASMEngineDuplicateInstall(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()

	engine.Start(ctx)
	defer engine.Stop(ctx)

	if err := engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", ""); err != nil {
		t.Fatalf("首次安装失败: %v", err)
	}

	if err := engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", ""); err != nil {
		t.Fatalf("重复安装不应报错: %v", err)
	}

	if engine.PluginCount() != 1 {
		t.Errorf("期望1个插件, 实际=%d", engine.PluginCount())
	}
}

func TestRegistryRouteManagement(t *testing.T) {
	registry := NewRegistry()

	routes := []RouteRegistration{
		{Method: "GET", Path: "/api/test", Handler: "testHandler"},
		{Method: "POST", Path: "/api/data", Handler: "dataHandler"},
	}

	registry.RegisterRoutes(1, "com.fayhub.test", routes)

	result := registry.GetRoutes(1, "com.fayhub.test")
	if len(result) != 2 {
		t.Errorf("期望2个路由, 实际=%d", len(result))
	}

	registry.UnregisterRoutes(1, "com.fayhub.test")
	result = registry.GetRoutes(1, "com.fayhub.test")
	if len(result) != 0 {
		t.Errorf("期望0个路由(已注销), 实际=%d", len(result))
	}
}

func TestRegistryAPIManagement(t *testing.T) {
	registry := NewRegistry()

	apis := []APIRegistration{
		{Method: "GET", Path: "/api/v1/test", Group: "test"},
	}

	registry.RegisterAPIs(1, "com.fayhub.test", apis)

	result := registry.GetAPIs(1, "com.fayhub.test")
	if len(result) != 1 {
		t.Errorf("期望1个API, 实际=%d", len(result))
	}

	registry.UnregisterAPIs(1, "com.fayhub.test")
	result = registry.GetAPIs(1, "com.fayhub.test")
	if len(result) != 0 {
		t.Errorf("期望0个API(已注销), 实际=%d", len(result))
	}
}

func TestRegistryTenantIsolation(t *testing.T) {
	registry := NewRegistry()

	routes1 := []RouteRegistration{
		{Method: "GET", Path: "/api/tenant1", Handler: "handler1"},
	}
	routes2 := []RouteRegistration{
		{Method: "GET", Path: "/api/tenant2", Handler: "handler2"},
	}

	registry.RegisterRoutes(1, "com.fayhub.test", routes1)
	registry.RegisterRoutes(2, "com.fayhub.test", routes2)

	result1 := registry.GetRoutes(1, "com.fayhub.test")
	if len(result1) != 1 || result1[0].Path != "/api/tenant1" {
		t.Error("租户1路由隔离失败")
	}

	result2 := registry.GetRoutes(2, "com.fayhub.test")
	if len(result2) != 1 || result2[0].Path != "/api/tenant2" {
		t.Error("租户2路由隔离失败")
	}
}

func TestWASMEngineWithRealWASMModule(t *testing.T) {
	wasmPath := "../../plugins/examples/hello-plugin/hello-plugin.wasm"
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Skipf("跳过: WASM文件不存在(%s): %v", wasmPath, err)
	}

	engine := NewWASMEngine()
	ctx := context.Background()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}
	defer engine.Stop(ctx)

	ds := newMockDataSource()
	manifestJSON := `{
		"name": "hello-plugin",
		"version": "1.0.0",
		"entry_point": "_start",
		"description": "FayHub示例插件",
		"permissions": ["log:write"],
		"routes": [],
		"apis": []
	}`

	ds.addResource("com.fayhub.hello", &PluginResource{
		Name:         "hello-plugin",
		Version:      "1.0.0",
		EntryPoint:   "_start",
		ManifestJSON: manifestJSON,
	})

	engine.SetDataSource(ds)

	_ = wasmBytes

	err = engine.Install(ctx, 1, "com.fayhub.hello", "1.0.0", "")
	if err != nil {
		t.Fatalf("安装插件失败: %v", err)
	}

	infos := engine.GetPluginInfos()
	if len(infos) != 1 {
		t.Fatalf("期望1个插件, 实际=%d", len(infos))
	}

	if infos[0].Name != "hello-plugin" {
		t.Errorf("期望name=hello-plugin, 实际=%s", infos[0].Name)
	}

	if infos[0].Status != "active" {
		t.Errorf("期望status=active, 实际=%s", infos[0].Status)
	}
}

func TestNoopEngine(t *testing.T) {
	engine := &NoopEngine{}
	ctx := context.Background()

	if err := engine.Install(ctx, 1, "com.fayhub.test", "1.0.0", ""); err != nil {
		t.Fatalf("NoopEngine Install失败: %v", err)
	}

	if err := engine.Uninstall(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("NoopEngine Uninstall失败: %v", err)
	}

	if err := engine.Enable(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("NoopEngine Enable失败: %v", err)
	}

	if err := engine.Disable(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("NoopEngine Disable失败: %v", err)
	}

	if err := engine.HealthCheck(ctx, 1, "com.fayhub.test"); err != nil {
		t.Fatalf("NoopEngine HealthCheck失败: %v", err)
	}

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("NoopEngine Start失败: %v", err)
	}

	if err := engine.Stop(ctx); err != nil {
		t.Fatalf("NoopEngine Stop失败: %v", err)
	}
}

func TestPluginKey(t *testing.T) {
	key := pluginKey(1, "com.fayhub.mall")
	expected := "t1_pcom.fayhub.mall"
	if key != expected {
		t.Errorf("期望key=%s, 实际=%s", expected, key)
	}
}
