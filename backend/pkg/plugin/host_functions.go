package plugin

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tetratelabs/wazero/api"
)

var pluginCache sync.Map

type DBQueryFunc func(ctx context.Context, tenantKey string, query string, args ...interface{}) ([]map[string]interface{}, error)
type DBExecFunc func(ctx context.Context, tenantKey string, query string, args ...interface{}) (int64, error)

var globalDBQuery DBQueryFunc
var globalDBExec DBExecFunc

func RegisterDBFuncs(queryFn DBQueryFunc, execFn DBExecFunc) {
	globalDBQuery = queryFn
	globalDBExec = execFn
}

type HostFunctions struct {
	pluginKey  string
	manifest   *Manifest
	sandboxCfg *SandboxConfig
	httpClient *http.Client
	logger     func(format string, args ...interface{})
	dbQueryCount  int
	dbExecCount   int
	httpReqCount  int
}

func NewHostFunctions(pluginKey string, manifest *Manifest, sandboxCfg *SandboxConfig) *HostFunctions {
	return &HostFunctions{
		pluginKey:  pluginKey,
		manifest:   manifest,
		sandboxCfg: sandboxCfg,
		logger: func(format string, args ...interface{}) {
			log.Printf("[HostFunc:"+pluginKey+"] "+format, args...)
		},
	}
}

func (h *HostFunctions) SetHTTPClient(client *http.Client) {
	h.httpClient = client
}

func (h *HostFunctions) HostLog(ctx context.Context, m api.Module, offset uint32, length uint32) {
	if !h.manifest.HasPermission("log:write") {
		h.logger("权限拒绝: 插件无 log:write 权限")
		return
	}

	buf, ok := m.Memory().Read(offset, length)
	if !ok {
		h.logger("读取插件内存失败: offset=%d, length=%d", offset, length)
		return
	}

	msg := string(buf)
	h.logger("插件日志: %s", msg)
}

func (h *HostFunctions) HostHTTPRequest(ctx context.Context, m api.Module, methodOffset uint32, methodLen uint32, urlOffset uint32, urlLen uint32, bodyOffset uint32, bodyLen uint32, resultOffset uint32) uint32 {
	if !h.manifest.HasPermission("http:get") && !h.manifest.HasPermission("http:post") {
		h.logger("权限拒绝: 插件无 HTTP 权限")
		return 0
	}

	if !h.sandboxCfg.AllowNetwork {
		h.logger("网络访问被沙箱禁止")
		return 0
	}

	if h.sandboxCfg.MaxHTTPRequests > 0 && h.httpReqCount >= h.sandboxCfg.MaxHTTPRequests {
		h.logger("HTTP请求次数已达上限: %d", h.sandboxCfg.MaxHTTPRequests)
		return 0
	}

	methodBytes, ok := m.Memory().Read(methodOffset, methodLen)
	if !ok {
		return 0
	}

	urlBytes, ok := m.Memory().Read(urlOffset, urlLen)
	if !ok {
		return 0
	}

	method := string(methodBytes)
	urlStr := string(urlBytes)

	methodUpper := strings.ToUpper(method)
	if methodUpper == "POST" || methodUpper == "PUT" || methodUpper == "DELETE" {
		if !h.manifest.HasPermission("http:post") {
			h.logger("权限拒绝: 插件无 http:post 权限")
			return 0
		}
	}
	if methodUpper == "GET" {
		if !h.manifest.HasPermission("http:get") {
			h.logger("权限拒绝: 插件无 http:get 权限")
			return 0
		}
	}

	h.logger("HTTP请求: %s %s", method, urlStr)

	var bodyReader io.Reader
	if bodyLen > 0 {
		bodyBytes, ok := m.Memory().Read(bodyOffset, bodyLen)
		if !ok {
			return 0
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, methodUpper, urlStr, bodyReader)
	if err != nil {
		h.logger("创建HTTP请求失败: %v", err)
		return 0
	}
	if bodyLen > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	client := h.httpClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		h.logger("HTTP请求失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger("读取HTTP响应失败: %v", err)
		return 0
	}

	h.httpReqCount++

	resultData := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(respBody),
	}
	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return 0
	}

	writeLen := uint32(len(resultJSON))
	if writeLen > resultOffset {
		return 0
	}

	if resultOffset == 0 || writeLen == 0 {
		return writeLen
	}

	if !m.Memory().Write(resultOffset, resultJSON) {
		return 0
	}

	return writeLen
}

func (h *HostFunctions) HostDBQuery(ctx context.Context, m api.Module, queryOffset uint32, queryLen uint32, resultOffset uint32, resultMaxLen uint32) uint32 {
	if !h.manifest.HasPermission("db:read") {
		h.logger("权限拒绝: 插件无 db:read 权限")
		return 0
	}

	if globalDBQuery == nil {
		h.logger("数据库查询功能未初始化")
		return 0
	}

	if h.sandboxCfg.MaxDBQueries > 0 && h.dbQueryCount >= h.sandboxCfg.MaxDBQueries {
		h.logger("数据库查询次数已达上限: %d", h.sandboxCfg.MaxDBQueries)
		return 0
	}

	queryBytes, ok := m.Memory().Read(queryOffset, queryLen)
	if !ok {
		return 0
	}

	query := string(queryBytes)
	h.logger("数据库查询: %s", query)

	rows, err := globalDBQuery(ctx, h.pluginKey, query)
	if err != nil {
		h.logger("数据库查询失败: %v", err)
		return 0
	}

	h.dbQueryCount++

	resultJSON, err := json.Marshal(rows)
	if err != nil {
		return 0
	}

	writeLen := uint32(len(resultJSON))
	if writeLen > resultMaxLen {
		h.logger("查询结果超出缓冲区: resultLen=%d, maxLen=%d", writeLen, resultMaxLen)
		writeLen = resultMaxLen
	}

	if resultOffset == 0 || writeLen == 0 {
		return writeLen
	}

	if !m.Memory().Write(resultOffset, resultJSON[:writeLen]) {
		return 0
	}

	return writeLen
}

func (h *HostFunctions) HostDBExec(ctx context.Context, m api.Module, queryOffset uint32, queryLen uint32) uint32 {
	if !h.manifest.HasPermission("db:write") {
		h.logger("权限拒绝: 插件无 db:write 权限")
		return 0
	}

	if globalDBExec == nil {
		h.logger("数据库执行功能未初始化")
		return 0
	}

	if h.sandboxCfg.MaxDBQueries > 0 && h.dbExecCount >= h.sandboxCfg.MaxDBQueries {
		h.logger("数据库执行次数已达上限: %d", h.sandboxCfg.MaxDBQueries)
		return 0
	}

	queryBytes, ok := m.Memory().Read(queryOffset, queryLen)
	if !ok {
		return 0
	}

	query := string(queryBytes)
	h.logger("数据库执行: %s", query)

	affected, err := globalDBExec(ctx, h.pluginKey, query)
	if err != nil {
		h.logger("数据库执行失败: %v", err)
		return 0
	}

	h.dbExecCount++
	return uint32(affected)
}

func (h *HostFunctions) HostCacheGet(ctx context.Context, m api.Module, keyOffset uint32, keyLen uint32, resultOffset uint32, resultMaxLen uint32) uint32 {
	if !h.manifest.HasPermission("cache:read") {
		return 0
	}

	keyBytes, ok := m.Memory().Read(keyOffset, keyLen)
	if !ok {
		return 0
	}

	cacheKey := h.pluginKey + ":" + string(keyBytes)

	val, ok := pluginCache.Load(cacheKey)
	if !ok {
		return 0
	}

	valueStr, ok := val.(string)
	if !ok {
		return 0
	}

	writeLen := uint32(len(valueStr))
	if writeLen > resultMaxLen {
		writeLen = resultMaxLen
	}

	if resultOffset == 0 || writeLen == 0 {
		return writeLen
	}

	if !m.Memory().Write(resultOffset, []byte(valueStr[:writeLen])) {
		return 0
	}

	return writeLen
}

func (h *HostFunctions) HostCacheSet(ctx context.Context, m api.Module, keyOffset uint32, keyLen uint32, valueOffset uint32, valueLen uint32) {
	if !h.manifest.HasPermission("cache:write") {
		return
	}

	keyBytes, ok := m.Memory().Read(keyOffset, keyLen)
	if !ok {
		return
	}

	valueBytes, ok := m.Memory().Read(valueOffset, valueLen)
	if !ok {
		return
	}

	cacheKey := h.pluginKey + ":" + string(keyBytes)
	pluginCache.Store(cacheKey, string(valueBytes))
}

func writeUint32LE(m api.Module, offset uint32, value uint32) bool {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, value)
	return m.Memory().Write(offset, buf)
}

func ReadStringFromMemory(m api.Module, offset uint32, length uint32) (string, error) {
	buf, ok := m.Memory().Read(offset, length)
	if !ok {
		return "", fmt.Errorf("读取内存失败: offset=%d, length=%d", offset, length)
	}
	return string(buf), nil
}
