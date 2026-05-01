package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fayhub/pkg/config"
	"fayhub/pkg/response"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type loadedPlugin struct {
	Key        string
	Module     api.Module
	Manifest   *Manifest
	SandboxCfg *SandboxConfig
	HostFuncs  *HostFunctions
	WasmBytes  []byte
	Status     string
}

type WASMEngine struct {
	mu         sync.RWMutex
	runtime    wazero.Runtime
	wasiCloser api.Closer
	plugins    map[string]*loadedPlugin
	registry   *Registry
	client     *http.Client
	dataSource PluginDataSource
	ginEngine  *gin.Engine
	ginRoutes  map[string][]ginRouteInfo
	routeIdx   *routeIndex
}

type ginRouteInfo struct {
	method  string
	path    string
	handler string
}

type routeIndex struct {
	mu       sync.RWMutex
	byMethod map[string]map[string]routeEntry
}

type routeEntry struct {
	pluginKey string
	handler   string
	pluginID  string
}

func NewWASMEngine() *WASMEngine {
	timeout := 30
	if config.GlobalConfig != nil && config.GlobalConfig.PluginEngine.HTTPTimeoutSec > 0 {
		timeout = config.GlobalConfig.PluginEngine.HTTPTimeoutSec
	}

	return &WASMEngine{
		plugins:   make(map[string]*loadedPlugin),
		registry:  NewRegistry(),
		ginRoutes: make(map[string][]ginRouteInfo),
		routeIdx: &routeIndex{
			byMethod: make(map[string]map[string]routeEntry),
		},
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (e *WASMEngine) SetDataSource(ds PluginDataSource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dataSource = ds
	log.Println("[WASMEngine] 数据源已注入")
}

func (e *WASMEngine) SetGinEngine(eng *gin.Engine) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ginEngine = eng
	log.Println("[WASMEngine] Gin引擎已注入")
}

func (e *WASMEngine) SetupPluginProxyRoutes() {
	if e.ginEngine == nil {
		return
	}

	proxyGroup := e.ginEngine.Group("/api/plugin-proxy/:pluginID")
	proxyGroup.Use(e.pluginProxyMiddleware())

	proxyGroup.Any("/*handler", func(c *gin.Context) {})

	log.Println("[WASMEngine] 插件代理路由组已注册: /api/plugin-proxy/:pluginID/*handler")
}

func (e *WASMEngine) pluginProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		pluginID := c.Param("pluginID")
		handlerPath := c.Param("handler")
		if len(handlerPath) > 0 && handlerPath[0] == '/' {
			handlerPath = handlerPath[1:]
		}
		method := c.Request.Method

		entry := e.lookupRoute(method, pluginID, handlerPath)
		if entry == nil {
			response.GinErrorWithData(c, 40400, "插件路由不存在或插件未启用", gin.H{"path": handlerPath, "method": method})
			c.Abort()
			return
		}

		tenantID, _ := c.Get("tenant_id")
		tid, ok := tenantID.(uint)
		if !ok {
			response.GinError(c, 40300, "无法识别租户")
			c.Abort()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.GinError(c, 40000, "读取请求体失败")
			c.Abort()
			return
		}

		payload := e.buildPluginPayload(c, body)

		result, callErr := e.Call(c.Request.Context(), tid, pluginID, entry.handler, payload)
		duration := time.Since(startTime)

		if callErr != nil {
			log.Printf("[WASMEngine] 插件调用失败: plugin=%s, handler=%s, duration=%v, error=%v",
				pluginID, entry.handler, duration, callErr)
			response.GinError(c, 50000, fmt.Sprintf("插件调用失败: %v", callErr))
			c.Abort()
			return
		}

		log.Printf("[WASMEngine] 插件调用成功: plugin=%s, handler=%s, duration=%v, resultSize=%d",
			pluginID, entry.handler, duration, len(result))

		c.Data(http.StatusOK, "application/json", result)
	}
}

func (e *WASMEngine) lookupRoute(method, pluginID, path string) *routeEntry {
	e.routeIdx.mu.RLock()
	defer e.routeIdx.mu.RUnlock()

	methodMap, ok := e.routeIdx.byMethod[method]
	if !ok {
		return nil
	}

	routePath := path
	if len(routePath) > 0 && routePath[0] == '/' {
		routePath = routePath[1:]
	}

	entry, exists := methodMap[pluginID+":"+routePath]
	if !exists {
		return nil
	}
	return &entry
}

func (e *WASMEngine) buildPluginPayload(c *gin.Context, body []byte) []byte {
	queryParams := c.Request.URL.Query()

	if len(queryParams) == 0 && len(body) == 0 {
		return []byte("{}")
	}

	payload := make(map[string]interface{})

	if len(body) > 0 {
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			payload["raw_body"] = string(body)
		}
	}

	if len(queryParams) > 0 {
		params := make(map[string]string)
		for k, v := range queryParams {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
		payload["query_params"] = params
	}

	headers := make(map[string]string)
	headers["content_type"] = c.GetHeader("Content-Type")
	headers["accept"] = c.GetHeader("Accept")
	headers["user_agent"] = c.GetHeader("User-Agent")
	headers["authorization"] = c.GetHeader("Authorization")
	payload["headers"] = headers

	result, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return body
	}
	return result
}

func (e *WASMEngine) injectRoutesToGin(tenantID uint, pluginID string, routes []RouteRegistration) {
	if len(routes) == 0 {
		return
	}

	key := pluginKey(tenantID, pluginID)
	var infos []ginRouteInfo

	e.routeIdx.mu.Lock()
	for _, route := range routes {
		handlerName := route.Handler
		if handlerName == "" {
			handlerName = route.Path
		}
		fullPath := "/api/plugin-proxy/" + pluginID + route.Path
		routePath := route.Path
		if len(routePath) > 0 && routePath[0] == '/' {
			routePath = routePath[1:]
		}

		infos = append(infos, ginRouteInfo{
			method:  route.Method,
			path:    fullPath,
			handler: handlerName,
		})

		methodMap := e.routeIdx.byMethod[route.Method]
		if methodMap == nil {
			methodMap = make(map[string]routeEntry)
			e.routeIdx.byMethod[route.Method] = methodMap
		}
		methodMap[pluginID+":"+routePath] = routeEntry{
			pluginKey: key,
			handler:   handlerName,
			pluginID:  pluginID,
		}
		log.Printf("[WASMEngine] 注册插件路由索引: %s %s -> %s", route.Method, fullPath, handlerName)
	}
	e.routeIdx.mu.Unlock()

	e.ginRoutes[key] = infos
}

func (e *WASMEngine) removeRoutesFromGin(tenantID uint, pluginID string) {
	key := pluginKey(tenantID, pluginID)
	if _, exists := e.ginRoutes[key]; exists {
		delete(e.ginRoutes, key)
		log.Printf("[WASMEngine] 移除插件路由记录: key=%s", key)
	}

	e.routeIdx.mu.Lock()
	for method, methodMap := range e.routeIdx.byMethod {
		for k, entry := range methodMap {
			if entry.pluginKey == key {
				delete(methodMap, k)
			}
		}
		if len(methodMap) == 0 {
			delete(e.routeIdx.byMethod, method)
		}
	}
	e.routeIdx.mu.Unlock()
}

func (e *WASMEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.runtime != nil {
		log.Println("[WASMEngine] 引擎已启动，跳过重复初始化")
		return nil
	}

	defaultCfg := DefaultSandboxConfig()

	runtimeConfig := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithMemoryLimitPages(defaultCfg.MemoryLimitPages)

	e.runtime = wazero.NewRuntimeWithConfig(ctx, runtimeConfig)

	var err error
	e.wasiCloser, err = wasi_snapshot_preview1.Instantiate(ctx, e.runtime)
	if err != nil {
		e.runtime.Close(ctx)
		e.runtime = nil
		return fmt.Errorf("实例化WASI模块失败: %v", err)
	}

	log.Println("[WASMEngine] WASM运行时初始化成功")
	return nil
}

func (e *WASMEngine) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.runtime == nil {
		return nil
	}

	for key, p := range e.plugins {
		if p.Module != nil {
			if err := p.Module.Close(ctx); err != nil {
				log.Printf("[WASMEngine] 关闭插件模块失败: key=%s, err=%v", key, err)
			}
		}
		delete(e.plugins, key)
	}

	if err := e.runtime.Close(ctx); err != nil {
		return fmt.Errorf("关闭WASM运行时失败: %v", err)
	}

	e.runtime = nil
	log.Println("[WASMEngine] WASM运行时已关闭")
	return nil
}

func (e *WASMEngine) Install(ctx context.Context, tenantID uint, pluginID string, version string, licenseKey string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.runtime == nil {
		return fmt.Errorf("WASM引擎未启动")
	}

	key := pluginKey(tenantID, pluginID)
	if _, exists := e.plugins[key]; exists {
		log.Printf("[WASMEngine] 插件已加载: key=%s，跳过重复安装", key)
		return nil
	}

	wasmBytes, manifest, sandboxCfg, err := e.loadPluginResources(ctx, pluginID, version)
	if err != nil {
		return fmt.Errorf("加载插件资源失败: %v", err)
	}

	hostFuncs := NewHostFunctions(key, manifest, sandboxCfg)
	hostFuncs.SetHTTPClient(e.client)

	if len(wasmBytes) == 0 {
		log.Printf("[WASMEngine] 无WASM模块数据，仅注册元数据: key=%s", key)
		e.plugins[key] = &loadedPlugin{
			Key:        key,
			Module:     nil,
			Manifest:   manifest,
			SandboxCfg: sandboxCfg,
			HostFuncs:  hostFuncs,
			WasmBytes:  nil,
			Status:     "active",
		}
		e.registry.RegisterRoutes(tenantID, pluginID, manifest.Routes)
		e.registry.RegisterAPIs(tenantID, pluginID, manifest.APIs)
		e.registry.RegisterMenus(tenantID, pluginID, manifest.Menus)
		e.injectRoutesToGin(tenantID, pluginID, manifest.Routes)
		log.Printf("[WASMEngine] 插件元数据注册成功(无WASM模块): key=%s", key)
		return nil
	}

	builder := e.runtime.NewHostModuleBuilder("env_" + key)
	builder.NewFunctionBuilder().
		WithParameterNames("offset", "length").
		WithFunc(hostFuncs.HostLog).
		Export("host_log")

	if manifest.HasPermission("http:get") || manifest.HasPermission("http:post") {
		builder.NewFunctionBuilder().
			WithParameterNames("method_offset", "method_len", "url_offset", "url_len", "body_offset", "body_len", "result_offset").
			WithResultNames("status").
			WithFunc(hostFuncs.HostHTTPRequest).
			Export("host_http_request")
	}

	if manifest.HasPermission("db:read") {
		builder.NewFunctionBuilder().
			WithParameterNames("query_offset", "query_len", "result_offset", "result_max_len").
			WithResultNames("result_len").
			WithFunc(hostFuncs.HostDBQuery).
			Export("host_db_query")
	}

	if manifest.HasPermission("db:write") {
		builder.NewFunctionBuilder().
			WithParameterNames("query_offset", "query_len").
			WithResultNames("affected").
			WithFunc(hostFuncs.HostDBExec).
			Export("host_db_exec")
	}

	if manifest.HasPermission("cache:read") {
		builder.NewFunctionBuilder().
			WithParameterNames("key_offset", "key_len", "result_offset", "result_max_len").
			WithResultNames("result_len").
			WithFunc(hostFuncs.HostCacheGet).
			Export("host_cache_get")
	}

	if manifest.HasPermission("cache:write") {
		builder.NewFunctionBuilder().
			WithParameterNames("key_offset", "key_len", "value_offset", "value_len").
			WithResultNames("").
			WithFunc(hostFuncs.HostCacheSet).
			Export("host_cache_set")
	}

	if _, err := builder.Instantiate(ctx); err != nil {
		return fmt.Errorf("实例化Host模块失败: %v", err)
	}

	moduleConfig := wazero.NewModuleConfig().
		WithName(key).
		WithEnv("PLUGIN_ID", pluginID).
		WithEnv("TENANT_ID", fmt.Sprintf("%d", tenantID)).
		WithEnv("VERSION", version)

	mod, err := e.runtime.InstantiateWithConfig(ctx, wasmBytes, moduleConfig)
	if err != nil {
		return fmt.Errorf("实例化WASM模块失败: %v", err)
	}

	e.plugins[key] = &loadedPlugin{
		Key:        key,
		Module:     mod,
		Manifest:   manifest,
		SandboxCfg: sandboxCfg,
		HostFuncs:  hostFuncs,
		WasmBytes:  wasmBytes,
		Status:     "active",
	}

	e.registry.RegisterRoutes(tenantID, pluginID, manifest.Routes)
	e.registry.RegisterAPIs(tenantID, pluginID, manifest.APIs)
	e.registry.RegisterMenus(tenantID, pluginID, manifest.Menus)
	e.injectRoutesToGin(tenantID, pluginID, manifest.Routes)

	log.Printf("[WASMEngine] 插件安装成功: key=%s, entry=%s", key, manifest.EntryPoint)
	return nil
}

func (e *WASMEngine) Uninstall(ctx context.Context, tenantID uint, pluginID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	if !exists {
		log.Printf("[WASMEngine] 插件未加载: key=%s", key)
		return nil
	}

	if p.Module != nil {
		if err := p.Module.Close(ctx); err != nil {
			log.Printf("[WASMEngine] 关闭插件模块失败: key=%s, err=%v", key, err)
		}
	}

	e.registry.UnregisterRoutes(tenantID, pluginID)
	e.registry.UnregisterAPIs(tenantID, pluginID)
	e.registry.UnregisterMenus(tenantID, pluginID)
	e.removeRoutesFromGin(tenantID, pluginID)

	delete(e.plugins, key)
	log.Printf("[WASMEngine] 插件卸载成功: key=%s", key)
	return nil
}

func (e *WASMEngine) Enable(ctx context.Context, tenantID uint, pluginID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	if !exists {
		return fmt.Errorf("插件未安装: key=%s", key)
	}

	if p.Status == "active" {
		return fmt.Errorf("插件已处于启用状态")
	}

	if len(p.WasmBytes) > 0 {
		moduleConfig := wazero.NewModuleConfig().
			WithName(key).
			WithEnv("PLUGIN_ID", pluginID).
			WithEnv("TENANT_ID", fmt.Sprintf("%d", tenantID))

		mod, err := e.runtime.InstantiateWithConfig(ctx, p.WasmBytes, moduleConfig)
		if err != nil {
			return fmt.Errorf("重新实例化WASM模块失败: %v", err)
		}
		p.Module = mod
	}

	p.Status = "active"

	e.registry.RegisterRoutes(tenantID, pluginID, p.Manifest.Routes)
	e.registry.RegisterAPIs(tenantID, pluginID, p.Manifest.APIs)
	e.registry.RegisterMenus(tenantID, pluginID, p.Manifest.Menus)
	e.injectRoutesToGin(tenantID, pluginID, p.Manifest.Routes)

	log.Printf("[WASMEngine] 插件启用成功: key=%s", key)
	return nil
}

func (e *WASMEngine) Disable(ctx context.Context, tenantID uint, pluginID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	if !exists {
		return fmt.Errorf("插件未安装: key=%s", key)
	}

	if p.Status == "disabled" {
		return fmt.Errorf("插件已处于禁用状态")
	}

	if p.Module != nil {
		if err := p.Module.Close(ctx); err != nil {
			log.Printf("[WASMEngine] 关闭插件模块失败(禁用): key=%s, err=%v", key, err)
		}
	}

	p.Module = nil
	p.Status = "disabled"

	e.registry.UnregisterRoutes(tenantID, pluginID)
	e.registry.UnregisterAPIs(tenantID, pluginID)
	e.registry.UnregisterMenus(tenantID, pluginID)
	e.removeRoutesFromGin(tenantID, pluginID)

	log.Printf("[WASMEngine] 插件禁用成功: key=%s", key)
	return nil
}

func (e *WASMEngine) Upgrade(ctx context.Context, tenantID uint, pluginID string, oldVersion string, newVersion string, licenseKey string) error {
	if err := e.Uninstall(ctx, tenantID, pluginID); err != nil {
		return fmt.Errorf("卸载旧版本失败: %v", err)
	}
	if err := e.Install(ctx, tenantID, pluginID, newVersion, licenseKey); err != nil {
		return fmt.Errorf("安装新版本失败: %v", err)
	}
	log.Printf("[WASMEngine] 插件升级成功: plugin=%s, %s→%s", pluginID, oldVersion, newVersion)
	return nil
}

func (e *WASMEngine) GetStatus(ctx context.Context, tenantID uint, pluginID string) (*PluginStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	if !exists {
		return nil, fmt.Errorf("插件未安装: key=%s", key)
	}

	return &PluginStatus{
		PluginID: pluginID,
		Version:  p.Manifest.Version,
		Status:   p.Status,
	}, nil
}

func (e *WASMEngine) ListPlugins(ctx context.Context, tenantID uint) ([]*InstalledPluginInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	prefix := fmt.Sprintf("t%d_", tenantID)
	result := make([]*InstalledPluginInfo, 0)

	for key, p := range e.plugins {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			info := &InstalledPluginInfo{
				PluginID:    p.Manifest.Name,
				Name:        p.Manifest.Name,
				Version:     p.Manifest.Version,
				Status:      p.Status,
				Description: p.Manifest.Description,
			}
			result = append(result, info)
		}
	}

	return result, nil
}

func (e *WASMEngine) HealthCheck(ctx context.Context, tenantID uint, pluginID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	if !exists {
		return fmt.Errorf("插件未安装: key=%s", key)
	}

	if p.Status != "active" {
		return fmt.Errorf("插件状态异常: %s", p.Status)
	}

	if p.Module == nil && len(p.WasmBytes) > 0 {
		return fmt.Errorf("插件模块为空")
	}

	return nil
}

// Call 底座向插件发起调用的核心入口
// 通过 wazero 调用插件 WASM 模块的导出函数，实现底座↔插件的双向通信
func (e *WASMEngine) Call(ctx context.Context, tenantID uint, pluginID string, functionName string, payload []byte) ([]byte, error) {
	e.mu.RLock()
	key := pluginKey(tenantID, pluginID)
	p, exists := e.plugins[key]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("插件未安装: key=%s", key)
	}

	if p.Status != "active" {
		return nil, fmt.Errorf("插件未启用，当前状态: %s", p.Status)
	}

	if p.Module == nil {
		return nil, fmt.Errorf("插件无WASM模块实例（纯元数据插件不支持函数调用）")
	}

	// 安全校验：被调用的函数必须在 Manifest 的 Routes/APIs 中声明
	if !e.isExportedFunction(p, functionName) {
		return nil, fmt.Errorf("函数未在Manifest中声明: %s（插件: %s）", functionName, pluginID)
	}

	// 查找导出函数
	fn := p.Module.ExportedFunction(functionName)
	if fn == nil {
		return nil, fmt.Errorf("WASM模块中不存在导出函数: %s", functionName)
	}

	// 将 payload 写入 WASM 线性内存
	memory := p.Module.Memory()
	if memory == nil {
		return nil, fmt.Errorf("插件WASM模块无共享内存")
	}

	// 分配内存区域：payload区 + 结果区
	// 布局: [payloadLen(4bytes)][payload(payloadSize)][resultLen(4bytes)][result(resultMaxLen)]
	payloadSize := uint32(len(payload))
	resultMaxLen := uint32(64 * 1024) // 64KB 结果缓冲区
	totalAlloc := 4 + payloadSize + 4 + resultMaxLen

	// 检查沙箱内存限制
	if p.SandboxCfg != nil && totalAlloc > p.SandboxCfg.MemoryLimitBytes() {
		return nil, fmt.Errorf("payload大小超过沙箱内存限制: %d > %d", totalAlloc, p.SandboxCfg.MemoryLimitBytes())
	}

	// 使用 wazero 的 Malloc 分配内存
	mallocFn := p.Module.ExportedFunction("malloc")
	if mallocFn == nil {
		return nil, fmt.Errorf("插件WASM模块未导出malloc函数")
	}

	// 分配 payload 缓冲区
	payloadPtrRes, err := mallocFn.Call(ctx, uint64(payloadSize))
	if err != nil {
		return nil, fmt.Errorf("分配payload内存失败: %v", err)
	}
	payloadPtr := uint32(payloadPtrRes[0])

	// 写入 payload 到 WASM 内存
	success := memory.Write(payloadPtr, payload)
	if !success {
		return nil, fmt.Errorf("写入payload到WASM内存失败")
	}

	// 分配结果缓冲区
	resultPtrRes, err := mallocFn.Call(ctx, uint64(resultMaxLen))
	if err != nil {
		return nil, fmt.Errorf("分配结果内存失败: %v", err)
	}
	resultPtr := uint32(resultPtrRes[0])

	freeFn := p.Module.ExportedFunction("free")
	if freeFn != nil {
		defer func() {
			freeCtx, freeCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer freeCancel()
			freeFn.Call(freeCtx, uint64(payloadPtr))
			freeFn.Call(freeCtx, uint64(resultPtr))
		}()
	}

	timeout := 30 * time.Second
	if p.SandboxCfg != nil && p.SandboxCfg.ExecutionTimeout > 0 {
		timeout = p.SandboxCfg.ExecutionTimeout
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	results, err := fn.Call(callCtx, uint64(payloadPtr), uint64(payloadSize), uint64(resultPtr), uint64(resultMaxLen))
	if err != nil {
		return nil, fmt.Errorf("调用插件函数失败 [%s]: %v", functionName, err)
	}

	// 读取返回结果
	resultLen := uint32(results[0])
	if resultLen == 0 {
		return []byte{}, nil
	}

	if resultLen > resultMaxLen {
		return nil, fmt.Errorf("插件返回数据超过缓冲区大小: %d > %d", resultLen, resultMaxLen)
	}

	resultData, success := memory.Read(resultPtr, resultLen)
	if !success {
		return nil, fmt.Errorf("读取插件返回数据失败")
	}

	// 复制结果（脱离 WASM 内存引用）
	result := make([]byte, resultLen)
	copy(result, resultData)

	log.Printf("[WASMEngine] 调用成功: plugin=%s, function=%s, payloadSize=%d, resultSize=%d",
		pluginID, functionName, payloadSize, resultLen)

	return result, nil
}

// isExportedFunction 检查函数名是否在 Manifest 的路由/API声明中注册
func (e *WASMEngine) isExportedFunction(p *loadedPlugin, functionName string) bool {
	if p.Manifest == nil {
		return false
	}

	// 入口函数始终允许调用
	if functionName == p.Manifest.EntryPoint {
		return true
	}

	// 检查 Routes 中声明的 Handler
	for _, route := range p.Manifest.Routes {
		if route.Handler == functionName {
			return true
		}
	}

	return false
}

func (e *WASMEngine) GetRegistry() *Registry {
	return e.registry
}

type PluginInfo struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
	EntryPoint  string   `json:"entry_point"`
	Permissions []string `json:"permissions"`
	HasModule   bool     `json:"has_module"`
}

func (e *WASMEngine) GetLoadedPlugins() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]string)
	for key, p := range e.plugins {
		result[key] = p.Status
	}
	return result
}

func (e *WASMEngine) GetManifest(tenantID uint, pluginID string) *Manifest {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	if p, ok := e.plugins[key]; ok {
		return p.Manifest
	}
	return nil
}

func (e *WASMEngine) GetPluginInfos() []PluginInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]PluginInfo, 0, len(e.plugins))
	for _, p := range e.plugins {
		info := PluginInfo{
			Key:         p.Key,
			Name:        p.Manifest.Name,
			Version:     p.Manifest.Version,
			Status:      p.Status,
			EntryPoint:  p.Manifest.EntryPoint,
			Permissions: p.Manifest.Permissions,
			HasModule:   p.Module != nil,
		}
		result = append(result, info)
	}
	return result
}

func (e *WASMEngine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.runtime != nil
}

func (e *WASMEngine) PluginCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.plugins)
}

func (e *WASMEngine) loadPluginResources(ctx context.Context, pluginID string, version string) ([]byte, *Manifest, *SandboxConfig, error) {
	if e.dataSource != nil {
		return e.loadFromDataSource(ctx, pluginID, version)
	}
	return e.loadFallbackResources(ctx, pluginID, version)
}

func (e *WASMEngine) loadFromDataSource(ctx context.Context, pluginID string, version string) ([]byte, *Manifest, *SandboxConfig, error) {
	res, err := e.dataSource.GetPluginResource(ctx, pluginID, version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("从数据源获取插件资源失败: %v", err)
	}

	manifest := &Manifest{
		Name:          res.Name,
		Version:       res.Version,
		EntryPoint:    res.EntryPoint,
		Description:   res.Description,
		MinAppVersion: res.MinAppVersion,
	}

	if res.ManifestJSON != "" {
		if jsonErr := json.Unmarshal([]byte(res.ManifestJSON), manifest); jsonErr != nil {
			log.Printf("[WASMEngine] 解析manifest_json失败，使用基础信息: %v", jsonErr)
			manifest.Name = res.Name
			manifest.Version = res.Version
			manifest.EntryPoint = res.EntryPoint
			manifest.Description = res.Description
			manifest.MinAppVersion = res.MinAppVersion
		}
	}

	if manifest.EntryPoint == "" {
		manifest.EntryPoint = "_start"
	}

	if manifest.Layout == "" {
		manifest.Layout = "embedded"
	}

	if valErr := manifest.Validate(); valErr != nil {
		return nil, nil, nil, fmt.Errorf("manifest校验失败: %v", valErr)
	}

	sandboxCfg := DefaultSandboxConfig()
	if res.SandboxConfig != "" {
		if jsonErr := json.Unmarshal([]byte(res.SandboxConfig), sandboxCfg); jsonErr != nil {
			log.Printf("[WASMEngine] 解析sandbox_config失败，使用默认配置: %v", jsonErr)
			sandboxCfg = DefaultSandboxConfig()
		}
	}

	var wasmBytes []byte
	if len(res.WASMBytes) > 0 {
		wasmBytes = res.WASMBytes
	} else if res.DownloadURL != "" {
		fetchedBytes, fetchErr := e.fetchWASMModule(ctx, res.DownloadURL)
		if fetchErr != nil {
			log.Printf("[WASMEngine] 下载WASM模块失败: %v", fetchErr)
		} else {
			wasmBytes = fetchedBytes
		}
	}

	if len(wasmBytes) > 0 && res.PackageHash != "" {
		hash := sha256.Sum256(wasmBytes)
		computedHash := hex.EncodeToString(hash[:])
		if computedHash != res.PackageHash {
			return nil, nil, nil, fmt.Errorf("WASM文件SHA256校验失败: 期望=%s, 实际=%s", res.PackageHash, computedHash)
		}
	}

	return wasmBytes, manifest, sandboxCfg, nil
}

func (e *WASMEngine) loadFallbackResources(_ context.Context, pluginID string, version string) ([]byte, *Manifest, *SandboxConfig, error) {
	manifest := &Manifest{
		Name:       fmt.Sprintf("plugin-%s", pluginID),
		Version:    version,
		EntryPoint: "_start",
	}

	sandboxCfg := DefaultSandboxConfig()
	return nil, manifest, sandboxCfg, nil
}

func (e *WASMEngine) fetchWASMModule(ctx context.Context, downloadURL string) ([]byte, error) {
	if downloadURL == "" {
		return nil, fmt.Errorf("下载地址为空")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载WASM模块失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载WASM模块失败: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取WASM模块数据失败: %v", err)
	}

	return data, nil
}
